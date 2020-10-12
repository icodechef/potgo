package potgo

import (
	"net/http"
	"path"
	"strings"
)

// Router 路由器
type Router struct {
	app      *Application
	prefix   string
	handlers []HandlerFunc
}

// Use 添加中间件到当前的 Router
func (r *Router) Use(middleware ...HandlerFunc) {
	r.handlers = append(r.handlers, middleware...)
}

// Group 创建指定路径前缀的路由分组
func (r *Router) Group(prefix string) *Router {
	return &Router{
		prefix:   path.Join(r.prefix, prefix),
		app:      r.app,
		handlers: r.handlers[:len(r.handlers):len(r.handlers)],
	}
}

func (r *Router) add(method, relativePath string, handlers []HandlerFunc) *Route {
	// 合并中间件
	m := make([]HandlerFunc, len(r.handlers)+len(handlers))
	copy(m, r.handlers)
	copy(m[len(r.handlers):], handlers)

	return r.app.addRoute(method, path.Join(r.prefix, relativePath), m)
}

// GET 注册一个 HTTP GET 方法的路由
func (r *Router) GET(relativePath string, handlers ...HandlerFunc) *Route {
	return r.add(http.MethodGet, relativePath, handlers)
}

// POST 注册一个 HTTP POST 方法的路由
func (r *Router) POST(relativePath string, handlers ...HandlerFunc) *Route {
	return r.add(http.MethodPost, relativePath, handlers)
}

// PUT 注册一个 HTTP PUT 方法的路由
func (r *Router) PUT(relativePath string, handlers ...HandlerFunc) *Route {
	return r.add(http.MethodPut, relativePath, handlers)
}

// DELETE 注册一个 HTTP DELETE 方法的路由
func (r *Router) DELETE(relativePath string, handlers ...HandlerFunc) *Route {
	return r.add(http.MethodDelete, relativePath, handlers)
}

// PATCH 注册一个 HTTP PATCH 方法的路由
func (r *Router) PATCH(relativePath string, handlers ...HandlerFunc) *Route {
	return r.add(http.MethodPatch, relativePath, handlers)
}

// HEAD 注册一个 HTTP HEAD 方法的路由
func (r *Router) HEAD(relativePath string, handlers ...HandlerFunc) *Route {
	return r.add(http.MethodHead, relativePath, handlers)
}

// OPTIONS 注册一个 HTTP OPTIONS 方法的路由
func (r *Router) OPTIONS(relativePath string, handlers ...HandlerFunc) *Route {
	return r.add(http.MethodOptions, relativePath, handlers)
}

// Any 注册实现响应所有 HTTP 请求的路由
func (r *Router) Any(relativePath string, handlers ...HandlerFunc) []*Route {
	routes := make([]*Route, len(methods))
	for i, method := range methods {
		routes[i] = r.add(method, relativePath, handlers)
	}
	return routes
}

// Match 注册指定响应 HTTP 请求的路由
func (r *Router) Match(methods string, relativePath string, handlers ...HandlerFunc) []*Route {
	mm := strings.Split(methods, ",")
	routes := make([]*Route, len(mm))
	for i, method := range mm {
		routes[i] = r.add(method, relativePath, handlers)
	}
	return routes
}

// Static 静态文件
//
// router.Static("/static", "/public")
// router.Static("/static", "./public/static")
func (r *Router) Static(relativePath, root string) {
	if root == "" {
		root = "."
	}
	prefix := path.Join(r.prefix, relativePath)
	fs := http.Dir(root)
	fileServer := http.StripPrefix(prefix, http.FileServer(fs))
	urlPattern := path.Join(prefix, "/{filepath:*}")

	r.GET(urlPattern, func(c *Context) error {
		file := c.Param("filepath")

		f, err := fs.Open(file)
		if err != nil {
			return r.app.notFoundHandler(c)
		}
		_ = f.Close()

		fileServer.ServeHTTP(c.Response.Writer, c.Request)
		return nil
	})
}

// File 输出指定的文件
func (r *Router) File(relativePath, filepath string) {
	r.GET(relativePath, func(c *Context) error {
		c.File(filepath)
		return nil
	})
}

// Layout 设置路由分组的视图布局
func (r *Router) Layout(layoutFile string) {
	r.Use(func(c *Context) error {
		c.Layout(layoutFile)
		return c.Next()
	})
}

// NoLayout 设置路由分组不使用视图布局
func (r *Router) NoLayout() {
	r.Use(func(c *Context) error {
		c.NoLayout()
		return c.Next()
	})
}
