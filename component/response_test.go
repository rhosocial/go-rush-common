package component

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
			response := Response{
				RequestID: requestID,
				Code:      code,
				Message:   message,
			}
			body, err := json.Marshal(response)
			if err != nil {
				t.Error(err)
				t.Fail()
			}
			w.Write(body)
		})
		defer teardownResponseNewServer(t)

		resp, err := http.Get(server.URL)
		defer resp.Body.Close()
		if err != nil {
			t.Error(err)
			t.Fail()
		}
		body, err := UnmarshalResponseBody(resp)
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
			response := Response{
				RequestID: requestID,
				Code:      code,
				Message:   message,
			}
			body, err := json.Marshal(response)
			if err != nil {
				t.Error(err)
				t.Fail()
			}
			w.Write(append(body, byte(65)))
		})
		defer teardownResponseNewServer(t)

		resp, err := http.Get(server.URL)
		defer resp.Body.Close()
		if err != nil {
			t.Error(err)
			t.Fail()
		}
		body, err := UnmarshalResponseBody(resp)
		assert.NotNil(t, err, "error(s) should be reported due to invalid JSON string.")
		assert.Nil(t, body, "the body of response should be nil.")
	})
}
