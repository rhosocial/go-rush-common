package component

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupRouterErrorHandler(useNextFunc func()) *gin.Engine {
	r := gin.New()
	r.Use(ErrorHandler())
	r.Use(func(c *gin.Context) {
		useNextFunc()
	})
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})
	r.GET("/ping_error", func(c *gin.Context) {
		panic(errors.New("ping error occurred"))
	})
	return r
}

func TestErrorHandler(t *testing.T) {
	useNext := false
	r := setupRouterErrorHandler(func() {
		useNext = true
	})

	t.Run("Error(s) not occurred", func(t *testing.T) {
		useNext = false
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/ping", nil)

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "pong", w.Body.String())
		assert.Equal(t, true, useNext)
	})

	t.Run("Error(s) occurred", func(t *testing.T) {
		useNext = false
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/ping_error", nil)

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, true, useNext)
	})
}
