package entity

import "net/url"

type UpstreamServer struct {
	URL    *url.URL
	Weight int
}