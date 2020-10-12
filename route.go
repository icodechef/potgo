package potgo

import (
	"strings"
)

// Route 包含注册路由的有关信息
type Route struct {
	method    string
	name      string
	path      string
	maxParams uint8
	handlers  []HandlerFunc
}

// NewRoute 创建路由
func NewRoute(method, path string, handlers []HandlerFunc) *Route {
	r := &Route{
		method:   method,
		handlers: handlers,
	}
	r.buildPathTemplate(path)
	return r
}

// Name 命名路由
func (r *Route) Name(name string) {
	r.name = name
}

func (r *Route) buildPathTemplate(path string) {
	r.maxParams = 0

	parts := strings.Split(path, "/")
	for i, s := range parts {
		if s == "" {
			continue
		}
		if s[0] != '{' || s[len(s)-1] != '}' {
			continue
		}

		r.maxParams++
		m := strings.IndexByte(s, ':')
		if m < 0 {
			parts[i] = ":" + s[1:len(s)-1]
		} else {
			parts[i] = ":" + s[1:m]
		}
	}
	r.path = strings.Join(parts, "/")
}
