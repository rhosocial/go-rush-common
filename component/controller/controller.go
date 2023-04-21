package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/rhosocial/go-rush-common/component/response"
)

type Base interface {
	RegisterActions(r *gin.Engine)
}

type GenericController struct {
	Base
}

func (c *GenericController) NewResponseBase(r *gin.Context, code uint32, message string) *response.Base {
	return response.NewBase(r, code, message)
}

func (c *GenericController) NewResponseGeneric(r *gin.Context, code uint32, message string, data any, ext any) *response.Generic[any, any] {
	return response.NewGeneric(r, code, message, data, ext)
}
