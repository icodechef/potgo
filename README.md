# Potgo

Potgo 是一个 Go 微框架，可以帮助您快速编写简单但功能强大的 Web 应用程序和 API。

## 安装

下载并安装

```bash
$ go get github.com/icodechef/potgo
```

在代码中导入

```go
import "github.com/icodechef/potgo"
```

## Quick start

创建 `main.go`

```go
package main

import (
	"github.com/icodechef/potgo"
)

func main() {
	app := potgo.New()

	app.GET("/", func(c *potgo.Context) error {
		return c.Text("hello world！")
	})

	app.Run(":8080")
}
```

运行这段代码并在浏览器中访问 `http://localhost:8080`

```bash
$ go run main.go
Listening on: http://localhost:8080. Press CTRL+C to shut down.
```

## 路由

在讲解路由前，我们了解一下 `HandlerFunc` 和 `Context` 上下文

### HandlerFunc

一个 `HandlerFunc`，用于处理请求的函数

```go
type HandlerFunc func(*Context) error
```

### Context 上下文

`HandlerFunc` 传入的参数为从 `sync.Pool` 中获取一个新上下文 `Context` 对象 。

### 基本路由

构建基本路由只需要一个 `路由路径` 与一个 `HandlerFunc`。

```go
func main() {
	app := potgo.New()

	app.GET("/hello", func(c *potgo.Context) error {
		return c.Text("hello world！")
	})
	// ...
}
```

### 可用的路由方法

路由器允许你注册能响应任何 HTTP 请求的路由如下:

```go
func handler(c *potgo.Context) error {
	return c.Text("Hello from method: %s and path: %s", c.Request.Method, c.Request.URL.Path)
}

func main() {
	app := potgo.New()

	// GET 路由
	app.GET("/", handler)

	// POST 路由
	app.POST("/", handler)

	// PUT 路由
	app.PUT("/", handler)

	// DELETE 路由
	app.DELETE("/", handler)

	// PATCH 路由
	app.PATCH("/", handler)

	// OPTIONS 路由
	app.OPTIONS("/", handler)

	// HEAD 路由
	app.HEAD("/", handler)
}
```

有时候可能需要注册一个可响应多个 HTTP 请求的路由，这时你可以使用 `Match` 方法，也可以使用 `Any` 方法注册一个实现响应所有 HTTP 请求的路由：

```go
func main() {
	// 响应多个 HTTP 请求的路由
	app.Match("GET,POST", "/", handlerFunc)

	// 用于所有 HTTP 方法
	app.Any("/", handlerFunc)
}
```

### 路由参数

有时需要在路由中捕获一些 URL 片段。例如，从 URL 中捕获用户的 ID，可以通过定义路由参数来执行此操作：

```go
func main() {
	app := potgo.New()

	app.GET("/user/{id}", func(c *potgo.Context) error {
		return c.Text("User: " + c.Param("id"))
	})

	// ...
}
```

也可以根据需要在路由中定义多个参数：

```go
func main() {
	app := potgo.New()

	app.GET("/user/{uid}/posts/{pid}", func(c *potgo.Context) error {
		return c.Text("User: " + c.Param("uid") + ", Post: " + c.Param("pid"))
	})

	// ...
}
```

路由参数都被放在 `{}` 内，如果没有设置正则约束，参数名称为括号内的字面量，所以 `{param}`，`param` 表示参数名称。

### 正则约束

可以在路由参数中约束参数的格式。`{}` 接受以 `:` 分隔的参数名称和定义参数应如何约束的正则表达式，格式为 `{param:regex}`：

```go
func main() {
	app := potgo.New()

	app.GET("/user/{name:[A-Za-z]+}", func(c *potgo.Context) error {
		return c.Text("User: " + c.Param("name"))
	})

	app.GET("/user/{id:[0-9]+}", func(c *potgo.Context) error {
		return c.Text("User: " + c.Param("id"))
	})

	// ...
}
```

### 匹配剩余字符

当已经匹配一部分 URL 片段，可以使用带 `*` 号路由参数匹配剩余 URL 片段，格式为 `{param:*}`。

```go
func main() {
	app := potgo.New()

	app.GET("/user/{id}/{action:*}", func(c *potgo.Context) error {
		return c.Text("Action: " + c.Param("action"))
	})

	// ...
}
```

### 路由命名

路由命名可以为指定路由生成 URL 或者重定向。通过在路由定义上链式调用 `Name` 方法可指定路由名称：

