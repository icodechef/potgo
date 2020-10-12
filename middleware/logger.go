package middleware

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/icodechef/potgo"
)

// Logger 返回日志中间件
func Logger() potgo.HandlerFunc {
	return LoggerWithWriter(os.Stdout)
}

// LoggerWithWriter 返回指定 io.Writer 的日志中间件
func LoggerWithWriter(out io.Writer) potgo.HandlerFunc {
	return func(c *potgo.Context) error {
		start := time.Now()
		err := c.Next()
		clientIP := c.ClientIP()
		elapsed := time.Now().Sub(start)

		if elapsed > time.Minute {
			elapsed = elapsed - elapsed%time.Second
		}

		fmt.Fprintf(out, "%v | %15s | %13v | %s %s %d %#v \n",
			time.Now().Format("2006/01/02 - 15:04:05"),
			clientIP,
			elapsed,
			c.Request.Method,
			c.Request.Proto,
			c.Response.Status(),
			c.Request.URL.String())

		return err
	}
}

// LoggerWithFile 返回写入文件的日志中间件
func LoggerWithFile(file string) potgo.HandlerFunc {
	f, err := os.Create(file)

	if err != nil {
		log.Fatalf("LoggerWithFile: open log file [%s] err: %v", file, err)
	}

	return LoggerWithWriter(f)
}
