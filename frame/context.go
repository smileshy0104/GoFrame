package frame

import (
	"frame/render"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"strings"
)

// Context 是请求处理的上下文，包含了请求和响应的引用。
// 它提供了一种在请求处理过程中传递请求特定数据、中断请求处理等方式。
type Context struct {
	W          http.ResponseWriter // W 用于向客户端发送响应。
	R          *http.Request       // R 包含了当前请求的所有信息。
	engine     *Engine             // engine 是一个指向Engine的指针，用于访问Engine中的HTMLRender。
	StatusCode int                 // StatusCode 用于记录响应的状态码。
	queryCache url.Values          // queryCache用于缓存查询参数。
	formCache  url.Values          // formCache用于缓存表单数据。
}

// Render函数用于向客户端发送响应，并设置响应的状态码。
func (c *Context) Render(statusCode int, r render.Render) error {
	//如果设置了statusCode，对header的修改就不生效了
	err := r.Render(c.W, statusCode)
	c.StatusCode = statusCode
	//多次调用 WriteHeader 就会产生这样的警告 superfluous response.WriteHeader
	return err
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
	// TODO 未进行封装的版本
	//c.W.Header().Set("Content-Type", "text/html; charset=utf-8")
	//err := c.engine.HTMLRender.Template.ExecuteTemplate(c.W, name, data)
	//if err != nil {
	//	return err
	//}
	//return nil

	//状态是200 默认不设置的话 如果调用了 write这个方法 实际上默认返回状态 200
	return c.Render(http.StatusOK, &render.HTML{
		Data:       data,
		IsTemplate: true,
		Template:   c.engine.HTMLRender.Template,
		Name:       name,
	})
}

// JSON函数用于向客户端发送JSON格式的响应。
func (c *Context) JSON(status int, data any) error {
	// TODO 未进行封装的版本
	//c.W.Header().Set("Content-Type", "application/json; charset=utf-8")
	//c.W.WriteHeader(status)
	//rsp, err := json.Marshal(data)
	//if err != nil {
	//	return err
	//}
	//_, err = c.W.Write(rsp)
	//if err != nil {
	//	return err
	//}
	//return nil
	return c.Render(status, &render.JSON{Data: data})
}

// XML函数用于向客户端发送XML格式的响应。
func (c *Context) XML(status int, data any) error {
	// TODO 未进行封装的版本
	//header := c.W.Header()
	//header["Content-Type"] = []string{"application/xml; charset=utf-8"}
	//c.W.WriteHeader(status)
	//err := xml.NewEncoder(c.W).Encode(data)
	//if err != nil {
	//	return err
	//}
	//return nil

	return c.Render(status, &render.XML{Data: data})
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
	// defer 语句用于在函数返回之前恢复之前的URL路径。
	defer func(old string) {
		c.R.URL.Path = old
	}(c.R.URL.Path)

	c.R.URL.Path = filepath

	http.FileServer(fs).ServeHTTP(c.W, c.R)
}

// Redirect函数用于重定向客户端到指定URL。（用在如果对应页面失效）
func (c *Context) Redirect(status int, location string) error {
	// TODO 未进行封装的版本
	//// 验证重定向状态码是否在指定的范围内，如果不是，则抛出异常。
	//if (status < http.StatusMultipleChoices || status > http.StatusPermanentRedirect) && status != http.StatusCreated {
	//	panic(fmt.Sprintf("Cannot redirect with status code %d", status))
	//}
	//// 调用http.Redirect函数，将客户端重定向到指定URL。
	//http.Redirect(c.W, c.R, location, status)

	return c.Render(status, &render.Redirect{
		Code:     status,
		Request:  c.R,
		Location: location,
	})
}

// String函数用于向客户端发送字符串格式的响应。
func (c *Context) String(status int, format string, values ...any) (err error) {
	// TODO 未进行封装的版本
	//plainContentType := "text/plain; charset=utf-8"
	//c.W.Header().Set("Content-Type", plainContentType)
	//c.W.WriteHeader(status)
	//if len(values) > 0 {
	//	_, err = fmt.Fprintf(c.W, format, values...)
	//	return
	//}
	//// 使用StringToBytes函数将字符串转换为字节数组，并写入响应体。
	//_, err = c.W.Write(StringToBytes(format))
	//return

	return c.Render(status, &render.String{Format: format, Data: values})
}

