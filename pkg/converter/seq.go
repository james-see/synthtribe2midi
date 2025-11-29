package converter

import (
	"errors"
	"fmt"
	"os"
)

// SeqConverter handles .seq file parsing and generation
type SeqConverter struct {
	device Device
}

// NewSeqConverter creates a new .seq converter
func NewSeqConverter(device Device) *SeqConverter {
	return &SeqConverter{device: device}
}

// ParseSeqFile reads a .seq file and returns a Pattern
func (s *SeqConverter) ParseSeqFile(filename string) (*Pattern, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read seq file: %w", err)
	}
	return s.ParseSeq(data)
}

// ParseSeq parses .seq data and returns a Pattern
func (s *SeqConverter) ParseSeq(data []byte) (*Pattern, error) {
	if s.device == nil {
		return nil, errors.New("no device configured")
	}
	return s.device.ParseSeq(data)
}

// GenerateSeq creates .seq data from a Pattern
func (s *SeqConverter) GenerateSeq(pattern *Pattern) ([]byte, error) {
	if s.device == nil {
		return nil, errors.New("no device configured")
	}
	return s.device.GenerateSeq(pattern)
}

// WriteSeqFile writes .seq data to a file
func (s *SeqConverter) WriteSeqFile(pattern *Pattern, filename string) error {
	data, err := s.GenerateSeq(pattern)
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}

// ValidateSeq validates .seq data structure
func (s *SeqConverter) ValidateSeq(data []byte) error {
	if len(data) < 32 {
		return errors.New("seq data too short: minimum 32 bytes required")
	}
	
	// Basic validation - check for reasonable step data
	for i := 0; i < 16 && i*2+1 < len(data); i++ {
		noteData := data[i*2]
		if noteData > 127 {
			return fmt.Errorf("invalid note value at step %d: %d (max 127)", i, noteData)
		}
	}
	
	return nil
}

