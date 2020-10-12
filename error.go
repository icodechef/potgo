package potgo

import "net/http"

// HTTPError HTTP 错误接口
type HTTPError interface {
	error
	Status() int
}

// Error
type httpError struct {
	Code    int    `json:"status" xml:"status"`
	Message string `json:"message" xml:"message"`
}

// NewHTTPError 创建 HttpError 实例
func NewHTTPError(status int, message ...string) HTTPError {
	e := &httpError{status, http.StatusText(status)}
	if len(message) > 0 {
		e.Message = message[0]
	}
	return e
}

// Error 返回错误信息
func (e *httpError) Error() string {
	return e.Message
}

// Status 返回 HTTP 状态码
func (e *httpError) Status() int {
	return e.Code
}
