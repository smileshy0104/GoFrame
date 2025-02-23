package frame

import (
	"html/template"
	"log"
	"net/http"
)

// Context 是请求处理的上下文，包含了请求和响应的引用。
// 它提供了一种在请求处理过程中传递请求特定数据、中断请求处理等方式。
type Context struct {
	W http.ResponseWriter // W 用于向客户端发送响应。
	R *http.Request       // R 包含了当前请求的所有信息。
}

// HTML函数用于向客户端发送HTML格式的响应。
// 参数status指定HTTP响应的状态码，
// 参数html是待发送的HTML内容字符串。
func (c *Context) HTML(status int, html string) {
	c.W.WriteHeader(status)
	c.W.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, err := c.W.Write([]byte(html))
	if err != nil {
		log.Println(err)
	}
}

// HTMLTemplate函数用于通过指定的模板文件生成HTML响应。
// 参数name是模板的名称，funcMap是模板函数映射，
// data是传递给模板的数据，fileName是可变参数，包含一个或多个模板文件的路径。
func (c *Context) HTMLTemplate(name string, funcMap template.FuncMap, data any, fileName ...string) {
	// 创建一个模板，并设置模板函数
	t := template.New(name)
	// 设置模板函数
	t.Funcs(funcMap)
	// 通过文件名解析模板文件
	t, err := t.ParseFiles(fileName...)
	if err != nil {
		log.Println(err)
		return
	}
	// 设置响应头，并执行模板渲染
	c.W.Header().Set("Content-Type", "text/html; charset=utf-8")
	// 执行模板渲染
	err = t.Execute(c.W, data)
	if err != nil {
		log.Println(err)
	}
}

// HTMLTemplateGlob函数用于通过匹配模式的模板文件生成HTML响应。
// 参数name是模板的名称，funcMap是模板函数映射，
// pattern是指定模板文件的模式字符串，data是传递给模板的数据。
func (c *Context) HTMLTemplateGlob(name string, funcMap template.FuncMap, pattern string, data any) {
	// 创建一个模板，并设置模板函数
	t := template.New(name)
	// 设置模板函数
	t.Funcs(funcMap)
	// 通过匹配模式解析模板文件
	t, err := t.ParseGlob(pattern)
	if err != nil {
		log.Println(err)
		return
	}
	// 设置响应头，并执行模板渲染
	c.W.Header().Set("Content-Type", "text/html; charset=utf-8")
	// 执行模板渲染
	err = t.Execute(c.W, data)
	if err != nil {
		log.Println(err)
	}
}
