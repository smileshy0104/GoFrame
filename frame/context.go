package frame

import (
	"encoding/json"
	"errors"
	"frame/binding"
	newlogger "frame/log"
	"frame/render"
	"html/template"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
)

// defaultMultipartMemory是multipart/form-data请求的最大内存限制，单位为字节。
const defaultMultipartMemory = 32 << 20 //32M

// Context 是请求处理的上下文，包含了请求和响应的引用。
// 它提供了一种在请求处理过程中传递请求特定数据、中断请求处理等方式。
type Context struct {
	W                     http.ResponseWriter // W 用于向客户端发送响应。
	R                     *http.Request       // R 包含了当前请求的所有信息。
	engine                *Engine             // engine 是一个指向Engine的指针，用于访问Engine中的HTMLRender。
	StatusCode            int                 // StatusCode 用于记录响应的状态码。
	queryCache            url.Values          // queryCache用于缓存查询参数。
	formCache             url.Values          // formCache用于缓存表单数据。
	DisallowUnknownFields bool                // DisallowUnknownFields用于设置是否允许未知字段。
	IsValidate            bool                // 是否进行验证
	sameSite              http.SameSite       // SameSite用于设置Cookie的SameSite属性。
	Logger                *newlogger.Logger   // logger用于记录日志。
	Keys                  map[string]any      // Keys是一个用于存储键值对的映射，用于在请求处理过程中传递请求特定数据。
	mu                    sync.RWMutex        // 同步读写锁
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

// initFormCache 初始化表单缓存。该方法确保每个请求只解析一次表单数据。
func (c *Context) initFormCache() {
	// 如果表单缓存为nil，则创建一个新的url.Values对象作为表单缓存。
	if c.formCache == nil {
		c.formCache = make(url.Values)
		req := c.R
		// 解析表单数据，如果出现错误且不是ErrNotMultipart错误，则记录错误信息。
		if err := req.ParseMultipartForm(defaultMultipartMemory); err != nil {
			if !errors.Is(err, http.ErrNotMultipart) {
				log.Println(err)
			}
		}
		// 将解析后的表单数据赋值给表单缓存。
		c.formCache = c.R.PostForm
	}
}

// GetPostForm 获取POST表单中指定键的第一个值。
func (c *Context) GetPostForm(key string) (string, bool) {
	// 获取POST表单中指定键的第一个值，并返回一个布尔值表示键是否存在。
	if values, ok := c.GetPostFormArray(key); ok {
		return values[0], ok
	}
	return "", false
}

// PostFormArray 以切片形式获取POST表单中指定键的值。
func (c *Context) PostFormArray(key string) (values []string) {
	// 以切片形式获取POST表单中指定键的值。
	values, _ = c.GetPostFormArray(key)
	return
}

// GetPostFormArray 获取POST表单中指定键的值切片。
func (c *Context) GetPostFormArray(key string) (values []string, ok bool) {
	c.initFormCache()
	values, ok = c.formCache[key]
	return
}

// GetPostFormMap 获取POST表单中指定键的值映射。
func (c *Context) GetPostFormMap(key string) (map[string]string, bool) {
	c.initFormCache()
	return c.get(c.formCache, key)
}

// PostFormMap 以映射形式获取POST表单中指定键的值。
func (c *Context) PostFormMap(key string) (dicts map[string]string) {
	dicts, _ = c.GetPostFormMap(key)
	return
}

// FormFile 通过表单字段名称获取上传的文件头信息。
// 参数 name: 表单字段名称。
// 返回值 *multipart.FileHeader: 文件头信息，包含文件的名称、大小和类型等。
func (c *Context) FormFile(name string) *multipart.FileHeader {
	// 获取通过表单字段上传的文件头信息。
	file, header, err := c.R.FormFile(name)
	if err != nil {
		log.Println(err)
	}
	// 关闭文件流。
	defer file.Close()
	return header
}

// FormFiles 获取通过表单字段上传的多个文件的头信息。
// 参数 name: 表单字段名称。
// 返回值 []*multipart.FileHeader: 文件头信息的切片，包含所有上传文件的名称、大小和类型等。
func (c *Context) FormFiles(name string) []*multipart.FileHeader {
	// 获取通过表单字段上传的多个文件的头信息。
	multipartForm, err := c.MultipartForm()
	if err != nil {
		return make([]*multipart.FileHeader, 0)
	}
	// 返回所有上传文件的头信息。
	return multipartForm.File[name]
}

// SaveUploadedFile 将上传的文件保存到指定路径。
// 参数 file: 文件头信息，包含待保存文件的名称、大小和类型等。
// 参数 dst: 文件保存的目标路径。
// 返回值 error: 保存过程中遇到的错误，如果没有错误则返回nil。
func (c *Context) SaveUploadedFile(file *multipart.FileHeader, dst string) error {
	// 将上传的文件保存到指定路径。
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()
	// 创建目标路径的文件。
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	// 将源文件内容复制到目标文件中。
	_, err = io.Copy(out, src)
	return err
}

// MultipartForm 解析请求中的multipart/form-data，以便处理文件上传。
// 返回值 *multipart.Form: 解析后的multipart表单。
// 返回值 error: 解析过程中遇到的错误，如果没有错误则返回nil。
func (c *Context) MultipartForm() (*multipart.Form, error) {
	// 解析请求中的multipart/form-data，以便处理文件上传。
	err := c.R.ParseMultipartForm(defaultMultipartMemory)
	return c.R.MultipartForm, err
}

// BindJson 将请求体中的JSON数据绑定到指定的对象。它通过设置JSON绑定器以不允许未知字段并启用验证来解析JSON。
func (c *Context) BindJson(obj any) error {
	// 使用JSON绑定器解析请求体中的JSON数据，并绑定到指定的对象。
	json := binding.JSON
	// 设置不允许未知字段和启用验证。
	json.DisallowUnknownFields = true
	// 设置是否启用验证。
	json.IsValidate = true
	return c.MustBindWith(obj, json)
}

// DealJson 解析请求体中的JSON数据并将其存储在传入的数据结构中。（未封装的Json处理器）
func (c *Context) DealJson(data any) error {
	// 获取请求体
	body := c.R.Body
	// 检查请求和请求体是否有效
	if c.R == nil || body == nil {
		return errors.New("invalid request")
	}
	// 创建一个JSON解码器
	decoder := json.NewDecoder(body)
	// 使用解码器将JSON数据解析到传入的数据结构中
	return decoder.Decode(data)
}

// BindXML 将请求体中的XML数据绑定到指定的对象。
func (c *Context) BindXML(obj any) error {
	return c.MustBindWith(obj, binding.XML)
}

// MustBindWith 使用指定的绑定器将请求数据绑定到对象。如果绑定失败，它会返回400错误。
func (c *Context) MustBindWith(obj any, bind binding.Binding) error {
	// 使用指定的绑定器将请求数据绑定到对象。
	if err := c.ShouldBind(obj, bind); err != nil {
		// 如果绑定失败，则返回400错误。
		c.W.WriteHeader(http.StatusBadRequest)
		return err
	}
	return nil
}

// ShouldBind 使用指定的绑定器尝试将请求数据绑定到对象，并返回任何绑定错误。
func (c *Context) ShouldBind(obj any, bind binding.Binding) error {
	// 使用指定的绑定器尝试将请求数据绑定到对象，并返回任何绑定错误。
	return bind.Bind(c.R, obj)
}

// Fail 发送一个失败的响应，包含指定的状态码和消息。
func (c *Context) Fail(code int, msg string) {
	c.String(code, msg)
}

// HandleWithError 处理响应，如果存在错误，则使用错误处理器处理；否则，发送带有状态码和对象的响应。
func (c *Context) HandleWithError(statusCode int, obj any, err error) {
	// 如果存在错误，则使用错误处理器处理，否则，发送带有状态码和对象的响应。
	if err != nil {
		// 使用错误处理器处理错误，并返回处理后的响应。
		code, data := c.engine.errorHandler(err)
		c.JSON(code, data)
		return
	}
	c.JSON(statusCode, obj)
}

// SetCookie 在响应中设置一个cookie。
func (c *Context) SetCookie(name, value string, maxAge int, path, domain string, secure, httpOnly bool) {
	if path == "" {
		path = "/"
	}
	http.SetCookie(c.W, &http.Cookie{
		Name:     name,
		Value:    url.QueryEscape(value),
		MaxAge:   maxAge,
		Path:     path,
		Domain:   domain,
		SameSite: c.sameSite,
		Secure:   secure,
		HttpOnly: httpOnly,
	})
}

// GetHeader 从请求中获取指定的头信息。
func (c *Context) GetHeader(key string) string {
	return c.R.Header.Get(key)
}

// TODO 认证支持————Basic认证（进行base64进行编码，存放到header中）
// Set 方法用于在Context对象中设置键值对。
// 它接受一个键和一个值作为参数，将它们添加到Context的Keys字典中。
// 如果Keys字典尚未初始化，则会先进行初始化。
// 使用互斥锁确保并发安全性。
func (c *Context) Set(key string, value string) {
	c.mu.Lock()
	if c.Keys == nil {
		c.Keys = make(map[string]any)
	}

	c.Keys[key] = value
	c.mu.Unlock()
}

// Get 方法用于从Context对象中获取与指定键关联的值。
// 它接受一个键作为参数，并返回对应的值以及一个布尔值表示该键是否存在。
// 如果键不存在，则返回值为0且exists为false。
// 使用读锁确保并发安全性。
func (c *Context) Get(key string) (value any, exists bool) {
	c.mu.RLock()
	value, exists = c.Keys[key]
	if !exists {
		value = 0
	}
	c.mu.RUnlock()
	return
}

// SetBasicAuth 设置请求的Basic认证信息。
func (c *Context) SetBasicAuth(username, password string) {
	// 设置请求的Basic认证信息。（将对应的用户名和密码存到请求头中）
	c.R.Header.Set("Authorization", "Basic "+BasicAuth(username, password))
}
