package component

import (
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

func ValidatePassword(passHash string, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(passHash), []byte(password)) == nil
}

func AuthRequired() gin.HandlerFunc {
	return authHandlerFunc
}

const HeaderXAuthorizationToken = "X-Authorization-Token"
const MessageMissingHeaderXAuthorizationToken = "empty authorization token"
const MessageInvalidAuthorizationToken = "invalid authorization token"

type HeaderAuthorizationToken struct {
	AuthorizationToken string `header:"X-Authorization-Token" binding:"required"`
}

func (a *HeaderAuthorizationToken) validate(password string) bool {
	return ValidatePassword(a.AuthorizationToken, password)
}

func authHandlerFunc(c *gin.Context) {
	h := HeaderAuthorizationToken{}
	err := c.ShouldBindHeader(&h)
	if err != nil {
		c.AbortWithStatusJSON(
			http.StatusForbidden,
			NewGenericResponse(c, 1, MessageMissingHeaderXAuthorizationToken, nil, nil),
		)
	} else if !h.validate("password") {
		c.AbortWithStatusJSON(
			http.StatusForbidden,
			NewGenericResponse(c, 1, MessageInvalidAuthorizationToken, nil, nil),
		)
	}
	c.Next()
}
