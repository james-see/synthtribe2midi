// Package devices provides device-specific format handlers
package devices

import (
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/james-see/synthtribe2midi/pkg/converter"
)

// TD3 device constants
const (
	TD3DeviceID     = 0x00 // TD-3 device ID in SysEx
	TD3Manufacturer = 0x20 // Behringer manufacturer ID (part 1)
	TD3ManufID2     = 0x32 // Behringer manufacturer ID (part 2)
	TD3ManufID3     = 0x00 // Behringer manufacturer ID (part 3)
	TD3ModelID      = 0x01 // TD-3 model ID
	MaxSteps        = 16
	MaxPatterns     = 64
)

// SysEx message types
const (
	SysExStart     = 0xF0
	SysExEnd       = 0xF7
	PatternDump    = 0x40
	PatternRequest = 0x41
)

// TD3 implements the Device interface for Behringer TD-3
type TD3 struct{}

// NewTD3 creates a new TD-3 device handler
func NewTD3() *TD3 {
	return &TD3{}
}

// Name returns the device name
func (t *TD3) Name() string {
	return "Behringer TD-3"
}

// ID returns the device ID
func (t *TD3) ID() uint8 {
	return TD3DeviceID
}

// ParseSeq parses a .seq file into a Pattern
// The .seq format is SynthTribe's pattern format
// Based on analysis from Acid-Injector project
func (t *TD3) ParseSeq(data []byte) (*converter.Pattern, error) {
	if len(data) < 32 {
		return nil, errors.New("seq data too short")
	}

	pattern := &converter.Pattern{
		Name:     "TD-3 Pattern",
		DeviceID: TD3DeviceID,
		Steps:    make([]converter.Step, 0, MaxSteps),
		Length:   MaxSteps,
	}

	// .seq format structure (based on Acid-Injector analysis):
	// Each step uses 2 bytes:
	// Byte 0: Note number (0-127) with flags in high bits
	// Byte 1: Attributes (accent, slide, tie, gate)
	
	for i := 0; i < MaxSteps && i*2+1 < len(data); i++ {
		noteData := data[i*2]
		attrData := data[i*2+1]

		step := converter.Step{
			Note:     noteData & 0x7F,        // Lower 7 bits = note
			Gate:     (attrData & 0x01) != 0, // Bit 0 = gate
			Accent:   (attrData & 0x02) != 0, // Bit 1 = accent
			Slide:    (attrData & 0x04) != 0, // Bit 2 = slide
			Tie:      (attrData & 0x08) != 0, // Bit 3 = tie
			Velocity: 100,                    // Default velocity
		}
		
		if step.Accent {
			step.Velocity = 127
		}
		
		pattern.Steps = append(pattern.Steps, step)
	}

	return pattern, nil
}

// GenerateSeq generates .seq data from a Pattern
func (t *TD3) GenerateSeq(pattern *converter.Pattern) ([]byte, error) {
	if pattern == nil {
		return nil, errors.New("nil pattern")
	}

	// Allocate buffer for 16 steps (2 bytes each)
	data := make([]byte, MaxSteps*2)

	for i := 0; i < MaxSteps; i++ {
		var step converter.Step
		if i < len(pattern.Steps) {
			step = pattern.Steps[i]
		}

		// Byte 0: Note number
		data[i*2] = step.Note & 0x7F

		// Byte 1: Attributes
		var attr uint8
		if step.Gate {
			attr |= 0x01
		}
		if step.Accent {
			attr |= 0x02
		}
		if step.Slide {
			attr |= 0x04
		}
		if step.Tie {
			attr |= 0x08
		}
		data[i*2+1] = attr
	}

	return data, nil
}

