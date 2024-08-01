//
// Copyright 2022 SkyAPM org
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package iris

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/SkyAPM/go2sky"
	h "github.com/SkyAPM/go2sky/plugins/http"
	"github.com/SkyAPM/go2sky/reporter"
	"github.com/kataras/iris"
)

func ExampleMiddleware() {
	// Use gRPC reporter for production
	re, err := reporter.NewLogReporter()
	if err != nil {
		log.Fatalf("new reporter error %v \n", err)
	}
	defer re.Close()

	tracer, err := go2sky.NewTracer("iris-server", go2sky.WithReporter(re))
	if err != nil {
		log.Fatalf("create tracer error %v \n", err)
	}

	// 创建一个新的 Iris 应用实例
	app := iris.New()

	//Use go2sky middleware with tracing
	app.Use(Middleware(tracer))

	app.Get("/Hello", func(ctx iris.Context) {
		// 写入响应内容
		ctx.Writef("Hello, Iris!")
	})

	go func() {
		// 这里使用 8080 端口，你可以根据需要更改端口号
		app.Run(iris.Addr(":8080"))
	}()
	// Wait for the server to start
	time.Sleep(time.Second)

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		request(tracer)
	}()
	wg.Wait()
	// Output:
}

func request(tracer *go2sky.Tracer, _ ...h.ClientOption) {
	//NewClient returns an HTTP Client with tracer
	client, err := h.NewClient(tracer)
	if err != nil {
		log.Fatalf("create client error %v \n", err)
	}

	request, err := http.NewRequest("GET", fmt.Sprintf("%s/Hello", "http://127.0.0.1:8080"), nil)
	if err != nil {
		log.Fatalf("unable to create http request: %+v\n", err)
	}

	res, err := client.Do(request)
	if err != nil {
		log.Fatalf("unable to do http request: %+v\n", err)
	}

	_ = res.Body.Close()
}
