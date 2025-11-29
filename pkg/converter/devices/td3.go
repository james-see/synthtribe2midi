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

	// TD3 SEQ file offsets (based on CraveSeq project)
	HeaderSize      = 32
	FillSize        = 4
	NotesOffset     = HeaderSize + FillSize           // 36
	AccentsOffset   = NotesOffset + 32                // 68
	SlidesOffset    = AccentsOffset + 32              // 100
	TripletOffset   = SlidesOffset + 32               // 132
	LengthOffset    = TripletOffset + 2               // 134
	ReservedOffset  = LengthOffset + 2                // 136
	TieOffset       = ReservedOffset + 2              // 138
	RestOffset      = TieOffset + 4                   // 142
	TD3SeqMinSize   = RestOffset + 4                  // 146 bytes minimum
)

// SysEx message types
const (
	SysExStart     = 0xF0
	SysExEnd       = 0xF7
	PatternDump    = 0x40
	PatternRequest = 0x41
)

// TD3 header magic bytes
var td3HeaderMagic = []byte{0x23, 0x98, 0x54, 0x76}

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
// Format based on https://github.com/claziss/CraveSeq
func (t *TD3) ParseSeq(data []byte) (*converter.Pattern, error) {
	// Check minimum size
	if len(data) < TD3SeqMinSize {
		return nil, fmt.Errorf("seq data too short: got %d bytes, need at least %d", len(data), TD3SeqMinSize)
	}

	// Verify header magic
	if data[0] != td3HeaderMagic[0] || data[1] != td3HeaderMagic[1] ||
		data[2] != td3HeaderMagic[2] || data[3] != td3HeaderMagic[3] {
		return nil, errors.New("invalid TD-3 seq file: wrong magic bytes")
	}

	// Get sequence length from file
	seqLength := int(data[LengthOffset])*16 + int(data[LengthOffset+1])
	if seqLength == 0 || seqLength > MaxSteps {
		seqLength = MaxSteps
	}

	// Parse tie and rest bitmasks (4 bytes each, little-endian nibble format)
	tie := uint32(data[TieOffset+1]) + uint32(data[TieOffset])<<4 +
		uint32(data[TieOffset+3])<<8 + uint32(data[TieOffset+2])<<12
	rest := uint32(data[RestOffset+1]) + uint32(data[RestOffset])<<4 +
		uint32(data[RestOffset+3])<<8 + uint32(data[RestOffset+2])<<12

	pattern := &converter.Pattern{
		Name:     "TD-3 Pattern",
		DeviceID: TD3DeviceID,
		Steps:    make([]converter.Step, seqLength),
		Length:   seqLength,
		Tempo:    120.0, // Default tempo
	}

	// Parse notes, accents, and slides
	for i := 0; i < seqLength; i++ {
		noteIdx := NotesOffset + i*2
		accentIdx := AccentsOffset + i*2
		slideIdx := SlidesOffset + i*2

		// Note value = high nibble * 16 + low nibble
		noteVal := int(data[noteIdx])*16 + int(data[noteIdx+1])

		// Convert to MIDI note (TD-3 octave 0 = MIDI octave 2, so add 24)
		// Actually the note value already encodes octave, so:
		// noteVal = note + octave*12, where octave starts from 0
		// MIDI note = noteVal + 24 (to shift to reasonable bass range)
		midiNote := uint8(noteVal + 24)
		if midiNote > 127 {
			midiNote = 127
		}

		// Check if this step is a rest
		isRest := (rest & (1 << i)) != 0

		// Check if this is a tie (sustain from previous note)
		isTie := (tie & (1 << i)) == 0 // 0 means sustain/tie

		// Accent is in second byte of accent pair
		hasAccent := (data[accentIdx+1] & 0x01) != 0

		// Slide is in second byte of slide pair
		hasSlide := (data[slideIdx+1] & 0x01) != 0

		step := converter.Step{
			Note:     midiNote,
			Gate:     !isRest,
			Accent:   hasAccent,
			Slide:    hasSlide,
			Tie:      isTie && i > 0, // Can't tie first note
			Velocity: 100,
		}

		if hasAccent {
			step.Velocity = 127
		}

		pattern.Steps[i] = step
	}

	return pattern, nil
}

