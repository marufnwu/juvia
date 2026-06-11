package api

import (
	"time"

	"github.com/gin-gonic/gin"
)

type responseEnvelope struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Error   *apiError   `json:"error,omitempty"`
	Meta    meta        `json:"meta"`
}

type apiError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	UserMessage string `json:"user_message"`
}

type meta struct {
	RequestID string    `json:"request_id"`
	Timestamp time.Time `json:"timestamp"`
}

func envelope(success bool, data interface{}, code, message string) responseEnvelope {
	err := (*apiError)(nil)
	if code != "" {
		err = &apiError{
			Code:        code,
			Message:     message,
			UserMessage: message,
		}
	}
	return responseEnvelope{
		Success: success,
		Data:    data,
		Error:   err,
		Meta: meta{
			RequestID: "", // set by middleware if needed
			Timestamp: time.Now().UTC(),
		},
	}
}

func success(data interface{}) responseEnvelope {
	return envelope(true, data, "", "")
}

func fail(code, message string) responseEnvelope {
	return envelope(false, nil, code, message)
}

func getUserID(c *gin.Context) int64 {
	uid, _ := c.Get("user_id")
	if id, ok := uid.(int64); ok {
		return id
	}
	return 0
}
