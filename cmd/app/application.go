package main

import (
	"starter/internal/app/controllers"
	Repository "starter/internal/app/repository"
	"starter/internal/app/services"
	"starter/internal/config"
)

type Application struct {
	db             config.DBPool
	crudRepo       Repository.CRUDRepository
	restCaller     services.RestCaller
	routes         Router
	userController controllers.UserController
	userRepository Repository.UserRepository
}

func NewApplication(
	db config.DBPool,
	crudRepo Repository.CRUDRepository,
	userRepository Repository.UserRepository,
	restCaller services.RestCaller,
	routes Router,
	userController controllers.UserController) *Application {
	return &Application{
		db:             db,
		crudRepo:       crudRepo,
		restCaller:     restCaller,
		routes:         routes,
		userController: userController,
		userRepository: userRepository,
	}
}
