package component

import (
	"github.com/gin-gonic/gin"
)

type ControllerInterface interface {
	RegisterActions(r *gin.Engine)
}

type GenericController struct {
	ControllerInterface
}
