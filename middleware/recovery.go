package middleware

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"runtime"

	"github.com/icodechef/potgo"
)

// Recovery 返回 Recovery 中间件
func Recovery() potgo.HandlerFunc {
	return RecoveryWithWriter(os.Stderr)
}

// RecoveryWithWriter 返回指定 io.Writer 的 Recovery 中间件
func RecoveryWithWriter(out io.Writer) potgo.HandlerFunc {
	var logger *log.Logger
	if out != nil {
		logger = log.New(out, "", log.LstdFlags)
	}

	return func(c *potgo.Context) error {
		defer func() {
			if err := recover(); err != nil {
				logger.Printf("[Recovery] panic recovered:\n%s\n%s\n", err, getCallStack(3))

				c.Status(http.StatusInternalServerError)
				c.Text(fmt.Sprintf("%v", err))
				c.Abort()
			}
		}()

		return c.Next()
	}
}

func getCallStack(skip int) string {
	buf := new(bytes.Buffer)
	for i := skip; ; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		fmt.Fprintf(buf, "%s:%d (%s)\n", file, line, runtime.FuncForPC(pc).Name())
	}
	return buf.String()
}
