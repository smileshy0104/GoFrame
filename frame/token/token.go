package token

import (
	"errors"
	"fmt"
	"frame"
	"github.com/golang-jwt/jwt/v4"
	"net/http"
	"time"
)

// JWTToken 常量用于定义JWT token的默认cookie名称
const JWTToken = "frame_token"

// JwtHandler 结构体定义了JWT的处理配置，包括算法、过期时间、密钥等配置项
type JwtHandler struct {
	// jwt的算法
	Alg string
	// 过期时间
	TimeOut time.Duration
	// 刷新token过期时间
	RefreshTimeOut time.Duration
	// 时间函数
	TimeFuc func() time.Time
	// Key
	Key []byte
	// 刷新key
	RefreshKey string
	// 私钥
	PrivateKey string
	// 发送cookie
	SendCookie bool
	// 用户认证函数相关内容（实际开发中通常会进行设置）
	Authenticator func(ctx *frame.Context) (map[string]any, error)

	CookieName     string
	CookieMaxAge   int64
	CookieDomain   string
	SecureCookie   bool
	CookieHTTPOnly bool
	Header         string
	AuthHandler    func(ctx *frame.Context, err error)
}

// JwtResponse 结构体用于封装生成的JWT token和刷新token
type JwtResponse struct {
	Token        string
	RefreshToken string
}

// LoginHandler 处理用户登录，验证用户凭据并生成JWT token
func (j *JwtHandler) LoginHandler(ctx *frame.Context) (*JwtResponse, error) {
	// 验证用户凭据，并返回用户数据
	data, err := j.Authenticator(ctx)
	if err != nil {
		return nil, err
	}
	// 设置默认的算法为HS256
	if j.Alg == "" {
		j.Alg = "HS256"
	}
	// A部分 Header部分
	// 通过jwt.GetSigningMethod方法获取指定的签名算法
	signingMethod := jwt.GetSigningMethod(j.Alg)
	// 通过jwt.New方法创建一个新的token对象
	token := jwt.New(signingMethod)

	// B部分 PayLoad部分
	claims := token.Claims.(jwt.MapClaims)
	// 设置PayLoad部分，存放用户数据data
	if data != nil {
		for key, value := range data {
			claims[key] = value
		}
	}
	// 设置过期时间（token）
	if j.TimeFuc == nil {
		j.TimeFuc = func() time.Time {
			return time.Now()
		}
	}
	expire := j.TimeFuc().Add(j.TimeOut)
	// 过期时间
	claims["exp"] = expire.Unix()
	// 发布时间
	claims["iat"] = j.TimeFuc().Unix()

	var tokenString string
	var tokenErr error
	// C部分 secret密钥部分
	// 根据不同的算法使用不同的密钥
	if j.usingPublicKeyAlgo() {
		// 使用私钥签名
		tokenString, tokenErr = token.SignedString(j.PrivateKey)
	} else {
		// 使用密钥签名
		tokenString, tokenErr = token.SignedString(j.Key)
	}
	if tokenErr != nil {
		return nil, tokenErr
	}
	// 返回创建的token
	jr := &JwtResponse{
		Token: tokenString,
	}
	// refreshToken 设置刷新token
	refreshToken, err := j.refreshToken(token)
	if err != nil {
		return nil, err
	}
	// 设置refreshToken
	jr.RefreshToken = refreshToken

	//发送存储cookie
	if j.SendCookie {
		if j.CookieName == "" {
			j.CookieName = JWTToken
		}
		if j.CookieMaxAge == 0 {
			j.CookieMaxAge = expire.Unix() - j.TimeFuc().Unix()
		}
		ctx.SetCookie(j.CookieName, tokenString, int(j.CookieMaxAge), "/", j.CookieDomain, j.SecureCookie, j.CookieHTTPOnly)
	}

	return jr, nil
}

