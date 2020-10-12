package potgo

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRoute_Name(t *testing.T) {
	route := &Route{}
	route.Name("user-edit")
	assert.Equal(t, "user-edit", route.name)
}

func TestRoute_BuildPathTemplate(t *testing.T) {
	route := &Route{}
	route.buildPathTemplate("/user/{id:[0-9]}/{action:*}")
	assert.Equal(t, 2, int(route.maxParams))
	assert.Equal(t, "/user/:id/:action", route.path)
}
