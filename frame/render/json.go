package render

import (
	"encoding/json"
	"net/http"
)

// JSON 结构体用于处理JSON格式的数据响应。
// 它封装了任何类型的数据（Data字段），并提供了渲染JSON数据到HTTP响应的方法。
type JSON struct {
	Data any // Data字段用于存储将被渲染为JSON的原始数据。
}

// Render 方法用于将JSON数据写入HTTP响应中。
// 参数w是http.ResponseWriter接口，用于向客户端写入响应。
// 参数code是HTTP状态码，例如200表示成功。
// 返回值error用于返回在JSON序列化或写入响应过程中发生的错误，如果没有错误则为nil。
func (j *JSON) Render(w http.ResponseWriter, code int) error {
	// 设置响应的内容类型为JSON。
	j.WriteContentType(w)
	// 写入HTTP状态码。
	w.WriteHeader(code)
	// 将Data字段序列化为JSON格式的字节切片。
	jsonData, err := json.Marshal(j.Data)
	if err != nil {
		// 如果序列化过程中发生错误，返回错误。
		return err
	}
	// 将序列化的JSON数据写入HTTP响应中。
	_, err = w.Write(jsonData)
	// 返回写入过程中可能发生的错误。
	return err
}

// WriteContentType 方法用于设置HTTP响应的内容类型。
// 参数w是http.ResponseWriter接口，用于向客户端写入响应。
func (j *JSON) WriteContentType(w http.ResponseWriter) {
	// 调用writeContentType函数设置响应头中的内容类型为JSON。
	writeContentType(w, "application/json; charset=utf-8")
}
