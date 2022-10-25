package ghttp

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type client struct {
	//超时时间
	timeOut time.Duration

	//Header请求头信息
	header map[string]string

	//cookies信息
	cookies []*http.Cookie

	//golang内置client
	client *http.Client

	//临时header请求头, 仅本次请求生效
	_header map[string]string

	//临时cookie, 仅本次请求生效
	_cookies []*http.Cookie

	//处理HTTP响应的Response
	buildResponse BuildResponse

	//加载日志处理
	loggerWriter io.Writer

	// 日志开关
	openLogger bool

	// 请求唯一id
	uniqueId string
}

//临时header设置，仅本次请求生效
func (c *client) HeaderCache(header map[string]string) *client {
	c._header = header
	return c
}

//追加请求头，全生命周期有效
func (c *client) AddHeader(header map[string]string) {
	for k, v := range header {
		c.header[k] = v
	}
}

//重置请求头，全生命周期有效
func (c *client) SetHeader(header map[string]string) {
	c.header = header
}

//临时cookie设置，仅本次请求有效效
func (c *client) CookiesCache(cookies []*http.Cookie) *client {
	c._cookies = cookies
	return c
}

//追加cookie，全生命周期有效
func (c *client) AddCookies(cookies []*http.Cookie) {
	for _, cookie := range cookies {
		c.cookies = append(c.cookies, cookie)
	}
}

//重设cookie，全生命周期有效
func (c *client) SetCookies(cookies []*http.Cookie) {
	c.cookies = cookies
}

// 设置请求的uniqueId
func (c *client) SetUniqueId(id string) {
	c.uniqueId = id
}

//初始化一个request
func (c *client) getRequest(method, url string, body io.Reader) (*http.Request, error) {
	request, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	for k, v := range c.header {
		request.Header.Set(k, v)
	}
	for k, v := range c._header {
		request.Header.Set(k, v)
	}
	c._header = nil

	if _, ok := request.Header["User-Agent"]; !ok {
		request.Header.Set("User-Agent", HTTP_USER_AGENT_CHROME_PC)
	}

	for _, v := range c.cookies {
		request.AddCookie(v)
	}
	for _, v := range c._cookies {
		request.AddCookie(v)
	}
	c._cookies = nil

	return request, nil
}

//封装http请求
func (c *client) doRequest(ctx context.Context, r *http.Request) (context.Context, *http.Response, error) {
	// 记录请求开始时间
	ctx = c.buildStartTime(ctx)
	response, err := c.client.Do(r)
	return ctx, response, err
}

//POST请求中,处理request的函数
func setRequestPostFrom(r *http.Request) {
	r.Header.Set("Content-Type", HTTP_CONTENT_TYPE_FROM_DATA)
}

//POST请求中,处理request的函数,设置`Content-Type` 为 json
func setRequestPostJson(r *http.Request) {
	r.Header.Set("Content-Type", HTTP_CONTENT_TYPE_JSON)
}

//POST请求中,处理request的函数,设置`Content-Type` 为 xml
func setRequestPostXml(r *http.Request) {
	r.Header.Set("Content-Type", HTTP_CONTENT_TYPE_XML)
}

//常规发起http请求
func (c *client) sendWithMethod(ctx context.Context, method, url string, body io.Reader, setContentType ContentTypeFunc) IResponse {
	request, err := c.getRequest(method, url, body)
	if err != nil {
		return c.logger(c.buildResponse(nil, nil, err))
	}
	if setContentType != nil {
		setContentType(request)
	}

	return c.logger(c.buildResponse(c.doRequest(ctx, request)))
}

//发起异步回调处理的请求
func (c *client) sendWithMethodCallback(ctx context.Context, method, url string, body io.Reader, setContentType ContentTypeFunc, callback func(response IResponse)) error {
	request, err := c.getRequest(method, url, body)
	if err != nil {
		return err
	}
	if setContentType != nil {
		setContentType(request)
	}

	go func() {
		callback(c.logger(c.buildResponse(c.doRequest(ctx, request))))
	}()
	return nil
}

//通用GET请求，可先使用GGET绑定参数，再调用此方法
func (c *client) Get(url string) IResponse {
	return c.sendWithMethod(nil, http.MethodGet, url, nil, nil)
}

//GET异步请求，使用回调函数
func (c *client) GetAsync(url string, call func(response IResponse)) error {
	return c.sendWithMethodCallback(nil, http.MethodGet, url, nil, nil, call)
}

//GET异步请求，使用回调接口
func (c *client) GetAsyncWithCallback(url string, call ICallBack) error {
	return c.GetAsync(url, call.ResponseCallback)
}

//post 的form请求
func (c *client) PostForm(url string, values url.Values) IResponse {
	var reader io.Reader
	var ctx context.Context

	ctx = nil

	if values != nil {
		reader = strings.NewReader(values.Encode())
		ctx = c.buildContext(values.Encode())
	}

	return c.sendWithMethod(ctx, http.MethodPost, url, reader, setRequestPostFrom)
}

//Post form 异步请求,使用回调函数
func (c *client) PostFormAsyn(url string, values url.Values, call func(response IResponse)) error {
	if call == nil {
		return errors.New("callback function is nil")
	}

	if values == nil {
		return errors.New("values is nil")
	}

	reader := strings.NewReader(values.Encode())
	ctx := c.buildContext(values.Encode())
	return c.sendWithMethodCallback(ctx, http.MethodPost, url, reader, setRequestPostFrom, call)
}

