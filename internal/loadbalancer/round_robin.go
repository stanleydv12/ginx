//go:build linux

package loadbalancer

import (
	"errors"
	"github.com/stanleydv12/ginx/internal/entity"
)

type RoundRobinLoadBalancer struct {
	currentServer int
	upstreamServers []entity.UpstreamServer	
}

func NewRoundRobinLoadBalancer(currentServer int, upstreamServers []entity.UpstreamServer) LoadBalancerHandler {
	return &RoundRobinLoadBalancer{
		currentServer: currentServer,
		upstreamServers: upstreamServers,
	}
}

func (l *RoundRobinLoadBalancer) SelectServer() (entity.UpstreamServer, error) {
	if len(l.upstreamServers) == 0 {
		return entity.UpstreamServer{}, errors.New("no upstream servers")
	}
	
	server := l.upstreamServers[l.currentServer]
	l.currentServer = (l.currentServer + 1) % len(l.upstreamServers)
	return server, nil
}

func (l *RoundRobinLoadBalancer) AddServer(server entity.UpstreamServer) error {
	l.upstreamServers = append(l.upstreamServers, server)
	return nil
}

func (l *RoundRobinLoadBalancer) RemoveServer(server entity.UpstreamServer) error {
	for i, s := range l.upstreamServers {
		if s == server {
			l.upstreamServers = append(l.upstreamServers[:i], l.upstreamServers[i+1:]...)
			return nil
		}
	}
	return nil
}