// ParseSyx parses a .syx SysEx file into a Pattern
func (t *TD3) ParseSyx(data []byte) (*converter.Pattern, error) {
	if len(data) < 10 {
		return nil, errors.New("syx data too short")
	}

	// Validate SysEx structure
	if data[0] != SysExStart {
		return nil, errors.New("invalid SysEx: missing start byte")
	}
	if data[len(data)-1] != SysExEnd {
		return nil, errors.New("invalid SysEx: missing end byte")
	}

	// Verify Behringer manufacturer ID
	if len(data) > 4 && data[1] == 0x00 && data[2] == TD3Manufacturer && data[3] == TD3ManufID2 {
		// Extended manufacturer ID format
		return t.parseBehringerSyx(data)
	}

	return nil, errors.New("unrecognized SysEx format")
}

// parseBehringerSyx parses Behringer-specific SysEx format
func (t *TD3) parseBehringerSyx(data []byte) (*converter.Pattern, error) {
	pattern := &converter.Pattern{
		Name:     "TD-3 SysEx Pattern",
		DeviceID: TD3DeviceID,
		Steps:    make([]converter.Step, 0, MaxSteps),
		Length:   MaxSteps,
	}

	// Skip header bytes (F0, manufacturer ID, device ID, model ID, command)
	// Pattern data typically starts around byte 8-10
	headerLen := 8
	if len(data) < headerLen+MaxSteps*2 {
		return nil, fmt.Errorf("syx data too short: got %d, need at least %d", len(data), headerLen+MaxSteps*2)
	}

	// Parse step data from SysEx payload
	// SysEx uses 7-bit encoding, so we need to handle nibblization
	for i := 0; i < MaxSteps; i++ {
		offset := headerLen + i*2
		if offset+1 >= len(data)-1 { // -1 for F7 end byte
			break
		}

		noteData := data[offset]
		attrData := data[offset+1]

		step := converter.Step{
			Note:     noteData & 0x7F,
			Gate:     (attrData & 0x01) != 0,
			Accent:   (attrData & 0x02) != 0,
			Slide:    (attrData & 0x04) != 0,
			Tie:      (attrData & 0x08) != 0,
			Velocity: 100,
		}

		if step.Accent {
			step.Velocity = 127
		}

		pattern.Steps = append(pattern.Steps, step)
	}

	return pattern, nil
}

// GenerateSyx generates .syx SysEx data from a Pattern
func (t *TD3) GenerateSyx(pattern *converter.Pattern) ([]byte, error) {
	if pattern == nil {
		return nil, errors.New("nil pattern")
	}

	// Calculate total message length
	// F0 + manufacturer (3 bytes) + device (1) + model (1) + command (1) + data + checksum + F7
	dataLen := MaxSteps * 2
	totalLen := 1 + 3 + 1 + 1 + 1 + dataLen + 1 + 1

	syx := make([]byte, totalLen)
	idx := 0

	// SysEx start
	syx[idx] = SysExStart
	idx++

	// Behringer manufacturer ID (extended format: 00 20 32)
	syx[idx] = 0x00
	idx++
	syx[idx] = TD3Manufacturer
	idx++
	syx[idx] = TD3ManufID2
	idx++

	// Device ID
	syx[idx] = TD3DeviceID
	idx++

	// Model ID
	syx[idx] = TD3ModelID
	idx++

	// Command (pattern dump)
	syx[idx] = PatternDump
	idx++

	// Pattern data
	var checksum uint8
	for i := 0; i < MaxSteps; i++ {
		var step converter.Step
		if i < len(pattern.Steps) {
			step = pattern.Steps[i]
		}

		// Note byte
		noteByte := step.Note & 0x7F
		syx[idx] = noteByte
		checksum ^= noteByte
		idx++

		// Attribute byte
		var attr uint8
		if step.Gate {
			attr |= 0x01
		}
		if step.Accent {
			attr |= 0x02
		}
		if step.Slide {
			attr |= 0x04
		}
		if step.Tie {
			attr |= 0x08
		}
		syx[idx] = attr
		checksum ^= attr
		idx++
	}

	// Checksum (XOR of all data bytes)
	syx[idx] = checksum & 0x7F
	idx++

	// SysEx end
	syx[idx] = SysExEnd

	return syx, nil
}

// Helper function to ensure binary package is used
var _ = binary.LittleEndian

