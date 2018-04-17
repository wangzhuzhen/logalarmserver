package routers

import (
	"github.com/gorilla/mux"
	"net/http"
	"gitlab.com/wangzhuzhen/EFK_component/Build_Images/logalarmserver/logalarmserver/log"
)

/* API 请求的重定向 router */
func NewRouter() *mux.Router {

	router := mux.NewRouter().StrictSlash(true)

	for _, route := range routes {
		var httphandler http.Handler

		httphandler = route.HandlerFunc
		httphandler = log.Logger(httphandler, route.Name)
		httphandler = CorsHeader(httphandler)

		router.
		Methods("OPTIONS").
			Path(route.Pattern).
			Name("cors").
			Handler(CorsHeader(httphandler))

		router.
		Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(httphandler)

	}
	return router
}


func CorsHeader(inner http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Access-Control-Allow-Origin",  "*" )
		//w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Method","POST, GET, HEAD, PUT, OPTIONS, DELETE")
		w.Header().Set("Access-Control-Max-Age", "3600")

		//w.Header().Set("Access-Control-Allow-Headers","Origin, X-Requested-With, X-HTTP-Method-Override,accept-charset,accept-encoding , Content-Type, Accept, Cookie")
		w.Header().Set("Access-Control-Allow-Headers","x-requested-with,Authorization,X-Custom-Header,accept,Origin,No-Cache,If-Modified-Since,Pragma, Last-Modified, Cache-Control, Expires, Content-Type, X-E4M-With")
		w.Header().Set("Content-Type","json/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		inner.ServeHTTP(w, r)
	})
}
