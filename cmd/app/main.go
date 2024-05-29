package main

import (
	"io"
	"net/http"
	"os"
	_ "starter/docs"
	"starter/internal/app/utils"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

var serverPort = "4000"

func Init(app *Application, appType string) *Application {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		logrus.Warn("No .env file found")
	}
	//logrus configuration
	loglevel := utils.GetEnvAsString("APPLICATION_LOG_LEVEL", "info")
	level, _ := logrus.ParseLevel(loglevel)
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(level)
	// Set up logging to both file and console
	logFile, err := os.OpenFile("application.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0666)
	if err != nil {
		logrus.Info("Failed to log to file, using default stderr")
	} else {
		mw := io.MultiWriter(os.Stdout, logFile)
		logrus.SetOutput(mw)
	}
	return app
}

// @title           Golang Starter Application
// @version         1.0
// @description     Swagger APIS for a starter Application

// @contact.email   nitin1494gupta@gmail.com
// @contact.name    Nitin Gupta
// @host      localhost:4000
// @BasePath  /
// @externalDocs.description  OpenAPI
func main() {
	appType := os.Args[1]
	app := Init(InitializeApplication(), appType)

	defer app.db.Close()

	logrus.Info("Loading gin server")
	//Setup routes and start service
	r := app.routes.SetupRouter()
	server := &http.Server{
		Addr:    ":" + serverPort,
		Handler: r,
	}
	logrus.Fatal(server.ListenAndServe())
}
