package groupmembership

import (
	"net/http"
	"strings"
)

func New() *groupMembership {
	return &groupMembership{}
}

var (
	cate2eusers = []string{
		"jseely@ntdev.microsoft.com",
		"amoor@microsoft.com",
		"aconrad@microsoft.com",
		"bswan@microsoft.com",
		"chrsims@microsoft.com",
		"dpoole@microsoft.com",
		"elwort@microsoft.com",
		"eptolla@microsoft.com",
		"hastlife@microsoft.com",
		"jdh@microsoft.com",
		"joal@microsoft.com",
		"johnrag@microsoft.com",
		"kedomico@microsoft.com",
		"mapraka@microsoft.com",
		"mpae@microsoft.com",
		"nikman@microsoft.com",
		"rochak@microsoft.com",
		"samcoop@microsoft.com",
		"stschn@microsoft.com",
		"suskuma@microsoft.com",
		"tudhadiw@microsoft.com",
	}
)

func (g *groupMembership) GetGroupMembership(w http.ResponseWriter, req *http.Request) (http.Header, error) {
	header := http.Header{}
	if strings.HasSuffix(req.Header.Get("User-Principal-Name"), "microsoft.com") {
		header.Add("Group-Membership", "Microsoft")
	}
	for _, user := range cate2eusers {
		if user == req.Header.Get("User-Principal-Name") {
			header.Add("Group-Membership", "AzureCATE2E")
		}
	}
	return header, nil
}

type groupMembership struct{}
