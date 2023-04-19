package component

import "github.com/gin-gonic/gin"

type Response struct {
	RequestID uint32 `json:"request_id"`
	Code      uint32 `json:"code"`
	Message   string `json:"message"`
}

type ResponseDataExtension struct {
	Data      any `json:"data,omitempty"`
	Extension any `json:"ext,omitempty"`
}

type GenericResponse struct {
	Response
	ResponseDataExtension
}

func NewGenericResponse(c *gin.Context, code uint32, message string, data any, extension any) *GenericResponse {
	r := GenericResponse{
		Response{
			RequestID: c.Value(ContextRequestID).(uint32),
			Code:      code,
			Message:   message,
		},
		ResponseDataExtension{
			Data:      data,
			Extension: extension,
		},
	}
	return &r
}
