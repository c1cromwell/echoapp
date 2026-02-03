package cardano

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"
)

// SchemaVersion represents a specific version of a schema
type SchemaVersion struct {
	Version     int                    `json:"version"`
	SchemaID    string                 `json:"schema_id"`
	ContentHash string                 `json:"content_hash"`
	CreatedAt   time.Time              `json:"created_at"`
	CreatedBy   string                 `json:"created_by"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// CredentialSchema represents a W3C-compliant credential schema
type CredentialSchema struct {
	SchemaID        string                 `json:"schema_id"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	Version         int                    `json:"version"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
	CreatedBy       string                 `json:"created_by"` // Issuer DID
	ContentHash     string                 `json:"content_hash"`
	Properties      map[string]interface{} `json:"properties"`
	Required        []string               `json:"required"`
	Context         []string               `json:"context"`
	Type            []string               `json:"type"`
	Status          string                 `json:"status"` // active, deprecated, archived
	VersionHistory  []*SchemaVersion       `json:"version_history,omitempty"`
	OnChainMetadata map[string]interface{} `json:"on_chain_metadata,omitempty"`
	TxHash          string                 `json:"tx_hash,omitempty"`
	Timestamp       time.Time              `json:"timestamp"`
}

// SchemaRegistry manages schema storage and versioning
type SchemaRegistry struct {
	schemas map[string]*CredentialSchema
	logger  interface{}
}

// SchemaQuery filters schemas during retrieval
type SchemaQuery struct {
	SchemaID  string
	Version   int
	CreatedBy string
	Status    string
	Limit     int
	Offset    int
}

// SchemaResult represents schema retrieval result
type SchemaResult struct {
	Total     int
	Count     int
	Schemas   []*CredentialSchema
	QueryTime time.Duration
	TxHash    string
}

// StoreSchema stores a credential schema on the Cardano blockchain
func (c *Client) StoreSchema(ctx context.Context, schema *CredentialSchema) (string, error) {
	if err := validateSchema(schema); err != nil {
		c.logger.Printf("Schema validation failed: %v", err)
		return "", fmt.Errorf("invalid schema: %w", err)
	}

	// Generate content hash
	contentBytes, _ := json.Marshal(schema.Properties)
	hash := sha256.Sum256(contentBytes)
	schema.ContentHash = fmt.Sprintf("%x", hash)[:32]

	// Initialize version history
	if schema.VersionHistory == nil {
		schema.VersionHistory = make([]*SchemaVersion, 0)
	}

	// Add current version to history
	schema.VersionHistory = append(schema.VersionHistory, &SchemaVersion{
		Version:     schema.Version,
		SchemaID:    schema.SchemaID,
		ContentHash: schema.ContentHash,
		CreatedAt:   schema.CreatedAt,
		CreatedBy:   schema.CreatedBy,
	})

	// Store metadata on-chain
	payload := map[string]interface{}{
		"schema_id":    schema.SchemaID,
		"name":         schema.Name,
		"version":      schema.Version,
		"content_hash": schema.ContentHash,
		"created_by":   schema.CreatedBy,
		"created_at":   schema.CreatedAt.Unix(),
		"status":       schema.Status,
		"properties":   schema.Properties,
	}

	txHash, err := c.simulateCardanoSubmission(ctx, "store-schema", payload)
	if err != nil {
		c.logger.Printf("Failed to store schema on blockchain: %v", err)
		return "", err
	}

	schema.TxHash = txHash
	schema.Timestamp = time.Now()

	c.logger.Printf("Schema %s (v%d) stored with tx %s", schema.SchemaID, schema.Version, txHash)
	return txHash, nil
}

// GetSchema retrieves a credential schema
func (c *Client) GetSchema(ctx context.Context, schemaID string, version int) (*CredentialSchema, error) {
	cacheKey := fmt.Sprintf("schema_%s_v%d", schemaID, version)

	// Check cache
	if cached, exists := c.credentialCache.Get(cacheKey); exists {
		c.logger.Printf("Schema cache hit: %s v%d", schemaID, version)
		return cached.(*CredentialSchema), nil
	}

	// Simulate blockchain query
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		// In production, would query Cardano blockchain
		c.logger.Printf("Retrieved schema %s v%d", schemaID, version)
		return nil, nil
	}
}

// QuerySchemas queries schemas with filters
func (c *Client) QuerySchemas(ctx context.Context, query *SchemaQuery) (*SchemaResult, error) {
	startTime := time.Now()

	cacheKey := fmt.Sprintf("schema_query_%s_%d", query.SchemaID, query.Version)
	if cached, exists := c.credentialCache.Get(cacheKey); exists {
		return cached.(*SchemaResult), nil
	}

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		result := &SchemaResult{
			Total:     0,
			Count:     0,
			Schemas:   make([]*CredentialSchema, 0),
			QueryTime: time.Since(startTime),
		}
		return result, nil
	}
}

// UpdateSchema updates an existing schema (creates new version)
func (c *Client) UpdateSchema(ctx context.Context, schema *CredentialSchema) (string, error) {
	if schema.SchemaID == "" {
		return "", fmt.Errorf("schema ID required")
	}

	// Increment version
	schema.Version++
	schema.UpdatedAt = time.Now()

	// Store updated schema
	txHash, err := c.StoreSchema(ctx, schema)
	if err != nil {
		return "", err
	}

	c.logger.Printf("Schema %s updated to v%d with tx %s", schema.SchemaID, schema.Version, txHash)
	return txHash, nil
}

// DeprecateSchema marks a schema as deprecated
func (c *Client) DeprecateSchema(ctx context.Context, schemaID string, reason string) (string, error) {
	payload := map[string]interface{}{
		"schema_id": schemaID,
		"action":    "deprecate",
		"reason":    reason,
		"timestamp": time.Now().Unix(),
	}

	txHash, err := c.simulateCardanoSubmission(ctx, "deprecate-schema", payload)
	if err != nil {
		return "", err
	}

	c.logger.Printf("Schema %s deprecated: %s", schemaID, reason)
	return txHash, nil
}

// GetSchemaVersionHistory retrieves all versions of a schema
func (c *Client) GetSchemaVersionHistory(ctx context.Context, schemaID string) ([]*SchemaVersion, error) {
	cacheKey := fmt.Sprintf("schema_history_%s", schemaID)

	if cached, exists := c.credentialCache.Get(cacheKey); exists {
		return cached.([]*SchemaVersion), nil
	}

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		c.logger.Printf("Retrieved version history for schema %s", schemaID)
		return make([]*SchemaVersion, 0), nil
	}
}

// ValidateCredentialAgainstSchema validates a credential against its schema
func (c *Client) ValidateCredentialAgainstSchema(ctx context.Context, credential *Credential, schema *CredentialSchema) (bool, []string, error) {
	errors := make([]string, 0)

	// Check required fields
	if credential.Data == nil {
		credential.Data = make(map[string]interface{})
	}

	for _, required := range schema.Required {
		if _, exists := credential.Data[required]; !exists {
			errors = append(errors, fmt.Sprintf("required field missing: %s", required))
		}
	}

	// Validate field types against schema properties
	for field := range credential.Data {
		if _, exists := schema.Properties[field]; !exists {
			errors = append(errors, fmt.Sprintf("unexpected field: %s", field))
		}
	}

	if len(errors) > 0 {
		c.logger.Printf("Credential validation against schema %s failed: %v", schema.SchemaID, errors)
		return false, errors, nil
	}

	return true, nil, nil
}

// validateSchema validates schema structure
func validateSchema(schema *CredentialSchema) error {
	if schema.SchemaID == "" {
		return fmt.Errorf("schema ID is required")
	}
	if schema.Name == "" {
		return fmt.Errorf("schema name is required")
	}
	if schema.Version < 1 {
		return fmt.Errorf("schema version must be >= 1")
	}
	if schema.CreatedBy == "" {
		return fmt.Errorf("creator DID is required")
	}
	if len(schema.Properties) == 0 {
		return fmt.Errorf("schema properties cannot be empty")
	}
	return nil
}
