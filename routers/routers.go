package routers

import (
	"github.com/gorilla/mux"
	"net/http"
	"github.com/wangzhuzhen/logalarmserver/log"
)

/* API 请求的重定向 router */
func NewRouter() *mux.Router {

	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {
		var httphandler http.Handler

		httphandler = route.HandlerFunc
		httphandler = log.Logger(httphandler, route.Name)

		router.
		Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(httphandler)
	}

	return router
}