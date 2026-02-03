package did

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"
)

// Repository defines the interface for DID data access
type Repository interface {
	// CreateDIDMapping creates a new DID-to-account mapping
	CreateDIDMapping(ctx context.Context, mapping *DIDMapping) error

	// GetDIDByID retrieves a DID mapping by DID
	GetDIDByID(ctx context.Context, did string) (*DIDMapping, error)

	// GetDIDByUserID retrieves a DID mapping by user ID
	GetDIDByUserID(ctx context.Context, userID string) (*DIDMapping, error)

	// UpdateDIDMapping updates an existing DID mapping
	UpdateDIDMapping(ctx context.Context, mapping *DIDMapping) error

	// DeleteDIDMapping deletes a DID mapping
	DeleteDIDMapping(ctx context.Context, did string) error

	// AddDevice adds a new device to a DID mapping
	AddDevice(ctx context.Context, did string, device *DeviceRegistration) error

	// GetDevice retrieves a device from a DID mapping
	GetDevice(ctx context.Context, did, deviceID string) (*DeviceRegistration, error)

	// UpdateDevice updates a device in a DID mapping
	UpdateDevice(ctx context.Context, did string, device *DeviceRegistration) error

	// RemoveDevice removes a device from a DID mapping
	RemoveDevice(ctx context.Context, did, deviceID string) error

	// ListDevices lists all devices for a DID
	ListDevices(ctx context.Context, did string) ([]DeviceRegistration, error)

	// StoreDIDDocument stores a DID document
	StoreDIDDocument(ctx context.Context, did string, document *DIDDocument) error

	// GetDIDDocument retrieves a stored DID document
	GetDIDDocument(ctx context.Context, did string) (*DIDDocument, error)

	// UpdateDIDDocument updates a stored DID document
	UpdateDIDDocument(ctx context.Context, did string, document *DIDDocument) error

	// RecordAnchor records a blockchain anchor for a DID
	RecordAnchor(ctx context.Context, did, txHash string, blockNum int64) error

	// GetAnchor retrieves blockchain anchor information for a DID
	GetAnchor(ctx context.Context, did string) (string, int64, time.Time, error)

	// Health checks database connectivity
	Health(ctx context.Context) error

	// Close closes the repository connection
	Close() error
}

// InMemoryRepository provides an in-memory implementation of Repository
type InMemoryRepository struct {
	mu             sync.RWMutex
	didMappings    map[string]*DIDMapping
	didDocuments   map[string]*DIDDocument
	anchors        map[string]*AnchorInfo
	userToDidIndex map[string]string
}

// AnchorInfo holds blockchain anchor information
type AnchorInfo struct {
	DID       string
	TxHash    string
	BlockNum  int64
	Timestamp time.Time
}

// NewInMemoryRepository creates a new in-memory repository
func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		didMappings:    make(map[string]*DIDMapping),
		didDocuments:   make(map[string]*DIDDocument),
		anchors:        make(map[string]*AnchorInfo),
		userToDidIndex: make(map[string]string),
	}
}

// CreateDIDMapping creates a new DID-to-account mapping
func (r *InMemoryRepository) CreateDIDMapping(ctx context.Context, mapping *DIDMapping) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.didMappings[mapping.DID]; exists {
		return NewDIDError(ErrCodeDIDAlreadyExists, fmt.Sprintf("DID mapping already exists: %s", mapping.DID), nil)
	}

	now := time.Now()
	mapping.CreatedAt = now
	mapping.UpdatedAt = now

	r.didMappings[mapping.DID] = mapping
	r.userToDidIndex[mapping.UserID] = mapping.DID

	return nil
}

// GetDIDByID retrieves a DID mapping by DID
func (r *InMemoryRepository) GetDIDByID(ctx context.Context, did string) (*DIDMapping, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	mapping, exists := r.didMappings[did]
	if !exists {
		return nil, NewDIDError(ErrCodeDIDNotFound, fmt.Sprintf("DID not found: %s", did), nil)
	}

	return mapping, nil
}

