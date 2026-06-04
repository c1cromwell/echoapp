package logging

// WO-33 / WO-294: Decentralized storage for encrypted blobs.
// Implementations live in pkg/storage/encblob; this package re-exports them for audit-log publishing.

import (
	"github.com/thechadcromwell/echoapp/pkg/storage/encblob"
)

// ErrStorageNotConfigured is returned when required environment credentials are absent.
var ErrStorageNotConfigured = encblob.ErrStorageNotConfigured

type (
	// IPFSStorage pins opaque ciphertext (client-encrypted batches or passport sync blobs).
	IPFSStorage         = encblob.Storage
	PinataIPFSStorage   = encblob.PinataStorage
	StorjIPFSStorage    = encblob.StorjStorage
	FallbackIPFSStorage = encblob.FallbackStorage
	StubIPFSStorage     = encblob.StubStorage
)

func NewPinataIPFSStorage() (*PinataIPFSStorage, error)     { return encblob.NewPinataStorage() }
func NewStorjIPFSStorage() (*StorjIPFSStorage, error)       { return encblob.NewStorjStorage() }
func NewFallbackIPFSStorage() (*FallbackIPFSStorage, error) { return encblob.NewFallbackStorage() }

func NewFallbackIPFSStorageWithProviders(primary, secondary IPFSStorage) *FallbackIPFSStorage {
	return encblob.NewFallbackStorageWithProviders(primary, secondary)
}
