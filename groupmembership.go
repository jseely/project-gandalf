package main

import (
	"net/http"
)

type GroupMembership interface {
	GetGroupMembership(req *http.Request) (*http.Request, error)
}
