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

type ResponseDataExtension[T1 interface{}, T2 interface{}] struct {
	Data      T1 `json:"data,omitempty"`
	Extension T2 `json:"ext,omitempty"`
}

type GenericResponse[T1 interface{}, T2 interface{}] struct {
	Response
	ResponseDataExtension[T1, T2]
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

func UnmarshalResponseDataExtension[T1 interface{}, T2 interface{}](resp *http.Response) (*GenericResponse[T1, T2], error) {
	if resp == nil {
		return nil, nil
	}
	body := make([]byte, resp.ContentLength)
	if _, err := resp.Body.Read(body); err != nil && err != io.EOF {
		return nil, err
	}
	var r GenericResponse[T1, T2]
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

func NewGenericResponse(c *gin.Context, code uint32, message string, data any, extension any) *GenericResponse[interface{}, interface{}] {
	r := GenericResponse[interface{}, interface{}]{
		Response{
			RequestID: c.Value(ContextRequestID).(uint32),
			Code:      code,
			Message:   message,
		},
		ResponseDataExtension[interface{}, interface{}]{
			Data:      data,
			Extension: extension,
		},
	}
	return &r
}
