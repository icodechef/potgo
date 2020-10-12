package potgo

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHTML(t *testing.T) {
	view := HTML("./testdata/views_1", ".html")
	err := view.Load()
	assert.Nil(t, err)
	tmpl := view.templates.Lookup("hello.html")
	assert.NotNil(t, tmpl)
	assert.Equal(t, "hello.html", tmpl.Name())
}

func TestHTMLEngine_Layout(t *testing.T) {
	view := HTML("./testdata/views_1", ".html")
	view.Layout("layout.html")
	assert.Equal(t, "layout.html", view.layout)
}

func TestHTMLEngine_Render(t *testing.T) {
	view := HTML("./testdata/views_1", ".html")
	_ = view.Load()

	res := httptest.NewRecorder()
	view.Render(res, "hello.html", "", map[string]string{"name": "world"}, &Context{})
	assert.Equal(t, "<h1>Hello, world</h1>", res.Body.String())
}

func TestHTMLEngine_Delims(t *testing.T) {
	r := New()
	view := HTML("./testdata/views_1", ".html")
	view.Delims("{%", "%}")
	r.RegisterView(view)

	r.GET("/test", func(c *Context) error {
		return c.View("delims.html", Map{
			"name": "world",
		})
	})

	ts := httptest.NewServer(r)
	defer ts.Close()

	res, err := http.Get(fmt.Sprintf("%s/test", ts.URL))
	if err != nil {
		fmt.Println(err)
	} else {
		resp, _ := ioutil.ReadAll(res.Body)
		assert.Equal(t, "<h1>Hello, world</h1>", string(resp))
	}
}

func formatAsDate(t time.Time) string {
	year, month, day := t.Date()
	return fmt.Sprintf("%d-%02d-%02d", year, month, day)
}

func TestHTMLEngine_Func(t *testing.T) {
	r := New()
	view := HTML("./testdata/views_2", ".html")
	view.AddFunc("formatAsDate", formatAsDate)
	r.RegisterView(view)

	r.GET("/test", func(c *Context) error {
		return c.View("func.html", map[string]interface{}{
			"now": time.Date(2020, 9, 15, 12, 0, 0, 0, time.Local),
		})
	})

	ts := httptest.NewServer(r)
	defer ts.Close()

	res, err := http.Get(fmt.Sprintf("%s/test", ts.URL))
	if err != nil {
		fmt.Println(err)
	} else {
		resp, _ := ioutil.ReadAll(res.Body)
		assert.Equal(t, "Date: 2020-09-15", string(resp))
	}
}
