package component

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

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

func UnmarshalResponseBody(resp *http.Response) (*Response, error) {
	if resp == nil {
		return nil, nil
	}
	body := make([]byte, resp.ContentLength)
	if _, err := resp.Body.Read(body); err != nil && err != io.EOF {
		return nil, err
	}
	var r Response
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, err
	}
	return &r, nil
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
