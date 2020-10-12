package potgo

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

const defaultMemory = 32 << 20 // 32 MB

// Context 上下文对象
type Context struct {
	app           *Application
	pKeys         []string
	pValues       []string
	handlers      []HandlerFunc
	Request       *http.Request
	Response      Response
	index         int
	mu            sync.RWMutex
	data          map[string]interface{}
	queryCache    url.Values
	postFormCache url.Values
	formCache     url.Values
	viewData      map[string]interface{}
	viewLayout    string
}

func (c *Context) reset(w http.ResponseWriter, r *http.Request) {
	c.Response.reset(w)
	c.Request = r
	c.handlers = nil
	c.pKeys = c.pKeys[0:0]
	c.index = -1
	c.queryCache = nil
	c.postFormCache = nil
	c.formCache = nil
	c.viewData = nil
	c.viewLayout = ""
}

// Next 调用与当前路由关联的其它 HandlerFunc
func (c *Context) Next() error {
	c.index++
	if c.index < len(c.handlers) {
		if err := c.handlers[c.index](c); err != nil {
			return err
		}
	}
	return nil
}

// Abort 跳过与当前路由关联的其它 HandlerFunc
func (c *Context) Abort() {
	c.index = len(c.handlers)
}

// IsAborted 是否已经中止
func (c *Context) IsAborted() bool {
	return c.index == len(c.handlers)
}

// Set 在上下文中保存数据
func (c *Context) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.data == nil {
		c.data = make(map[string]interface{})
	}
	c.data[key] = value
}

// Get 从上下文中检索数据
func (c *Context) Get(key string) (value interface{}, exists bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	value, exists = c.data[key]
	return
}

//  +-----------------------------------------------------------+
//  | Request and Post Data                                     |
//  +-----------------------------------------------------------+

// Param 获取路径中的参数
func (c *Context) Param(key string) string {
	for i, n := range c.pKeys {
		if n == key {
			return c.pValues[i]
		}
	}
	return ""
}

func (c *Context) getQuery() url.Values {
	if c.queryCache == nil {
		c.queryCache = c.Request.URL.Query()
	}
	return c.queryCache
}

// Query 获取查询字符串参数的第一个值
func (c *Context) Query(key string) string {
	return c.QueryDefault(key, "")
}

// QueryDefault 获取查询字符串参数的第一个值，如果参数不存在可以返回指定的默认值
func (c *Context) QueryDefault(key string, defaultValue string) string {
	if vs, ok := c.getQuery()[key]; ok && len(vs) > 0 {
		return vs[0]
	}
	return defaultValue
}

// QuerySlice 获取查询字符串参数切片
func (c *Context) QuerySlice(key string) []string {
	if vs := c.getQuery()[key]; len(vs) > 0 {
		return vs
	}
	return []string{}
}

// QueryMap 获取查询字符串参数映射
func (c *Context) QueryMap(key string) map[string]string {
	m := make(map[string]string)
	for k, v := range c.getQuery() {
		if i := strings.IndexByte(k, '['); i >= 1 && k[0:i] == key {
			if j := strings.IndexByte(k[i+1:], ']'); j >= 1 {
				m[k[i+1:i+j]] = v[0]
			}
		}
	}
	return m
}

// PostValue 获取 POST 参数的第一个值
func (c *Context) PostValue(key string) string {
	return c.PostValueDefault(key, "")
}

// PostValueDefault 获取 POST 参数的第一个值，如果参数不存在可以返回指定的默认值
func (c *Context) PostValueDefault(key string, defaultValue string) string {
	if vs := c.getPostForm()[key]; len(vs) > 0 {
		return vs[0]
	}
	return defaultValue
}

// PostValueSlice 获取 POST 参数切片
func (c *Context) PostValueSlice(key string) []string {
	if vs := c.getPostForm()[key]; len(vs) > 0 {
		return vs
	}
	return []string{}
}

// PostValueMap 获取 POST 参数映射
func (c *Context) PostValueMap(key string) map[string]string {
	m := make(map[string]string)
	for k, v := range c.getPostForm() {
		if i := strings.IndexByte(k, '['); i >= 1 && k[0:i] == key {
			if j := strings.IndexByte(k[i+1:], ']'); j >= 1 {
				m[k[i+1:i+j]] = v[0]
			}
		}
	}
	return m
}

