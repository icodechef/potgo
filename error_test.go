package potgo

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestNewHttpError(t *testing.T) {
	e := NewHTTPError(http.StatusNotFound)
	assert.Equal(t, http.StatusNotFound, e.Status())
	assert.Equal(t, http.StatusText(http.StatusNotFound), e.Error())

	e = NewHTTPError(http.StatusNotFound, "foo")
	assert.Equal(t, http.StatusNotFound, e.Status())
	assert.Equal(t, "foo", e.Error())

	s, _ := json.Marshal(e)
	assert.Equal(t, `{"status":404,"message":"foo"}`, string(s))
}
