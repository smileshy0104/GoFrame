package render

import (
	"encoding/xml"
	"net/http"
)

// XML 结构体用于表示待渲染成XML格式的数据。
// 它包含一个名为Data的字段，可以是任何类型的，代表待编码为XML的数据。
type XML struct {
	Data any
}

// Render 方法负责将XML格式的数据渲染到HTTP响应中。
// 参数w是http.ResponseWriter类型，用于写入HTTP响应。
// 参数code是int类型，代表HTTP响应的状态码。
// 该方法首先设置响应的内容类型，然后写入状态码，最后将XML数据编码到响应中。
// 返回值是error类型，如果编码过程中发生错误，则返回该错误。
func (x *XML) Render(w http.ResponseWriter, code int) error {
	// 设置响应的内容类型为XML格式。
	x.WriteContentType(w)
	// 写入HTTP响应的状态码。
	w.WriteHeader(code)
	// 使用xml.NewEncoder创建一个新的XML编码器，并将XML数据编码到响应中。
	// 如果编码过程中发生错误，该错误将被返回。
	err := xml.NewEncoder(w).Encode(x.Data)
	return err
}

// WriteContentType 方法用于设置HTTP响应的内容类型为XML格式。
// 参数w是http.ResponseWriter类型，用于写入HTTP响应。
func (s *XML) WriteContentType(w http.ResponseWriter) {
	// 调用writeContentType函数设置响应的内容类型为"application/xml; charset=utf-8"。
	writeContentType(w, "application/xml; charset=utf-8")
}
