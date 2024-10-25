package models

import "sync"

type Links struct {
	M     sync.RWMutex
	Links map[string]string
}

func NewLinks() *Links {
	return &Links{
		M:     sync.RWMutex{},
		Links: make(map[string]string),
	}
}
