package middleware

import (
	"bytes"
	"github.com/icodechef/potgo"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLogger(t *testing.T) {
	buf := new(bytes.Buffer)

	app := potgo.New()
	app.Use(LoggerWithWriter(buf))
	app.GET("/test", func(c *potgo.Context) error {
		return nil
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	app.ServeHTTP(res, req)

	assert.Contains(t, buf.String(), "200")
	assert.Contains(t, buf.String(), "GET")
	assert.Contains(t, buf.String(), "/test")
}
