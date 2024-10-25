package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/RexArseny/url_shortener/internal/app/args"
	"github.com/RexArseny/url_shortener/internal/app/controllers"
	"github.com/RexArseny/url_shortener/internal/app/usecases"
)

func main() {
	interactor := usecases.NewInteractor()
	ctrl := controllers.NewController(interactor)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /", controllers.RecoverMiddleware(ctrl.CreateShortLink))
	mux.HandleFunc(fmt.Sprintf("GET /{%s}", controllers.ID), controllers.RecoverMiddleware(ctrl.GetShortLink))

	log.Printf("start service on %s:%d", args.DefaultDomain, args.DefaultPort)

	err := http.ListenAndServe(fmt.Sprintf("%s:%d", args.DefaultDomain, args.DefaultPort), mux)
	if err != nil {
		panic(err)
	}
}
