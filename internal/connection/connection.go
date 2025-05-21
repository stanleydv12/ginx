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
	State string
}