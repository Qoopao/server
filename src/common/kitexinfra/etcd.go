package kitexinfra

import (
	"os"
	"strings"

	"github.com/cloudwego/kitex/pkg/discovery"
	"github.com/cloudwego/kitex/pkg/registry"
	etcd "github.com/kitex-contrib/registry-etcd"
)

// EtcdEndpointsFromEnv centralizes the runtime endpoint configuration used by
// Kitex registry and resolver. Service discovery itself is provided by Kitex's
// etcd extension; this helper only maps environment configuration to that API.
func EtcdEndpointsFromEnv() []string {
	raw := os.Getenv("ETCD_ENDPOINTS")
	if raw == "" {
		return []string{"localhost:2379"}
	}

	parts := strings.Split(raw, ",")
	endpoints := make([]string, 0, len(parts))
	for _, part := range parts {
		if endpoint := strings.TrimSpace(part); endpoint != "" {
			endpoints = append(endpoints, endpoint)
		}
	}
	if len(endpoints) == 0 {
		return []string{"localhost:2379"}
	}
	return endpoints
}

func NewEtcdRegistryFromEnv() (registry.Registry, error) {
	return etcd.NewEtcdRegistry(EtcdEndpointsFromEnv())
}

func NewEtcdResolverFromEnv() (discovery.Resolver, error) {
	return etcd.NewEtcdResolver(EtcdEndpointsFromEnv())
}
