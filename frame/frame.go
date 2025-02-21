package frame

import (
	"log"
	"net/http"
)

// HandlerFunc 定义了处理器函数的类型，它接收一个 http.ResponseWriter 和一个 *http.Request 作为参数
type HandlerFunc func(w http.ResponseWriter, r *http.Request)

// router 是路由管理的结构体，包含一组路由组
type router struct {
	groups []*routerGroup
}

// Group 方法用于创建一个新的路由组，并将其添加到 router 的 groups 列表中
func (r *router) Group(name string) *routerGroup {
	g := &routerGroup{groupName: name, handlerMap: make(map[string]HandlerFunc)}
	r.groups = append(r.groups, g)
	return g
}

// routerGroup 代表一个路由组，包含组名和一组处理器函数映射
type routerGroup struct {
	groupName  string
	handlerMap map[string]HandlerFunc
}

// Add 方法用于向路由组中添加一个新的路由和对应的处理器函数
func (r *routerGroup) Add(name string, handlerFunc HandlerFunc) {
	r.handlerMap[name] = handlerFunc
}

// Engine 是框架的核心结构体，包含一个 router 实例
type Engine struct {
	*router
}

// New 函数用于创建并返回一个新的 Engine 实例
func New() *Engine {
	return &Engine{
		&router{},
	}
}

// Run 方法用于启动 HTTP 服务器，监听指定端口，并为所有路由组和处理器函数设置路由
func (e *Engine) Run() {
	groups := e.router.groups
	for _, g := range groups {
		for name, handle := range g.handlerMap {
			http.HandleFunc("/"+g.groupName+name, handle)
		}
	}
	err := http.ListenAndServe(":8111", nil)
	if err != nil {
		log.Fatal(err)
	}
}