// GenerateSeq generates .seq data from a Pattern
func (t *TD3) GenerateSeq(pattern *converter.Pattern) ([]byte, error) {
	if pattern == nil {
		return nil, errors.New("nil pattern")
	}

	// Allocate full TD3 seq buffer
	data := make([]byte, TD3SeqMinSize)

	// Write header magic
	copy(data[0:4], td3HeaderMagic)

	// Write device name "TD-3" in UTF-16LE
	data[4] = 0x00
	data[5] = 0x00
	data[6] = 0x00
	data[7] = 0x08 // Device name length
	data[8] = 0x00
	data[9] = 0x54 // 'T'
	data[10] = 0x00
	data[11] = 0x44 // 'D'
	data[12] = 0x00
	data[13] = 0x2d // '-'
	data[14] = 0x00
	data[15] = 0x33 // '3'

	// Version info
	data[16] = 0x00
	data[17] = 0x00
	data[18] = 0x00
	data[19] = 0x0a
	data[20] = 0x00
	data[21] = 0x31 // '1'
	data[22] = 0x00
	data[23] = 0x2e // '.'
	data[24] = 0x00
	data[25] = 0x33 // '3'
	data[26] = 0x00
	data[27] = 0x2e // '.'
	data[28] = 0x00
	data[29] = 0x37 // '7'
	data[30] = 0x00
	data[31] = 0x00

	// Fill/length field
	data[32] = 0x00
	data[33] = 0x70 // 112 bytes remaining
	data[34] = 0x00
	data[35] = 0x00

	seqLength := len(pattern.Steps)
	if seqLength > MaxSteps {
		seqLength = MaxSteps
	}

	var tie, rest uint32

	for i := 0; i < MaxSteps; i++ {
		var step converter.Step
		if i < len(pattern.Steps) {
			step = pattern.Steps[i]
		}

		// Convert MIDI note back to TD-3 format (subtract 24)
		noteVal := int(step.Note) - 24
		if noteVal < 0 {
			noteVal = 0
		}

		// Write note (2 bytes: high nibble, low nibble)
		data[NotesOffset+i*2] = byte(noteVal / 16)
		data[NotesOffset+i*2+1] = byte(noteVal % 16)

		// Write accent (2 bytes, flag in second byte)
		if step.Accent {
			data[AccentsOffset+i*2+1] = 0x01
		}

		// Write slide (2 bytes, flag in second byte)
		if step.Slide {
			data[SlidesOffset+i*2+1] = 0x01
		}

		// Build tie bitmask (1 = new note, 0 = sustain)
		if !step.Tie {
			tie |= (1 << i)
		}

		// Build rest bitmask
		if !step.Gate {
			rest |= (1 << i)
		}
	}

	// Write sequence length
	data[LengthOffset] = byte(seqLength / 16)
	data[LengthOffset+1] = byte(seqLength % 16)

	// Write tie bitmask (4 bytes, nibble format)
	data[TieOffset] = byte((tie >> 4) & 0x0F)
	data[TieOffset+1] = byte(tie & 0x0F)
	data[TieOffset+2] = byte((tie >> 12) & 0x0F)
	data[TieOffset+3] = byte((tie >> 8) & 0x0F)

	// Write rest bitmask (4 bytes, nibble format)
	data[RestOffset] = byte((rest >> 4) & 0x0F)
	data[RestOffset+1] = byte(rest & 0x0F)
	data[RestOffset+2] = byte((rest >> 12) & 0x0F)
	data[RestOffset+3] = byte((rest >> 8) & 0x0F)

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
		Tempo:    120.0,
	}

	// Skip header bytes (F0, manufacturer ID, device ID, model ID, command)
	headerLen := 8
	if len(data) < headerLen+MaxSteps*2 {
		return nil, fmt.Errorf("syx data too short: got %d, need at least %d", len(data), headerLen+MaxSteps*2)
	}

	// Parse step data from SysEx payload
	for i := 0; i < MaxSteps; i++ {
		offset := headerLen + i*2
		if offset+1 >= len(data)-1 {
			break
		}

		noteData := data[offset]
		attrData := data[offset+1]

		step := converter.Step{
			Note:     (noteData & 0x7F) + 24, // Add octave offset
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

		// Note byte (subtract octave offset)
		noteVal := step.Note
		if noteVal >= 24 {
			noteVal -= 24
		}
		noteByte := noteVal & 0x7F
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