// usingPublicKeyAlgo 判断是否使用公钥算法
func (j *JwtHandler) usingPublicKeyAlgo() bool {
	switch j.Alg {
	case "RS256", "RS512", "RS384":
		return true
	}
	return false
}

// refreshToken 生成刷新token
func (j *JwtHandler) refreshToken(token *jwt.Token) (string, error) {
	// 将token的Claims部分转换为jwt.MapClaims类型，以便于后续操作。
	claims := token.Claims.(jwt.MapClaims)
	// 更新claims中的"exp"字段，设置新的过期时间。
	claims["exp"] = j.TimeFuc().Add(j.RefreshTimeOut).Unix()

	// 初始化tokenString变量，用于存储生成的token字符串。
	// 初始化tokenErr变量，用于存储生成token过程中可能出现的错误。
	var tokenString string
	var tokenErr error

	// 根据是否使用公钥算法来决定使用PrivateKey还是Key对token进行签名。
	if j.usingPublicKeyAlgo() {
		// 使用PrivateKey对token进行签名，生成token字符串。
		tokenString, tokenErr = token.SignedString(j.PrivateKey)
	} else {
		// 使用Key对token进行签名，生成token字符串。
		tokenString, tokenErr = token.SignedString(j.Key)
	}

	// 如果生成token过程中出现错误，则返回空字符串和错误信息。
	if tokenErr != nil {
		return "", tokenErr
	}

	// 返回生成的token字符串和nil错误，表示token成功生成。
	return tokenString, nil

}

// LogoutHandler 处理用户退出登录，通过清除cookie实现
// TODO 在实际开发中我们把对应信息存入redis中，logout时将对应的token从redis中删除
func (j *JwtHandler) LogoutHandler(ctx *frame.Context) error {
	// 如果配置了发送Cookie，则执行以下操作
	if j.SendCookie {
		// 如果Cookie名称未设置，则使用默认名称JWTToken
		if j.CookieName == "" {
			j.CookieName = JWTToken
		}
		// 清除Cookie，通过设置过期时间为负值表示立即过期
		// 这里解释了为什么要清除Cookie：可能是为了安全原因，比如避免跟踪或泄露信息
		ctx.SetCookie(j.CookieName, "", -1, "/", j.CookieDomain, j.SecureCookie, j.CookieHTTPOnly)
		// 操作成功，返回nil表示无错误发生
		return nil
	}
	// 如果不发送Cookie，也返回nil，表示无操作执行
	return nil

}

