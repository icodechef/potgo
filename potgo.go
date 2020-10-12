package potgo

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"
)

// methods 列出路由器支持的所有 HTTP 方法
var methods = [...]string{
	http.MethodGet,
	http.MethodPost,
	http.MethodPut,
	http.MethodDelete,
	http.MethodPatch,
	http.MethodHead,
	http.MethodOptions,
}

// HandlerFunc 用于处理请求的函数
type HandlerFunc func(*Context) error

// ErrorHandlerFunc 用于错误处理的函数
type ErrorHandlerFunc func(*Context, string, int)

// Map 定义类型为 `map [string] interface {}` 的通用 map
type Map map[string]interface{}

// Application 负责管理应用程序
type Application struct {
	Router          // 内嵌类型
	pool            sync.Pool
	routes          []*Route
	trees           map[string]*node
	maxParams       uint8
	context         *Context
	notFoundHandler HandlerFunc
	errorHandler    ErrorHandlerFunc
	view            ViewEngine
}

// New 创建一个新的 Application
func New() *Application {
	app := &Application{
		routes: make([]*Route, 0),
	}
	app.pool.New = func() interface{} {
		return &Context{
			app:     app,
			pValues: make([]string, app.maxParams),
		}
	}
	app.app = app
	app.prefix = "/"

	app.NotFound(NotFoundHandler())
	app.Error(ErrorHandler())

	return app
}

// RegisterView 注册视图
func (app *Application) RegisterView(view ViewEngine) error {
	app.view = view
	return app.view.Load()
}

// Run http.ListenAndServe(addr, app) 的快捷方式
func (app *Application) Run(addr string) error {
	listeningOn(addr)
	return http.ListenAndServe(addr, app)
}

// RunTLS http.ListenAndServeTLS(addr, certFile, keyFile, app) 的快捷方式
func (app *Application) RunTLS(addr, certFile, keyFile string) error {
	listeningOn(addr)
	return http.ListenAndServeTLS(addr, certFile, keyFile, app)
}

// RunWithGracefulShutdown 带优雅停止的启动
func (app *Application) RunWithGracefulShutdown(addr string, timeout time.Duration) {
	listeningOn(addr)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: app,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	GracefulShutdown(srv, timeout)
}

// ServeHTTP 处理 HTTP 请求
func (app *Application) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c := app.pool.Get().(*Context)
	c.reset(w, req)

	if root := app.trees[req.Method]; root != nil {
		value := root.getRoute(req.URL.Path, c.pValues)

		if value.handlers != nil {
			c.pKeys = value.pKeys
			c.handlers = value.handlers
		}
	}

	if c.handlers == nil {
		c.handlers = append(c.handlers, app.notFoundHandler)
	}

	if err := c.Next(); err != nil {
		app.handleError(c, err)
	}

	app.pool.Put(c)
}

// addRoute 添加路由
func (app *Application) addRoute(method, path string, handlers []HandlerFunc) *Route {
	if app.trees == nil {
		app.trees = make(map[string]*node)
	}
	root, ok := app.trees[method]
	if !ok {
		root = new(node)
		app.trees[method] = root
	}

	r := NewRoute(method, path, handlers)
	root.addRoute(path, r)
	app.routes = append(app.routes, r)

	if r.maxParams > app.maxParams {
		app.maxParams = r.maxParams
	}

	return r
}

// NotFound 添加 NotFound 处理程序
func (app *Application) NotFound(handler HandlerFunc) {
	app.notFoundHandler = handler
}

// NotFoundHandler 默认 NotFound 处理程序
func NotFoundHandler() HandlerFunc {
	return func(c *Context) error {
		http.Error(c.Response.Writer, "404 page not found", http.StatusNotFound)
		return nil
	}
}

// handleError 处理错误
func (app *Application) handleError(c *Context, err error) {
	if httpError, ok := err.(HTTPError); ok {
		app.errorHandler(c, httpError.Error(), httpError.Status())
	} else {
		app.errorHandler(c, err.Error(), http.StatusInternalServerError)
	}
}

// NotFound 添加错误处理程序
func (app *Application) Error(handler ErrorHandlerFunc) {
	app.errorHandler = handler
}

// ErrorHandler 默认错误处理程序
func ErrorHandler() ErrorHandlerFunc {
	return func(c *Context, error string, code int) {
		http.Error(c.Response.Writer, error, code)
	}
}

// URL 使用命名的路由和参数值创建 URL
func (app *Application) URL(name string, pairs ...interface{}) string {
	for _, r := range app.routes {
		if r.name == name {
			s := r.path
			l := len(pairs)
			for i := 0; i < l; i += 2 {
				name := fmt.Sprintf(":%v", pairs[i])
				value := ""
				if i < l-1 {
					value = url.QueryEscape(fmt.Sprint(pairs[i+1]))
				}
				s = strings.Replace(s, name, value, -1)
			}
			return s
		}
	}
	return ""
}

// listeningOn 打印启动信息
func listeningOn(addr string) {
	if len(addr) > 0 && addr[0] == ':' {
		addr = "http://localhost" + addr
	}

	interruptKey := "CTRL"
	if runtime.GOOS == "darwin" {
		interruptKey = "CMD"
	}

	fmt.Fprintf(os.Stdout, "Listening on: %s. Press %s+C to shut down.\n", addr, interruptKey)
}

// GracefulShutdown 优雅停止
func GracefulShutdown(srv *http.Server, timeout time.Duration) {
	// 等待中断信息
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Printf("shutting down server with a timeout of %s\n", timeout)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server was shut down unexpectedly: %v", err)
	} else {
		log.Println("server was shut down gracefully")
	}
}
