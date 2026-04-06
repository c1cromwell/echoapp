package services

import "testing"

func TestServiceRegistry(t *testing.T) {
	expectedServices := []string{
		"gateway", "identity", "relay", "trust", "rewards",
		"contacts", "metagraph-gateway", "notification", "media",
		"log-publisher", "digital-evidence",
	}
	for _, name := range expectedServices {
		if _, exists := ServiceRegistry[name]; !exists {
			t.Errorf("Service %s not found in registry", name)
		}
	}
}

func TestServicePorts(t *testing.T) {
	tests := map[string]int{
		"gateway":           8000,
		"identity":          8001,
		"relay":             8002,
		"trust":             8003,
		"rewards":           8004,
		"contacts":          8005,
		"metagraph-gateway": 8006,
		"notification":      8007,
		"media":             8008,
		"log-publisher":     8009,
		"digital-evidence":  8010,
	}
	for name, expectedPort := range tests {
		svc, ok := ServiceRegistry[name]
		if !ok {
			t.Errorf("Service %s not found", name)
			continue
		}
		if svc.Port != expectedPort {
			t.Errorf("%s port: got %d, want %d", name, svc.Port, expectedPort)
		}
	}
}

func TestServiceRegistryCount(t *testing.T) {
	if len(ServiceRegistry) != 11 {
		t.Errorf("expected 11 services in registry, got %d", len(ServiceRegistry))
	}
}

func TestServiceRoles(t *testing.T) {
	for name, svc := range ServiceRegistry {
		if svc.Role == "" {
			t.Errorf("service %s has empty role", name)
		}
	}
}
