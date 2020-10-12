package potgo

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNew(t *testing.T) {
	r := New()

	r.GET("/users", func(c *Context) error {
		fmt.Fprint(c.Response.Writer, "GET:ok")
		return nil
	})
	r.POST("/users", func(c *Context) error {
		fmt.Fprint(c.Response.Writer, "POST:ok")
		return nil
	})

	req, _ := http.NewRequest("GET", "/users", nil)
	res := httptest.NewRecorder()
	r.ServeHTTP(res, req)
	assert.Equal(t, "GET:ok", res.Body.String())

	req, _ = http.NewRequest("POST", "/users", nil)
	res = httptest.NewRecorder()
	r.ServeHTTP(res, req)
	assert.Equal(t, "POST:ok", res.Body.String())
}

func TestApplication_URL(t *testing.T) {
	r := New()

	h := func(c *Context) error { return nil }

	r.GET("/user/{id}", h).Name("user.info")

	g := r.Group("/group")
	g.GET("/user/{id}/posts/{pid}", h).Name("group.user.post")

	assert.Equal(t, "/user/10", r.URL("user.info", "id", 10))
	assert.Equal(t, "/group/user/10/posts/1", r.URL("group.user.post", "id", 10, "pid", 1))
	assert.Equal(t, "/group/user/:id/posts/1", r.URL("group.user.post", "pid", 1))
	assert.Equal(t, "", r.URL("undefined_name"))
}

func TestApplication_RegisterView(t *testing.T) {
	r := New()

	view := HTML("./testdata/views_1", ".html")
	_ = r.RegisterView(view)

	assert.NotNil(t, r.view)
}

func TestViewAndLayout(t *testing.T) {
	r := New()
	view := HTML("./testdata/views_1", ".html")
	view.Layout("layout.html")
	_ = r.RegisterView(view)

	assert.Equal(t, "layout.html", view.getLayout(""))

	r.GET("/hello", func(c *Context) error {
		c.ViewData("name", "world")
		return c.View("hello.html")
	})

	r.GET("/greeting", func(c *Context) error {
		c.ViewData("name", "world")
		c.NoLayout()
		return c.View("hello.html")
	})

	r.GET("/hello/{name}", func(c *Context) error {
		c.ViewData("name", c.Param("name"))
		return c.View("hello.html")
	})

	users := r.Group("/users")
	{
		users.GET("/{id}", func(c *Context) error {
			c.ViewData("id", c.Param("id"))
			return c.View("users.html")
		})
	}

	user := r.Group("/user")
	user.Layout("layout_user.html")
	{
		user.GET("/login", func(c *Context) error {
			c.NoLayout()
			return c.View("login.html")
		})

		user.GET("/{id}/info", func(c *Context) error {
			c.ViewData("id", c.Param("id"))
			return c.View("users.html")
		})
	}

	other := r.Group("/other")
	other.NoLayout()
	{
		other.GET("/foo", func(c *Context) error {
			c.ViewData("name", "foo")
			return c.View("hello.html")
		})

		other.GET("/bar", func(c *Context) error {
			c.ViewData("name", "bar")
			c.Layout("layout_other.html")
			return c.View("hello.html")
		})
	}

	// 视图布局
	req, _ := http.NewRequest("GET", "/hello", nil)
	res := httptest.NewRecorder()
	r.ServeHTTP(res, req)
	assert.Equal(t, `<main id="Content"><h1>Hello, world</h1></main>`, res.Body.String())

	// 没有视图布局
	req, _ = http.NewRequest("GET", "/greeting", nil)
	res = httptest.NewRecorder()
	r.ServeHTTP(res, req)
	assert.Equal(t, `<h1>Hello, world</h1>`, res.Body.String())

	// 又有视图布局
	req, _ = http.NewRequest("GET", "/hello/world", nil)
	res = httptest.NewRecorder()
	r.ServeHTTP(res, req)
	assert.Equal(t, `<main id="Content"><h1>Hello, world</h1></main>`, res.Body.String())

	// 分组
	req, _ = http.NewRequest("GET", "/users/2", nil)
	res = httptest.NewRecorder()
	r.ServeHTTP(res, req)
	assert.Equal(t, `<main id="Content"><h1>user: 2</h1></main>`, res.Body.String())

	req, _ = http.NewRequest("GET", "/user/login", nil)
	res = httptest.NewRecorder()
	r.ServeHTTP(res, req)
	assert.Equal(t, `<h1>login</h1>`, res.Body.String())

	req, _ = http.NewRequest("GET", "/user/6/info", nil)
	res = httptest.NewRecorder()
	r.ServeHTTP(res, req)
	assert.Equal(t, `<main id="user"><h1>user: 6</h1></main>`, res.Body.String())

	req, _ = http.NewRequest("GET", "/other/foo", nil)
	res = httptest.NewRecorder()
	r.ServeHTTP(res, req)
	assert.Equal(t, `<h1>Hello, foo</h1>`, res.Body.String())

	req, _ = http.NewRequest("GET", "/other/bar", nil)
	res = httptest.NewRecorder()
	r.ServeHTTP(res, req)
	assert.Equal(t, `<main id="other"><h1>Hello, bar</h1></main>`, res.Body.String())
}