// TODO 获取QUERY查询请求 `http://xxx.com/user/add?id=1&age=20&username=张三`
// initQueryCache 初始化查询缓存，确保queryCache字段被正确设置。
func (c *Context) initQueryCache() {
	// 如果queryCache字段为nil，则创建一个新的url.Values对象作为查询缓存。
	if c.queryCache == nil {
		if c.R != nil {
			c.queryCache = c.R.URL.Query()
		} else {
			c.queryCache = url.Values{}
		}
	}
}

// GetQueryArray 返回指定键的查询参数值数组及是否存在标志。
// key: 要检索的查询参数的键。
func (c *Context) GetQueryArray(key string) (values []string, ok bool) {
	c.initQueryCache()
	values, ok = c.queryCache[key]
	return
}

// DefaultQuery 返回指定键的查询参数值，如果不存在则返回默认值。
// key: 要检索的查询参数的键。
// defaultValue: 如果键不存在时返回的默认值。
func (c *Context) DefaultQuery(key, defaultValue string) string {
	array, ok := c.GetQueryArray(key)
	if !ok {
		return defaultValue
	}
	return array[0]
}

// GetQuery 返回指定键的单个查询参数值。
// key: 要检索的查询参数的键。
func (c *Context) GetQuery(key string) string {
	c.initQueryCache()
	return c.queryCache.Get(key)
}

// QueryArray 返回指定键的查询参数值数组。
// key: 要检索的查询参数的键。
func (c *Context) QueryArray(key string) (values []string) {
	c.initQueryCache()
	values, _ = c.queryCache[key]
	return
}

// TODO 获取QUERYMAP查询请求 `http://localhost:8080/queryMap?user[id]=1&user[name]=张三`
// QueryMap函数用于获取指定查询参数键对应的值的映射。
// 主要用于处理查询参数中包含结构的情况，例如：user[id]=1&user[name]=张三。
// 参数:
//
//	key - 查询参数的键名。
//
// 返回值:
//
//	一个映射，键为结构字段名，值为字段对应的值。
func (c *Context) QueryMap(key string) (dicts map[string]string) {
	dicts, _ = c.GetQueryMap(key)
	return
}

// GetQueryMap函数用于获取指定查询参数键对应的值的映射，并返回一个布尔值表示键是否存在。
// 参数:
//
//	key - 查询参数的键名。
//
// 返回值:
//
//	一个映射，键为结构字段名，值为字段对应的值。
//	一个布尔值，表示指定的键是否存在于查询参数中。
func (c *Context) GetQueryMap(key string) (map[string]string, bool) {
	c.initQueryCache()
	return c.get(c.queryCache, key)
}

// get函数用于从map中获取指定键对应的值，并返回一个布尔值表示键是否存在。
// 参数:
//
//	m - 一个映射，存储了查询参数的键值对，其中键是查询参数的完整键名（包含结构字段名），值是字段对应的值。
//	key - 查询参数的键名。
//
// 返回值:
//
//	一个映射，键为结构字段名，值为字段对应的值。
//	一个布尔值，表示指定的键是否存在于查询参数中。
func (c *Context) get(m map[string][]string, key string) (map[string]string, bool) {
	dicts := make(map[string]string)
	exist := false
	// 遍历查询参数映射，查找包含指定键的键值对。
	for k, value := range m {
		// 如果键以"["开头，则表示键值对包含结构字段名。
		if i := strings.IndexByte(k, '['); i >= 1 && k[0:i] == key {
			// 如果键以"]"结尾，则表示键值对包含字段名。
			if j := strings.IndexByte(k[i+1:], ']'); j >= 1 {
				// 从键中提取字段名，并将其添加到映射中。
				exist = true
				dicts[k[i+1:][:j]] = value[0]
			}
		}
	}
	return dicts, exist
}
