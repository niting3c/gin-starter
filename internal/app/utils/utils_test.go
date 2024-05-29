package utils

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupRouter() *gin.Engine {
	router := gin.Default()
	return router
}

func TestStringContains(t *testing.T) {
	tests := []struct {
		slice []string
		str   string
		want  bool
	}{
		{[]string{"apple", "banana", "cherry"}, "banana", true},
		{[]string{"apple", "banana", "cherry"}, "mango", false},
		{[]string{}, "mango", false},
	}

	for _, tt := range tests {
		t.Run(tt.str, func(t *testing.T) {
			got := StringContains(tt.slice, tt.str)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestConvertToInt64(t *testing.T) {
	tests := []struct {
		name    string
		val     interface{}
		want    int64
		wantErr bool
	}{
		{"int64", int64(123), 123, false},
		{"int", int(123), 123, false},
		{"int32", int32(123), 123, false},
		{"int16", int16(123), 123, false},
		{"int8", int8(123), 123, false},
		{"float64", float64(123.456), 123, false},
		{"float32", float32(123.456), 123, false},
		{"string", "123", 123, false},
		{"string invalid", "abc", 0, true},
		{"unsupported", []int{1, 2, 3}, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ConvertToInt64(tt.val)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestGetEnvAsInt(t *testing.T) {
	const envName = "TEST_ENV_INT"
	const defaultValue = 42

	os.Unsetenv(envName)
	got := GetEnvAsInt(envName, defaultValue)
	assert.Equal(t, defaultValue, got, "Expected default value when environment variable is not set")

	expectedValue := 100
	os.Setenv(envName, strconv.Itoa(expectedValue))
	got = GetEnvAsInt(envName, defaultValue)
	assert.Equal(t, expectedValue, got, "Expected value should match the environment variable")

	os.Setenv(envName, "not_an_int")
	got = GetEnvAsInt(envName, defaultValue)
	assert.Equal(t, defaultValue, got, "Expected default value when environment variable is not a valid integer")

	os.Unsetenv(envName)
}

func TestGetEnvAsString(t *testing.T) {
	const envName = "TEST_ENV_STRING"
	const defaultValue = "default"

	os.Unsetenv(envName)
	got := GetEnvAsString(envName, defaultValue)
	assert.Equal(t, defaultValue, got, "Expected default value when environment variable is not set")

	expectedValue := "test_value"
	os.Setenv(envName, expectedValue)
	got = GetEnvAsString(envName, defaultValue)
	assert.Equal(t, expectedValue, got, "Expected value should match the environment variable")
	os.Unsetenv(envName)
}

func TestPagination_GetOffset(t *testing.T) {
	p := Pagination{Page: 2, Limit: 10}
	expectedOffset := 10
	assert.Equal(t, expectedOffset, p.GetOffset(), "Offset calculation is incorrect")
}

func TestPagination_GetLimit(t *testing.T) {
	p := Pagination{}
	defaultLimit := 10
	assert.Equal(t, defaultLimit, p.GetLimit(), "Default limit should be 10")

	p.Limit = 20
	assert.Equal(t, 20, p.GetLimit(), "Set limit should be respected")
}

func TestPagination_GetPage(t *testing.T) {
	p := Pagination{}
	defaultPage := 1
	assert.Equal(t, defaultPage, p.GetPage(), "Default page should be 1")

	p.Page = 2
	assert.Equal(t, 2, p.GetPage(), "Set page should be respected")
}

func TestPagination_GetSort(t *testing.T) {
	p := Pagination{}
	defaultSort := "\"id\" desc"
	assert.Equal(t, defaultSort, p.GetSort(), "Default sort should be '\"id\" desc'")

	p.Sort = "\"name\" asc"
	assert.Equal(t, "\"name\" asc", p.GetSort(), "Set sort should be respected")
}

func TestPaginateQueryExtractor(t *testing.T) {
	router := gin.Default()
	router.GET("/test", func(c *gin.Context) {
		validSortFields := []string{"id", "name"}
		pagination, errMsg := PaginateQueryExtractor(c, validSortFields)
		if errMsg != nil {
			c.JSON(errMsg.StatusCode, gin.H{"error": errMsg.Message})
			return
		}
		c.JSON(200, pagination)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test?page=2&per_page=15&sort=name&sortDesc=true", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code, "Expected successful response")
}

func TestRespondJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := setupRouter()

	router.GET("/testRespondJSON", func(c *gin.Context) {
		RespondJSON(c, http.StatusOK, gin.H{"message": "ok"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/testRespondJSON", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"message":"ok"}`, w.Body.String())
}

func TestErrorResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := setupRouter()

	router.GET("/testErrorResponse", func(c *gin.Context) {
		ErrorResponse(c, http.StatusBadRequest, "error occurred")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/testErrorResponse", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.JSONEq(t, `{"error":"error occurred"}`, w.Body.String())
}

func TestErrorResponseWithRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := setupRouter()

	router.GET("/testErrorResponseWithRequestID", func(c *gin.Context) {
		c.Set("RequestID", "12345")
		ErrorResponse(c, http.StatusBadRequest, "error with request id")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/testErrorResponseWithRequestID", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.JSONEq(t, `{"error":"error with request id"}`, w.Body.String())
}

func TestError(t *testing.T) {
	m := ErrorMessage{Message: "hello", StatusCode: 500}
	assert.Equal(t, "hello", m.Message)
}

func TestPaginationExtractor(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		router := setupRouter()

		router.GET("/testPaginateExtractor", func(c *gin.Context) {
			PaginateQueryExtractor(c, []string{})
		})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/testPaginateExtractor?per_page=-1&page=-1", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})
	t.Run("Failure", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		router := setupRouter()

		router.GET("/testPaginateExtractor", func(c *gin.Context) {
			_, err := PaginateQueryExtractor(c, []string{"allowed"})
			if err != nil {
				ErrorResponse(c, err.StatusCode, err.Message)
			}
		})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/testPaginateExtractor?per_page=-1&page=-1&sort=b", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
	t.Run("Failure 2", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		router := setupRouter()

		router.GET("/testPaginateExtractor", func(c *gin.Context) {
			_, err := PaginateQueryExtractor(c, []string{"b"})
			if err != nil {
				ErrorResponse(c, err.StatusCode, err.Error())
			}
		})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/testPaginateExtractor?per_page=-1&page=-1&sort=b&sortDesc=false", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Error message check", func(t *testing.T) {
		err := ErrorMessage{
			StatusCode: 0,
			Message:    "error",
		}
		assert.Equal(t, "error", err.Error())
	})
}

func Test_Int_contains(t *testing.T) {
	t.Run("Check pass", func(t *testing.T) {
		assert.Equal(t, true, IntContains([]int64{1, 2}, 1))
		assert.Equal(t, false, IntContains([]int64{1, 2}, 4))
	})
}