func (c *Context) getPostForm() url.Values {
	if c.postFormCache == nil {
		_ = c.Request.ParseMultipartForm(defaultMemory) // 32M
		c.postFormCache = c.Request.PostForm
	}
	return c.postFormCache
}

// FormValue 获取表单参数的第一个值
func (c *Context) FormValue(key string) string {
	return c.FormValueDefault(key, "")
}

// FormValueDefault 获取表单参数的第一个值，如果参数不存在可以返回指定的默认值
func (c *Context) FormValueDefault(key string, defaultValue string) string {
	if vs := c.getForm()[key]; len(vs) > 0 {
		return vs[0]
	}
	return defaultValue
}

// FormValueSlice 获取表单参数切片
func (c *Context) FormValueSlice(key string) []string {
	if vs := c.getForm()[key]; len(vs) > 0 {
		return vs
	}
	return []string{}
}

// FormValueMap 获取表单参数映射
func (c *Context) FormValueMap(key string) map[string]string {
	m := make(map[string]string)
	for k, v := range c.getForm() {
		if i := strings.IndexByte(k, '['); i >= 1 && k[0:i] == key {
			if j := strings.IndexByte(k[i+1:], ']'); j >= 1 {
				m[k[i+1:i+j]] = v[0]
			}
		}
	}
	return m
}

func (c *Context) getForm() url.Values {
	if c.formCache == nil {
		_ = c.Request.ParseMultipartForm(defaultMemory) // 32M
		c.formCache = c.Request.Form
	}
	return c.formCache
}

// ReadForm 从 HTTP 请求中提取数据填充给定的结构体的各个字段
func (c *Context) ReadForm(ptr interface{}) error {
	return readFormData(c.Request, ptr)
}

// readFormData 此方法参考自 《The Go Programming Language》 12.7. Accessing Str uct Field Tags 的示例
func readFormData(req *http.Request, ptr interface{}) error {
	if err := req.ParseForm(); err != nil {
		return err
	}

	rv := reflect.ValueOf(ptr)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.New("ptr must be a pointer")
	}

	v := rv.Elem() // the struct variable
	if v.Kind() != reflect.Struct {
		return errors.New("ptr must be a pointer to a struct")
	}
	// Build map of fields keyed by effective name.
	fields := make(map[string]reflect.Value)
	for i := 0; i < v.NumField(); i++ {
		fieldInfo := v.Type().Field(i) // a reflect.StructField
		tag := fieldInfo.Tag           // a reflect.StructTag
		name := tag.Get("http")        // 标签
		if name == "" {
			name = strings.ToLower(fieldInfo.Name)
		}
		fields[name] = v.Field(i)
	}

	// 填充给定的结构体的各个字段
	for name, values := range req.Form {
		var fieldKey string
		if i := strings.IndexByte(name, '['); i >= 1 { // 如果是 field[key] 这种形式的
			start := i + 1
			if j := strings.IndexByte(name[start:], ']'); j >= 1 {
				fieldKey, name = name[start:start+j], name[0:i]
			}
		}
		f := fields[name]
		if !f.IsValid() {
			continue // ignore unrecognized HTTP parameters
		}

		for _, value := range values {
			switch f.Kind() {
			case reflect.Slice:
				elem := reflect.New(f.Type().Elem()).Elem()
				if err := populateFieldValue(elem, value); err != nil {
					return fmt.Errorf("%s: %v", name, err)
				}
				f.Set(reflect.Append(f, elem))
			case reflect.Map:
				if f.IsNil() {
					f.Set(reflect.MakeMap(f.Type()))
				}
				key := reflect.New(f.Type().Key()).Elem()
				key.SetString(fieldKey)
				elem := reflect.New(f.Type().Elem()).Elem()
				if err := populateFieldValue(elem, value); err != nil {
					return fmt.Errorf("%s: %v", name, err)
				}
				f.SetMapIndex(key, elem)
			default:
				if err := populateFieldValue(f, value); err != nil {
					return fmt.Errorf("%s: %v", name, err)
				}
			}
		}
	}
	return nil
}