```go
func main() {
	app := potgo.New()

	app.GET("/user/{id}", func(c *potgo.Context) error {
		return c.Text("User: " + c.Param("id"))
	}).Name("user.profile")

	hello := app.GET("/hello", func(c *potgo.Context) error {
		return c.Text("Hello world!")
	})
	hello.Name("hello")
	
	app.Run(":8080")
}
```

注意：路由命名必须是唯一的

### 路由分组

一组路由可以用前缀路径进行分组，组之间共享相同的中间件和视图布局，组内可以嵌套组。

使用 `Group` 方法进行路由分组：

```go
func main() {
	app := potgo.New()

	v1 := app.Group("/v1")
	{
		v1.GET("/login", func(c *potgo.Context) error {
			return c.Text("v1.login")
		})

		v1.GET("/submit", func(c *potgo.Context) error {
			return c.Text("v1.submit")
		})
	}

	v2 := app.Group("/v2")
	{
		v2.GET("/login", func(c *potgo.Context) error {
			return c.Text("v2.login")
		})

		v2.GET("/submit", func(c *potgo.Context) error {
			return c.Text("v2.submit")
		})
	}

	app.Run(":8080")
}
```

## 中间件

### 定义中间件

中间件的定义与路由的 `HandlerFunc` 一致，处理的输入是 `Context` 对象。

```go
func Hello() potgo.HandlerFunc {
	return func(c *potgo.Context) error {
		err := c.Next()
		return err
	}
}
```

`c.Next()` 表示等待执行其他的中间件或用户的 `HandlerFunc`。

### 使用中间件

使用 `Use` 方法注册中间件

```go
import (
	"github.com/icodechef/potgo"
	"log"
)

// Logger 自定义日志访问中间件
func Logger() potgo.HandlerFunc {
	return func(c *potgo.Context) error {
		log.Println("开始记录")
		err := c.Next()
		log.Println("记录结束")
		return err
	}
}

func main() {
	app := potgo.New()

	app.Use(Logger())

	app.GET("/", func(c *potgo.Context) error {
		log.Println("访问中")
		return c.Text("Hello World")
	})

	app.Run(":8080")
}
```

运行这段代码并在浏览器中访问 `http://localhost:8080`，然后查看控制台可以得到以下输出：

```bash
$ go run main.go
Listening on: http://localhost:8080. Press CTRL+C to shut down.
2020/09/23 00:27:55 开始记录
2020/09/23 00:27:55 访问中
2020/09/23 00:27:55 记录结束
```

### 路由分组中间件

```go
package main

import (
	"github.com/icodechef/potgo"
	"log"
	"net/http"
)

func Logger() potgo.Handler {
	return func(c *potgo.Context) error {
		log.Println("开始记录")
		err := c.Next()
		log.Println("记录结束")
		return err
	}
}

func main() {
	app := potgo.New()

	app.GET("/", func(c *potgo.Context) error {
		log.Println("访问 / 中")
		return c.Text("Hello World")
	})

	api := app.Group("/api")
	api.Use(Logger())
	{
		api.GET("/users", func(c *potgo.Context) error {
			log.Println("访问 /api/users 中")
			return c.Text("Hello, Pot")
		})
	}

	app.Run(":8080")
}
```

运行这段代码并在浏览器中分别访问 `http://localhost:8080`、 `http://localhost:8080/api/users`，然后查看控制台可以得到以下输出：

```bash
$ go run main.go
Listening on: http://localhost:8080. Press CTRL+C to shut down.
2020/09/23 00:34:39 访问 / 中
2020/09/23 00:34:58 开始记录
2020/09/23 00:34:58 访问 /api/users 中
2020/09/23 00:34:58 记录结束
```

### 前置 & 后置 中间件

中间件是在请求之前或之后执行，取决于中间件本身，也就是说 `c.Next()` 的位置。

```go
func Logger() potgo.HandlerFunc {
	return func(c *potgo.Context) error {
		// 处理当前中间件的逻辑
		log.Println("记录中")

		// 处理下一个中间件
		err := c.Next()
		return err
	}
}
```

调整一下顺序：

```go
func Logger() potgo.HandlerFunc {
	return func(c *potgo.Context) error {
		// 先处理下一个中间件
		err := c.Next()
		
		// 然后再处理当前中间件的逻辑
		log.Println("记录中")
		return err
	}
}
```

## 请求

### 接收请求

