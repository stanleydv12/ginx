//go:build linux

package connection

import "github.com/stanleydv12/ginx/internal/entity"

type Connection struct {
	ClientFD int
	ClientAddress string
	UpstreamFD int
	UpstreamServer entity.UpstreamServer
	Request entity.HTTPRequest
	Response entity.HTTPResponse
	State ConnectionState
}

type ConnectionState string

const (
    StateClientAccepted ConnectionState = "client_accepted"
    StateRequestReceived ConnectionState = "request_received"
    StateConnectingUpstream ConnectionState = "connecting_upstream"
    StateForwardingRequest ConnectionState = "forwarding_request"
    StateWaitingResponse ConnectionState = "waiting_response"
    StateSendingResponse ConnectionState = "sending_response"
    StateCompleted ConnectionState = "completed"
    StateError ConnectionState = "error"
)