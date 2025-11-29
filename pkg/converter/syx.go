package converter

import (
	"errors"
	"fmt"
	"os"
)

// SysEx constants
const (
	SysExStart = 0xF0
	SysExEnd   = 0xF7
)

// SyxConverter handles .syx file parsing and generation
type SyxConverter struct {
	device Device
}

// NewSyxConverter creates a new .syx converter
func NewSyxConverter(device Device) *SyxConverter {
	return &SyxConverter{device: device}
}

// ParseSyxFile reads a .syx file and returns a Pattern
func (s *SyxConverter) ParseSyxFile(filename string) (*Pattern, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read syx file: %w", err)
	}
	return s.ParseSyx(data)
}

// ParseSyx parses .syx data and returns a Pattern
func (s *SyxConverter) ParseSyx(data []byte) (*Pattern, error) {
	if s.device == nil {
		return nil, errors.New("no device configured")
	}
	
	// Validate SysEx structure first
	if err := s.ValidateSyx(data); err != nil {
		return nil, err
	}
	
	return s.device.ParseSyx(data)
}

// GenerateSyx creates .syx data from a Pattern
func (s *SyxConverter) GenerateSyx(pattern *Pattern) ([]byte, error) {
	if s.device == nil {
		return nil, errors.New("no device configured")
	}
	return s.device.GenerateSyx(pattern)
}

// WriteSyxFile writes .syx data to a file
func (s *SyxConverter) WriteSyxFile(pattern *Pattern, filename string) error {
	data, err := s.GenerateSyx(pattern)
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}

// ValidateSyx validates .syx data structure
func (s *SyxConverter) ValidateSyx(data []byte) error {
	if len(data) < 2 {
		return errors.New("syx data too short")
	}
	
	if data[0] != SysExStart {
		return fmt.Errorf("invalid SysEx: expected start byte 0x%02X, got 0x%02X", SysExStart, data[0])
	}
	
	if data[len(data)-1] != SysExEnd {
		return fmt.Errorf("invalid SysEx: expected end byte 0x%02X, got 0x%02X", SysExEnd, data[len(data)-1])
	}
	
	// Check all data bytes are 7-bit (valid MIDI data)
	for i := 1; i < len(data)-1; i++ {
		if data[i] > 127 {
			return fmt.Errorf("invalid SysEx: byte at position %d is > 127 (0x%02X)", i, data[i])
		}
	}
	
	return nil
}

// ExtractManufacturerID extracts the manufacturer ID from SysEx data
func ExtractManufacturerID(data []byte) ([]byte, error) {
	if len(data) < 4 {
		return nil, errors.New("syx data too short for manufacturer ID")
	}
	
	if data[0] != SysExStart {
		return nil, errors.New("invalid SysEx start")
	}
	
	// Check if extended manufacturer ID (starts with 0x00)
	if data[1] == 0x00 {
		if len(data) < 5 {
			return nil, errors.New("syx data too short for extended manufacturer ID")
		}
		return data[1:4], nil
	}
	
	// Single byte manufacturer ID
	return data[1:2], nil
}

// IsBehringerSyx checks if the SysEx data is from a Behringer device
func IsBehringerSyx(data []byte) bool {
	if len(data) < 5 {
		return false
	}
	
	// Behringer extended manufacturer ID: 00 20 32
	return data[0] == SysExStart &&
		data[1] == 0x00 &&
		data[2] == 0x20 &&
		data[3] == 0x32
}