// GetDIDByUserID retrieves a DID mapping by user ID
func (r *InMemoryRepository) GetDIDByUserID(ctx context.Context, userID string) (*DIDMapping, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	did, exists := r.userToDidIndex[userID]
	if !exists {
		return nil, NewDIDError(ErrCodeDIDNotFound, fmt.Sprintf("No DID found for user: %s", userID), nil)
	}

	mapping, exists := r.didMappings[did]
	if !exists {
		return nil, NewDIDError(ErrCodeDIDNotFound, fmt.Sprintf("DID mapping not found: %s", did), nil)
	}

	return mapping, nil
}

// UpdateDIDMapping updates an existing DID mapping
func (r *InMemoryRepository) UpdateDIDMapping(ctx context.Context, mapping *DIDMapping) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.didMappings[mapping.DID]; !exists {
		return NewDIDError(ErrCodeDIDNotFound, fmt.Sprintf("DID not found: %s", mapping.DID), nil)
	}

	mapping.UpdatedAt = time.Now()
	r.didMappings[mapping.DID] = mapping

	return nil
}

// DeleteDIDMapping deletes a DID mapping
func (r *InMemoryRepository) DeleteDIDMapping(ctx context.Context, did string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	mapping, exists := r.didMappings[did]
	if !exists {
		return NewDIDError(ErrCodeDIDNotFound, fmt.Sprintf("DID not found: %s", did), nil)
	}

	delete(r.didMappings, did)
	delete(r.userToDidIndex, mapping.UserID)
	delete(r.didDocuments, did)
	delete(r.anchors, did)

	return nil
}

// AddDevice adds a new device to a DID mapping
func (r *InMemoryRepository) AddDevice(ctx context.Context, did string, device *DeviceRegistration) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	mapping, exists := r.didMappings[did]
	if !exists {
		return NewDIDError(ErrCodeDIDNotFound, fmt.Sprintf("DID not found: %s", did), nil)
	}

	// Check if device already exists
	for _, d := range mapping.Devices {
		if d.DeviceID == device.DeviceID {
			return NewDIDError(ErrCodeDeviceAlreadyExists, fmt.Sprintf("Device already exists: %s", device.DeviceID), nil)
		}
	}

	device.CreatedAt = time.Now()
	device.LastUsedAt = time.Now()
	device.IsActive = true

	mapping.Devices = append(mapping.Devices, *device)
	mapping.UpdatedAt = time.Now()

	return nil
}

// GetDevice retrieves a device from a DID mapping
func (r *InMemoryRepository) GetDevice(ctx context.Context, did, deviceID string) (*DeviceRegistration, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	mapping, exists := r.didMappings[did]
	if !exists {
		return nil, NewDIDError(ErrCodeDIDNotFound, fmt.Sprintf("DID not found: %s", did), nil)
	}

	for _, device := range mapping.Devices {
		if device.DeviceID == deviceID {
			return &device, nil
		}
	}

	return nil, NewDIDError(ErrCodeDeviceNotFound, fmt.Sprintf("Device not found: %s", deviceID), nil)
}

// UpdateDevice updates a device in a DID mapping
func (r *InMemoryRepository) UpdateDevice(ctx context.Context, did string, device *DeviceRegistration) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	mapping, exists := r.didMappings[did]
	if !exists {
		return NewDIDError(ErrCodeDIDNotFound, fmt.Sprintf("DID not found: %s", did), nil)
	}

	found := false
	for i, d := range mapping.Devices {
		if d.DeviceID == device.DeviceID {
			device.LastUsedAt = time.Now()
			mapping.Devices[i] = *device
			found = true
			break
		}
	}

	if !found {
		return NewDIDError(ErrCodeDeviceNotFound, fmt.Sprintf("Device not found: %s", device.DeviceID), nil)
	}

	mapping.UpdatedAt = time.Now()
	return nil
}

// RemoveDevice removes a device from a DID mapping
func (r *InMemoryRepository) RemoveDevice(ctx context.Context, did, deviceID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	mapping, exists := r.didMappings[did]
	if !exists {
		return NewDIDError(ErrCodeDIDNotFound, fmt.Sprintf("DID not found: %s", did), nil)
	}

	newDevices := make([]DeviceRegistration, 0)
	found := false

	for _, device := range mapping.Devices {
		if device.DeviceID != deviceID {
			newDevices = append(newDevices, device)
		} else {
			found = true
		}
	}

	if !found {
		return NewDIDError(ErrCodeDeviceNotFound, fmt.Sprintf("Device not found: %s", deviceID), nil)
	}

	mapping.Devices = newDevices
	mapping.UpdatedAt = time.Now()

	return nil
}

