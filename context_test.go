package potgo

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestContext_Reset(t *testing.T) {
	c := &Context{}
	c.index = 1
	assert.Nil(t, c.Request)
	assert.Nil(t, c.Response.Writer)

	req, _ := http.NewRequest("GET", "/test", nil)
	c.reset(httptest.NewRecorder(), req)
	assert.NotNil(t, c.Request)
	assert.NotNil(t, c.Response.Writer)
	assert.Equal(t, -1, c.index)
	assert.Len(t, c.pKeys, 0)
	assert.Len(t, c.handlers, 0)
}

func TestContext_Param(t *testing.T) {
	c := &Context{}
	c.pKeys = []string{"id", "name"}
	c.pValues = []string{"12", "foo"}

	assert.Equal(t, "foo", c.Param("name"))
	assert.Equal(t, "", c.Param("undefined"))
}

func TestContext_SetGet(t *testing.T) {
	c := &Context{}
	c.Set("foo", "bar")

	value, ok := c.Get("foo")
	assert.Equal(t, "bar", value)
	assert.True(t, ok)

	value, ok = c.Get("abc")
	assert.Nil(t, value)
	assert.False(t, ok)
}

func getNextHandler(tag string) HandlerFunc {
	return func(c *Context) error {
		_, _ = fmt.Fprintf(c.Response.Writer, "<%v>", tag)
		err := c.Next()
		_, _ = fmt.Fprintf(c.Response.Writer, "</%v>", tag)
		return err
	}
}

func getAbortHandler(tag string) HandlerFunc {
	return func(c *Context) error {
		_, _ = fmt.Fprintf(c.Response.Writer, "<%v/>", tag)
		c.Abort()
		return nil
	}
}

func TestContext_Next(t *testing.T) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)

	c := &Context{}
	c.reset(res, req)
	c.handlers = []HandlerFunc{
		getNextHandler("h1"),
		getNextHandler("h2"),
		getNextHandler("h3"),
	}

	assert.Nil(t, c.Next())
	assert.Equal(t, "<h1><h2><h3></h3></h2></h1>", res.Body.String())
}

func TestContext_Abort(t *testing.T) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)

	c := &Context{}
	c.reset(res, req)
	c.handlers = []HandlerFunc{
		getNextHandler("h1"),
		getAbortHandler("h2"),
		getNextHandler("h3"),
	}

	assert.Nil(t, c.Next())
	assert.Equal(t, "<h1><h2/></h1>", res.Body.String())
}

func TestContext_Query(t *testing.T) {
	req, _ := http.NewRequest("GET", "/?foo=bar&page=2&id[]=10&id[]=20&name=foo&name=bar&a[s]=1&a[f]=2", nil)

	c := &Context{}
	c.reset(httptest.NewRecorder(), req)

	assert.Equal(t, "bar", c.Query("foo"))
	assert.Equal(t, "2", c.Query("page"))
	assert.Equal(t, "", c.Query("undefined"))
	assert.Equal(t, "10", c.QueryDefault("undefined", "10"))
	assert.Equal(t, 2, len(c.QuerySlice("name")))

	ids := c.QuerySlice("id[]")
	assert.Equal(t, 2, len(ids))
	assert.Equal(t, "20", ids[1])

	a := c.QueryMap("a")
	if v, ok := a["f"]; ok {
		assert.Equal(t, "2", v)
	}
}

