package Repository

import (
	"starter/internal/app/models"
	"starter/internal/app/utils"

	"github.com/jackc/pgx/v5"
	"github.com/sirupsen/logrus"
)

const USER = "users"

func NewUserRepository(crudRepository CRUDRepository) UserRepository {
	return &UserRepoHandler{
		crudRepository: crudRepository,
	}
}

//go:generate mockery --name UserRepository
type UserRepository interface {
	Create(user *models.User) (*models.User, *utils.ErrorMessage)
	Delete(emailId string) *utils.ErrorMessage
	Get(emailId string) (*models.User, *utils.ErrorMessage)
	UpdatePassword(email string, hashedPass string, salt string) *utils.ErrorMessage

	UpdateUserSelfDetails(currentEmail string, user *models.User) *utils.ErrorMessage
	ListAllUsers() ([]interface{}, *utils.ErrorMessage)
	GetUserByID(id int64) (*models.User, *utils.ErrorMessage)
}

type UserRepoHandler struct {
	crudRepository CRUDRepository
}

func (u *UserRepoHandler) Create(user *models.User) (*models.User, *utils.ErrorMessage) {
	logrus.Debug("Creating User")
	query := `INSERT INTO "public"."users" ("userEmailId", "encrypted_password", "inserted_at", "updated_at", "userDisplayName","userFirstName","userLastName","userRole","stored_salt")
			VALUES ($1, $2, $3, $4, $5,$6 ,$7,$8,$9) RETURNING "id"`
	id, err := u.crudRepository.Create(query, USER, user.UserEmailId, user.EncryptedPassword, user.InsertedAt, user.UpdatedAt, user.UserDisplayName, user.UserFirstName, user.UserLastName, user.UserRole, user.StoredSalt)
	v, _ := id.(int64)
	user.ID = v
	return user, err
}

func (u *UserRepoHandler) Delete(emailId string) *utils.ErrorMessage {
	logrus.Debug("Getting User from EmailId:", emailId)
	query := `DELETE FROM "public"."users" WHERE "userEmailId"=$1`
	if err := u.crudRepository.Delete(query, USER, emailId); err != nil {
		logrus.Error("Failed to delete User from database")
		return err
	}
	return nil
}

func (u *UserRepoHandler) GetUserByID(id int64) (*models.User, *utils.ErrorMessage) {
	query := `SELECT "id", "userEmailId", "inserted_at", "updated_at",
       			   "userDisplayName", "userFirstName", "userLastName", "userRole" 
			FROM "public"."users" 
			WHERE "id"=$1;`
	user, err := u.crudRepository.GetOne(query, USER, userMapperWithoutPassword, id)
	v, _ := user.(*models.User)
	return v, err
}

func (u *UserRepoHandler) Get(emailId string) (*models.User, *utils.ErrorMessage) {
	logrus.Debug("Getting User from EmailId:", emailId)
	query := `SELECT "id","userEmailId","encrypted_password","inserted_at","updated_at","userDisplayName","userFirstName","userLastName","userRole","stored_salt"  FROM "public"."users" WHERE "userEmailId"=$1;`
	user, err := u.crudRepository.GetOne(query, USER, userMapper, emailId)
	v, _ := user.(*models.User)
	return v, err
}

func (u *UserRepoHandler) UpdatePassword(email string, hashedPass string, salt string) *utils.ErrorMessage {
	query := `UPDATE "public"."users" 
			SET "encrypted_password"=$2, 
			    "stored_salt"=$3
			WHERE "userEmailId"=$1;`
	return u.crudRepository.Update(query, USER, email, hashedPass, salt)
}

func (u *UserRepoHandler) UpdateUserSelfDetails(currentEmail string, user *models.User) *utils.ErrorMessage {
	query := `UPDATE  "public"."users"  
		   SET     "userEmailId"= $1,
		           "updated_at"=NOW(),
		           "userDisplayName"=$2,
		           "userFirstName"=$3,
		           "userLastName"=$4
		   WHERE "userEmailId"=$5;`
	return u.crudRepository.Update(query, USER,
		user.UserEmailId, user.UserDisplayName,
		user.UserFirstName, user.UserLastName,
		currentEmail)
}

func (u *UserRepoHandler) ListAllUsers() ([]interface{}, *utils.ErrorMessage) {
	query := `SELECT "id", "userEmailId", "inserted_at", "updated_at",
       			   "userDisplayName", "userFirstName", "userLastName", "userRole" 
			FROM "public"."users";`
	return u.crudRepository.Get(query, USER, userMapperWithoutPassword)
}

var userMapperWithoutPassword = func(row pgx.Row) (interface{}, error) {
	var user models.User
	err := row.Scan(&user.ID, &user.UserEmailId, &user.InsertedAt, &user.UpdatedAt, &user.UserDisplayName, &user.UserFirstName, &user.UserLastName, &user.UserRole)
	if err != nil {
		logrus.Errorf("Failed to scan user: %v", err)
	}
	return &user, err
}

var userMapper = func(row pgx.Row) (interface{}, error) {
	var user models.User
	err := row.Scan(&user.ID, &user.UserEmailId, &user.EncryptedPassword, &user.InsertedAt, &user.UpdatedAt, &user.UserDisplayName, &user.UserFirstName, &user.UserLastName, &user.UserRole, &user.StoredSalt)
	if err != nil {
		logrus.Errorf("Failed to scan user: %v", err)
	}
	return &user, err
}
