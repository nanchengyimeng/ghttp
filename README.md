### 这是一个二次开发的项目，它用来发起curl

>不知道怎么说~~~

---

### 以下是对构造器和client客户端的说明，想直接上手的小可爱，直接看最后

------

### 构造器的用法

> 构造器用来生成一个client，此外，构造器在生成client时，允许设置一些常规参数。我来举个栗子 ^-^。

> 构造器对应的设置，针对生成的client时始终有效的

1. 首先你需要得到一个构造器，然后才可以设置基准参数

```go
builder := ghttp.NewClientBuilder()
```

> 这个构造器默认开启的日志并输出到控制台，默认关闭https证书验证，使用默认的响应处理器

> 构造器提供的方法均为可选

```go
//初始化 ClientBuilder
func NewClientBuilder() *ClientBuilder {
	return &ClientBuilder{
		skipVerify:    true,                    //  跳过证书验证
		openJar:       false,                   //  关闭cookiejar
		buildResponse: DefaultBuildResponse,    //  使用默认的响应封装
		loggerWriter:  os.Stdout,               //  日志默认输出至控制台
		openLogger:    true,                    //  默认启用日志输出
	}
}
``` 

2. 设置请求超时时间（可选）

```go
//设置超时时间为10s
builder.SetTimeOut(10 * time.Second)
```

3. 设置请求代理 ()

```go
//设置请求代理 
builder.SetProxyUrl("代理服务器地址")
```

4. 获取一个client

> Build方法返回一个client，client的描述看 Client的用法

> 如果存在以上可选设置，如超时时间等，会在build时设置到client中，b

```go
builder.Build()
```
### Client的用法


-------

## 尝试一下

#### Get请求

1. 常规GET请求

> 常规请求是同步的

```go
package main

import (
	"fmt"
	"github.com/nanchengyimeng/ghttp"
	"log"
)

func main() {
	//请求地址
	requestUrl := "http://test.com/debug/tt/tt"

	//设置入参
	params := make(map[string]string)
	params["name"] = "tt"

	//获取一个client构造器
	builder := ghttp.NewClientBuilder()

	//构造一个client
	client, err := builder.Build()
	if err != nil {
		log.Fatalln(err.Error())
	}

	//使用GGet方法，设置Get请求的参数
	requestUrl = ghttp.GGet(requestUrl, params)
	iResponse := client.Get(requestUrl)

	//获取响应的错误信息、状态码、内容、内容长度，此外iResponse还封装了cookie信息、header信息、响应的Request、响应的Response
	fmt.Println(iResponse.Error(), iResponse.StatusCode(), iResponse.Content(), iResponse.ContentLength())
}
```

#### PostForm请求

1. 异步匿名函数请求

> 这种请求方式需要主程使用waitGroup，确保携程执行完毕

```go
package main

import (
	"fmt"
	"github.com/nanchengyimeng/ghttp"
	"log"
	"sync"
)

func main() {
	//请求地址
	requestUrl := "http://test.com/debug/tt/tt"

	//设置入参
	params := make(map[string]string)
	params["name"] = "tt"

	//获取一个client构造器
	builder := ghttp.NewClientBuilder()

	//构造一个client
	client, err := builder.Build()
	if err != nil {
		log.Fatalln(err.Error())
	}

	//使用GPostData方法，设置PostForm请求的参数
	data := ghttp.GPostData(params)

	//使用Wg确保异步函数正确结束
	var wg sync.WaitGroup

	//异步的post请求
	wg.Add(1)
	err = client.PostFormAsyn(requestUrl, data, func(response ghttp.IResponse) {
		defer wg.Done()
		fmt.Println(response.StatusCode())
	})
	if err != nil {
		panic(err)
	}

	//循环是为了确保异步函数执行完毕，可
	wg.Wait()
}
```


2. 异步interface请求

> 有时候我们需要发起一个很复杂的请求，例如限制并发请求数，我们可以用对应请求类型的异步回调

> 异步回调允许你自己控制响应的结果

> 异步回调方法需要一个实现了ghttp.ICallBack接口的结构体

```go
package main

import (
	"github.com/nanchengyimeng/ghttp"
	"sync"
)

// MyResponse 实现了ghttp.ICallBack的回调接口，我们可以通过该struct，处理更复杂的逻辑
type MyResponse struct {
	ErrChan    chan error       //接收错误信息
	ResultChan chan []byte      //接收请求结果
	MaxRequest chan interface{} //限制最大并发请求数
	TWg        sync.WaitGroup
}

// ResponseCallback http响应的具体处理函数
func (t *MyResponse) ResponseCallback(response ghttp.IResponse) {
	//程序结束时，告知waitGroup结束
	defer t.TWg.Done()
	//程序结束时，告知RequestWait结束
	defer t.RequestDone()
	if response.Error() != nil {
		t.ErrChan <- response.Error()
		return
	}

	t.ResultChan <- response.Content()
}

// RequestWait 阻塞等待当前请求
func (t *MyResponse) RequestWait() {
	t.MaxRequest <- struct{}{}
}

// RequestDone 释放当前请求的占用
func (t *MyResponse) RequestDone() {
	<-t.MaxRequest
}

func main() {
	requestUrl := "http://test.com/debug/tt/tt"
	params := make(map[string]string)
	params["name"] = "tt"

	//获取一个client构造器
	builder := ghttp.NewClientBuilder()

	//获得一个client
	client, err := builder.Build()
	if err != nil {
		panic(err)
	}

	//拼接query参数
	getUrl := ghttp.GGet(requestUrl, params)

	//初始化一个自定义响应接收者
	tt := &MyResponse{
		ErrChan:    make(chan error, 1),
		ResultChan: make(chan []byte, 1000),
		MaxRequest: make(chan interface{}, 3),  //设置并行的请求数为3
	}

	//发起一百次异步请求
	for i := 0; i < 100; i++ {
		tt.TWg.Add(1)

		//控制并行的请求，等待其他请求执行完毕
		tt.RequestWait()

		err = client.GetAsyncWithCallback(getUrl, tt)
		if err != nil {
			panic(err.Error())
		}
	}

	tt.TWg.Wait()

	close(tt.ErrChan)
	close(tt.ResultChan)
	close(tt.MaxRequest)
}

```