// ListDevices lists all devices for a DID
func (r *InMemoryRepository) ListDevices(ctx context.Context, did string) ([]DeviceRegistration, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	mapping, exists := r.didMappings[did]
	if !exists {
		return nil, NewDIDError(ErrCodeDIDNotFound, fmt.Sprintf("DID not found: %s", did), nil)
	}

	devices := make([]DeviceRegistration, len(mapping.Devices))
	copy(devices, mapping.Devices)

	return devices, nil
}

// StoreDIDDocument stores a DID document
func (r *InMemoryRepository) StoreDIDDocument(ctx context.Context, did string, document *DIDDocument) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if document == nil {
		return NewDIDError(ErrCodeInvalidDocument, "DID document cannot be nil", nil)
	}

	r.didDocuments[did] = document
	return nil
}

// GetDIDDocument retrieves a stored DID document
func (r *InMemoryRepository) GetDIDDocument(ctx context.Context, did string) (*DIDDocument, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	doc, exists := r.didDocuments[did]
	if !exists {
		return nil, NewDIDError(ErrCodeDIDNotFound, fmt.Sprintf("DID document not found: %s", did), nil)
	}

	return doc, nil
}

// UpdateDIDDocument updates a stored DID document
func (r *InMemoryRepository) UpdateDIDDocument(ctx context.Context, did string, document *DIDDocument) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if document == nil {
		return NewDIDError(ErrCodeInvalidDocument, "DID document cannot be nil", nil)
	}

	if _, exists := r.didDocuments[did]; !exists {
		return NewDIDError(ErrCodeDIDNotFound, fmt.Sprintf("DID document not found: %s", did), nil)
	}

	document.Updated = time.Now()
	r.didDocuments[did] = document

	return nil
}

// RecordAnchor records a blockchain anchor for a DID
func (r *InMemoryRepository) RecordAnchor(ctx context.Context, did, txHash string, blockNum int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.anchors[did] = &AnchorInfo{
		DID:       did,
		TxHash:    txHash,
		BlockNum:  blockNum,
		Timestamp: time.Now(),
	}

	return nil
}

// GetAnchor retrieves blockchain anchor information for a DID
func (r *InMemoryRepository) GetAnchor(ctx context.Context, did string) (string, int64, time.Time, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	anchor, exists := r.anchors[did]
	if !exists {
		return "", 0, time.Time{}, NewDIDError(ErrCodeDocumentNotAnchored, fmt.Sprintf("No anchor found for DID: %s", did), nil)
	}

	return anchor.TxHash, anchor.BlockNum, anchor.Timestamp, nil
}

// Health checks repository connectivity
func (r *InMemoryRepository) Health(ctx context.Context) error {
	return nil
}

// Close closes the repository
func (r *InMemoryRepository) Close() error {
	return nil
}

// DatabaseRepository provides a database-backed implementation of Repository
type DatabaseRepository struct {
	db *sql.DB
	mu sync.RWMutex
}

// NewDatabaseRepository creates a new database repository
func NewDatabaseRepository(db *sql.DB) *DatabaseRepository {
	return &DatabaseRepository{
		db: db,
	}
}

