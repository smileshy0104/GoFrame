package main

import (
	"fmt"
	"frame"
)

// User 结构体
type User struct {
	Name      string   `xml:"name" json:"name" msgo:"required"`
	Age       int      `xml:"age" json:"age" validate:"required,max=50,min=18"`
	Addresses []string `json:"addresses"`
	Email     string   `json:"email" msgo:"required"`
}

// Log 中间件
func Log(next frame.HandlerFunc) frame.HandlerFunc {
	return func(ctx *frame.Context) {
		fmt.Println("打印请求参数")
		next(ctx)
		fmt.Println("返回执行时间")
	}
}

func main() {
	engine := frame.New()
	g := engine.Group("user")

	// 使用 Use 方法添加一个中间件，该中间件会在处理请求之前和之后分别执行一些操作。
	// 通用级别中间件
	g.Use(func(next frame.HandlerFunc) frame.HandlerFunc {
		return func(ctx *frame.Context) {
			fmt.Println("pre handler")
			next(ctx)
			fmt.Println("post handler")
		}
	})

	g.Get("/hello", func(context *frame.Context) {
		fmt.Fprintln(context.W, "GET test test")
	})

	g.Post("/hello", func(context *frame.Context) {
		fmt.Fprintln(context.W, "POST test test")
	})

	g.Get("/hello/*/get", func(context *frame.Context) {
		fmt.Fprintln(context.W, "/hello/*/get test test")
	}, Log)

	g.Get("/html", func(context *frame.Context) {
		context.HTML(200, "<h1>测试测试</h1>")
	}, Log)

	// TODO 页面渲染相关代码
	g.Get("/htmlTemplate", func(context *frame.Context) {
		context.HTMLTemplate("index.html", nil, "", "tpl/index.html")
	}, Log)

	g.Get("/htmlTemplate1", func(context *frame.Context) {
		user := User{Name: "yyds"}
		context.HTMLTemplate("login.html", nil, user, "tpl/login.html", "tpl/header.html")
	}, Log)

	g.Get("/htmlTemplateGlob", func(context *frame.Context) {
		user := User{Name: "yyds"}
		// 匹配所有以.html结尾的文件
		context.HTMLTemplateGlob("login.html", nil, "tpl/*.html", user)
	}, Log)

	// 提前加载模板
	engine.LoadTemplateGlob("tpl/*.html")

	// 模板渲染
	g.Get("/template", func(context *frame.Context) {
		user := User{Name: "yyds"}
		err := context.Template("login.html", user)
		if err != nil {
			fmt.Println(err)
		}
	}, Log)
	engine.Run()
}
