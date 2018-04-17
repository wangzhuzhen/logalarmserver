package tracing

import (
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"net/http"
	"context"
	"io/ioutil"
	"errors"
	"encoding/json"
)

//HTTPClient wraps an http.Client with tracing instrumentation.
type HTTPClient struct {
	Tracer	opentracing.Tracer
	Client	*http.Client
}

func (c *HTTPClient) GetJson(ctx context.Context, endpoint string, url string, out interface{}) error{
	req, err := http.NewRequest("GET", url, nil)
	if err != nil{
		return err
	}
	req = req.WithContext(ctx)
	req, ht := nethttp.TraceRequest(c.Tracer, req, nethttp.OperationName("HTTP GET: "+endpoint))
	defer ht.Finish()

	res, err := c.Client.Do(req)
	if err != nil{
		return err
	}
	defer res.Body.Close()

	if res.StatusCode >= 400{
		body, err := ioutil.ReadAll(res.Body)
		if err != nil{
			return err
		}
		return errors.New(string(body))
	}
	decoder := json.NewDecoder(res.Body)
	return decoder.Decode(out)
}

