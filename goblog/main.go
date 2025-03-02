package main

import (
	"errors"
	"fmt"
	"frame"
	newerror "frame/error"
	newlogger "frame/log"
	"log"
	"net/http"
)

// User 结构体
type User struct {
	Name      string   `xml:"name" json:"name" binding:"required"`
	Age       int      `xml:"age" json:"age" validate:"required,max=50,min=10"`
	Addresses []string `json:"addresses"`
	Email     string   `json:"email" binding:"required"`
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
	//engine := frame.New()
	// 创建一个Engine实例，并设置日志记录器为默认日志记录器。
	engine := frame.Default()
	// 设置错误处理函数，用于处理框架的错误。
	engine.RegisterErrorHandler(func(err error) (int, any) {
		switch e := err.(type) {
		case *BlogResponse:
			return http.StatusOK, e.Response()
		default:
			return http.StatusInternalServerError, "500 error"
		}
	})

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

	// 通用级别中间件（日志中间件）
	g.Use(frame.Logging)

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

	// JSON渲染
	g.Get("/json", func(ctx *frame.Context) {
		_ = ctx.JSON(http.StatusOK, &User{
			Name: "Json渲染测试",
		})
	})

	// XML渲染
	g.Get("/xml", func(ctx *frame.Context) {
		_ = ctx.XML(http.StatusOK, &User{
			Name: "XML渲染测试",
		})
	})

	// String渲染
	g.Get("/string", func(ctx *frame.Context) {
		ctx.String(http.StatusOK, "%s 渲染 %s \n", "string", "go微服务框架")
	})

	// 文件渲染(默认文件名)
	g.Get("/excel", func(ctx *frame.Context) {
		ctx.File("tpl/test.xlsx")
	})

	// 文件渲染(指定文件名)
	g.Get("/excelName", func(ctx *frame.Context) {
		ctx.FileAttachment("tpl/test.xlsx", "指定文件名称.xlsx")
	})

	// 文件渲染(指定文件目录)
	g.Get("/excelFs", func(ctx *frame.Context) {
		//ctx.FileAttachment("tpl/test.xlsx", "哈哈.xlsx")
		ctx.FileFromFS("test1.xlsx", http.Dir("tpl"))
	})

	// 页面重定向
	g.Get("/redirect", func(ctx *frame.Context) {
		ctx.Redirect(http.StatusMovedPermanently, "https://www.baidu.com")
	})

	// GetQuery获取请求参数 http://localhost:8111/user/get_query?id=1
	g.Get("/get_query", func(ctx *frame.Context) {
		id := ctx.GetQuery("id")
		fmt.Printf("id: %v , ok: %v \n", id, true)
	})

	// GetQueryArray获取请求参数 http://localhost:8111/user/get_query_array?id=1&id=2
	g.Get("/get_query_array", func(ctx *frame.Context) {
		id, ok := ctx.GetQueryArray("id")
		fmt.Printf("id: %v , ok: %v \n", id, ok)
	})

	// GetQueryArray获取请求参数 http://localhost:8111/user/get_default_query?id=1
	g.Get("/get_default_query", func(ctx *frame.Context) {
		id := ctx.DefaultQuery("id", "999")
		fmt.Printf("id: %v  \n", id)
	})

	// GetQueryMap获取请求参数 http://localhost:8111/user/get_query_map?user[id]=1&user[name]=张三
	g.Get("/get_query_map", func(ctx *frame.Context) {
		m, _ := ctx.GetQueryMap("user")
		ctx.JSON(http.StatusOK, m)
	})

	// GetPostForm/GetPostFormArray获取请求参数 通过form-data输入
	g.Post("/form_post", func(ctx *frame.Context) {
		//m, _ := ctx.GetPostForm("user") // 单个获取
		m, _ := ctx.GetPostFormArray("user")
		ctx.JSON(http.StatusOK, m)
	})

	// GetPostFormMap获取请求参数 通过form-data输入 user[id]=1&user[name]=张三
	g.Post("/form_post_map", func(ctx *frame.Context) {
		m, _ := ctx.GetPostFormMap("user") // 单个获取
		ctx.JSON(http.StatusOK, m)
	})

	// FormFiles获取请求参数 通过form-data输入
	g.Post("/form_post_file", func(ctx *frame.Context) {
		files := ctx.FormFiles("file")
		for _, file := range files {
			err := ctx.SaveUploadedFile(file, "./upload/"+file.Filename)
			if err != nil {
				fmt.Println(err)
			}
		}
		ctx.JSON(http.StatusOK, "上传成功！")
	})

	g.Post("/jsonParam0", func(ctx *frame.Context) {
		user := &User{}
		err := ctx.DealJson(user)
		if err == nil {
			ctx.JSON(http.StatusOK, user)
		} else {
			log.Println(err)
		}
	})

	// JSON参数绑定
	/**
		[
	    {
	        "name": "张三",
	        "age": 10,
	        "addresses": [
	            "北京",
	            "杭州"
	        ],
	        "email": "www.baidu.com"
	    }
	]
	*/
	g.Post("/jsonParam", func(ctx *frame.Context) {
		user := make([]User, 0)
		//ctx.DisallowUnknownFields = true
		//ctx.IsValidate = true
		err := ctx.BindJson(&user)
		if err == nil {
			ctx.JSON(http.StatusOK, user)
		} else {
			log.Println(err)
		}
	})

	// XML参数绑定
	/**
	<User>
	<name>张三</name>
	<age>20</age>
	</User>
	*/
	g.Post("/xmlParam", func(ctx *frame.Context) {
		user := &User{}
		//user := User{}
		err := ctx.BindXML(user)
		if err == nil {
			ctx.JSON(http.StatusOK, user)
		} else {
			log.Println(err)
		}
	})

	// TODO 封装日志记录器
	engine.Logger.Level = newlogger.LevelDebug
	//engine.Logger.Formatter = &newlogger.JsonFormatter{
	//	TimeDisplay: true,
	//}
	//logger.Outs = append(logger.Outs, msLog.FileWriter("./log/log.log"))
	engine.Logger.LogFileSize = 1 << 10
	//engine.Logger.SetLogPath("./log")

	// 内置日志包
	g.Get("/log_test", func(ctx *frame.Context) {

		// 三种不同级别的日志输出
		//log.Println("log_test")
		//log.Fatal("log_test")
		//log.Panic("log_test")

		// 调用自定义的logger
		//ctx.Logger.Debug("log_test")
		//ctx.Logger.Info("log_test")
		//ctx.Logger.Error("log_test")

		// TODO 未封装日志记录器
		//logger := newlogger.Default()
		//// 指定展示的格式（默认展示text格式）
		//logger.Formatter = &newlogger.JsonFormatter{
		//	TimeDisplay: true,
		//}
		//logger.SetLogPath("./log")
		//logger.LogFileSize = 1 << 10
		//logger.WithFields(newlogger.Fields{
		//	"name": "yyds",
		//	"age":  18,
		//	"sex":  "男",
		//}).Debug("我是debug日志")
		ctx.Logger.Info("我是info日志")
		ctx.Logger.Error("我是err日志")

	})

	var u *User
	g.Post("/xmlParamErr", func(ctx *frame.Context) {
		u.Age = 10
		user := &User{}
		err := ctx.BindXML(user)
		if err == nil {
			ctx.JSON(http.StatusOK, user)
		} else {
			log.Println(err)
		}
		//err := newerror.Default()
		//err.Result(func(Error *newerror.MsError) {
		//	ctx.Logger.Error(Error.Error())
		//	ctx.JSON(http.StatusInternalServerError, user)
		//})
		//a(1, err)
		//b(1, err)
		//c(1, err)
		//ctx.JSON(http.StatusOK, user)
		//err := login()
		//ctx.HandleWithError(http.StatusOK, user, err)
	})
	engine.Run()
}

type BlogResponse struct {
	Success bool
	Code    int
	Data    any
	Msg     string
}
type BlogNoDataResponse struct {
	Success bool
	Code    int
	Msg     string
}

func (b *BlogResponse) Error() string {
	return b.Msg
}

func (b *BlogResponse) Response() any {
	if b.Data == nil {
		return &BlogNoDataResponse{
			Success: false,
			Code:    -999,
			Msg:     "账号密码错误",
		}
	}
	return b
}

func login() *BlogResponse {
	return &BlogResponse{
		Success: false,
		Code:    -999,
		Data:    nil,
		Msg:     "账号密码错误",
	}
}

func a(param int, msError *newerror.MsError) {
	if param == 1 {
		//发生错误的时候，放入一个地方 然后进行统一处理
		err := errors.New("a error")
		msError.Put(err)
	}
}

func b(param int, msError *newerror.MsError) {
	if param == 1 {
		err2 := errors.New("b error")
		msError.Put(err2)
	}
}

func c(param int, msError *newerror.MsError) {
	if param == 1 {
		err2 := errors.New("c error")
		msError.Put(err2)
	}
}
