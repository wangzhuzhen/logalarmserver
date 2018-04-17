package tracing

import (
	"net/http"
	"github.com/opentracing/opentracing-go"
	"gitlab.com/wangzhuzhen/EFK_component/Build_Images/logalarmserver/logalarmserver/pkg/DHTracer/log"
	"go.uber.org/zap"
)

type DicHandleRequests struct{
	HandleRequest	func(http.ResponseWriter, *http.Request)
}

type WebMuxServer struct {
	Logger 				log.Factory
	Tracer				opentracing.Tracer
	handleRequests		map[string]DicHandleRequests
}

type DHTracer struct{
	Logger			log.Factory
	Tracer			opentracing.Tracer
	ActiveSpan      opentracing.Span
}

func RunWebMuxServer(serverName string, methods map[string]DicHandleRequests) (http.Handler, *WebMuxServer){
	zLogger, _ := zap.NewDevelopment()
	logger := log.NewFactory(zLogger.With(zap.String("service", serverName)))
	tracer := InitWithServiceName(serverName)
	server := &WebMuxServer{
		Tracer:	tracer ,
		Logger: logger,
	}
	sliceMethods := make([]string, len(methods))
	sliceRequests := make(map[string]DicHandleRequests)
	//for i := 0; i < len(methods); i++{
	//	sliceMethods[i] = methods[i]
	//}
	index := 0
	for k,v := range methods{
		sliceMethods[index] = k
		sliceRequests[k] = v
		index = index + 1
	}

	handle := server.createServeMux(sliceMethods, sliceRequests)
	return handle, server
}

func CreateDHTracer(serverName string) *DHTracer{
	zLogger, _ := zap.NewDevelopment()
	logger := log.NewFactory(zLogger.With(zap.String("service", serverName)))
	tracer := InitWithServiceName(serverName)
	tracerServer := &DHTracer{
		Tracer:	tracer ,
		Logger: logger,
	}
	//tracerServer.createServeMux()
	return tracerServer
}

func (s *DHTracer) CreateServeMux(methods map[string]DicHandleRequests) http.Handler{

	sliceMethods := make([]string, len(methods))
	sliceRequests := make(map[string]DicHandleRequests)
	//for i := 0; i < len(methods); i++{
	//	sliceMethods[i] = methods[i]
	//}
	index := 0
	for k,v := range methods{
		sliceMethods[index] = k
		sliceRequests[k] = v
		index = index + 1
	}

	handle := NewServeMux(s.Tracer)
	for k, v := range sliceRequests{
		handle.Handle(k, http.HandlerFunc(v.HandleRequest))
	}

	return handle
}

func (s *WebMuxServer) createServeMux(methods []string, requests map[string]DicHandleRequests) http.Handler{
	handle := NewServeMux(s.Tracer)

	//for _, method := range methods{
	//	handle.Handle(method, http.HandlerFunc(s.HandleRequset))
	//}

	for k, v := range requests{
		handle.Handle(k, http.HandlerFunc(v.HandleRequest))
	}

	return handle
}


