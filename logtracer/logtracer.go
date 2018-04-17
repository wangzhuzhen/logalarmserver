package logtracer

import (
	"gitlab.com/wangzhuzhen/EFK_component/Build_Images/logalarmserver/logalarmserver/pkg/DHTracer/tracing"
	"net/http"
	"context"
	"github.com/opentracing/opentracing-go"
)

// 一般是API的入口处调用，作为 Tracer 的入口
// 对应产生父 Tracer、父Span
// spanName 作为span的名称，可以用服务名称或者自定义的名称
// mesg 作为到对应入口处的添加日志信息
func TracerAndSpan_Entry(spanName string, mesg string, r *http.Request, Tracer *tracing.DHTracer) (*tracing.DHTracer, context.Context) {

	var tagOptions map[string]interface{}
	tagOptions = make(map[string]interface{})

	//tagOptions["Url"]=r.URL.Path
	//tagOptions["Method"]=r.Method
	//tagOptions["ServiceName"]="logalarmserver"
	tagOptions["URL.Path"]=r.URL.Path
	tagOptions["Host"]=r.Host
	tagOptions["RequestURI"]=r.RequestURI
	tagOptions["RemoteAddr"]=r.RemoteAddr
	tagOptions["Method"]=r.Method
	tagOptions["ServiceName"]="logalarmserver"

	var logOptions map[string]interface{}
	logOptions = make(map[string]interface{})
	logOptions["Info"]= mesg

	Tracer,ctx := tracing.CreateTraceWithContext(spanName,tagOptions,logOptions,r,Tracer)


	// 以下是用于出本次Handler 调用另一个 Handler 的时候用
	//request,_ := http.NewRequest("GET","http://localhost:8989/",nil)
	//tracing.SendTraceResponse(ctx,Tracer,request)
	return Tracer,ctx
}

// 一般是链路最底层API的出口处调用，作为 API 调用的出口 Tracer
// 对应产生子 Tracer、子Span
// spanName 作为span的名称，可以用服务名称或者自定义的名称
// mesg 作为到对应入口处的添加日志信息
// unusual 作为出口标志，显示 API 是否是按期望的正常路径出口返回
func TracerAndSpan_Leave(spanName string, mesg string, unusual bool, r *http.Request, Tracer  *tracing.DHTracer)  *tracing.DHTracer {

	var tagOptions map[string]interface{}
	tagOptions = make(map[string]interface{})

	tagOptions["URL.Path"]=r.URL.Path
	tagOptions["Host"]=r.Host
	tagOptions["RequestURI"]=r.RequestURI
	tagOptions["RemoteAddr"]=r.RemoteAddr
	tagOptions["Method"]=r.Method
	tagOptions["ServiceName"]="logalarmserver"
	tagOptions["error"]=unusual

	var logOptions map[string]interface{}
	logOptions = make(map[string]interface{})
	logOptions["Info"]= mesg

	tracer := tracing.CreateChildSpanInTrace(spanName,tagOptions,logOptions,Tracer)
	return tracer
}


func GetLogTracerID(tracer opentracing.Tracer, span opentracing.Span) string{
	tracerId := tracing.GetTraceID(tracer, span)
	if tracerId == "" {
		return "Get_null"
	}

	return tracerId
}