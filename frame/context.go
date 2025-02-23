package frame

import (
	"encoding/json"
	"encoding/xml"
	"html/template"
	"log"
	"net/http"
	"net/url"
)

// Context 是请求处理的上下文，包含了请求和响应的引用。
// 它提供了一种在请求处理过程中传递请求特定数据、中断请求处理等方式。
type Context struct {
	W      http.ResponseWriter // W 用于向客户端发送响应。
	R      *http.Request       // R 包含了当前请求的所有信息。
	engine *Engine             // engine 是一个指向Engine的指针，用于访问Engine中的HTMLRender。
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

// Template函数用于通过指定的模板文件生成HTML响应。
func (c *Context) Template(name string, data any) error {
	c.W.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := c.engine.HTMLRender.Template.ExecuteTemplate(c.W, name, data)
	if err != nil {
		return err
	}
	return nil
}

// JSON函数用于向客户端发送JSON格式的响应。
func (c *Context) JSON(status int, data any) error {
	c.W.Header().Set("Content-Type", "application/json; charset=utf-8")
	c.W.WriteHeader(status)
	rsp, err := json.Marshal(data)
	if err != nil {
		return err
	}
	_, err = c.W.Write(rsp)
	if err != nil {
		return err
	}
	return nil
}

// XML函数用于向客户端发送XML格式的响应。
func (c *Context) XML(status int, data any) error {
	header := c.W.Header()
	header["Content-Type"] = []string{"application/xml; charset=utf-8"}
	c.W.WriteHeader(status)
	err := xml.NewEncoder(c.W).Encode(data)
	if err != nil {
		return err
	}
	return nil
}

// File函数用于将指定文件发送给客户端。
func (c *Context) File(fileName string) {
	// ServeFile函数用于将指定文件发送给客户端。
	http.ServeFile(c.W, c.R, fileName)
}

// FileAttachment函数用于将指定文件作为附件发送给客户端。(指定文件名字)
func (c *Context) FileAttachment(filepath, filename string) {
	if isASCII(filename) {
		// 设置Content-Disposition头，指定附件的名称。
		c.W.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	} else {
		c.W.Header().Set("Content-Disposition", `attachment; filename*=UTF-8''`+url.QueryEscape(filename))
	}
	http.ServeFile(c.W, c.R, filepath)
}

// filepath 是相对文件系统的路径（从对应文件系统目录获取文件）
func (c *Context) FileFromFS(filepath string, fs http.FileSystem) {
	defer func(old string) {
		c.R.URL.Path = old
	}(c.R.URL.Path)

	c.R.URL.Path = filepath

	http.FileServer(fs).ServeHTTP(c.W, c.R)
}
