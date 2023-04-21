package response

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rhosocial/go-rush-common/component/logger"
)

type Base struct {
	RequestID uint32 `json:"request_id"`
	Code      uint32 `json:"code"`
	Message   string `json:"message"`
}

type DataAndExtension[T1 interface{}, T2 interface{}] struct {
	Data      T1 `json:"data,omitempty"`
	Extension T2 `json:"ext,omitempty"`
}

type Generic[T1 interface{}, T2 interface{}] struct {
	Base
	DataAndExtension[T1, T2]
}

func UnmarshalResponseBodyBase(resp *http.Response) (*Base, error) {
	if resp == nil {
		return nil, nil
	}
	body := make([]byte, resp.ContentLength)
	if _, err := resp.Body.Read(body); err != nil && err != io.EOF {
		return nil, err
	}
	var r Base
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

func UnmarshalResponseBodyBaseWithDataAndExtension[T1 interface{}, T2 interface{}](resp *http.Response) (*Generic[T1, T2], error) {
	if resp == nil {
		return nil, nil
	}
	body := make([]byte, resp.ContentLength)
	if _, err := resp.Body.Read(body); err != nil && err != io.EOF {
		return nil, err
	}
	var r Generic[T1, T2]
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

func NewBase(c *gin.Context, code uint32, message string) *Base {
	r := Base{
		RequestID: c.Value(logger.ContextRequestID).(uint32),
		Code:      code,
		Message:   message,
	}
	return &r
}

func NewGeneric(c *gin.Context, code uint32, message string, data any, extension any) *Generic[interface{}, interface{}] {
	r := Generic[interface{}, interface{}]{
		Base{
			RequestID: c.Value(logger.ContextRequestID).(uint32),
			Code:      code,
			Message:   message,
		},
		DataAndExtension[interface{}, interface{}]{
			Data:      data,
			Extension: extension,
		},
	}
	return &r
}
