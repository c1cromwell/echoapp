package services

// ServiceDef defines a microservice
type ServiceDef struct {
	Name string
	Port int
	Role string
}

// NewServiceDef creates a service definition
func NewServiceDef(name string, port int, role string) ServiceDef {
	return ServiceDef{Name: name, Port: port, Role: role}
}

// ServiceRegistry holds all service configurations (v2.1)
var ServiceRegistry = map[string]ServiceDef{
	"gateway":           NewServiceDef("gateway", 8000, "Load balancer, TLS termination, rate limiting"),
	"identity":          NewServiceDef("identity-service", 8001, "Registration, DID management, credential caching"),
	"relay":             NewServiceDef("message-relay", 8002, "WebSocket relay, offline queue, APNs push"),
	"trust":             NewServiceDef("trust-service", 8003, "Trust score computation, tier caching"),
	"rewards":           NewServiceDef("rewards-service", 8004, "Reward validation, batching, AtomicAction submission"),
	"contacts":          NewServiceDef("contacts-service", 8005, "Contact list, block list, search"),
	"metagraph-gateway": NewServiceDef("metagraph-gateway", 8006, "L1/L0 submission (v3 types), snapshot listening, anchoring, FeeTransaction"),
	"notification":      NewServiceDef("notification-service", 8007, "APNs push, in-app notifications"),
	"media":             NewServiceDef("media-service", 8008, "Encrypted media upload/download"),
	"log-publisher":     NewServiceDef("log-publisher", 8009, "Batch encryption, IPFS submission, CID indexing"),
	"digital-evidence":  NewServiceDef("digital-evidence", 8010, "Enterprise audit fingerprinting, media verification, Smart Checkmark"),
}