路由处理程序（HandlerFunc）可以通过 `Context.Request` 获取请求信息，实际上 `Context.Request` 就是 `*http.Request`。

例如，获取请求 URL 信息：

```go
func main() {
	app := potgo.New()

	app.GET("/", func(c *potgo.Context) error {
		host := c.Request.URL.Host
		path := c.Request.URL.Path
		method := c.Request.Method
		// ...
	})

	app.Run(":8080")
}
``` 

### 获取路径中的参数

```go
func main() {
    app := potgo.New()
    
    app.GET("/users/{id}", func(c *potgo.Context) error {
        return c.Text("User: " + c.Param("id"))
    })
    
    app.GET("/users/{uid}/posts/{pid}", func(c *potgo.Context) error {
        uid := c.Param("uid")
        pid := c.Param("pid")
        return c.Text("User: " + uid + " Post: " + pid)
    })
    
    app.Run(":8080")
}
```

### 获取查询字符串参数

通过 `Query` 方法获取查询字符串参数

```go
func main() {
    app := potgo.New()
    
    // 请求的地址为：/hello?name=world
    app.GET("/hello", func(c *potgo.Context) error {
        return c.Text("Hello, " + c.Query("name"))
    })
    
    app.Run(":8080")
}
```

如果查询字符串参数不存在时，可以通过 `QueryDefault` 方法的第二个参数指定默认值

```go
func main() {
	app := potgo.New()

	// 请求的地址为：/hello
	app.GET("/hello", func(c *potgo.Context) error {
		return c.Text("Hello, " + c.QueryDefault("name", "world"))
	})

	app.Run(":8080")
}
```

### 获取 POST 参数

通过 `PostValue` 方法获取 POST 参数，注意，此方法会忽略查询字符串参数

```go
func main() {
    app := potgo.New()
    
    app.POST("/login", func(c *potgo.Context) error {
        username := c.PostValue("username")
        password := c.PostValue("password")
        //
    })
}
```

同样，POST 参数不存在时，可以通过 `PostValueDefault` 方法的第二个参数指定默认值。

### 获取表单参数

通过 `FormValue` 方法获取表单参数

```go
func main() {
    app := potgo.New()
    
    app.POST("/login", func(c *potgo.Context) error {
        username := c.FormValue("username")
        password := c.FormValue("password")
        //
    })
}
```

同样，表单参数不存在时，可以通过 `FormValueDefault` 方法的第二个参数指定默认值。

### 存储上传文件

#### 上传单个文件

创建 `index.html`

```html
<!-- 此文件位置 public/index.html -->
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>上传单个文件</title>
</head>
<body>
<h1>上传单个文件</h1>

<form action="/upload" method="post" enctype="multipart/form-data">
    Files: <input type="file" name="file"><br>
    <input type="submit" value="Submit">
</form>
</body>
</html>
```

创建路由，使用 `FormFile` 方法访问上传的单文件，然后使用 `store` 方法把上传文件移动到指定的目录

```go
func main() {
	app := potgo.New()

	app.Static("/", "./public") // 访问 index.html

	app.POST("/upload", func(c *potgo.Context) error {
		f, _ := c.FormFile("file")
		f.Store("./uploads")
		return c.Text("File: %s uploaded!", f.File.Filename)
	})

	app.Run(":8080")
}
```

如果不想自动生成文件名，那么可以使用 `StoreAs` 方法，它接受路径和文件名作为其参数

```go
func main() {
	app := potgo.New()

	app.Static("/", "./public") // 访问 index.html

	app.POST("/upload", func(c *potgo.Context) error {
		f, _ := c.FormFile("file")
		f.StoreAs("./uploads", "test.csv")
		return c.Text("File: %s uploaded!", f.File.Filename)
	})

	app.Run(":8080")
}
```

#### 上传多个文件

创建 `index.html`

```html
<!-- 此文件位置 public/index.html -->
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>上传多个文件</title>
</head>
<body>
<h1>上传多个文件</h1>

<form action="/upload" method="post" enctype="multipart/form-data">
    <input type="file" name="files" multiple><br>
    <input type="submit" value="Submit">
</form>
</body>
</html>
```

创建路由，使用 `FormFiles` 方法访问上传的多文件，然后使用 `store` 方法把上传文件移动到指定的目录

