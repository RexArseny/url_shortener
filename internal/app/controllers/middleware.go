package controllers

import (
	"log"
	"net/http"
	"runtime/debug"
)

func RecoverMiddleware(handler func(res http.ResponseWriter, req *http.Request)) func(res http.ResponseWriter, req *http.Request) {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic: %s; stacktrace: %s", err, debug.Stack())
			}
		}()
		handler(res, req)
	})
}