// TODO （不需要用户重新走对应的登陆逻辑）
// RefreshHandler 处理token刷新请求，验证刷新token并生成新的JWT token（不需要用户重新走对应的登陆逻辑）
// RefreshHandler 是一个处理JWT刷新请求的方法。
// 它从上下文中获取刷新令牌，验证并生成一个新的访问令牌。
// 参数: ctx *frame.Context - 包含请求上下文的指针。
// 返回值: *JwtResponse - 包含新生成的访问令牌和刷新令牌的响应对象。
//
//	error - 如果操作失败，返回错误。
func (j *JwtHandler) RefreshHandler(ctx *frame.Context) (*JwtResponse, error) {
	// 从上下文中获取刷新令牌。
	rToken, ok := ctx.Get(j.RefreshKey)
	if !ok {
		return nil, errors.New("refresh token is null")
	}
	// 如果算法未设置，则默认为HS256。
	if j.Alg == "" {
		j.Alg = "HS256"
	}
	// 解析token（通过rToken获取对应的token信息）
	t, err := jwt.Parse(rToken.(string), func(token *jwt.Token) (interface{}, error) {
		// 根据使用的算法返回相应的密钥。
		if j.usingPublicKeyAlgo() {
			return j.PrivateKey, nil
		} else {
			return j.Key, nil
		}
	})
	if err != nil {
		return nil, err
	}
	// B部分
	// 获取PayLoad部分，并设置过期时间。
	claims := t.Claims.(jwt.MapClaims)
	// 如果未设置时间函数，则使用当前时间。
	if j.TimeFuc == nil {
		j.TimeFuc = func() time.Time {
			return time.Now()
		}
	}
	// 计算新的过期时间。
	expire := j.TimeFuc().Add(j.TimeOut)
	// 过期时间
	claims["exp"] = expire.Unix()
	claims["iat"] = j.TimeFuc().Unix()
	// C部分 secret
	var tokenString string
	var tokenErr error
	// 根据使用的算法生成新的访问令牌。
	if j.usingPublicKeyAlgo() {
		tokenString, tokenErr = t.SignedString(j.PrivateKey)
	} else {
		tokenString, tokenErr = t.SignedString(j.Key)
	}
	if tokenErr != nil {
		return nil, tokenErr
	}
	// 创建包含新访问令牌的响应对象。
	jr := &JwtResponse{
		Token: tokenString,
	}
	// refreshToken
	refreshToken, err := j.refreshToken(t)
	if err != nil {
		return nil, err
	}
	// 将新的刷新令牌添加到响应对象中。
	jr.RefreshToken = refreshToken
	// 当SendCookie为true时，设置cookie
	if j.SendCookie {
		// 如果CookieName为空，则将其设置为JWTToken。
		if j.CookieName == "" {
			j.CookieName = JWTToken
		}
		// 如果CookieMaxAge为0，则计算并设置其值为令牌到期时间与当前时间的差值。
		if j.CookieMaxAge == 0 {
			j.CookieMaxAge = expire.Unix() - j.TimeFuc().Unix()
		}
		// 设置cookie相关内容（在实际开发中我们把对应信息存入redis中，再进行获取）。
		ctx.SetCookie(j.CookieName, tokenString, int(j.CookieMaxAge), "/", j.CookieDomain, j.SecureCookie, j.CookieHTTPOnly)
	}
	// 返回包含新访问令牌和刷新令牌的响应对象。
	return jr, nil
}

// AuthInterceptor jwt登录中间件，验证请求中的token
func (j *JwtHandler) AuthInterceptor(next frame.HandlerFunc) frame.HandlerFunc {
	// 返回一个处理JWT认证的中间件函数
	return func(ctx *frame.Context) {
		// 如果Header名称未设置，则默认为"Authorization"
		if j.Header == "" {
			j.Header = "Authorization"
		}

		// 从请求头中获取token
		token := ctx.R.Header.Get(j.Header)

		// 如果token为空且配置了发送Cookie，则从Cookie中获取token
		if token == "" && j.SendCookie {
			cookie, err := ctx.R.Cookie(j.CookieName)
			if err != nil {
				// 如果Cookie获取失败，则调用错误处理函数
				j.AuthErrorHandler(ctx, err)
				return
			}
			token = cookie.String()
		}

		// 如果token仍然为空，则调用错误处理函数并返回
		if token == "" {
			j.AuthErrorHandler(ctx, errors.New("token is null"))
			return
		}

		// 解析token（可以成功解析出对应内容）
		t, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
			// 根据使用的算法选择密钥
			if j.usingPublicKeyAlgo() {
				return j.PrivateKey, nil
			} else {
				return j.Key, nil
			}
		})
		if err != nil {
			// 如果token解析失败，则调用错误处理函数
			j.AuthErrorHandler(ctx, err)
			return
		}

		// 获取token中的声明
		claims := t.Claims.(jwt.MapClaims)

		// 将声明设置到上下文中
		//ctx.Set("jwt_claims", claims)
		//ctx.Set("jwt_claims", claims)
		fmt.Println(claims)

		// 调用下一个中间件或处理函数
		next(ctx)
	}

}

// AuthErrorHandler 处理认证错误，如果没有设置自定义的错误处理函数，则返回401状态码
func (j *JwtHandler) AuthErrorHandler(ctx *frame.Context, err error) {
	if j.AuthHandler == nil {
		ctx.W.WriteHeader(http.StatusUnauthorized)
	} else {
		j.AuthHandler(ctx, err)
	}
}