```go
func main() {
	app := potgo.New()

	app.Static("/", "./public") // 访问 index.html

	app.POST("/upload", func(c *potgo.Context) error {
		f, _ := c.FormFiles("files")
		f.Store("./uploads")
		return c.Text("%d files uploaded!", len(f.Files))
	})

	app.Run(":8080")
}
```

实际上 `FormFiles` 方法返回的对象包含一个 `Files` 切片，此切片的值与 `FormFile` 方法的返回值一样。

```go
func main() {
	app := potgo.New()

	app.Static("/", "./public") // 访问 index.html

	app.POST("/upload", func(c *potgo.Context) error {
		f, _ := c.FormFiles("files")

		for _, file := range f.Files {
			if _, err := file.Store("./uploads"); err != nil {
				c.Status(http.StatusBadRequest)
				return c.Text("upload file err: %s", err.Error())
			}
		}

		return c.Text("%d files uploaded!", len(f.Files))
	})

	app.Run(":8080")
}
```

### 绑定请求参数

```go
type Account struct {
	Username  string    `http:"username"`
	Password  string    `http:"password"`
}

func main() {
	app := potgo.New()

	app.Static("/", "./public")

	app.POST("/login", func(c *potgo.Context) error {
		var account Account
		if err := c.ReadForm(&account); err != nil {
			return err
		}

		log.Println(account.Username)
		log.Println(account.Password)

		return nil
	})

	app.Run(":8080")
}
```

## 响应

路由处理程序（HandlerFunc）可以通过 `Context.Response` 设置响应。 
`Context.Response` 包含 `http.ResponseWriter`，通过 `Context.Response.Writer` 进行访问。

### 设置 HTTP 状态码

```go
func main() {
	app := potgo.New()

	app.GET("/", func(c *potgo.Context) error {
		c.Status(http.StatusOK)
		// 或者
		// c.Response.WriteHeader(http.StatusOK)
		return c.Text("Hello world!")
	})

	app.Run(":8080")
}
```

### 添加响应头

```go
func main() {
	app := potgo.New()

	app.GET("/", func(c *potgo.Context) error {
		c.Header("Content-Type", "text/plain; charset=utf-8")
		c.Header("X-Header", "Header Value")
		return c.Text("Hello world!")
	})

	app.Run(":8080")
}
```

或者，你可以使用 `Context.WithHeaders` 方法来指定要添加到响应的头映射：

```go
func main() {
	app := potgo.New()

	app.GET("/", func(c *potgo.Context) error {
		c.WithHeaders(map[string]string{
			"Content-Type": "text/plain; charset=utf-8",
			"X-Header": "Header Value",
		})
		return c.Text("Hello world!")
	})

	app.Run(":8080")
}
```

### 重定向

如果要重定向到另一个指定的 URL，可以使用 `Context.Redirect` 方法。默认情况， HTTP 状态码是 `302`。

```go
func main() {
	app := potgo.New()

	app.GET("/", func(c *potgo.Context) error {
		return c.Redirect("/login")
	})

	app.GET("/login", func(c *potgo.Context) error {
		return c.Text("Login")
	})

	app.Run(":8080")
}
```

### 重定向到命名路由

一旦为路由指定了名称，就可以使用 `Context.RouteRedirect` 重定向到该路由

```go
func main() {
	app := potgo.New()

	app.GET("/user/{id}", func(c *potgo.Context) error {
		return c.Text("User: " + c.Param("id"))
	}).Name("user") // 命名路由

	app.GET("/", func(c *potgo.Context) error {
		return c.RouteRedirect("user", "id", 10)
	})

	app.Run(":8080")
}
```

### Cookie

设置 cookie

```go
func main()  {
	app := potgo.New()

	app.GET("/", func(c *potgo.Context) error {
		c.SetCookie(&http.Cookie{
			Name: "foo",
			Value: "bar",
			Expires: time.Now().Add(24 * time.Hour),
		})
		return c.Text("hello world!")
	})

	app.Run(":8080")
}
```

获取 cookie

```go
func main()  {
	app := potgo.New()

	app.GET("/", func(c *potgo.Context) error {
		foo, _ := c.GetCookie("foo")
		return c.Text(foo)
	})

	app.Run(":8080")
}
```

### 视图响应

```go
func main() {
	app := potgo.New()

	view := potgo.HTML("./views", ".html")
	app.RegisterView(view)

	app.GET("/greet", func(c *potgo.Context) error {
		return c.View("greeting.html")
	})

	app.Run(":8080")
}
```

### JSON 响应

