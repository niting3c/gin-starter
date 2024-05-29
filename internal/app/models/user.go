package models

import (
	"fmt"
	"net/http"
	"net/mail"
	"starter/internal/app/constants"
	"starter/internal/app/utils"
	"time"
)

type User struct {
	ID                int64     `json:"id"`
	UserEmailId       string    `json:"userEmailId"`
	EncryptedPassword string    `json:"encrypted_password"`
	InsertedAt        time.Time `json:"inserted_at"`
	UpdatedAt         time.Time `json:"updated_at"`
	UserDisplayName   string    `json:"userDisplayName"`
	UserFirstName     string    `json:"userFirstName"`
	UserLastName      string    `json:"userLastName"`
	UserRole          string    `json:"userRole"`
	StoredSalt        string    `json:"stored_salt"`
}

type UserResponseDto struct {
	AdminRole       bool   `json:"adminRole"`
	CanViewLogsRole bool   `json:"canViewLogsRole"`
	TesterRole      bool   `json:"testerRole"`
	UserDisplayName string `json:"userDisplayName"`
	UserEmailId     string `json:"userEmailId"`
	UserFirstName   string `json:"userFirstName"`
	UserLastName    string `json:"userLastName"`
	ViewerRole      bool   `json:"viewerRole"`
	UserId          int64  `json:"userId"`
	UserRole        string `json:"userRole"`
}

type UserRequestDto struct {
	UserEmailId     string `json:"userEmailId"`
	UserPassword    string `json:"userPassword"`
	UserDisplayName string `json:"userDisplayName"`
	UserFirstName   string `json:"userFirstName"`
	UserLastName    string `json:"userLastName"`
	UserRole        string `json:"userRole"`
}

func (c *UserRequestDto) Validate() *utils.ErrorMessage {
	if len(c.UserEmailId) == 0 {
		return &utils.ErrorMessage{Message: constants.INVALID_EMAILID, StatusCode: http.StatusBadRequest}
	}
	_, err := mail.ParseAddress(c.UserEmailId)
	if err != nil {
		return &utils.ErrorMessage{Message: constants.INVALID_EMAILID, StatusCode: http.StatusBadRequest}
	}
	if len(c.UserPassword) < 8 {
		return &utils.ErrorMessage{Message: constants.INVALID_PASSWORD, StatusCode: http.StatusBadRequest}
	}
	if len(c.UserDisplayName) == 0 {
		return &utils.ErrorMessage{Message: fmt.Sprintf(constants.EMPTY_FIELD, "UserDisplayName"), StatusCode: http.StatusBadRequest}
	}
	if len(c.UserFirstName) == 0 {
		return &utils.ErrorMessage{Message: fmt.Sprintf(constants.EMPTY_FIELD, "UserFirstName"), StatusCode: http.StatusBadRequest}
	}
	if len(c.UserLastName) == 0 {
		return &utils.ErrorMessage{Message: fmt.Sprintf(constants.EMPTY_FIELD, "UserLastName"), StatusCode: http.StatusBadRequest}
	}
	if len(c.UserRole) == 0 {
		return &utils.ErrorMessage{Message: fmt.Sprintf(constants.EMPTY_FIELD, "UserRole"), StatusCode: http.StatusBadRequest}
	}
	return nil
}
