package routing

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	sf "github.com/sanek1/metrics-collector/internal/flags/server"
	h "github.com/sanek1/metrics-collector/internal/handlers"
	ss "github.com/sanek1/metrics-collector/internal/storage/server"
	v "github.com/sanek1/metrics-collector/internal/validation"
	l "github.com/sanek1/metrics-collector/pkg/logging"
)

type Router struct {
	router         *gin.Engine
	l              *l.ZapLogger
	middleware     *v.MiddlewareController
	middlewareHash *v.Secret
	storage        ss.Storage
	s              *h.Storage
	opt            *sf.ServerOptions
}

func NewRouting(s ss.Storage, opt *sf.ServerOptions, logger *l.ZapLogger) *Router {
	c := &Router{
		l:       logger,
		router:  gin.Default(),
		storage: s,
		opt:     opt,
	}

	c.s = h.NewStorage(s, logger)
	c.middleware = v.NewValidation(c.s, logger)
	c.middlewareHash = v.NewHash(opt.CryptoKey)
	return c
}

func (r *Router) InitRouting() http.Handler {
	if r.opt.CryptoKey != "" {
		r.router.Use(r.middlewareHash.HashMiddleware())
	}

	r.router.Use(v.GzipMiddleware())
	r.router.Use(func(c *gin.Context) {
		if c.Request.Method == "POST" && c.FullPath() == "/update/:metricType/:metricName/:metricValue" {
			metricType := c.Param("metricType")
			metricValue := c.Param("metricValue")

			if metricType != "counter" && metricType != "gauge" {
				c.AbortWithStatus(http.StatusBadRequest)
				return
			}
			if metricType == "counter" {
				if _, err := parseCounterValue(metricValue); err != nil {
					c.AbortWithStatus(http.StatusBadRequest)
					return
				}
			}
			if metricType == "gauge" {
				if _, err := parseGaugeValue(metricValue); err != nil {
					c.AbortWithStatus(http.StatusBadRequest)
					return
				}
			}
		}

		c.Next()
	})

	r.router.POST("/update/:metricType/:metricName/:metricValue", r.s.MetricHandler)
	r.router.POST("/updates/", r.s.MetricHandler)
	r.router.POST("/update/", r.s.MetricHandler)
	r.router.POST("/value/", r.s.GetMetricsByValueHandler)
	r.router.POST("/", gin.WrapF(h.NotImplementedHandler))
	r.router.GET("/", r.s.MainPageHandler)
	r.router.GET("/ping", r.s.PingDBHandler)
	r.router.GET("/:metricValue/:metricType/:metricName", r.s.GetMetricsByNameHandler)
	r.router.NoRoute(gin.WrapF(h.NotImplementedHandler))
	return r.router
}

func parseCounterValue(value string) (int64, error) {
	intValue, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid counter value: %w", err)
	}

	if intValue < 0 {
		return 0, fmt.Errorf("counter value must be non-negative")
	}
	return intValue, nil
}

func parseGaugeValue(value string) (float64, error) {
	floatValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid gauge value: %w", err)
	}
	return floatValue, nil
}
