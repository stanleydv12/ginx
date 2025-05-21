//go:build linux

package loadbalancer

import (
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/stanleydv12/ginx/internal/config"
	"github.com/stanleydv12/ginx/internal/entity"
	"github.com/stanleydv12/ginx/pkg/logger"
)

type LoadBalancerHandler interface {
	SelectServer() (entity.UpstreamServer, error)
	AddServer(server entity.UpstreamServer) error
	RemoveServer(server entity.UpstreamServer) error
}

func NewLoadBalancer(cfg *config.ServerConfig) (LoadBalancerHandler, error) {
	upstreamServers := []entity.UpstreamServer{}

	for _, server := range cfg.Server.UpstreamServers {

		// Ensure the URL has a scheme
		if !strings.Contains(server, "://") {
			server = "http://" + server
		}

		url, err := url.Parse(server)
		if err != nil {
			return nil, err
		}

		// Validate host
        if url.Host == "" {
            return nil, fmt.Errorf("missing host in URL: %s", server)
        }

        // Only try to resolve if it's not an IP address
        host := url.Hostname()
        if net.ParseIP(host) == nil {
            logger.Info("Resolving hostname", "host", host)
            resolvedIPs, err := net.LookupHost(host)
            if err != nil {
                logger.Error("Failed to resolve hostname", "host", host, "error", err)
                return nil, fmt.Errorf("failed to resolve %s: %v", host, err)
            }
            if len(resolvedIPs) == 0 {
                return nil, fmt.Errorf("no IP addresses found for %s", host)
            }
            // Replace host with the first resolved IP, keeping the original port
            if url.Port() != "" {
                url.Host = net.JoinHostPort(resolvedIPs[0], url.Port())
            } else {
                url.Host = resolvedIPs[0]
            }
            logger.Debug("Resolved hostname", "host", host, "ip", resolvedIPs[0])
        }

		upstreamServers = append(upstreamServers, entity.UpstreamServer{
			URL: url,
		})

		logger.Info("Added upstream server", "url", url)
	}
	switch cfg.Server.LoadBalancer {
	case "round_robin":
		return NewRoundRobinLoadBalancer(0, upstreamServers), nil
	default:
		return nil, fmt.Errorf("unsupported load balancer type: %s", cfg.Server.LoadBalancer)
	}
}
