package rpc

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"
)

// FrameHttpClient 是一个用于HTTP请求的客户端结构体，包含了一个HTTP客户端和一个服务映射表
type FrameHttpClient struct {
	client     http.Client
	serviceMap map[string]FrameService
}

// NewHttpClient 创建并返回一个新的FrameHttpClient实例
// 该函数初始化了一个HTTP客户端，设置了超时和连接池的相关参数，以优化网络请求的性能
func NewHttpClient() *FrameHttpClient {
	//Transport 请求分发  协程安全 连接池
	client := http.Client{
		Timeout: time.Duration(3) * time.Second,
		Transport: &http.Transport{
			MaxIdleConnsPerHost:   5,
			MaxConnsPerHost:       100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
	// 返回新的FrameHttpClient实例，初始化服务映射表
	return &FrameHttpClient{client: client, serviceMap: make(map[string]FrameService)}
}

// GetRequest 创建一个GET请求。
// 如果有参数，则将它们附加到URL中。
func (c *FrameHttpClient) GetRequest(method string, url string, args map[string]any) (*http.Request, error) {
	// 将参数转换为查询字符串并附加到URL
	if args != nil && len(args) > 0 {
		url = url + "?" + c.toValues(args)
	}
	// 创建HTTP请求
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	return req, nil
}

// FormRequest 创建一个用于表单提交的请求。
// 参数被编码为表单格式并作为请求体。
func (c *FrameHttpClient) FormRequest(method string, url string, args map[string]any) (*http.Request, error) {
	// 创建HTTP请求，并将参数作为表单数据传递
	req, err := http.NewRequest(method, url, strings.NewReader(c.toValues(args)))
	if err != nil {
		return nil, err
	}
	return req, nil
}

// JsonRequest 创建一个JSON请求。
// 参数被序列化为JSON格式并作为请求体。
func (c *FrameHttpClient) JsonRequest(method string, url string, args map[string]any) (*http.Request, error) {
	// 将参数序列化为JSON格式
	jsonStr, _ := json.Marshal(args)
	// 创建HTTP请求，并将JSON数据作为请求体传递
	req, err := http.NewRequest(method, url, bytes.NewReader(jsonStr))
	if err != nil {
		return nil, err
	}
	return req, nil
}

// Response 方法用于获取HTTP请求的响应数据。
// 该方法接收一个http.Request对象作为参数，表示要发送的HTTP请求。
// 方法返回一个字节切片，包含服务器的响应内容，以及一个错误对象，用于处理可能发生的错误。
// 此方法的设计目的是为了处理HTTP响应，将请求发送到服务器并接收响应数据。
func (c *FrameHttpClientSession) Response(req *http.Request) ([]byte, error) {
	return c.responseHandle(req)
}

// Get 方法用于发起一个GET请求，并获取响应数据。
// 该方法接受一个URL和一个参数映射，将参数映射转换为查询字符串并附加到URL上。
func (c *FrameHttpClientSession) Get(url string, args map[string]any) ([]byte, error) {
	// 如果有查询参数，将它们转换为URL查询字符串格式并附加到URL末尾。
	if args != nil && len(args) > 0 {
		url = url + "?" + c.toValues(args)
	}
	// 打印最终构造的GET请求URL，用于调试和日志记录。
	log.Println(url)

	// 创建GET请求。
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// 处理响应。
	return c.responseHandle(request)
}

// PostForm 执行一个POST请求，发送表单数据。
func (c *FrameHttpClientSession) PostForm(url string, args map[string]any) ([]byte, error) {
	// 创建一个POST请求，使用表单数据。
	request, err := http.NewRequest("POST", url, strings.NewReader(c.toValues(args)))
	if err != nil {
		return nil, err
	}
	// 处理响应。
	return c.responseHandle(request)
}

// PostJson 执行一个POST请求，发送JSON数据。
func (c *FrameHttpClientSession) PostJson(url string, args map[string]any) ([]byte, error) {
	// 将args转换为JSON格式。
	marshal, _ := json.Marshal(args)
	// 创建一个POST请求，使用JSON数据。
	request, err := http.NewRequest("POST", url, bytes.NewReader(marshal))
	if err != nil {
		return nil, err
	}
	// 处理响应。
	return c.responseHandle(request)
}

// responseHandle 是 FrameHttpClientSession 类中的一个方法，用于处理 HTTP 响应。
// 它接收一个 HTTP 请求指针作为参数，并返回响应的字节切片和一个错误对象（如果有的话）。
func (c *FrameHttpClientSession) responseHandle(request *http.Request) ([]byte, error) {
	// 调用 ReqHandler 处理请求，这可能是为了在发送请求前进行一些定制的处理。
	c.ReqHandler(request)

	// 使用客户端发送 HTTP 请求。
	response, err := c.client.Do(request)
	if err != nil {
		// 如果发送请求或获取响应时发生错误，返回错误。
		return nil, err
	}

	// 检查 HTTP 响应状态码是否为 200（OK）。
	if response.StatusCode != http.StatusOK {
		// 如果状态码不是 200，构造一个错误信息并返回。
		info := fmt.Sprintf("response status is %d", response.StatusCode)
		return nil, errors.New(info)
	}

	// 创建一个缓冲区读取器来读取响应体。
	reader := bufio.NewReader(response.Body)

	// 关闭响应体的文件描述符，以释放资源。
	defer response.Body.Close()

	// 创建一个缓冲区用于存储读取的响应数据。
	var buf = make([]byte, 127)
	var body []byte

	// 循环读取响应体的数据，直到全部读取完毕或遇到错误。
	for {
		n, err := reader.Read(buf)
		if err != nil && err != io.EOF {
			// 如果读取时发生非 EOF 错误，返回错误。
			return nil, err
		}
		if err == io.EOF || n == 0 {
			// 如果读取到 EOF 或没有读取到任何数据，退出循环。
			break
		}
		// 将读取到的数据追加到 body 中。
		body = append(body, buf[:n]...)
		if n < len(buf) {
			// 如果读取的数据量小于缓冲区大小，说明已经读取完毕，退出循环。
			break
		}
	}

	// 返回读取到的响应体数据。
	return body, nil
}

func (c *FrameHttpClient) toValues(args map[string]any) string {
	if args != nil && len(args) > 0 {
		paraFrame := url.Values{}
		for k, v := range args {
			paraFrame.Set(k, fmt.Sprintf("%v", v))
		}
		return paraFrame.Encode()
	}
	return ""
}

type HttpConfig struct {
	Protocol string
	Host     string
	Port     int
}

const (
	HTTP  = "http"
	HTTPS = "https"
)
const (
	GET      = "GET"
	POSTForm = "POST_FORM"
	POSTJson = "POST_JSON"
)

type FrameService interface {
	Env() HttpConfig
}

type FrameHttpClientSession struct {
	*FrameHttpClient
	ReqHandler func(req *http.Request)
}

func (c *FrameHttpClient) RegisterHttpService(name string, service FrameService) {
	c.serviceMap[name] = service
}

func (c *FrameHttpClient) Session() *FrameHttpClientSession {
	return &FrameHttpClientSession{
		c, nil,
	}
}
func (c *FrameHttpClientSession) Do(service string, method string) FrameService {
	frameService, ok := c.serviceMap[service]
	if !ok {
		panic(errors.New("service not found"))
	}
	//找到service里面的Field 给其中要调用的方法 赋值
	t := reflect.TypeOf(frameService)
	v := reflect.ValueOf(frameService)
	if t.Kind() != reflect.Pointer {
		panic(errors.New("service not pointer"))
	}
	tVar := t.Elem()
	vVar := v.Elem()
	fieldIndex := -1
	for i := 0; i < tVar.NumField(); i++ {
		name := tVar.Field(i).Name
		if name == method {
			fieldIndex = i
			break
		}
	}
	if fieldIndex == -1 {
		panic(errors.New("method not found"))
	}
	tag := tVar.Field(fieldIndex).Tag
	rpcInfo := tag.Get("Framerpc")
	if rpcInfo == "" {
		panic(errors.New("not Framerpc tag"))
	}
	split := strings.Split(rpcInfo, ",")
	if len(split) != 2 {
		panic(errors.New("tag Framerpc not valid"))
	}
	methodType := split[0]
	path := split[1]
	httpConfig := frameService.Env()

	f := func(args map[string]any) ([]byte, error) {
		if methodType == GET {
			return c.Get(httpConfig.Prefix()+path, args)
		}
		if methodType == POSTForm {
			return c.PostForm(httpConfig.Prefix()+path, args)
		}
		if methodType == POSTJson {
			return c.PostJson(httpConfig.Prefix()+path, args)
		}
		return nil, errors.New("no match method type")
	}
	fValue := reflect.ValueOf(f)
	vVar.Field(fieldIndex).Set(fValue)
	return frameService
}

func (c HttpConfig) Prefix() string {
	if c.Protocol == "" {
		c.Protocol = HTTP
	}
	switch c.Protocol {
	case HTTP:
		return fmt.Sprintf("http://%s:%d", c.Host, c.Port)
	case HTTPS:
		return fmt.Sprintf("https://%s:%d", c.Host, c.Port)
	}
	return ""

}