func TestContext_Post(t *testing.T) {
	req, _ := http.NewRequest("POST", "/?page=2&id=&ids[]=3&ids[]=4",
		strings.NewReader("username=foo&email=bar@example.com&a[write]=1&a[game]=0&orders[]=2&orders[]=8"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	c := &Context{}
	c.reset(httptest.NewRecorder(), req)

	// 不包含 URL 参数
	assert.Equal(t, "", c.PostValue("page"))

	assert.Equal(t, "foo", c.PostValue("username"))
	assert.Equal(t, "", c.PostValue("undefined"))
	assert.Equal(t, "10", c.PostValueDefault("undefined", "10"))
	assert.Equal(t, 2, len(c.PostValueSlice("orders[]")))

	a := c.PostValueMap("a")
	if v, ok := a["game"]; ok {
		assert.Equal(t, "0", v)
	}
}

func TestContext_Form(t *testing.T) {
	req, _ := http.NewRequest("POST", "/?page=2&id[]=3&id[]=4&a[read]=1",
		strings.NewReader("username=foo&email=bar@example.com&a[write]=1&a[game]=0&id[]=1&id[]=2"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	c := &Context{}
	c.reset(httptest.NewRecorder(), req)

	assert.Equal(t, "2", c.FormValue("page"))
	assert.Equal(t, "foo", c.FormValue("username"))
	assert.Equal(t, "", c.FormValue("undefined"))
	assert.Equal(t, "10", c.FormValueDefault("undefined", "10"))
	assert.Equal(t, 4, len(c.FormValueSlice("id[]")))

	a := c.FormValueMap("a")
	if v, ok := a["read"]; ok {
		assert.Equal(t, "1", v)
	}
}

type testFormData struct {
	Labels   []string        `http:"l"`
	Username string          `http:"username"`
	Email    string          `http:"email"`
	Orders   []int           `http:"orders[]"`
	Like     map[string]bool `http:"like"`
}

func TestContextReadForm(t *testing.T) {
	var data testFormData
	c := &Context{}

	http.HandleFunc("/test", func(resp http.ResponseWriter, req *http.Request) {
		if err := c.ReadForm(&data); err != nil {
			return
		}
	})
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/test?l=golang&l=programming",
		strings.NewReader("username=foo&email=bar@example.com&like[write]=1&like[game]=0&orders[]=2&orders[]=8"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	c.reset(res, req)

	http.DefaultServeMux.ServeHTTP(res, req)

	assert.Equal(t, "bar@example.com", data.Email)
	assert.Equal(t, 2, len(data.Labels))

	val, ok := data.Like["write"]
	if ok {
		assert.Equal(t, true, val)
	}
}

func TestContextClientIP(t *testing.T) {
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Forwarded-For", "192.168.100.1")
	req.Header.Set("X-Real-IP", "192.168.100.2")
	req.RemoteAddr = "192.168.100.3"

	c := &Context{}
	c.reset(httptest.NewRecorder(), req)

	assert.Equal(t, "192.168.100.1", c.ClientIP())
	req.Header.Del("X-Forwarded-For")
	assert.Equal(t, "192.168.100.2", c.ClientIP())
	req.Header.Del("X-Real-IP")
	assert.Equal(t, "192.168.100.3", c.ClientIP())
	req.RemoteAddr = "192.168.100.3:8080"
	assert.Equal(t, "192.168.100.3", c.ClientIP())
}

func TestContextSetCookie(t *testing.T) {
	c := &Context{}
	c.reset(httptest.NewRecorder(), nil)
	c.SetCookie(&http.Cookie{
		Name:   "foo",
		Value:  "bar",
		Path:   "",
		Domain: "localhost",
		MaxAge: 1,
		// Expires: time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Secure:   true,
	})
	assert.Equal(t, "foo=bar; Path=/; Domain=localhost; Max-Age=1; HttpOnly; Secure", c.Response.Writer.Header().Get("Set-Cookie"))
}

func TestContextRedirect(t *testing.T) {
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	c := &Context{}
	c.reset(res, req)

	err := c.Redirect("/login", http.StatusTemporaryRedirect)
	assert.Equal(t, http.StatusTemporaryRedirect, res.Code)
	assert.Nil(t, err)

	res = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/test", nil)
	c = &Context{}
	c.reset(res, req)

	err = c.Redirect("/logout", http.StatusAccepted)
	assert.NotNil(t, err)
}

func TestContext_RouteRedirect(t *testing.T) {
	r := New()
	h := func(c *Context) error { return nil }
	r.GET("/user/{id}", h).Name("user")
	r.GET("/users", h).Name("users")
	c := &Context{app: r}

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	c.reset(res, req)
	err := c.RouteRedirect("user", "id", 10)
	assert.Nil(t, err)
	assert.Equal(t, "/user/10", res.Header().Get("Location"))
	assert.Equal(t, http.StatusFound, res.Code)

	res = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/test", nil)
	c.reset(res, req)
	err = c.RouteRedirect("user", "id", 10, http.StatusTemporaryRedirect)
	assert.Equal(t, http.StatusTemporaryRedirect, res.Code)

	res = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/test", nil)
	c.reset(res, req)
	err = c.RouteRedirect("users")
	assert.Equal(t, "/users", res.Header().Get("Location"))
	assert.Equal(t, http.StatusFound, res.Code)

	res = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/test", nil)
	c.reset(res, req)
	err = c.RouteRedirect("user", "id", 10, "302")
	assert.Equal(t, "invalid redirect status code", err.Error())
}
