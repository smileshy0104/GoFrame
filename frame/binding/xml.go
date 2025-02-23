package binding

import (
	"encoding/xml"
	"net/http"
)

// xmlBinding 定义了一个用于处理XML数据绑定的结构体。
type xmlBinding struct {
}

// Name 返回绑定类型的名字。
// 该方法满足 binding.Interface 接口的要求。
func (xmlBinding) Name() string {
	return "xml"
}

// Bind 将HTTP请求中的XML数据绑定到指定的对象。
// 该方法满足 binding.Interface 接口的要求。
// 参数:
//
//	r: HTTP请求对象，用于获取请求体。
//	obj: 任何类型的对象，将请求体中的XML数据解码到该对象中。
//
// 返回值:
//
//	如果请求体为空，则返回nil。
//	如果解码过程中发生错误，则返回该错误。
//	如果解码成功，则调用validate函数进行数据验证，返回验证结果。
func (b xmlBinding) Bind(r *http.Request, obj any) error {
	// 检查请求体是否为空，为空则直接返回nil。
	if r.Body == nil {
		return nil
	}
	// 创建一个新的XML解码器。
	decoder := xml.NewDecoder(r.Body)
	// 使用解码器将请求体中的数据解码到obj对象中。
	if err := decoder.Decode(obj); err != nil {
		return err
	}
	// 调用validate函数对解码后的数据进行验证。
	return validate(obj)
}
