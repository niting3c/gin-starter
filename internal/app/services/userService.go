package services

import (
	"net/http"
	"starter/internal/app/models"
	Repository "starter/internal/app/repository"
	"starter/internal/app/utils"
)

//go:generate mockery --name UserService
type UserService interface {
	GetUserByEmail(emailId string) (*models.User, *utils.ErrorMessage)
}

type userHandler struct {
	aesKey   string
	userRepo Repository.UserRepository
}

func NewUserService(userRepo Repository.UserRepository) UserService {
	aesKey := utils.GetEnvAsString("AES_KEY", "1234567812345678")
	return &userHandler{userRepo: userRepo, aesKey: aesKey}
}

func (us *userHandler) GetUserByEmail(emailId string) (*models.User, *utils.ErrorMessage) {
	user, err := us.userRepo.Get(emailId)
	if err != nil {
		return nil, &utils.ErrorMessage{StatusCode: http.StatusInternalServerError, Message: "Failed to fetch user details, please try again."}
	}
	return user, nil
}
