package ghttp

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"time"
)

// NewClientBuilder 初始化
func NewClientBuilder() *ClientBuilder {
	return &ClientBuilder{
		skipVerify:    true,
		openJar:       false,
		buildResponse: DefaultBuildResponse,
		loggerWriter:  os.Stdout,
		openLogger:    true,
	}
}

type ClientBuilder struct {

	//超时时间
	timeOut time.Duration

	//代理http url
	proxy string

	//tls私钥证书
	tlsPath []*TlsPath

	//cert root 证书
	certPool []string

	//是否跳过HTTPS证书校验(默认跳过)
	skipVerify bool

	//client发起HTTP请求时,header信息
	//默认会携带User-agent信息
	header map[string]string

	//client发起 HTTP请求时,自动携带的cookie
	cookie []*http.Cookie

	//重定向函数
	checkRedirect CheckRedirect

	//client是否开启cookieJar功能
	//默认不开启
	openJar bool

	//jarOptions 配置
	jarOptions *cookiejar.Options

	//处理HTTP的 response的回调函数
	//默认使用 `response.go`中的 `BuildResponse` 函数
	buildResponse BuildResponse

	//日志写入io, 默认stdout
	loggerWriter io.Writer

	// 日志开关
	openLogger bool
}

func (builder *ClientBuilder) SetTimeOut(t time.Duration) *ClientBuilder {
	builder.timeOut = t
	return builder
}

func (builder *ClientBuilder) SetProxyUrl(u string) *ClientBuilder {
	builder.proxy = u
	return builder
}

func (builder *ClientBuilder) SetTls(tlsPath []*TlsPath) *ClientBuilder {
	builder.tlsPath = tlsPath
	return builder
}

func (builder *ClientBuilder) SetCert(cert []string) *ClientBuilder {
	builder.certPool = cert
	return builder
}

func (builder *ClientBuilder) SkipVerify(skip bool) *ClientBuilder {
	builder.skipVerify = skip
	return builder
}

func (builder *ClientBuilder) SetCookie(cookie []*http.Cookie) *ClientBuilder {
	builder.cookie = cookie
	return builder
}

func (builder *ClientBuilder) CheckRedirect(checkRedirect CheckRedirect) *ClientBuilder {
	builder.checkRedirect = checkRedirect
	return builder
}

func (builder *ClientBuilder) SetHeader(header map[string]string) *ClientBuilder {
	builder.header = header
	return builder
}

func (builder *ClientBuilder) Jar(options *cookiejar.Options) *ClientBuilder {
	builder.openJar = true
	builder.jarOptions = options
	return builder
}

func (builder *ClientBuilder) BuildResponse(build BuildResponse) *ClientBuilder {
	builder.buildResponse = build
	return builder
}

func (builder *ClientBuilder) SetLoggerWriter(writer io.Writer) *ClientBuilder {
	builder.loggerWriter = writer
	return builder
}

func (builder *ClientBuilder) SetLoggerOpen(open bool) *ClientBuilder {
	builder.openLogger = open
	return builder
}

// Build 构造 client
func (builder *ClientBuilder) Build() (*client, error) {
	var (
		err         error
		proxy       *url.URL
		x509KeyPair tls.Certificate
	)

	if builder.buildResponse == nil {
		return nil, errors.New("clint not set BuildResponse")
	}

	if builder.proxy != "" {
		proxy, err = url.Parse(builder.proxy)
		if err != nil {
			return nil, err
		}
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: builder.skipVerify,
	}
	if builder.tlsPath != nil {
		certificates := make([]tls.Certificate, len(builder.tlsPath))
		for i, path := range builder.tlsPath {
			x509KeyPair, err = tls.LoadX509KeyPair(path.CertFile, path.KeyFile)
			if err != nil {
				return nil, err
			}
			certificates[i] = x509KeyPair
		}
		tlsConfig.Certificates = certificates
	}

	if builder.certPool != nil {
		tlsConfig.RootCAs = x509.NewCertPool()
		for _, certFile := range builder.certPool {
			if ca, err := ioutil.ReadFile(certFile); err != nil {
				return nil, err
			} else {
				if ok := tlsConfig.RootCAs.AppendCertsFromPEM(ca); !ok {
					return nil, fmt.Errorf("load:%s cert fail", certFile)
				}
			}
		}
	}
	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	if proxy != nil {
		transport.Proxy = http.ProxyURL(proxy)
	}
	c := &client{
		client: &http.Client{
			Transport:     transport,
			Timeout:       builder.timeOut,
			CheckRedirect: builder.checkRedirect,
		},
		header:        builder.header,
		cookies:       builder.cookie,
		buildResponse: builder.buildResponse,
		loggerWriter:  builder.loggerWriter,
		openLogger:    builder.openLogger,
	}

	if builder.openJar {
		jar, err := cookiejar.New(builder.jarOptions)
		if err != nil {
			return nil, err
		}
		c.client.Jar = jar
	}
	return c, nil
}