// CreateDIDMapping creates a new DID-to-account mapping in the database
func (r *DatabaseRepository) CreateDIDMapping(ctx context.Context, mapping *DIDMapping) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	query := `
		INSERT INTO did_mappings (did, user_id, account_id, created_at, updated_at, is_active, primary_device)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	now := time.Now()
	_, err := r.db.ExecContext(ctx, query,
		mapping.DID,
		mapping.UserID,
		mapping.AccountID,
		now,
		now,
		mapping.IsActive,
		mapping.PrimaryDevice,
	)

	if err != nil {
		return NewDIDError(ErrCodeDatabaseError, "Failed to create DID mapping", err)
	}

	return nil
}

// GetDIDByID retrieves a DID mapping by DID from the database
func (r *DatabaseRepository) GetDIDByID(ctx context.Context, did string) (*DIDMapping, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	query := `
		SELECT did, user_id, account_id, created_at, updated_at, is_active, primary_device
		FROM did_mappings
		WHERE did = $1
	`

	mapping := &DIDMapping{}
	err := r.db.QueryRowContext(ctx, query, did).Scan(
		&mapping.DID,
		&mapping.UserID,
		&mapping.AccountID,
		&mapping.CreatedAt,
		&mapping.UpdatedAt,
		&mapping.IsActive,
		&mapping.PrimaryDevice,
	)

	if err == sql.ErrNoRows {
		return nil, NewDIDError(ErrCodeDIDNotFound, fmt.Sprintf("DID not found: %s", did), nil)
	}
	if err != nil {
		return nil, NewDIDError(ErrCodeDatabaseError, "Failed to get DID mapping", err)
	}

	return mapping, nil
}

// GetDIDByUserID retrieves a DID mapping by user ID from the database
func (r *DatabaseRepository) GetDIDByUserID(ctx context.Context, userID string) (*DIDMapping, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	query := `
		SELECT did, user_id, account_id, created_at, updated_at, is_active, primary_device
		FROM did_mappings
		WHERE user_id = $1
	`

	mapping := &DIDMapping{}
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&mapping.DID,
		&mapping.UserID,
		&mapping.AccountID,
		&mapping.CreatedAt,
		&mapping.UpdatedAt,
		&mapping.IsActive,
		&mapping.PrimaryDevice,
	)

	if err == sql.ErrNoRows {
		return nil, NewDIDError(ErrCodeDIDNotFound, fmt.Sprintf("No DID found for user: %s", userID), nil)
	}
	if err != nil {
		return nil, NewDIDError(ErrCodeDatabaseError, "Failed to get DID mapping", err)
	}

	return mapping, nil
}

// UpdateDIDMapping updates an existing DID mapping in the database
func (r *DatabaseRepository) UpdateDIDMapping(ctx context.Context, mapping *DIDMapping) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	query := `
		UPDATE did_mappings
		SET user_id = $1, account_id = $2, updated_at = $3, is_active = $4, primary_device = $5
		WHERE did = $6
	`

	mapping.UpdatedAt = time.Now()
	_, err := r.db.ExecContext(ctx, query,
		mapping.UserID,
		mapping.AccountID,
		mapping.UpdatedAt,
		mapping.IsActive,
		mapping.PrimaryDevice,
		mapping.DID,
	)

	if err != nil {
		return NewDIDError(ErrCodeDatabaseError, "Failed to update DID mapping", err)
	}

	return nil
}

// DeleteDIDMapping deletes a DID mapping from the database
func (r *DatabaseRepository) DeleteDIDMapping(ctx context.Context, did string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	query := `DELETE FROM did_mappings WHERE did = $1`
	_, err := r.db.ExecContext(ctx, query, did)

	if err != nil {
		return NewDIDError(ErrCodeDatabaseError, "Failed to delete DID mapping", err)
	}

	return nil
}

// AddDevice adds a new device to a DID mapping in the database
func (r *DatabaseRepository) AddDevice(ctx context.Context, did string, device *DeviceRegistration) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	query := `
		INSERT INTO devices (did, device_id, device_name, public_key, created_at, is_active, is_secure_enclave)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.ExecContext(ctx, query,
		did,
		device.DeviceID,
		device.DeviceName,
		device.PublicKey,
		time.Now(),
		true,
		device.IsSecureEnclave,
	)

	if err != nil {
		return NewDIDError(ErrCodeDatabaseError, "Failed to add device", err)
	}

	return nil
}

// GetDevice retrieves a device from the database
func (r *DatabaseRepository) GetDevice(ctx context.Context, did, deviceID string) (*DeviceRegistration, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	query := `
		SELECT device_id, device_name, public_key, created_at, is_active, is_secure_enclave
		FROM devices
		WHERE did = $1 AND device_id = $2
	`

	device := &DeviceRegistration{}
	err := r.db.QueryRowContext(ctx, query, did, deviceID).Scan(
		&device.DeviceID,
		&device.DeviceName,
		&device.PublicKey,
		&device.CreatedAt,
		&device.IsActive,
		&device.IsSecureEnclave,
	)

	if err == sql.ErrNoRows {
		return nil, NewDIDError(ErrCodeDeviceNotFound, fmt.Sprintf("Device not found: %s", deviceID), nil)
	}
	if err != nil {
		return nil, NewDIDError(ErrCodeDatabaseError, "Failed to get device", err)
	}

	return device, nil
}

