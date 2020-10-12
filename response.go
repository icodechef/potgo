package potgo

import (
	"bufio"
	"net"
	"net/http"
)

const (
	defaultStatus = http.StatusOK
	noWritten     = -1
)

// Response 包装一个 http.ResponseWriter 并实现其要使用的接口
type Response struct {
	Writer http.ResponseWriter
	status int
	size   int
}

// reset 重置 Response
func (res *Response) reset(writer http.ResponseWriter) {
	res.Writer = writer
	res.size = noWritten
	res.status = defaultStatus
}

// Reset 重置 Response
func (res *Response) Reset(w http.ResponseWriter) bool {
	if res.Written() {
		return false
	}
	h := w.Header()
	for k := range h {
		h[k] = nil
	}

	res.reset(w)
	return true
}

// Size 返回写入数据的大小
func (res *Response) Size() int {
	return res.size
}

// Status 返回状态码
func (res *Response) Status() int {
	return res.status
}

// WriteHeader 发送 HTTP status code
func (res *Response) WriteHeader(code int) {
	if code > 0 && res.status != code {
		res.status = code
	}
}

// Written 是否已经向客户端写入数据
func (res *Response) Written() bool {
	return res.size > noWritten
}

// Write 向客户端写入回复的数据
func (res *Response) Write(b []byte) (int, error) {
	res.tryWriteHeader()
	n, err := res.Writer.Write(b)
	res.size += n
	return n, err
}

func (res *Response) tryWriteHeader() {
	if !res.Written() {
		res.size = 0
		if res.status == 0 {
			res.status = defaultStatus
		}
		res.Writer.WriteHeader(res.status)
	}
}

// Flush 刷新输出缓冲
func (res *Response) Flush() {
	if flusher, ok := res.Writer.(http.Flusher); ok {
		flusher.Flush()
	}
}

// Hijack 实现 http.Hijacker 接口
func (res *Response) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return res.Writer.(http.Hijacker).Hijack()
}
