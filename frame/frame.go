package frame

import (
	"fmt"
	"log"
	"net/http"
)

const ANY = "ANY"

// HandlerFunc 定义了处理器函数的类型，它接收一个 http.ResponseWriter 和一个 *http.Request 作为参数
// type HandlerFunc func(w http.ResponseWriter, r *http.Request)

// HandlerFunc 定义了一个处理函数的类型，它接受一个 Context 指针作为参数。
// 这种类型通常用于定义路由、中间件等处理程序。
type HandlerFunc func(ctx *Context)

// Context 是请求处理的上下文，包含了请求和响应的引用。
// 它提供了一种在请求处理过程中传递请求特定数据、中断请求处理等方式。
type Context struct {
	W http.ResponseWriter // W 用于向客户端发送响应。
	R *http.Request       // R 包含了当前请求的所有信息。
}

// router 是路由管理的结构体，包含一组路由组
type router struct {
	groups []*routerGroup // 路由组的列表
}

// Group 方法用于创建一个新的路由组，并将其添加到 router 的 groups 列表中
func (r *router) Group(name string) *routerGroup {
	g := &routerGroup{
		groupName:        name,
		handlerMap:       make(map[string]map[string]HandlerFunc),
		handlerMethodMap: make(map[string][]string),
	}
	r.groups = append(r.groups, g)
	return g
}

// handle 是一个用于在路由组中注册处理程序的方法。
// 它接受三个参数：name（路由的名称）、method（HTTP 方法）和 handlerFunc（处理程序）。
// 该方法的主要作用是将处理程序与路由名称和HTTP方法关联起来，以便正确处理相应的HTTP请求。
func (r *routerGroup) handle(name string, method string, handlerFunc HandlerFunc) {
	// 检查 handlerMap 中是否已存在该路由名称。
	_, ok := r.handlerMap[name]
	// 如果不存在，则创建一个新map，用于存储该路由名称对应的处理程序。
	if !ok {
		r.handlerMap[name] = make(map[string]HandlerFunc)
	}
	// 将处理程序与路由名称和HTTP方法关联起来。
	r.handlerMap[name][method] = handlerFunc

	// 将路由名称添加到 handlerMethodMap 中，以便按HTTP方法进行索引。
	r.handlerMethodMap[method] = append(r.handlerMethodMap[method], name)
}

// Any 添加一个处理所有HTTP方法的路由
// 参数:
//
//	name: 路由的名称或路径
//	handlerFunc: 处理路由请求的处理函数
func (r *routerGroup) Any(name string, handlerFunc HandlerFunc) {
	r.handle(name, "ANY", handlerFunc)
}

// Handle 添加一个处理特定HTTP方法的路由
// 参数:
//
//	name: 路由的名称或路径
//	method: HTTP方法，如 GET, POST 等
//	handlerFunc: 处理路由请求的处理函数
//
// 注意: 会对method的有效性做校验
func (r *routerGroup) Handle(name string, method string, handlerFunc HandlerFunc) {
	//method有效性做校验
	r.handle(name, method, handlerFunc)
}

// Get 添加一个处理GET请求的路由
// 参数:
//
//	name: 路由的名称或路径
//	handlerFunc: 处理路由请求的处理函数
func (r *routerGroup) Get(name string, handlerFunc HandlerFunc) {
	r.handle(name, http.MethodGet, handlerFunc)
}

// Post 添加一个处理POST请求的路由
func (r *routerGroup) Post(name string, handlerFunc HandlerFunc) {
	r.handle(name, http.MethodPost, handlerFunc)
}

// Delete 添加一个处理DELETE请求的路由
func (r *routerGroup) Delete(name string, handlerFunc HandlerFunc) {
	r.handle(name, http.MethodDelete, handlerFunc)
}

// Put 添加一个处理PUT请求的路由
func (r *routerGroup) Put(name string, handlerFunc HandlerFunc) {
	r.handle(name, http.MethodPut, handlerFunc)
}

// Patch 添加一个处理PATCH请求的路由
func (r *routerGroup) Patch(name string, handlerFunc HandlerFunc) {
	r.handle(name, http.MethodPatch, handlerFunc)
}

// Options 添加一个处理OPTIONS请求的路由
func (r *routerGroup) Options(name string, handlerFunc HandlerFunc) {
	r.handle(name, http.MethodOptions, handlerFunc)
}

// Head 添加一个处理HEAD请求的路由
func (r *routerGroup) Head(name string, handlerFunc HandlerFunc) {
	r.handle(name, http.MethodHead, handlerFunc)
}

// routerGroup 代表一个路由组，包含组名和一组处理器函数映射
type routerGroup struct {
	groupName        string                            // 路由组的名称
	handlerMap       map[string]map[string]HandlerFunc // 路由和处理器函数的映射
	handlerMethodMap map[string][]string               // 路由和处理器函数的映射
}

// Add 方法用于向路由组中添加一个新的路由和对应的处理器函数
//func (r *routerGroup) Add(name string, handlerFunc HandlerFunc) {
//	r.handlerMap[name] = handlerFunc
//}

// Engine 是框架的核心结构体，包含一个 router 实例
type Engine struct {
	*router // 使用嵌套结构体，将 router 实例作为 Engine 的字段
}

// New 函数用于创建并返回一个新的 Engine 实例
func New() *Engine {
	return &Engine{
		&router{},
	}
}

// ServeHTTP是Engine的HTTP服务处理函数。
// 它根据请求的URL和方法选择相应的处理程序。
// 参数w用于发送HTTP响应，r包含当前请求的所有信息。
func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 获取所有路由组。
	groups := e.router.groups
	// 遍历每个路由组。
	for _, g := range groups {
		// 遍历当前路由组中的所有处理程序映射。
		for name, methodHandle := range g.handlerMap {
			// 构建完整的URL路径。
			url := "/" + g.groupName + name
			// 检查请求的URL是否与构建的URL匹配。
			if r.RequestURI == url {
				// 创建一个Context实例，包含请求和响应写入器。
				ctx := &Context{
					W: w,
					R: r,
				}
				// 检查是否有ANY方法处理程序。
				_, ok := methodHandle[ANY]
				if ok {
					// 如果有，执行ANY方法处理程序。
					methodHandle[ANY](ctx)
					return
				}
				// 获取请求的方法。
				method := r.Method
				fmt.Println(method)
				// 根据请求方法获取相应的处理程序。
				handler, ok := methodHandle[method]
				if ok {
					// 如果找到，执行相应的处理程序。
					handler(ctx)
					return
				}
				// 如果没有找到合适的处理程序，发送405方法不允许的响应。
				w.WriteHeader(http.StatusMethodNotAllowed)
				fmt.Fprintln(w, method+" not allowed")
				return
			} else {
				// 如果请求的URL与任何路由都不匹配，发送404未找到的响应。
				w.WriteHeader(http.StatusNotFound)
				fmt.Fprintln(w, r.RequestURI+" not found")
				return
			}
		}
	}
}

// Run 启动 HTTP 服务器，监听指定的端口。
func (e *Engine) Run() {
	// 将 Engine 实例注册为 HTTP 服务器的处理程序
	http.Handle("/", e)
	// 监听 8111 端口并启动服务器
	err := http.ListenAndServe(":8111", nil)
	if err != nil {
		log.Fatal(err)
	}
}
