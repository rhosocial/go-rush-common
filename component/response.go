package component

import "github.com/gin-gonic/gin"

type GenericResponse struct {
	RequestID uint32 `json:"request_id,omitempty"`
	Code      uint32 `json:"code,omitempty"`
	Message   string `json:"message,omitempty"`
	Data      any    `json:"data,omitempty"`
	Extension any    `json:"ext,omitempty"`
}

func NewGenericResponse(c *gin.Context, code uint32, message string, data any, extension any) *GenericResponse {
	r := GenericResponse{
		RequestID: c.Value(ContextRequestID).(uint32),
		Code:      code,
		Message:   message,
		Data:      data,
		Extension: extension,
	}
	return &r
}