// UpdateDevice updates a device in the database
func (r *DatabaseRepository) UpdateDevice(ctx context.Context, did string, device *DeviceRegistration) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	query := `
		UPDATE devices
		SET device_name = $1, is_active = $2, last_used_at = $3
		WHERE did = $4 AND device_id = $5
	`

	_, err := r.db.ExecContext(ctx, query,
		device.DeviceName,
		device.IsActive,
		time.Now(),
		did,
		device.DeviceID,
	)

	if err != nil {
		return NewDIDError(ErrCodeDatabaseError, "Failed to update device", err)
	}

	return nil
}

// RemoveDevice removes a device from the database
func (r *DatabaseRepository) RemoveDevice(ctx context.Context, did, deviceID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	query := `DELETE FROM devices WHERE did = $1 AND device_id = $2`
	_, err := r.db.ExecContext(ctx, query, did, deviceID)

	if err != nil {
		return NewDIDError(ErrCodeDatabaseError, "Failed to remove device", err)
	}

	return nil
}

// ListDevices lists all devices for a DID from the database
func (r *DatabaseRepository) ListDevices(ctx context.Context, did string) ([]DeviceRegistration, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	query := `
		SELECT device_id, device_name, public_key, created_at, is_active, is_secure_enclave
		FROM devices
		WHERE did = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, did)
	if err != nil {
		return nil, NewDIDError(ErrCodeDatabaseError, "Failed to list devices", err)
	}
	defer rows.Close()

	devices := make([]DeviceRegistration, 0)
	for rows.Next() {
		device := DeviceRegistration{}
		if err := rows.Scan(
			&device.DeviceID,
			&device.DeviceName,
			&device.PublicKey,
			&device.CreatedAt,
			&device.IsActive,
			&device.IsSecureEnclave,
		); err != nil {
			return nil, NewDIDError(ErrCodeDatabaseError, "Failed to scan device", err)
		}
		devices = append(devices, device)
	}

	return devices, nil
}

// StoreDIDDocument stores a DID document in the database
func (r *DatabaseRepository) StoreDIDDocument(ctx context.Context, did string, document *DIDDocument) error {
	// Implementation would depend on JSON storage strategy
	return nil
}

// GetDIDDocument retrieves a stored DID document from the database
func (r *DatabaseRepository) GetDIDDocument(ctx context.Context, did string) (*DIDDocument, error) {
	// Implementation would depend on JSON retrieval strategy
	return nil, nil
}

// UpdateDIDDocument updates a stored DID document in the database
func (r *DatabaseRepository) UpdateDIDDocument(ctx context.Context, did string, document *DIDDocument) error {
	// Implementation would depend on JSON update strategy
	return nil
}

// RecordAnchor records a blockchain anchor in the database
func (r *DatabaseRepository) RecordAnchor(ctx context.Context, did, txHash string, blockNum int64) error {
	query := `
		INSERT INTO anchors (did, tx_hash, block_num, created_at)
		VALUES ($1, $2, $3, $4)
	`

	_, err := r.db.ExecContext(ctx, query, did, txHash, blockNum, time.Now())
	if err != nil {
		return NewDIDError(ErrCodeDatabaseError, "Failed to record anchor", err)
	}

	return nil
}

// GetAnchor retrieves blockchain anchor information from the database
func (r *DatabaseRepository) GetAnchor(ctx context.Context, did string) (string, int64, time.Time, error) {
	query := `
		SELECT tx_hash, block_num, created_at
		FROM anchors
		WHERE did = $1
	`

	var txHash string
	var blockNum int64
	var timestamp time.Time

	err := r.db.QueryRowContext(ctx, query, did).Scan(&txHash, &blockNum, &timestamp)
	if err == sql.ErrNoRows {
		return "", 0, time.Time{}, NewDIDError(ErrCodeDocumentNotAnchored, fmt.Sprintf("No anchor found for DID: %s", did), nil)
	}
	if err != nil {
		return "", 0, time.Time{}, NewDIDError(ErrCodeDatabaseError, "Failed to get anchor", err)
	}

	return txHash, blockNum, timestamp, nil
}

// Health checks database connectivity
func (r *DatabaseRepository) Health(ctx context.Context) error {
	return r.db.PingContext(ctx)
}

// Close closes the database connection
func (r *DatabaseRepository) Close() error {
	return r.db.Close()
}
