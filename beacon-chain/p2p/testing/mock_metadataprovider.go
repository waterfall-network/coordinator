package testing

import (
	"gitlab.waterfall.network/waterfall/protocol/coordinator/proto/prysm/v1alpha1/metadata"
)

// MockMetadataProvider is a fake implementation of the MetadataProvider interface.
type MockMetadataProvider struct {
	Data metadata.Metadata
}

// Metadata --
func (m *MockMetadataProvider) Metadata() metadata.Metadata {
	return m.Data
}

// MetadataSeq --
func (m *MockMetadataProvider) MetadataSeq() uint64 {
	return m.Data.SequenceNumber()
}
