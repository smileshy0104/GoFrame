package frame

import (
	"encoding/base64"
	"net/http"
)

// TODO 用于basic认证

// Accounts 结构体用于管理用户账户信息和认证逻辑
type Accounts struct {
	UnAuthHandler func(ctx *Context) // UnAuthHandler 是一个处理未授权访问的回调函数
	Users         map[string]string  // Users 存储用户名和密码的映射关系
	Realm         string             // Realm 是认证领域，用于HTTP认证
}

// BasicAuth 是一个中间件，用于处理基本的HTTP认证
// 它接收一个 HandlerFunc 类型的 next 参数，返回一个经过认证处理的 HandlerFunc
func (a *Accounts) BasicAuth(next HandlerFunc) HandlerFunc {
	return func(ctx *Context) {
		// 从请求中获取用户名和密码（调用ctx中的Basic认证请求）
		username, password, ok := ctx.R.BasicAuth()
		if !ok {
			a.unAuthHandler(ctx)
			return
		}
		// 检查用户名和密码是否匹配
		pwd, exist := a.Users[username]
		if !exist {
			a.unAuthHandler(ctx)
			return
		}
		// 检查密码是否匹配
		if pwd != password {
			a.unAuthHandler(ctx)
			return
		}
		// 将用户名存储到上下文中，以便后续的处理中使用
		ctx.Set("user", username)
		next(ctx)
	}
}

// unAuthHandler 处理未授权的访问请求
// 当用户未提供有效的认证信息时调用此方法
func (a *Accounts) unAuthHandler(ctx *Context) {
	// 如果用户提供了自定义的未授权处理函数，则调用该函数
	if a.UnAuthHandler != nil {
		a.UnAuthHandler(ctx)
	} else {
		//
		ctx.W.Header().Set("WWW-Authenticate", a.Realm)
		ctx.W.WriteHeader(http.StatusUnauthorized)
	}
}

// BasicAuth 函数生成基本认证的 base64 编码字符串
// 它接收用户名和密码作为参数，返回 base64 编码后的字符串
func BasicAuth(username, password string) string {
	// 创建一个包含用户名和密码的字符串，并以冒号分隔
	auth := username + ":" + password
	// 使用标准库的 base64 编码算法对字符串进行编码，并返回编码后的字符串
	return base64.StdEncoding.EncodeToString([]byte(auth))
}
