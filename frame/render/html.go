package render

import (
	"frame/internal/bytesconv"
	"html/template"
	"net/http"
)

// HTML 结构体用于存储 HTML 响应的相关信息。
// 它包括响应数据、模板名称、是否使用模板等。
type HTML struct {
	Data       any                // 存储传递给模板的数据。
	Name       string             // 存储模板的名称。
	Template   *template.Template // 存储编译后的模板。
	IsTemplate bool               // 标识是否使用模板渲染响应。
}

// HTMLRender 结构体用于定义 HTML 渲染器。
// 它包含一个指向编译后模板的指针。
type HTMLRender struct {
	Template *template.Template
}

// Render 方法用于渲染 HTML 响应。
// 参数 w: 用于写入响应的 http.ResponseWriter 对象。
// 参数 code: HTTP 响应码。
// 返回值 error: 渲染过程中可能发生的错误。
func (h *HTML) Render(w http.ResponseWriter, code int) error {
	// 写入响应的 Content-Type。
	h.WriteContentType(w)
	// 写入 HTTP 响应码。
	w.WriteHeader(code)
	// 根据 IsTemplate 标识来决定是否使用模板渲染。
	if h.IsTemplate {
		// 使用模板渲染，并处理可能发生的错误。
		err := h.Template.ExecuteTemplate(w, h.Name, h.Data)
		return err
	}
	// 不使用模板时，直接写入数据。
	_, err := w.Write(bytesconv.StringToBytes(h.Data.(string)))
	return err
}

// WriteContentType 方法用于写入响应的 Content-Type。
// 参数 w: 用于写入响应的 http.ResponseWriter 对象。
func (h *HTML) WriteContentType(w http.ResponseWriter) {
	// 调用 writeContentType 函数写入 "text/html; charset=utf-8"。
	writeContentType(w, "text/html; charset=utf-8")
}
