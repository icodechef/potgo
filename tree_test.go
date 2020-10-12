package potgo

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var routeVal string

func makeRoute(val string) *Route {
	return &Route{
		handlers: []HandlerFunc{func(c *Context) error {
			routeVal = val
			return nil
		}},
	}
}

type testPaths []struct {
	path     string
	route    string
	nilRoute bool
	pKeys    []string
}

func makeParamValues() []string {
	return make([]string, 20)
}

func checkRequests(t *testing.T, root *node, paths testPaths) {
	for _, n := range paths {
		value := root.getRoute(n.path, makeParamValues())

		if value.handlers == nil {
			if !n.nilRoute {
				t.Errorf("无法找到匹配的路由：%s", n.path)
			}
			continue
		} else if n.nilRoute {
			t.Errorf("无法找到匹配的路由：%s", n.path)
			continue
		} else {
			_ = value.handlers[0](nil)

			if routeVal != n.route {
				t.Errorf("错误的路由匹配 ：%s (%s ！= %s)", n.path, n.route, routeVal)
				continue
			}
		}

		if value.pKeys != nil {
			if len(value.pKeys) != len(n.pKeys) {
				t.Errorf("路由参数长度不匹配 ：%s", n.path)
				continue
			} else {
				for i, v := range value.pKeys {
					if v != n.pKeys[i] {
						t.Errorf("路由参数不匹配 ：%s", n.path)
						break
					}
				}
			}
		}
	}
}

func TestTreeAddAndGet(t *testing.T) {
	root := &node{}

	routes := [...]string{
		"/user/{id}/add",
		"/user/{id}/edit",
		"/user/{name}/add",
		"/user/{name}/del",
		"/user/10/",
		"/user/10",
		"/users/10/",
		"/users/{id}/",
		"/users/{id}/edit1",
		"/users/{id}/edit2",
		"/users/{name}/del",
		"/posts/list",
		"/page/{page:[0-9]+}",
		"/user/{id}/{name}/add",
		"/users/{id}",
		"/user/{id}/{name}/{title}/edit1",
		"/user/{id}/{name}/{title}/edit2",
		"/user/{id}/{name}/{title}/{files:*}/php",
		"/user/{id}/{name}/{title}/{file:*}",
		"/{id}",
	}

	for _, path := range routes {
		root.addRoute(path, makeRoute(path))
	}

	// fmt.Println(root.print(0))

	checkRequests(t, root, testPaths{
		{path: "/user/10/add", route: "/user/{id}/add", nilRoute: false, pKeys: []string{"id"}},
		{path: "/user/10/edit", route: "/user/{id}/edit", nilRoute: false, pKeys: []string{"id"}},
		{path: "/users/10/", route: "/users/10/", nilRoute: false, pKeys: []string{}},
		{path: "/users/99/", route: "/users/{id}", nilRoute: false, pKeys: []string{"id"}},
		{path: "/page/12", route: "/page/{page:[0-9]+}", nilRoute: false, pKeys: []string{"page"}},
		{path: "/foo/bar", route: "", nilRoute: true, pKeys: []string{}},
		{path: "/user/12/foo/hello/css/style.css", route: "/user/{id}/{name}/{title}/{file:*}", nilRoute: false, pKeys: []string{"id", "name", "title", "file"}},
		{path: "/user/12/foo/hello/css/style.css/php", route: "/user/{id}/{name}/{title}/{file:*}", nilRoute: false, pKeys: []string{"id", "name", "title", "file"}},
	})
}

func TestTreeParamAndWildcard(t *testing.T) {
	root := &node{}

	root.addRoute("/user/{uid:[0-9]+}/edit/{pid}", &Route{})
	root.addRoute("/src/{file:*}", &Route{})

	pValues := makeParamValues()
	value := root.getRoute("/user/12/edit/3", pValues)

	assert.Equal(t, "uid", value.pKeys[0])
	assert.Equal(t, "pid", value.pKeys[1])
	assert.Equal(t, "12", pValues[0])
	assert.Equal(t, "3", pValues[1])

	value = root.getRoute("/src/public/css/style.css", pValues)
	assert.Equal(t, "file", value.pKeys[0])
	assert.Equal(t, "public/css/style.css", pValues[0])

	value = root.getRoute("/user/foo/edit/3", pValues)
	assert.Nil(t, value.handlers)
}
