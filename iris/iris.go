package iris

import (
	"fmt"
	"github.com/SkyAPM/go2sky"
	"github.com/kataras/iris"
	"github.com/kataras/iris/context"
	agentv3 "skywalking.apache.org/repo/goapi/collect/language/agent/v3"
	"strconv"
)

const componentIDGINHttpServer = 5006

//Middleware iris middleware return HandlerFunc  with tracing.
func Middleware(tracer *go2sky.Tracer) context.Handler {
	if tracer == nil {
		return func(c iris.Context) {
			c.Next()
		}
	}

	return func(c iris.Context) {
		span, ctx, err := tracer.CreateEntrySpan(c.Request().Context(), getOperationName(c), func(key string) (string, error) {
			return c.Request().Header.Get(key), nil
		})
		if err != nil {
			c.Next()
			return
		}
		span.SetComponent(componentIDGINHttpServer)
		span.Tag(go2sky.TagHTTPMethod, c.Request().Method)
		span.Tag(go2sky.TagURL, c.Request().Host+c.Request().URL.Path)
		span.SetSpanLayer(agentv3.SpanLayer_Http)

		c.Request().WithContext(ctx)
		c.Next()

		span.Tag(go2sky.TagStatusCode, strconv.Itoa(c.ResponseWriter().StatusCode()))
		span.End()
	}
}

func getOperationName(c iris.Context) string {
	return fmt.Sprintf("/%s%s", c.Request().Method, c.Path())
}
