package main

import (
	"net/http"
	"log"
	"github.com/wangzhuzhen/logalarmserver/routers"
)

func main() {
	println("Starting Logalarm Server")

	router := routers.NewRouter()
	log.Fatal(http.ListenAndServe(":8080", router))
}


