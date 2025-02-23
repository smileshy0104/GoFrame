package binding

import "net/http"

// Binding 定义了将HTTP请求数据绑定到Go对象的接口。
// 它包括两个方法：Name和Bind。
// Name 方法返回绑定类型的名称。
// Bind 方法负责将HTTP请求的数据绑定到指定的Go对象实例。
type Binding interface {
	Name() string
	Bind(*http.Request, any) error
}

// JSON 和 XML 是 Binding 接口的两个实现示例。
// 这里通过具体实现（jsonBinding 和 xmlBinding）来实例化它们。
var (
	// JSON 用于JSON数据格式的绑定。
	JSON = jsonBinding{}
	// XML 用于XML数据格式的绑定。
	XML = xmlBinding{}
)
