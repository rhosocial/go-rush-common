package component

import (
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
	t.Run("Generating 5 password hashes", func(t *testing.T) {
		for i := 0; i < 5; i++ {
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

func setupRouter() *gin.Engine {
	r := gin.New()
	r.Use(AuthRequired())
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})
	return r
}

func TestUserAuthRequired(t *testing.T) {
	r := setupRouter()

	t.Run("Missing header", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/ping", nil)

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Header exists, but with wrong value", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/ping", nil)
		req.Header.Add(HeaderXAuthorizationToken, "1")

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Header exists with correct value", func(t *testing.T) {
		result, _ := bcrypt.GenerateFromPassword([]byte("password"), 13)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/ping", nil)
		req.Header.Add(HeaderXAuthorizationToken, string(result))

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}
