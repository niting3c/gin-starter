//go:build wireinject
// +build wireinject

package main

import (
	"starter/internal/app/controllers"
	Repository "starter/internal/app/repository"
	"starter/internal/app/services"
	"starter/internal/config"

	"github.com/google/wire"
)

func InitializeApplication() *Application {
	wire.Build(config.ConnectDB,
		Repository.NewUserRepository,
		services.NewUserService,
		controllers.NewUserController,
		controllers.NewInternalController,
		Repository.NewCRUDRepository,
		services.NewDefaultRestCaller,
		NewRouter,
		NewApplication,
	)
	return nil // The actual return value will be filled in by Wire
}
