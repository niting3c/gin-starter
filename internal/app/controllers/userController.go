package controllers

import (
	"net/http"
	"starter/internal/app/constants"
	"starter/internal/app/middlewares"
	"starter/internal/app/services"
	"starter/internal/app/utils"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

//go:generate mockery --name UserController
type UserController interface {
	GetUserByEmail(c *gin.Context)
}

type userController struct {
	aesKey      string
	userService services.UserService
}

// GetUserByEmail Gets the user details by Email
// @Summary Gets the user details by Email
// @Description Gets the user details by Email
// @Produce json
// @Tags User
// @Param email path string true "User email"
// @Success 200 {object} models.UserResponseDto
// @Failure 400 {object} utils.ErrorMessage
// @Failure 500 {object} utils.ErrorMessage
// @Router /user/{email} [get]
func (uc *userController) GetUserByEmail(c *gin.Context) {
	email := c.Param("email")
	if len(email) == 0 {
		utils.ErrorResponse(c, http.StatusBadRequest, constants.INVALID_ID)
		return
	}
	user, svcErr := uc.userService.GetUserByEmail(email)
	if svcErr != nil {
		utils.ErrorResponse(c, svcErr.StatusCode, svcErr.Message)
		return
	}
	utils.RespondJSON(c, http.StatusOK, user)
}

func NewUserController(userService services.UserService) UserController {
	aesKey := utils.GetEnvAsString("AES_KEY", "1234567812345678")
	return &userController{aesKey: aesKey,
		userService: userService}
}

func SetupUserRoute(router *gin.Engine, userController UserController,
	limiter *rate.Limiter, ) {
	userRoutes := router.Group("/user")
	userRoutes.Use(middlewares.RateLimitMiddleware(limiter))
	userRoutes.Use(middlewares.TimeoutMiddleware())
	userRoutes.GET("/:email", userController.GetUserByEmail)
}
