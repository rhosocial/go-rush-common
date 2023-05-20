package response

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var server *httptest.Server

func setupResponseNewServer(t *testing.T, handler func(w http.ResponseWriter, r *http.Request)) {
	server = httptest.NewServer(http.HandlerFunc(handler))
}

func teardownResponseNewServer(t *testing.T) {
	server.Close()
}

func TestUnmarshalResponseBody(t *testing.T) {
	t.Run("Normal Case", func(t *testing.T) {
		requestID := uint32(time.Now().Unix())
		code := requestID + 1
		message := fmt.Sprintf("message_%d", code+1)
		setupResponseNewServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			response := Base{
				RequestID: requestID,
				Code:      code,
				Message:   message,
			}
			body, err := json.Marshal(response)
			if err != nil {
				t.Error(err)
				t.Fail()
			}
			_, err = w.Write(body)
			if err != nil {
				t.Fatal(err)
				return
			}
		})
		defer teardownResponseNewServer(t)

		resp, err := http.Get(server.URL)
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				t.Fatal(err)
			}
		}(resp.Body)
		if err != nil {
			t.Error(err)
			t.Fail()
		}
		body, err := UnmarshalResponseBodyBase(resp)
		if err != nil && err != io.EOF {
			t.Error(err)
			t.Fail()
		}
		if body == nil {
			t.Error("the body of response should not be nil.")
			t.Fail()
		}
		assert.Equal(t, requestID, body.RequestID)
		assert.Equal(t, code, body.Code)
		assert.Equal(t, message, body.Message)
	})
	t.Run("Bad Case", func(t *testing.T) {
		requestID := uint32(time.Now().Unix())
		code := requestID + 1
		message := fmt.Sprintf("message_%d", code+1)
		setupResponseNewServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			response := Base{
				RequestID: requestID,
				Code:      code,
				Message:   message,
			}
			body, err := json.Marshal(response)
			if err != nil {
				t.Error(err)
				t.Fail()
			}
			_, err = w.Write(append(body, byte(65)))
			if err != nil {
				t.Fatal(err)
				return
			}
		})
		defer teardownResponseNewServer(t)

		resp, err := http.Get(server.URL)
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				t.Fatal(err)
			}
		}(resp.Body)
		if err != nil {
			t.Error(err)
			t.Fail()
		}
		body, err := UnmarshalResponseBodyBase(resp)
		assert.NotNil(t, err, "error(s) should be reported due to invalid JSON string.")
		assert.Nil(t, body, "the body of response should be nil.")
	})
}

type StructWithScalar struct {
	Integer int    `json:"integer"`
	String  string `json:"string"`
	Boolean bool   `json:"boolean"`
}

type NestedStructWithScalarAndStruct struct {
	Struct  StructWithScalar `json:"struct"`
	String  string           `json:"string"`
	Integer int              `json:"integer"`
	Boolean bool             `json:"boolean"`
}

func TestUnmarshalResponseDataExtensionBody(t *testing.T) {
	t.Run("Normal Case 1: Data with scalar and Extension with any nil", func(t *testing.T) {
		requestID := uint32(time.Now().Unix())
		code := requestID + 1
		message := fmt.Sprintf("message_%d", code+1)
		data := code + 2
		setupResponseNewServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			response := Generic[int, any]{
				Base{
					RequestID: requestID,
					Code:      code,
					Message:   message,
				},
				DataAndExtension[int, any]{
					Data:      int(data),
					Extension: nil,
				},
			}
			body, err := json.Marshal(response)
			if err != nil {
				t.Error(err)
				t.Fail()
			}
			_, err = w.Write(body)
			if err != nil {
				t.Fatal(err)
				return
			}
		})
		defer teardownResponseNewServer(t)

		resp, err := http.Get(server.URL)
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				t.Fatal(err)
			}
		}(resp.Body)
		if err != nil {
			t.Error(err)
			t.Fail()
		}
		body, err := UnmarshalResponseBodyBaseWithDataAndExtension[int, any](resp)
		if err != nil && err != io.EOF {
			t.Error(err)
			t.Fail()
		}
		if body == nil {
			t.Error("the body of response should not be nil.")
			t.Fail()
		}
		assert.Equal(t, requestID, body.RequestID)
		assert.Equal(t, code, body.Code)
		assert.Equal(t, message, body.Message)
		assert.Equal(t, data, uint32(body.Data))
		assert.Nil(t, body.Extension)
	})
	t.Run("Normal Case 2: Data with struct and Extension with any nil", func(t *testing.T) {
		requestID := uint32(time.Now().Unix())
		code := requestID + 1
		message := fmt.Sprintf("message_%d", code+1)
		data := StructWithScalar{
			1, "2", true,
		}
		setupResponseNewServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			response := Generic[StructWithScalar, any]{
				Base{
					RequestID: requestID,
					Code:      code,
					Message:   message,
				},
				DataAndExtension[StructWithScalar, any]{
					Data:      data,
					Extension: nil,
				},
			}
			body, err := json.Marshal(response)
			if err != nil {
				t.Error(err)
				t.Fail()
			}
			_, err = w.Write(body)
			if err != nil {
				t.Fatal(err)
				return
			}
		})
		defer teardownResponseNewServer(t)

		resp, err := http.Get(server.URL)
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				t.Fatal(err)
			}
		}(resp.Body)
		if err != nil {
			t.Error(err)
			t.Fail()
		}
		body, err := UnmarshalResponseBodyBaseWithDataAndExtension[StructWithScalar, any](resp)
		if err != nil && err != io.EOF {
			t.Error(err)
			t.Fail()
		}
		if body == nil {
			t.Error("the body of response should not be nil.")
			t.Fail()
		}
		assert.Equal(t, requestID, body.RequestID)
		assert.Equal(t, code, body.Code)
		assert.Equal(t, message, body.Message)
		assert.Equal(t, data, body.Data)
		assert.Nil(t, body.Extension)
	})
	t.Run("Normal Case 3: Data with nested struct and Extension with any nil", func(t *testing.T) {
		requestID := uint32(time.Now().Unix())
		code := requestID + 1
		message := fmt.Sprintf("message_%d", code+1)
		data := NestedStructWithScalarAndStruct{
			StructWithScalar{
				2, "3", false,
			}, "1", 2, true,
		}
		setupResponseNewServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			response := Generic[NestedStructWithScalarAndStruct, any]{
				Base{
					RequestID: requestID,
					Code:      code,
					Message:   message,
				},
				DataAndExtension[NestedStructWithScalarAndStruct, any]{
					Data:      data,
					Extension: nil,
				},
			}
			body, err := json.Marshal(response)
			if err != nil {
				t.Error(err)
				t.Fail()
			}
			_, err = w.Write(body)
			if err != nil {
				t.Fatal(err)
				return
			}
		})
		defer teardownResponseNewServer(t)

		resp, err := http.Get(server.URL)
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				t.Fatal(err)
			}
		}(resp.Body)
		if err != nil {
			t.Error(err)
			t.Fail()
		}
		body, err := UnmarshalResponseBodyBaseWithDataAndExtension[NestedStructWithScalarAndStruct, any](resp)
		if err != nil && err != io.EOF {
			t.Error(err)
			t.Fail()
		}
		if body == nil {
			t.Error("the body of response should not be nil.")
			t.Fail()
		}
		assert.Equal(t, requestID, body.RequestID)
		assert.Equal(t, code, body.Code)
		assert.Equal(t, message, body.Message)
		assert.Equal(t, data, body.Data)
		assert.Nil(t, body.Extension)
	})
}
