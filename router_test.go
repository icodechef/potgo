package potgo

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRouter_Use(t *testing.T) {
	r := &Router{}
	r.Use(func(c *Context) error {
		return nil
	})

	assert.Len(t, r.handlers, 1)

	r.Use(func(c *Context) error {
		return nil
	}, func(c *Context) error {
		return nil
	})

	assert.Len(t, r.handlers, 3)
}

func TestRouter_Add(t *testing.T) {
	r := &Router{
		app: New(),
	}
	r.add("GET", "/hello/{name}", []HandlerFunc{func(c *Context) error {
		return nil
	}})

	_, ok := r.app.trees["GET"]
	assert.True(t, ok)
	assert.Equal(t, 1, int(r.app.maxParams))
}

func TestRouter_Group(t *testing.T) {
	r := New()
	r.Use(func(c *Context) error {
		return c.Next()
	})

	api := r.Group("/api")
	api.Use(func(c *Context) error {
		return c.Next()
	})
	{
		api.GET("/login", func(c *Context) error {
			fmt.Fprint(c.Response.Writer, "login")
			return nil
		})

		logout := api.GET("/logout", func(c *Context) error {
			fmt.Fprint(c.Response.Writer, "logout")
			return nil
		})

		assert.Equal(t, "/api/logout", logout.path)
	}

	assert.Len(t, api.handlers, 2)

	req, _ := http.NewRequest("GET", "/api/login", nil)
	res := httptest.NewRecorder()
	r.ServeHTTP(res, req)
	assert.Equal(t, "login", res.Body.String())
}

func TestRouter_Match(t *testing.T) {
	r := New()

	routes := r.Match("GET,POST", "/hello", func(c *Context) error {
		fmt.Fprint(c.Response.Writer, "hello")
		return nil
	})
	assert.Len(t, routes, 2)

	req, _ := http.NewRequest("GET", "/hello", nil)
	res := httptest.NewRecorder()
	r.ServeHTTP(res, req)
	assert.Equal(t, "hello", res.Body.String())

	req, _ = http.NewRequest("POST", "/hello", nil)
	res = httptest.NewRecorder()
	r.ServeHTTP(res, req)
	assert.Equal(t, "hello", res.Body.String())
}

func TestRouter_Static(t *testing.T) {
	r := New()
	r.Static("/static", "./testdata/static")

	req, _ := http.NewRequest("GET", "/static/style.css", nil)
	res := httptest.NewRecorder()
	r.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)

	req, _ = http.NewRequest("GET", "/static/common.js", nil)
	res = httptest.NewRecorder()
	r.ServeHTTP(res, req)
	assert.Equal(t, http.StatusNotFound, res.Code)

	r.File("/favicon.ico", "./testdata/images/favicon.ico")

	req, _ = http.NewRequest("GET", "/favicon.ico", nil)
	res = httptest.NewRecorder()
	r.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
}
