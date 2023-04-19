package component

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestValidatePassword(t *testing.T) {
	t.Run("Generating 10 password hashes", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			result, _ := bcrypt.GenerateFromPassword([]byte("password"), i+4)
			t.Log(string(result))
		}
	})
	t.Run("Validate Password", func(t *testing.T) {
		for i := 0; i < 5; i++ {
			random := strconv.FormatUint(rand.Uint64(), 10) + strconv.FormatUint(rand.Uint64(), 10)
			result, _ := bcrypt.GenerateFromPassword([]byte(random), i+4)
			if !ValidatePassword(string(result), random) {
				t.Error("Password compared not equal.")
			}
		}
	})
}

func setupRouterAuthRequired(useNextFunc func()) *gin.Engine {
	r := gin.New()
	r.Use(AppendRequestID(), AuthRequired(), func(c *gin.Context) {
		useNextFunc()
	})
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})
	return r
}

func TestUserAuthRequired(t *testing.T) {
	useNext := false
	r := setupRouterAuthRequired(func() {
		useNext = true
	})

	t.Run("Missing header", func(t *testing.T) {
		useNext = false
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/ping", nil)

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		body := GenericResponse{}
		assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		assert.Equal(t, MessageMissingHeaderXAuthorizationToken, body.Message)
		assert.Equal(t, false, useNext)
	})

	t.Run("Header exists, but with wrong value", func(t *testing.T) {
		useNext = false
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/ping", nil)
		req.Header.Add(HeaderXAuthorizationToken, "1")

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		body := GenericResponse{}
		assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		assert.Equal(t, MessageInvalidAuthorizationToken, body.Message)
		assert.Equal(t, false, useNext)
	})

	t.Run("Header exists with correct value", func(t *testing.T) {
		result, _ := bcrypt.GenerateFromPassword([]byte("password"), 13)
		useNext = false
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/ping", nil)
		req.Header.Add(HeaderXAuthorizationToken, string(result))

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "pong", w.Body.String())
		assert.Equal(t, true, useNext)
	})
}
