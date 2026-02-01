package httpapi

import (
	"net"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
)

func requestAuditFields(req *http.Request) (requestID *string, userAgent *string, sourceIP net.IP) {
	if rid := middleware.GetReqID(req.Context()); rid != "" {
		requestID = &rid
	}
	if ua := req.UserAgent(); ua != "" {
		userAgent = &ua
	}

	remote := req.RemoteAddr
	host, _, err := net.SplitHostPort(remote)
	if err == nil {
		remote = host
	}
	sourceIP = net.ParseIP(remote)
	return requestID, userAgent, sourceIP
}
