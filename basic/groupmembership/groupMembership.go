package groupmembership

import (
	"net/http"
)

func New() *groupMembership {
	return &groupMembership{}
}

func (g *groupMembership) GetGroupMembership(w http.ResponseWriter, req *http.Request) (http.Header, error) {
	header := http.Header{}
	header.Add("Group-Membership", "AzureCATE2E")
	header.Add("Group-Membership", "TestGroup")
	return header, nil
}

type groupMembership struct{}
