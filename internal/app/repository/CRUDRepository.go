package Repository

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"starter/internal/app/constants"
	"starter/internal/app/utils"
	"starter/internal/config"

	"github.com/jackc/pgx/v5"
	"github.com/sirupsen/logrus"
)

//go:generate mockery --name CRUDRepository
type CRUDRepository interface {
	BeginTransaction() (pgx.Tx, *utils.ErrorMessage)
	CommitTransaction(tx pgx.Tx, objectType string) *utils.ErrorMessage
	RollBackTransaction(tx pgx.Tx, objectType string) *utils.ErrorMessage
	Delete(query string, objectType string, args ...any) *utils.ErrorMessage
	Update(query string, objectType string, args ...any) *utils.ErrorMessage
	Create(query string, objectType string, args ...any) (interface{}, *utils.ErrorMessage)
	GetOne(query string, objectType string, mapper utils.RowMapperFunc, args ...any) (interface{}, *utils.ErrorMessage)
	Get(query string, objectType string, mapper utils.RowMapperFunc, args ...any) ([]interface{}, *utils.ErrorMessage)
	GetWithPagination(countSQL string, objectType string, finalSQL string, mapper utils.RowMapperFunc, pagination *utils.Pagination, args ...any) (*utils.Pagination, *utils.ErrorMessage)
}

type crudRepository struct {
	db   config.DBPool
	lock bool
}

func NewCRUDRepository(db config.DBPool) CRUDRepository {
	return &crudRepository{
		db: db,
	}
}

func (crud *crudRepository) CheckAndResetDBConnection() {
	for crud.lock {
	}

	if crud.db == nil {
		crud.lock = true
		logrus.Warn("DB connection is nil, resetting it")
		crud.db = config.ConnectDB()
		crud.lock = false
	}
	if err := crud.db.Ping(context.Background()); err != nil {
		crud.lock = true
		logrus.Warn("DB connection is stale, resetting it")
		crud.db = config.ConnectDB()
		crud.lock = false
	}
}

func (crud *crudRepository) Delete(query string, objectType string, args ...any) *utils.ErrorMessage {
	logrus.Debugf("Deleting %s object in database", objectType)
	// Begin a transaction
	tx, txErr := crud.BeginTransaction()
	if txErr != nil {
		return txErr
	}
	// Execute the DELETE statement within the transaction
	cmdTag, err := tx.Exec(context.Background(), query, args...)
	if err != nil {
		logrus.Errorf("Failed to delete %s from database: %v", objectType, err)
		return crud.RollBackTransaction(tx, objectType)
	}
	logrus.Infof("Rows Affected by delete:%v", cmdTag.RowsAffected())
	if cmdTag.RowsAffected() == 0 {
		logrus.Warnf("No %s found with the given criteria to delete", objectType)
		return nil
	}
	// Commit the transaction
	return crud.CommitTransaction(tx, objectType)
}
func (crud *crudRepository) Create(query string, objectType string, args ...any) (interface{}, *utils.ErrorMessage) {
	var id interface{}
	logrus.Debugf("Creating %s object in database", objectType)
	// Begin a transaction
	tx, txErr := crud.BeginTransaction()
	if txErr != nil {
		return -1, txErr
	}
	if err := tx.QueryRow(context.Background(), query, args...).Scan(&id); err != nil {
		logrus.Errorf("Failed to create %s in database: %v", objectType, err)
		logrus.Errorf("Rollign back Create transaction for %s", objectType)
		crud.RollBackTransaction(tx, objectType)
		return -1, &utils.ErrorMessage{StatusCode: http.StatusInternalServerError,
			Message: fmt.Sprintf(constants.FAILED_TO_CREATE_OBJ, objectType)}
	}

	cErr := crud.CommitTransaction(tx, objectType)
	if cErr != nil {
		return -1, crud.RollBackTransaction(tx, objectType)
	}
	return id, nil
}
func (crud *crudRepository) BeginTransaction() (pgx.Tx, *utils.ErrorMessage) {
	crud.CheckAndResetDBConnection()
	tx, err := crud.db.Begin(context.Background())
	if err != nil {
		logrus.Errorf("Failed to begin transaction: %v", err)
		return nil, &utils.ErrorMessage{
			StatusCode: http.StatusInternalServerError,
			Message:    constants.FAILED_BEGIN_TRANSACTION,
		}
	}
	return tx, nil
}

func (crud *crudRepository) CommitTransaction(tx pgx.Tx, objectType string) *utils.ErrorMessage {
	if err := tx.Commit(context.Background()); err != nil {
		logrus.Errorf("Failed to commit transaction for %s update: %v", objectType, err)
		return &utils.ErrorMessage{
			StatusCode: http.StatusInternalServerError,
			Message:    fmt.Sprintf(constants.COMMIT_FAILED, objectType),
		}
	}
	return nil
}

