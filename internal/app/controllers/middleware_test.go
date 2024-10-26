package controllers

import (
	"net/http"
	"testing"
)

func TestRecoverMiddleware(t *testing.T) {
	test1 := func(res http.ResponseWriter, req *http.Request) {
		panic("test")
	}

	test2 := func(res http.ResponseWriter, req *http.Request) {
	}

	RecoverMiddleware(test1)
	RecoverMiddleware(test2)
}
