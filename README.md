- 小白练手所写

- GET
```go
package test

import (
	"fmt"
	"github.com/nanchengyimeng/ghttp"
)

func Get() {
	builder := ghttp.NewClientBuilder()
	client, err := builder.Build()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	requestUrl := "http://test.com/debug/tt/tt"
	params := make(map[string]string)
	params["name"] = "test"
	params["age"] = "18"
	requestUrl = ghttp.GGet(requestUrl, params)
	response := client.Get(requestUrl)

	fmt.Println(response.StatusCode())
	fmt.Println(response.Error())
	fmt.Println(string(response.Content()))
}
```

- GET
```go
package test

import (
	"fmt"
	"github.com/nanchengyimeng/ghttp"
)

func Get() {
	builder := ghttp.NewClientBuilder()
	client, err := builder.Build()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	requestUrl := "http://test.com/debug/tt/tt"
	params := make(map[string]string)
	params["name"] = "test"
	params["age"] = "18"
	requestUrl = ghttp.GGet(requestUrl, params)
	response := client.Get(requestUrl)

	fmt.Println(response.StatusCode())
	fmt.Println(response.Error())
	fmt.Println(string(response.Content()))
}
```

- 异步get请求示例

```go
package main

import (
	"fmt"
	"github.com/nanchengyimeng/ghttp"
	"sync"
)

func main() {
	requestUrl := "http://test.com/debug/tt/tt"
	params := make(map[string]string)
	params["name"] = "lgz"
	params["age"] = "18"

	builder := ghttp.NewClientBuilder()

	client, err := builder.Build()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	getUrl := ghttp.GGet(requestUrl, params)

	tt := &Test{
		ErrChan: make(chan error, 1),
		ResultChan: make(chan []byte, 1000),
	}

	//发起十次异步请求
	for i := 0; i < 10; i++ {
		tt.TWg.Add(1)
		err = client.GetAsyncWithCallback(getUrl, tt)
		if err != nil {
			panic(err.Error())
		}
	}

	fmt.Println("等待响应中")
	tt.TWg.Wait()
	fmt.Println("响应等待完毕")

	for data := range tt.ResultChan {
		fmt.Println(string(data))
	}
	
}

type Test struct{
	ErrChan chan error
	ResultChan chan []byte

	TWg sync.WaitGroup
}

func (t *Test) ResponseCallback(response ghttp.IResponse) {
	defer t.TWg.Done()
	if response.Error() != nil {
		t.ErrChan<- response.Error()
		return
	}

	t.ResultChan<- response.Content()
}

```