package tracing

import (
	"github.com/uber/jaeger-lib/metrics"
	"github.com/opentracing/opentracing-go"
	"gitlab.com/wangzhuzhen/EFK_component/Build_Images/logalarmserver/logalarmserver/pkg/DHTracer/log"
	"time"
	"fmt"
	"github.com/uber/jaeger-client-go/rpcmetrics"

	"go.uber.org/zap"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"

	"context"
	"net/http"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"io/ioutil"
)

type jaegerLoggerAdapter struct {
	logger log.Logger
}

func (l jaegerLoggerAdapter) Error(msg string) {
	l.logger.Error(msg)
}

func (l jaegerLoggerAdapter) Infof(msg string, args ...interface{}) {
	l.logger.Info(fmt.Sprintf(msg, args...))
}


// Init creates a new instance of Jaeger tracer.
func Init(serviceName string, metricsFactory metrics.Factory, logger log.Factory) opentracing.Tracer{
	cfg := config.Configuration{
		Sampler:&config.SamplerConfig{ //采样配置
			Type: "const",	//采样类型为固定间隔时间采样
			Param: 1,
		},
		Reporter:&config.ReporterConfig{ //提交到代理配置
			LogSpans: false,
			BufferFlushInterval: 1 * time.Second,
			LocalAgentHostPort:AgentUDPAddr,
		},
	}
	tracer, _, err := cfg.New(
		serviceName,
		config.Logger(jaegerLoggerAdapter{logger.Bg()}),
		config.Observer(rpcmetrics.NewObserver(metricsFactory, rpcmetrics.DefaultNameNormalizer)),
	)
	if err != nil{
		logger.Bg().Fatal("cannot initialize Jaeger Tracer", zap.Error(err))
	}
	return tracer
}

func InitWithServiceName(serviceName string) opentracing.Tracer{
	cfg := config.Configuration{
		Disabled:false,
		Sampler:&config.SamplerConfig{//采样配置
			Type: "const",//采样类型为固定间隔时间采样
			Param: 1,
		},
		Reporter:&config.ReporterConfig{//提交到代理配置
			LogSpans: false,
			BufferFlushInterval: 1 * time.Second,
			LocalAgentHostPort:AgentUDPAddr,
		},
	}
	tracer, _, err := cfg.New(
		serviceName,
	)
	if err != nil{
		fmt.Println("cannot initialize Jaeger Tracer")
	}
	return tracer
}

//根据上下文创建Trace和Span
func CreateTraceWithContext(serviceName string, tagOptions map[string]interface{}, logOptions map[string]interface{}, r *http.Request, parentTrace *DHTracer) (*DHTracer, context.Context){
	ctx := r.Context()
	parentSpan := opentracing.SpanFromContext(ctx)
	m_parentTrace := parentTrace
	var span opentracing.Span
	if parentTrace == nil {
		traceServer := CreateDHTracer(serviceName)
		span = traceServer.Tracer.StartSpan(serviceName)
		traceServer.ActiveSpan = span
		m_parentTrace = traceServer
	}else {
		if parentSpan != nil{
			parentCtx, _ := parentTrace.Tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
			span = parentTrace.Tracer.StartSpan(serviceName, opentracing.ChildOf(parentCtx))
			m_parentTrace.ActiveSpan = span
		} else {
			traceServer := CreateDHTracer(serviceName)
			span = traceServer.Tracer.StartSpan(serviceName)
			traceServer.ActiveSpan = span
			m_parentTrace = traceServer
		}
	}

	for k, v := range tagOptions {
		span.SetTag(k, v)
	}

	for k, v := range logOptions {
		span.LogKV(k, v)
	}

	ext.SpanKindRPCClient.Set(span)
	defer  span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)
	return m_parentTrace, ctx
}

//用于创建父子关系Span，无需上下文
func CreateChildSpanInTrace(serviceName string, tagOptions map[string]interface{}, logOptions map[string]interface{}, parentTrace *DHTracer) *DHTracer{
	m_parentTrace := parentTrace

	var span opentracing.Span
	if parentTrace == nil {
		traceServer := CreateDHTracer(serviceName)
		span = traceServer.Tracer.StartSpan(serviceName)
		traceServer.ActiveSpan = span
		m_parentTrace = traceServer
	}else{
		span = parentTrace.Tracer.StartSpan(serviceName, opentracing.ChildOf(m_parentTrace.ActiveSpan.Context()))
		m_parentTrace.ActiveSpan = span
	}

	for k, v := range tagOptions {
		span.SetTag(k, v)
	}

	for k, v := range logOptions {
		span.LogKV(k, v)
	}

	ext.SpanKindRPCClient.Set(span)
	defer  span.Finish()

	return m_parentTrace
}

// 用于新的请求，将上层带有tracer信息的ctx传递下去
func SendTraceResponse(ctx context.Context, dhTrace *DHTracer, r *http.Request){
	clientC := &http.Client{Transport:&nethttp.Transport{}}
	r = r.WithContext(ctx)
	//r.Header.Set("Connection", "keep-alive")

	// 将tracer包裹在request中
	r, ht := nethttp.TraceRequest(dhTrace.Tracer, r)
	defer ht.Finish()
	response, _ := clientC.Do(r)
	defer response.Body.Close()
	if response.StatusCode == 200 {
		body, _ := ioutil.ReadAll(response.Body)
		fmt.Println(string(body))
	}else{
		m_tracer, _ := dhTrace.Tracer.(*jaeger.Tracer)
		var tagOptions map[string]interface{}
		tagOptions = make(map[string]interface{})
		tagOptions["error"]=true
		tagOptions["parent"]=m_tracer.GetServiceName()


		var logOptions map[string]interface{}
		logOptions = make(map[string]interface{})
		logOptions["errorCode"]=response.StatusCode
		serviceName := fmt.Sprintf("%s_%s", m_tracer.GetServiceName(), "Response")
		CreateTraceWithContext(serviceName, tagOptions, logOptions, r, dhTrace)
	}
}

func GetTraceID(tracer opentracing.Tracer, span opentracing.Span) string{
	_, ok := tracer.(*opentracing.NoopTracer)
	if ok {
		return ""
	}
	_, ok = tracer.(*jaeger.Tracer)
	if ok {
		return span.Context().(jaeger.SpanContext).TraceID().String()
	}
	return ""
}