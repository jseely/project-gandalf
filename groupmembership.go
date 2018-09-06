package main

import (
	"net/http"
)

type GroupMembership interface {
	GetGroupMembership(w http.ResponseWriter, req *http.Request) (http.Header, error)
}
