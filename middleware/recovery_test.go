package middleware

import (
	"bytes"
	"github.com/icodechef/potgo"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRecovery(t *testing.T) {
	buf := new(bytes.Buffer)

	app := potgo.New()
	app.Use(RecoveryWithWriter(buf))

	app.GET("/test", func(c *potgo.Context) error {
		panic("abc")
		return c.Text("xyz")
	})

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)

	app.ServeHTTP(res, req)

	assert.Contains(t, buf.String(), "panic recovered")
	assert.Equal(t, http.StatusInternalServerError, res.Code)
	assert.Equal(t, "abc", res.Body.String())
	assert.NotContains(t, "xyz", res.Body.String())
}