```go
func main() {
	app := potgo.New()

	app.GET("/api", func(c *potgo.Context) error {
		return c.JSON(potgo.Map{
			"lang": "golang",
			"city": "gz",
		})
	})

	app.Run(":8080")
}
```

### 文本响应

```go
func main() {
	app := potgo.New()

	app.GET("/hello", func(c *potgo.Context) error {
		return c.Text("Hello world!")
	})

	app.Run(":8080")
}
```

### 文件响应

`File` 方法用于直接在用户浏览器显示一个图片之类的文件，而不是下载。

```go
func main() {
	app := potgo.New()

	app.GET("/", func(c *potgo.Context) error {
		return c.File("./greeting.txt")
	})

	app.Run(":8080")
}
```

### 文件下载

```go
func main() {
	app := potgo.New()

	app.GET("/download", func(c *potgo.Context) error {
		return c.Attachment("./data/fruits.csv", "fruits-02.csv")
	})

	app.Run(":8080")
}
```

### 流下载

```go
func main() {
	app := potgo.New()

	app.GET("/streamdownload", func(c *potgo.Context) error {
		f, err := os.Open("./data/fruits.csv")
		if err != nil {
			return err
		}
		return c.StreamAttachment(f, "fruits-01.csv")
	})

	app.Run(":8080")
}
```

## 视图

### 创建视图

Potgo 的视图功能，默认使用 Go 语言的模板引擎库 `html/template`。

使用 `Potgo.HTML(视图目录, 文件后缀)` 创建一个视图实例，然后使用 `RegisterView(ViewEngine)` 方法注册视图并预编译模板。

```go
func main() {
	app := potgo.New()

	view := potgo.HTML("./views", ".html")
	app.RegisterView(view)

	// ...
	
	app.Run(":8080")
}
```

要渲染或执行视图，在路由处理程序（Handler）中使用 `Context.View` 方法。

```go
func main() {
	app := potgo.New()

	view := potgo.HTML("./views", ".html")
	app.RegisterView(view)

	app.GET("/", func(c *potgo.Context) error {
		return c.View("welcome.html", map[string]interface{} {
			"name": "World",
		})
	})

	app.Run(":8080")
}
```

如你所见， 传递给 `Context.View` 方法的第一个参数对应 `./views` 目录中视图文件的名称。第二个参数是可供视图使用的数据映射。

编写视图文件 `welcome.html`

```html
<!-- 此文件位置 views/welcome.html -->
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Hello</title>
</head>
<body>

<h1>Hello, {{.name}}</h1>

</body>
</html>
```

运行这段代码并在浏览器中访问 `http://localhost:8080`

渲染的结果将如下所示：

```html

<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Hello</title>
</head>
<body>

<h1>Hello, World</h1>

</body>
</html>
```

### 向视图传递参数

正如在前面的示例中所看到的，可以将一组数据传递给视图：

```go
c.View("welcome.html", map[string]interface{} {
    "name": "World",
})
```

您还可以使用 `Context.ViewData` 方法将参数添加到视图中:

```go
func main() {
	app := potgo.New()

	view := potgo.HTML("./views", ".html")
	app.RegisterView(view)

	app.GET("/", func(c *potgo.Context) error {
		c.ViewData("name", "World")
		return c.View("welcome.html")
	})

	app.Run(":8080")
}
```

### 自定义渲染分隔符

默认分隔符为 `{{` 和 `}}`

编写 `welcome.html`

```html
<!-- 此文件位置 views/welcome.html -->
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Hello</title>
</head>
<body>

<h1>Hello, <{.name}}></h1>

</body>
</html>
```

编写 `main.go`，使用 `Delims` 方法自定义渲染分隔符

```go
func main() {
	app := potgo.New()

	view := potgo.HTML("./views", ".html")
	view.Delims("<{", "}>")
	app.RegisterView(view)

	app.GET("/", func(c *potgo.Context) error {
		c.ViewData("name", "World")
		return c.View("welcome.html")
	})

	app.Run(":8080")
}
```

### 自定义模板函数

编写 `welcome.html`

```html
<!-- 此文件位置 views/welcome.html -->
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Hello</title>
</head>
<body>

<h1>{{.name | greet}}</h1>

</body>
</html>
```

编写 `main.go`，使用 `AddFunc` 方法添加自定义模板函数