func (crud *crudRepository) RollBackTransaction(tx pgx.Tx, objectType string) *utils.ErrorMessage {
	if rollbackErr := tx.Rollback(context.Background()); rollbackErr != nil {
		logrus.Errorf("Failed to rollback transaction for %s update: %v", objectType, rollbackErr)
	}

	return &utils.ErrorMessage{
		StatusCode: http.StatusInternalServerError,
		Message:    fmt.Sprintf(constants.FAILED_TO_UPDATE_OBJ, objectType),
	}
}

func (crud *crudRepository) Update(query string, objectType string, args ...any) *utils.ErrorMessage {
	logrus.Debugf("Updating %s object in database", objectType)
	// Begin a transaction
	tx, txerr := crud.BeginTransaction()
	if txerr != nil {
		return txerr
	}
	// Execute the UPDATE statement within the transaction
	cmdTag, err := tx.Exec(context.Background(), query, args...)
	if err != nil {
		logrus.Errorf("Failed to update %s in database: %v", objectType, err)
		// If an error occurs, rollback the transaction
		return crud.RollBackTransaction(tx, objectType)
	}

	// Check if any row was actually updated
	if cmdTag.RowsAffected() == 0 {
		// No rows affected, might want to handle this as an error or just a no-op
		logrus.Warnf("No %s found with the given criteria to update", objectType)
		return &utils.ErrorMessage{
			StatusCode: http.StatusNotFound,
			Message:    fmt.Sprintf(constants.NO_ROWS_AFFECTED, objectType),
		}
	}
	return crud.CommitTransaction(tx, objectType)
}

func (crud *crudRepository) GetWithPagination(countSQL string, objectType string, finalSQL string, mapper utils.RowMapperFunc, pagination *utils.Pagination, args ...any) (*utils.Pagination, *utils.ErrorMessage) {
	crud.CheckAndResetDBConnection()
	ctx := context.Background()
	rows, err := crud.db.Query(ctx, finalSQL, args...)
	if err != nil {
		logrus.Errorf("Failed to execute query: %v for %s", err, objectType)
		return nil, &utils.ErrorMessage{StatusCode: http.StatusInternalServerError, Message: "Failed to execute query"}
	}
	defer rows.Close()
	var results []interface{}
	for rows.Next() {
		item, err := mapper(rows)
		if err != nil {
			logrus.Errorf("Failed to map row: %v", err)
			return nil, &utils.ErrorMessage{StatusCode: http.StatusInternalServerError, Message: constants.FAILED_SCAN}
		}
		results = append(results, item)
	}

	var totalRows int64
	err = crud.db.QueryRow(ctx, countSQL).Scan(&totalRows)
	if err != nil {
		logrus.Errorf("Failed to count total rows: %v", err)
		return nil, &utils.ErrorMessage{StatusCode: http.StatusInternalServerError, Message: "Failed to count total rows"}
	}

	pagination.TotalRows = totalRows
	pagination.TotalPages = int(math.Ceil(float64(totalRows) / float64(pagination.GetLimit())))
	pagination.Rows = results
	return pagination, nil
}

func (crud *crudRepository) Get(query string, objectType string, mapper utils.RowMapperFunc, args ...any) ([]interface{}, *utils.ErrorMessage) {
	crud.CheckAndResetDBConnection()
	ctx := context.Background()
	rows, err := crud.db.Query(ctx, query, args...)
	if err != nil {
		logrus.Errorf("Failed to execute query: %v for %v", err, objectType)
		return nil, &utils.ErrorMessage{StatusCode: http.StatusInternalServerError, Message: "Failed to execute query"}
	}
	defer rows.Close()

	var results []interface{}
	for rows.Next() {
		item, err := mapper(rows)
		if err != nil {
			logrus.Errorf("Failed to map row: %v", err)
			return nil, &utils.ErrorMessage{StatusCode: http.StatusInternalServerError, Message: "Failed to scan row"}
		}
		results = append(results, item)
	}
	return results, nil
}

func (crud *crudRepository) GetOne(query string, objectType string, mapper utils.RowMapperFunc, args ...any) (interface{}, *utils.ErrorMessage) {
	crud.CheckAndResetDBConnection()
	row := crud.db.QueryRow(context.Background(), query, args...)
	item, err := mapper(row)
	if err != nil {
		logrus.Debugf("Adding Query : %v \n", query)
		logrus.Errorf("Failed to execute query or map row: %v", err)
		if err == pgx.ErrNoRows {
			return nil, &utils.ErrorMessage{StatusCode: http.StatusNotFound, Message: fmt.Sprintf("No %s found with the given criteria", objectType)}
		}
		return nil, &utils.ErrorMessage{StatusCode: http.StatusInternalServerError, Message: fmt.Sprintf(constants.FAILED_SCAN)}
	}
	return item, nil
}
