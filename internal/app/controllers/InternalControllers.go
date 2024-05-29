package controllers

import (
	"context"
	"net/http"
	_ "starter/docs"
	"starter/internal/app/middlewares"
	"starter/internal/app/services"
	"starter/internal/app/utils"
	"starter/internal/config"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"golang.org/x/time/rate"
)

//go:generate mockery --name InternalController
type InternalController interface {
	SetLogLevel(c *gin.Context)
	HealthCheck(c *gin.Context)
}

type internal struct {
	db          config.DBPool
	userService services.UserService
}

func NewInternalController(db config.DBPool, userService services.UserService) InternalController {
	return &internal{db: db, userService: userService}
}

// SetLogLevel Sets Logrus Log level
// @Summary Set Log Level
// @Description Set Log Level
// @Produce json
// @Tags Internal
// @Param level path string true "Log Level"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} utils.ErrorMessage
// @Router /log/{level} [put]
func (i *internal) SetLogLevel(c *gin.Context) {
	level := c.Param("level")
	switch level {
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
	case "info":
		logrus.SetLevel(logrus.InfoLevel)
	default:
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid log level")
		return
	}
	utils.RespondJSON(c, http.StatusOK, gin.H{"message": "Log level set to " + level})
}

// HealthCheck Checks the health of the service
// @Summary Health Check
// @Description Checks the health of the service
// @Produce json
// @Tags Internal
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} utils.ErrorMessage
// @Router /health [get]
func (i *internal) HealthCheck(c *gin.Context) {
	err := i.db.Ping(context.Background())
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Database not available")
		return
	}
	utils.RespondJSON(c, http.StatusOK, gin.H{"status": "up"})
}

func SetupInternalRoute(router *gin.Engine, internalController InternalController, limiter *rate.Limiter) {
	swagger := router.Group("/swagger")

	swagger.GET("/*any", ginSwagger.WrapHandler(swaggerFiles.Handler,
		ginSwagger.DefaultModelsExpandDepth(-1)))

	internalRoutes := router.Group("/internal")
	internalRoutes.Use(middlewares.RateLimitMiddleware(limiter))

	pprof.Register(internalRoutes, "/pprof")
	internalRoutes.GET("/metrics", gin.WrapH(promhttp.Handler()))
	internalRoutes.PUT("/log/:level", internalController.SetLogLevel)
	internalRoutes.GET("/health", internalController.HealthCheck)
}
