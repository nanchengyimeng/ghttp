package ghttp

import (
	"net/http"
	"net/url"
	"strings"
)

type ContentTypeFunc func(request *http.Request)

// ICallBack 异步回调的接口
type ICallBack interface {
	ResponseCallback(IResponse)
}

// CheckRedirect request重定向的回调函数
type CheckRedirect func(req *http.Request, via []*http.Request) error

// TlsPath 证书文件地址
type TlsPath struct {
	//cert (pem) 路径
	CertFile string
	//key 路径
	KeyFile string
}

// GCookie 构造一个简单的HTTP请求cookie
func GCookie(simple map[string]string) []*http.Cookie {
	if len(simple) == 0 {
		return nil
	}
	cookies := make([]*http.Cookie, 0, len(simple))
	for k, v := range simple {
		cookies = append(cookies, &http.Cookie{
			Name:  k,
			Value: v,
		})
	}
	return cookies
}

// GGet 构造一个简单的GET请求协议
func GGet(strUrl string, values map[string]string) string {
	if strUrl == "" || values == nil {
		return strUrl
	}
	var buf strings.Builder
	buf.WriteString(strUrl)
	buf.WriteByte('?')
	i := 0
	for k, v := range values {
		if i > 0 {
			buf.WriteByte('&')
		}
		buf.WriteString(url.QueryEscape(k))
		buf.WriteByte('=')
		buf.WriteString(url.QueryEscape(v))
		i++
	}
	return buf.String()
}

// GPostData 构造一个简单的POST body
// 在form-data中，我没有允许一个值为数组的key，因为我觉得，这样好恶心，还不如直接调用json请求方式
func GPostData(values map[string]string) url.Values {
	if values == nil {
		return nil
	}
	value := make(url.Values, len(values))
	for k, v := range values {
		value.Add(k, v)
	}
	return value
}
