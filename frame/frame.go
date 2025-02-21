package frame

import (
	"fmt"
	"log"
	"net/http"
)

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
		handlerMap:       make(map[string]HandlerFunc),
		handlerMethodMap: make(map[string][]string),
	}
	r.groups = append(r.groups, g)
	return g
}

// Any 为当前路由组添加一个处理所有HTTP方法的路由。
// 该方法接收路由的名称和处理函数作为参数。
// 它将路由的处理函数注册到handlerMap中，并将该路由的名称添加到处理所有方法的handlerMethodMap中。
func (r *routerGroup) Any(name string, handlerFunc HandlerFunc) {
	r.handlerMap[name] = handlerFunc
	r.handlerMethodMap["ANY"] = append(r.handlerMethodMap["ANY"], name)
}

// Get 为当前路由组添加一个处理GET请求的路由。
// 该方法接收路由的名称和处理函数作为参数。
// 它将路由的处理函数注册到handlerMap中，并将该路由的名称添加到处理GET请求的handlerMethodMap中。
func (r *routerGroup) Get(name string, handlerFunc HandlerFunc) {
	r.handlerMap[name] = handlerFunc
	r.handlerMethodMap["GET"] = append(r.handlerMethodMap["GET"], name)
}

// Post 为当前路由组添加一个处理POST请求的路由。
// 该方法接收路由的名称和处理函数作为参数。
// 它将路由的处理函数注册到handlerMap中，并将该路由的名称添加到处理POST请求的handlerMethodMap中。
func (r *routerGroup) Post(name string, handlerFunc HandlerFunc) {
	r.handlerMap[name] = handlerFunc
	r.handlerMethodMap["POST"] = append(r.handlerMethodMap["POST"], name)
}

// routerGroup 代表一个路由组，包含组名和一组处理器函数映射
type routerGroup struct {
	groupName        string                 // 路由组的名称
	handlerMap       map[string]HandlerFunc // 路由和处理器函数的映射
	handlerMethodMap map[string][]string    // 路由和处理器函数的映射
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

// ServeHTTP 是 Engine 类型的 HTTP 服务处理函数。
// 它根据请求的 URL 和方法找到并执行相应的处理程序。
func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 获取所有路由组
	groups := e.router.groups
	// 遍历每个路由组
	for _, g := range groups {
		// 遍历当前路由组中的所有处理程序映射
		for name, handle := range g.handlerMap {
			// 构造完整的 URL 路径
			url := "/" + g.groupName + name
			// 如果请求的 URI 与构造的 URL 匹配，则创建上下文并尝试执行处理程序
			if r.RequestURI == url {
				ctx := &Context{
					W: w,
					R: r,
				}
				// 检查是否有 ANY 方法的处理程序
				if g.handlerMethodMap["ANY"] != nil {
					for _, v := range g.handlerMethodMap["ANY"] {
						if name == v {
							handle(ctx)
							return
						}
					}
				}
				// 获取请求的方法并打印
				method := r.Method
				fmt.Println(method)
				// 根据请求方法获取对应的处理程序列表
				routers := g.handlerMethodMap[method]
				if routers != nil {
					for _, v := range routers {
						if name == v {
							handle(ctx)
							return
						}
					}
				}
				// 如果没有找到允许的方法，返回 405 方法不允许错误
				w.WriteHeader(405)
				fmt.Fprintln(w, method+" not allowed")
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
