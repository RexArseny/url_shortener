package controllers

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/RexArseny/url_shortener/internal/app/usecases"
)

const Id = "id"

type Controller struct {
	interactor usecases.Interactor
}

func NewController(interactor usecases.Interactor) Controller {
	return Controller{
		interactor: interactor,
	}
}

func (c *Controller) CreateShortLink(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		log.Printf("incorrect method %s; should be POST; request: %v", req.Method, req)
		c.error(res, fmt.Errorf("method should be POST and not %s", req.Method))
		return
	}

	data, err := io.ReadAll(req.Body)
	if err != nil {
		log.Printf("can not read body %v; request: %v", req.Body, req)
		c.error(res, fmt.Errorf("service error"))
		return
	}

	result, err := c.interactor.CreateShortLink(string(data))
	if err != nil {
		log.Printf("can not create short link %s; request: %v", err, req)
		c.error(res, fmt.Errorf("service error"))
		return
	}

	if result == nil || *result == "" {
		log.Printf("short link is empty; request: %v", req)
		c.error(res, fmt.Errorf("service error"))
		return
	}

	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusCreated)
	res.Write([]byte(*result))
}

func (c *Controller) GetShortLink(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		log.Printf("incorrect method %s; should be GET; request: %v", req.Method, req)
		c.error(res, fmt.Errorf("method should be GET and not %s", req.Method))
		return
	}

	data := req.PathValue(Id)

	result, err := c.interactor.GetShortLink(data)
	if err != nil {
		log.Printf("can not get short link %s; request: %v", err, req)
		c.error(res, fmt.Errorf("service error"))
		return
	}

	if result == nil || *result == "" {
		log.Printf("short link is empty; request: %v", req)
		c.error(res, fmt.Errorf("service error"))
		return
	}

	res.Header().Set("Location", *result)
	res.WriteHeader(http.StatusTemporaryRedirect)
}

func (c *Controller) error(res http.ResponseWriter, err error) {
	res.WriteHeader(http.StatusBadRequest)
	res.Write([]byte(err.Error()))
}
