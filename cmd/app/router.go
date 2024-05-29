package main

import (
	"net/http"
	"starter/internal/app/controllers"
	"starter/internal/app/middlewares"
	"starter/internal/app/utils"
	"starter/internal/config"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

//go:generate mockery --name Router
type Router interface {
	SetupRouter() *gin.Engine
}
type router struct {
	db                 config.DBPool
	userController     controllers.UserController
	internalController controllers.InternalController
}

func NewRouter(db config.DBPool, internalController controllers.InternalController, userController controllers.UserController) Router {

	return &router{
		db:                 db,
		userController:     userController,
		internalController: internalController,
	}
}

func (r *router) SetupRouter() *gin.Engine {
	gin.SetMode(utils.GetEnvAsString("GIN_MODE", gin.DebugMode))
	ginRouter := gin.New()
	ginRouter.HandleMethodNotAllowed = true
	//Setting up middlewares
	// Global middlewares
	ginRouter.Use(middlewares.RequestIDMiddleware())

	//logrus configuration
	loglevel := utils.GetEnvAsString("APPLICATION_LOG_LEVEL", "info")
	level, _ := logrus.ParseLevel(loglevel)
	log := logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{})
	log.SetLevel(level)
	logrus.Infof("Setting level:%v", level)
	ginRouter.Use(middlewares.LoggerMiddleware(log))

	ginRouter.Use(gin.Recovery())

	// Get the rate limit, with a default value
	reqPerSec := utils.GetEnvAsInt("RATE_LIMIT_PER_SEC", 100)
	// Set up rate limiter
	limiter := rate.NewLimiter(rate.Limit(reqPerSec), reqPerSec)

	//Setup Internal Route
	controllers.SetupInternalRoute(ginRouter, r.internalController, limiter)
	//Setup User controller router
	controllers.SetupUserRoute(ginRouter, r.userController, limiter)
	return ginRouter
}
func testResponse(c *gin.Context) {
	c.String(http.StatusRequestTimeout, "timeout")
}