func populateFieldValue(v reflect.Value, value string) error {
	switch v.Kind() {
	case reflect.String:
		v.SetString(value)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if value == "" {
			value = "0"
		}
		i, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		v.SetInt(i)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if value == "" {
			value = "0"
		}
		i, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return err
		}
		v.SetUint(i)

	case reflect.Float32, reflect.Float64:
		if value == "" {
			value = "0"
		}
		i, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		v.SetFloat(i)

	case reflect.Bool:
		if value == "" {
			value = "false"
		}
		b, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		v.SetBool(b)

	default:
		return fmt.Errorf("unsupported kind %s", v.Type())
	}
	return nil
}

// MultipartForm 取得 multipart/form-data 编码的表单数据
func (c *Context) MultipartForm() (*multipart.Form, error) {
	err := c.Request.ParseMultipartForm(defaultMemory)
	return c.Request.MultipartForm, err
}

// UploadedFile 单个上传的文件
type UploadedFile struct {
	File *multipart.FileHeader
	Name string
}

// FormFile 返回给定键的上传文件的 UploadedFile
func (c *Context) FormFile(name string) (*UploadedFile, error) {
	f, fh, err := c.Request.FormFile(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return &UploadedFile{
		File: fh,
		Name: name,
	}, nil
}

// Store 把上传文件移动到指定的目录
func (f *UploadedFile) Store(path string) (int64, error) {
	return storeFormFile(f.File, filepath.Join(path, f.File.Filename))
}

// StoreAs 将表单文件上传到指定的目录，并指定文件名
func (f *UploadedFile) StoreAs(path, filename string) (int64, error) {
	return storeFormFile(f.File, filepath.Join(path, filename))
}

// UploadedFiles 多个上传的文件
type UploadedFiles struct {
	Form  *multipart.Form
	Files []*UploadedFile
	Name  string
}

// FormFiles 返回给定键的上传文件的 UploadedFiles
func (c *Context) FormFiles(name string) (*UploadedFiles, error) {
	form, err := c.MultipartForm()
	if err != nil {
		return nil, err
	}
	f := &UploadedFiles{
		Form: form,
		Name: name,
	}
	for _, file := range form.File[name] {
		f.Files = append(f.Files, &UploadedFile{
			File: file,
			Name: name,
		})
	}
	return f, nil
}

// Store 将多个上传文件上传到指定的目录
func (f *UploadedFiles) Store(path string) (written int64, err error) {
	for _, file := range f.Files {
		n, err := file.Store(path)
		if err != nil {
			return written, err
		}

		written += n
	}
	return written, nil
}

func storeFormFile(file *multipart.FileHeader, dst string) (int64, error) {
	// 源文件
	src, err := file.Open()
	if err != nil {
		return 0, err
	}
	defer src.Close()
	// 目标文件
	out, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer out.Close()
	return io.Copy(out, src)
}

// ClientIP 返回用户 IP
func (c *Context) ClientIP() string {
	ip := c.Request.Header.Get("X-Forwarded-For")
	if ip = strings.TrimSpace(strings.Split(ip, ",")[0]); ip == "" {
		if ip = strings.TrimSpace(c.Request.Header.Get("X-Real-Ip")); ip == "" {
			ip = strings.TrimSpace(c.Request.RemoteAddr)
			// 存在冒号才使用 SplitHostPort，不然会出错
			if colon := strings.LastIndex(ip, ":"); colon != -1 {
				ip, _, _ = net.SplitHostPort(ip)
			}
		}
	}
	return ip
}

//  +-----------------------------------------------------------+
//  | Response                                                  |
//  +-----------------------------------------------------------+

// Status 设置 HTTP 状态码(HTTP Status Code)
func (c *Context) Status(code int) {
	c.Response.WriteHeader(code)
}

// Header 设置 HTTP Header
func (c *Context) Header(name string, value string) {
	if value == "" {
		c.Response.Writer.Header().Del(name)
		return
	}
	c.Response.Writer.Header().Add(name, value)
}

// WithHeaders 添加 HTTP 头
func (c *Context) WithHeaders(headers map[string]string) {
	for key, value := range headers {
		c.Header(key, value)
	}
}

// ContentType 设置 Content-Type
func (c *Context) ContentType(value string) {
	header := c.Response.Writer.Header()
	if header.Get("Content-Type") == "" {
		header.Set("Content-Type", value)
	}
}

// SetCookie 设置 cookie
func (c *Context) SetCookie(cookie *http.Cookie) {
	if cookie.Path == "" {
		cookie.Path = "/"
	}
	http.SetCookie(c.Response.Writer, cookie)
}

// GetCookie 获取 cookie
func (c *Context) GetCookie(name string) (string, error) {
	cookie, err := c.Request.Cookie(name)
	if err != nil {
		return "", err
	}
	val, _ := url.QueryUnescape(cookie.Value)
	return val, nil
}

// File 读取文件内容并输出
func (c *Context) File(filepath string) error {
	http.ServeFile(c.Response.Writer, c.Request, filepath)
	return nil
}

// Text 将给定的字符串写入响应主体
func (c *Context) Text(format string, data ...interface{}) (err error) {
	c.ContentType("text/plain; charset=utf-8")
	_, err = c.Write([]byte(fmt.Sprintf(format, data...)))
	return
}

// JSON 将指定结构作为 JSON 写入响应主体
func (c *Context) JSON(obj interface{}) (err error) {
	c.ContentType("application/json; charset=utf-8")
	jsonBytes, err := json.Marshal(obj)
	if err != nil {
		return
	}

	_, err = c.Write(jsonBytes)
	return
}

// Write 向客户端写入数据，c.Response.Writer.Write(b) 的快捷方式
func (c *Context) Write(b []byte) (int, error) {
	return c.Response.Write(b)
}

// WriteWithStatus 发送HTTP状态，并向客户端写入数据
func (c *Context) WriteWithStatus(statusCode int, b []byte) (int, error) {
	c.Status(statusCode)
	return c.Write(b)
}

// StreamAttachment 流下载
func (c *Context) StreamAttachment(r io.Reader, filename string) (err error) {
	c.Header("content-disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	_, err = io.Copy(c.Response.Writer, r)
	return
}

// Attachment 文件下载
func (c *Context) Attachment(filepath, filename string) error {
	c.Header("content-disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	return c.File(filepath)
}

// Redirect 重定向
func (c *Context) Redirect(path string, code ...int) error {
	var statusCode int
	if len(code) > 0 {
		statusCode = code[0]
		if statusCode < http.StatusMultipleChoices || statusCode > http.StatusPermanentRedirect {
			return errors.New("invalid redirect status code")
		}
	} else {
		statusCode = http.StatusFound
	}

	c.Abort()
	http.Redirect(c.Response.Writer, c.Request, path, statusCode)
	return nil
}

// RouteRedirect 重定向到命名路由
func (c *Context) RouteRedirect(name string, pairs ...interface{}) error {
	var statusCode int
	if l := len(pairs); l > 0 {
		if l%2 == 1 {
			switch pairs[l-1].(type) {
			case int:
				statusCode = pairs[l-1].(int) // 最后一个为状态码
				pairs = pairs[0 : l-2]
			default:
				return errors.New("invalid redirect status code")
			}
		}
	}

	if statusCode == 0 {
		statusCode = http.StatusFound
	}

	return c.Redirect(c.URL(name, pairs...), statusCode)
}

// URL 使用命名的路由和参数值创建 URL
func (c *Context) URL(name string, pairs ...interface{}) string {
	return c.app.URL(name, pairs...)
}

//  +-----------------------------------------------------------+
//  | View                                                      |
//  +-----------------------------------------------------------+

// ViewData 添加视图数据
func (c *Context) ViewData(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.viewData == nil {
		c.viewData = make(map[string]interface{})
	}
	c.viewData[key] = value
}

// Layout 设置当前 HandlerFunc 使用的视图布局文件
func (c *Context) Layout(layout string) {
	c.viewLayout = layout
}

// NoLayout 设置当前 HandlerFunc 不使用视图布局
func (c *Context) NoLayout() {
	c.viewLayout = NoLayout
}

// View 渲染视图
func (c *Context) View(name string, optionalData ...map[string]interface{}) error {
	if c.app.view == nil {
		return errors.New("view engine is missing, pls use `RegisterView`")
	}
	c.ContentType("text/html; charset=utf-8")
	// 合并视图数据
	if len(optionalData) > 0 {
		data := optionalData[0]
		for k, v := range data {
			c.ViewData(k, v)
		}
	}
	return c.app.view.Render(c.Response.Writer, name, c.viewLayout, c.viewData, c)
}
