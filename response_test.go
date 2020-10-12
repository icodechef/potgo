package potgo

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestResponse_Write(t *testing.T) {
	res := &Response{}
	res.reset(httptest.NewRecorder())

	assert.Equal(t, http.StatusOK, res.status)
	assert.Equal(t, http.StatusOK, res.Status())

	assert.Equal(t, -1, res.size)
	assert.Equal(t, -1, res.Size())
	assert.False(t, res.Written())

	res.Write([]byte("foo"))

	assert.True(t, res.Written())
	assert.Equal(t, len("foo"), res.Size())

	res.Write([]byte("bar"))

	assert.True(t, res.Written())
	assert.Equal(t, len("foobar"), res.Size())
}

func TestResponse_WriteHeader(t *testing.T) {
	res := &Response{}
	res.reset(httptest.NewRecorder())

	res.WriteHeader(http.StatusNotFound)
	assert.False(t, res.Written())
	assert.Equal(t, http.StatusNotFound, res.Status())

	res.WriteHeader(-1)
	assert.Equal(t, http.StatusNotFound, res.Status())
}

func TestResponse_Reset(t *testing.T) {
	res := &Response{}
	res.reset(httptest.NewRecorder())

	res.WriteHeader(http.StatusFound)
	res.Writer.Header().Set("foo", "bar")

	assert.Equal(t, http.StatusFound, res.Status())
	assert.False(t, res.Written())
	assert.Equal(t, "bar", res.Writer.Header().Get("foo"))

	res.Reset(httptest.NewRecorder())
	_, _ = res.Write([]byte("foo"))

	assert.Equal(t, http.StatusOK, res.Status())
	assert.True(t, res.Written())
	assert.Equal(t, len("foo"), res.Size())
	assert.Equal(t, "", res.Writer.Header().Get("foo"))

	n := res.Reset(httptest.NewRecorder())

	assert.False(t, n)
}
