package ghttp

import (
	"context"
	"io/ioutil"
	"net/http"
)

type BuildResponse func(ctx context.Context, resp *http.Response, err error) (context.Context, IResponse)

type IResponse interface {
	//返回当前请求的错误
	Error() error

	//返回当前请求的http状态码
	StatusCode() int

	//返回当前请求的header信息
	Header() http.Header

	//返回HTTP内容长度
	ContentLength() int64

	//返回HTTP内容
	Content() []byte

	//返回HTTP包中response的信息
	Resp() *http.Response

	//返回这次请求的request信息
	Request() *http.Request

	//根据名称返回cookie值
	Cookie(name string) *http.Cookie
}

type HttpResponse struct {
	err             error
	ResponseContent []byte
	httpResp        *http.Response
}

func (h *HttpResponse) Error() error {
	return h.err
}

func (h *HttpResponse) StatusCode() int {
	if h.httpResp == nil {
		return 0
	}
	return h.httpResp.StatusCode
}

func (h *HttpResponse) Header() http.Header {
	if h.httpResp == nil {
		return nil
	}
	return h.httpResp.Header
}

func (h *HttpResponse) ContentLength() int64 {
	if h.httpResp == nil {
		return 0
	}
	return h.httpResp.ContentLength
}

func (h *HttpResponse) Content() []byte {
	return h.ResponseContent
}

func (h *HttpResponse) Resp() *http.Response {
	return h.httpResp
}

func (h *HttpResponse) Request() *http.Request {
	if h.httpResp == nil {
		return nil
	}
	return h.httpResp.Request
}

func (h *HttpResponse) Cookie(name string) *http.Cookie {
	for _, cookie := range h.httpResp.Cookies() {
		if cookie.Name == name {
			return cookie
		}
	}
	return nil
}

//默认的HTTP响应构造器
func DefaultBuildResponse(ctx context.Context, resp *http.Response, err error) (context.Context, IResponse) {
	iResponse := new(HttpResponse)
	if err != nil {
		iResponse.err = err
		return ctx, iResponse
	}

	iResponse.httpResp = resp
	responseContent, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		iResponse.err = err
		return ctx, iResponse
	}
	iResponse.ResponseContent = responseContent
	_ = resp.Body.Close()

	return ctx, iResponse
}