//Post form 异步请求,使用接口回调
func (c *client) PostFormAsynWithCallback(url string, values url.Values, call ICallBack) error {
	return c.PostFormAsyn(url, values, call.ResponseCallback)
}

//post 的bytes请求
func (c *client) PostBytes(url string, value []byte, req func(request *http.Request)) IResponse {
	if value == nil {
		return c.logger(c.buildResponse(nil, nil, errors.New("PostBytes value is nil")))
	}
	reader := bytes.NewReader(value)
	ctx := c.buildContext(string(value))
	return c.sendWithMethod(ctx, http.MethodPost, url, reader, req)
}

//post 的bytes请求
func (c *client) PostBytesAsyn(url string, value []byte, req func(request *http.Request), call func(response IResponse)) error {
	if call == nil {
		return errors.New("callback function is nil")
	}

	if value == nil {
		return errors.New("value is nil")
	}

	reader := bytes.NewReader(value)
	ctx := c.buildContext(string(value))
	return c.sendWithMethodCallback(ctx, http.MethodPost, url, reader, req, call)
}

//post 的json请求
func (c *client) PostJson(url string, value interface{}) IResponse {
	if value == nil {
		return c.logger(c.buildResponse(nil, nil, errors.New("PostJson value is nil")))
	}
	by, err := json.Marshal(value)
	if err != nil {
		return c.logger(c.buildResponse(nil, nil, err))
	}
	return c.PostBytes(url, by, setRequestPostJson)
}

//Post json 异步请求,使用回调函数
func (c *client) PostJsonAsyn(url string, value interface{}, call func(response IResponse)) error {
	if call == nil {
		return errors.New("callback function is nil")
	}
	if value == nil {
		return errors.New("value is nil")
	}
	by, err := json.Marshal(value)
	if err != nil {
		return errors.New("value json encode error: " + err.Error())
	}
	return c.PostBytesAsyn(url, by, setRequestPostJson, call)
}

//Post json 异步请求,使用接口回调
func (c *client) PostJsonAsynWithCallback(url string, values interface{}, call ICallBack) error {
	return c.PostJsonAsyn(url, values, call.ResponseCallback)
}

//post 的xml请求
func (c *client) PostXml(url string, value interface{}) IResponse {
	if value == nil {
		return c.logger(c.buildResponse(nil, nil, errors.New("PostJson value is nil")))
	}
	by, err := xml.Marshal(value)
	if err != nil {
		return c.logger(c.buildResponse(nil, nil, err))
	}
	return c.PostBytes(url, by, setRequestPostXml)
}

//Post xml 异步请求,使用回调函数
func (c *client) PostXmlAsyn(url string, value interface{}, call func(response IResponse)) error {
	if call == nil {
		return errors.New("callback function is nil")
	}
	if value == nil {
		return errors.New("value is nil")
	}
	by, err := json.Marshal(value)
	if err != nil {
		return errors.New("value json encode error: " + err.Error())
	}
	return c.PostBytesAsyn(url, by, setRequestPostXml, call)
}

//Post xml 异步请求,使用接口回调
func (c *client) PostXmlAsynWithCallback(url string, values interface{}, call ICallBack) error {
	return c.PostXmlAsyn(url, values, call.ResponseCallback)
}

//post 的multipart请求
func (c *client) PostMultipart(url string, body IMultipart) IResponse {
	return c.sendWithMethod(nil, http.MethodPost, url, body, func(request *http.Request) {
		request.Header.Set("Content-Type", body.ContentType())
	})
}

//post 的multipart请求,使用回调函数
func (c *client) PostMultipartAsyn(url string, body IMultipart, call func(response IResponse)) error {
	if call == nil {
		return errors.New("callback function is nil")
	}
	return c.sendWithMethodCallback(nil, http.MethodPost, url, body, func(request *http.Request) {
		request.Header.Set("Content-Type", body.ContentType())
	}, call)
}

//post 的multipart请求,使用接口回调
func (c *client) PostMultipartAsynWithCallback(url string, body IMultipart, call ICallBack) error {
	return c.PostMultipartAsyn(url, body, call.ResponseCallback)
}

//设置请求上下文，用于日志记录
func (c *client) buildContext(body string) context.Context {
	ctx := context.Background()
	return context.WithValue(ctx, "body", body)
}

//设置请求时间到ctx
func (c *client) buildStartTime(ctx context.Context) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	startTime := time.Now()
	return context.WithValue(ctx, "startTime", startTime)
}

// log记录
func (c *client) logger(ctx context.Context, resp IResponse) IResponse {
	// 日志关闭不记录
	if !c.openLogger {
		return resp
	}

	logger := log.New(c.loggerWriter, "curl   ", log.LstdFlags)

	//统计请求耗时
	startTime := ctx.Value("startTime").(time.Time)
	costTime := time.Since(startTime).String()

	//请求错误处理
	if resp.Error() != nil {
		//远程服务异常时，不再记录日志，该异常交由调用方自行处理
		return resp
	}

	header, _ := json.Marshal(resp.Request().Header)
	if ctx == nil {
		logger.Printf("\t%s\t%s %s %s  header:%s\tparams:%s\tresponse:%s", costTime, c.uniqueId, resp.Request().Method, resp.Request().URL, string(header), "", string(resp.Content()))
		return resp
	}

	body := ctx.Value("body")
	var bodyStr string
	if body != nil {
		bodyStr = body.(string)
	}

	logger.Printf("\t%s\t%s %s %s  header:%s\tparams:%s\tresponse:%s", costTime, c.uniqueId, resp.Request().Method, resp.Request().URL, string(header), bodyStr, string(resp.Content()))
	return resp
}
