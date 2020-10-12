package potgo

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

// NoLayout 不使用视图布局常量
const NoLayout = "no.layout"

// ViewEngine 视图引擎接口
type ViewEngine interface {
	Load() error
	Render(io.Writer, string, string, interface{}, *Context) error
}

// HTMLEngine HTML 引擎
type HTMLEngine struct {
	templates *template.Template
	path      string
	extension string
	left      string
	right     string
	layout    string
	funcMap   template.FuncMap
}

var _ ViewEngine = &HTMLEngine{}

// HTML 创建 HTML 对象
func HTML(path string, extension string) *HTMLEngine {
	return &HTMLEngine{
		path:      path,
		extension: extension,
		left:      "{{",
		right:     "}}",
		layout:    "",
		funcMap:   make(template.FuncMap),
	}
}

// Layout 设置视图布局文件
func (v *HTMLEngine) Layout(layout string) {
	v.layout = layout
}

// Delims 设置视图动作的左右限定符
func (v *HTMLEngine) Delims(left, right string) {
	v.left, v.right = left, right
}

// AddFunc 添加视图函数
func (v *HTMLEngine) AddFunc(name string, callable interface{}) {
	v.funcMap[name] = callable
}

// Func 设置视图函数
func (v *HTMLEngine) Func(funcMap template.FuncMap) {
	for name, value := range funcMap {
		v.funcMap[name] = value
	}
}

// Load 加载视图文件下的所有视图文件
func (v *HTMLEngine) Load() (err error) {
	v.templates = template.New("").Delims(v.left, v.right).Funcs(v.funcMap).Funcs(template.FuncMap{
		"content": func() (string, error) {
			return "", nil
		},
		"route": func() (string, error) {
			return "", nil
		},
	})

	// 遍历视图目录
	err = filepath.Walk(v.path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info == nil || info.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(v.path, path)
		if err != nil {
			return err
		}

		ext := filepath.Ext(path)

		if ext == v.extension {
			b, err := ioutil.ReadFile(path)

			if err != nil {
				return err
			}

			s := string(b)
			name := filepath.ToSlash(rel)

			tmpl := v.templates.New(name)
			_, err = tmpl.Parse(s)

			if err != nil {
				return err
			}
		}

		return nil
	})

	return
}

// Render 渲染视图
func (v *HTMLEngine) Render(w io.Writer, name string, layout string, data interface{}, c *Context) error {
	t := v.templates.Lookup(name)
	if t == nil {
		return fmt.Errorf("template: %s not found", name)
	}
	commonFunc := template.FuncMap{
		"route": func(name string, pairs ...interface{}) (string, error) {
			return c.URL(name, pairs...), nil
		},
	}
	t.Funcs(commonFunc)

	if layout = v.getLayout(layout); layout != "" { // 视图布局
		lt := v.templates.Lookup(layout)
		if lt == nil {
			return fmt.Errorf("layout: %s not found", layout)
		}
		lt.Funcs(commonFunc)
		lt.Funcs(template.FuncMap{
			"content": func() (template.HTML, error) {
				buf := new(bytes.Buffer)
				err := t.Execute(buf, data) // 当前视图
				return template.HTML(buf.String()), err
			},
		})
		return lt.Execute(w, data)
	}
	return t.Execute(w, data)
}

func (v *HTMLEngine) getLayout(layout string) string {
	if layout == NoLayout {
		return ""
	}
	if layout == "" && v.layout != "" {
		return v.layout
	}
	return layout
}
