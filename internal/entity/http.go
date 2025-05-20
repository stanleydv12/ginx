//go:build linux

package entity

type HTTPRequest struct {
	Method string
	Path   string
	Protocol string
	Headers map[string]string
	Body   []byte
	Raw    []byte
}

type HTTPResponse struct {
	StatusCode int
	Headers map[string]string
	Body   []byte
	Raw    []byte
}