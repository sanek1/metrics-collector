//go:build ignore
// +build ignore

// Package routing provides API routing functionality
// @title Metrics Collector API
// @version 1.0
// @description Сервис для сбора и хранения метрик
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@metrics-collector.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /
// @schemes http https
package routing

import (
	_ "github.com/sanek1/metrics-collector/docs"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// InitSwaggerUI initializes Swagger UI routes
// @title Metrics Collector API
// @version 1.0
// @description API для сбора и обработки метрик
// @host localhost:8080
// @BasePath /
func InitSwaggerUI(r *gin.Engine) {
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
}