```go
func greet(s string) string {
	return "Hello, " + s + "!"
}

func main() {
	app := potgo.New()

	view := potgo.HTML("./views", ".html")
	view.AddFunc("greet", greet)
	app.RegisterView(view)

	app.GET("/", func(c *potgo.Context) error {
		c.ViewData("name", "World")
		return c.View("welcome.html")
	})

	app.Run(":8080")
}
```

### 视图布局

大多数 web 应用在不同的页面中使用相同的布局方式，因此我们使用布局视图来重复使用。

创建布局视图 `layout.html`

```html
<!-- 此视图文件位置 views/layout.html -->
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
</head>
<body>
{{ content }}
</body>
</html>
```

创建布局视图嵌套的 `content` 视图 `welcome.html`

```html
<!-- 此视图文件位置 views/welcome.html -->
<h1>Hello, {{ .name }}</h1>
```

使用 `Layout` 方法设定视图布局

```go
func main() {
	app := potgo.New()

	view := potgo.HTML("./views", ".html")
	view.Layout("layout.html")
	app.RegisterView(view)

	app.GET("/", func(c *potgo.Context) error {
		c.ViewData("name", "World")
		return c.View("welcome.html")
	})

	app.Run(":8080")
}
```

运行这段代码并在浏览器中访问 `http://localhost:8080`

渲染的结果将如下所示：

```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
</head>
<body>
<h1>Hello, World</h1>
</body>
</html>
```

## 生成 URL

### 生成指定路由的 URL

```go
func main() {
	app := potgo.New()

	app.GET("/users/{id}", func(c *potgo.Context) error {
		c.URL("user", "id", 10) // 返回 /user/10
		return nil
	}).Name("user")

	app.URL("user", "id", 10) // 返回 /users/10
	
	// ...
}
```

`URL` 方法第一个参数为 `路由名称`。如果是有定义参数的路由，可以把参数作为 `URL` 方法的第二个参数开始以`键值对`形式传入，格式为 `参数键, 参数值, 参数键, 参数值...`，指定的参数将会自动插入到 URL 中对应的位置。

### 在视图中生成 URL

`route` 是内置模板函数，用于生成指定路由的 URL

编辑 `main.go`

```go
func main() {
	app := potgo.New()

	view := potgo.HTML("./views", ".html")
	app.RegisterView(view)

	app.GET("/", func(c *potgo.Context) error {
		return c.View("users.html")
	})

	app.GET("/user/{id}", func(c *potgo.Context) error {
		return c.Text(c.Request.URL.String())
	}).Name("user.profile") // 路由命名

	app.Run(":8080")
}
```

创建视图 `users.html`

```html
<!-- 此视图文件位置 views/users.html -->
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Dashboard</title>
</head>
<body>
<a href="{{ route "user.profile" "id" 10 }}">User 10</a>
</body>
</html>
```

运行这段代码并在浏览器中访问 `http://localhost:8080`

```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Dashboard</title>
</head>
<body>
<a href="/user/10">User 10</a>
</body>
</html>
```

## 错误处理

### 页面未找到

使用 `NotFound` 自定义错误处理程序

```go
func main() {
	app := potgo.New()

	view := potgo.HTML("./views", ".html")
	app.RegisterView(view)

	app.NotFound(func(c *potgo.Context) error {
		return c.View("404.html")
	})

	app.Run(":8080")
}
```

### HandlerFunc 错误处理

使用 `Error` 自定义错误处理程序

```go
func main() {
	app := potgo.New()

	view := potgo.HTML("./views", ".html")
	app.RegisterView(view)

	app.Error(func(c *potgo.Context, message string, code int) {
		c.Status(code)
		c.ViewData("message", message)
		c.View("500.html")
	})

	app.GET("/", func(c *potgo.Context) error {
		return errors.New("Oops!")
	})

	app.Run(":8080")
}
```

## 优雅停止

带优雅停止的启动

```go
func main()  {
	app := potgo.New()

	app.GET("/", func(c *potgo.Context) error {
		time.Sleep(5 * time.Second)
		return c.Text("Hello world!")
	})

	app.RunWithGracefulShutdown(":8080", 10 * time.Second)
}
```

自定义优雅停止

```go
func main()  {
	app := potgo.New()

	app.GET("/", func(c *potgo.Context) error {
		time.Sleep(5 * time.Second)
		return c.Text("Hello world!")
	})

	srv := &http.Server{
		Addr:    ":8080",
		Handler: app,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	potgo.GracefulShutdown(srv,  10 * time.Second)
}
```
