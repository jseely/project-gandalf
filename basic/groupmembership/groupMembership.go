package groupmembership

import (
	"net/http"
)

func New() *groupMembership {
	return &groupMembership{}
}

func (g *groupMembership) GetGroupMembership(req *http.Request) (*http.Request, error) {
	req.Header.Add("Group-Membership", "AzureCATE2E")
	req.Header.Add("Group-Membership", "TestGroup")
	return req, nil
}

type groupMembership struct{}
