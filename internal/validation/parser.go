package validation

import (
	"github.com/gin-gonic/gin"
	"github.com/sanek1/metrics-collector/internal/handlers"
)

type Parser struct {
	services *handlers.Services
}

func NewParser(s *handlers.Services) *Parser {
	return &Parser{services: s}
}

func (p *Parser) HandleMetrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		metrics, err := p.services.ParseMetricsServices(c)
		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			c.Abort()
			return
		}
		c.Set("metrics", metrics)
		c.Next()
	}
}
