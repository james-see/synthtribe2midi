package devices

import (
	"testing"

	"github.com/james-see/synthtribe2midi/pkg/converter"
)

func TestTD3Name(t *testing.T) {
	td3 := NewTD3()
	if td3.Name() != "Behringer TD-3" {
		t.Errorf("Name() = %q, want %q", td3.Name(), "Behringer TD-3")
	}
}

func TestTD3ID(t *testing.T) {
	td3 := NewTD3()
	if td3.ID() != TD3DeviceID {
		t.Errorf("ID() = %d, want %d", td3.ID(), TD3DeviceID)
	}
}

func TestTD3GenerateSeq(t *testing.T) {
	td3 := NewTD3()

	pattern := &converter.Pattern{
		Name:   "Test",
		Length: 16,
		Steps: []converter.Step{
			{Note: 60, Gate: true, Accent: false, Slide: false, Velocity: 100},  // C3
			{Note: 62, Gate: true, Accent: true, Slide: false, Velocity: 127},   // D3
			{Note: 64, Gate: true, Accent: false, Slide: true, Velocity: 100},   // E3
			{Note: 65, Gate: true, Accent: false, Slide: false, Tie: true, Velocity: 100}, // F3
		},
	}

	data, err := td3.GenerateSeq(pattern)
	if err != nil {
		t.Fatalf("GenerateSeq() error = %v", err)
	}

	// Should be TD3SeqMinSize (146 bytes)
	if len(data) != TD3SeqMinSize {
		t.Errorf("GenerateSeq() data length = %d, want %d", len(data), TD3SeqMinSize)
	}

	// Check header magic
	if data[0] != 0x23 || data[1] != 0x98 || data[2] != 0x54 || data[3] != 0x76 {
		t.Errorf("Header magic = %02X %02X %02X %02X, want 23 98 54 76",
			data[0], data[1], data[2], data[3])
	}

	// Check first note (C3 = MIDI 60, stored as 60-24=36 = 0x24 -> nibbles 02 04)
	noteVal := int(data[NotesOffset])*16 + int(data[NotesOffset+1])
	expectedNote := 60 - 24 // 36
	if noteVal != expectedNote {
		t.Errorf("Step 0 note value = %d, want %d", noteVal, expectedNote)
	}
}

func TestTD3ParseSeq(t *testing.T) {
	td3 := NewTD3()

	// Create a valid TD3 seq file structure
	data := make([]byte, TD3SeqMinSize)

	// Header magic
	data[0] = 0x23
	data[1] = 0x98
	data[2] = 0x54
	data[3] = 0x76

	// Device info
	data[4] = 0x00
	data[5] = 0x00
	data[6] = 0x00
	data[7] = 0x08
	data[8] = 0x00
	data[9] = 0x54  // 'T'
	data[10] = 0x00
	data[11] = 0x44 // 'D'
	data[12] = 0x00
	data[13] = 0x2d // '-'
	data[14] = 0x00
	data[15] = 0x33 // '3'

	// Fill remaining header
	for i := 16; i < 32; i++ {
		data[i] = 0x00
	}

	// Fill bytes
	data[32] = 0x00
	data[33] = 0x70

	// Set sequence length to 4
	data[LengthOffset] = 0x00
	data[LengthOffset+1] = 0x04

	// Set tie bitmask (all notes are new, no ties) = 0xFFFF
	data[TieOffset] = 0x0F
	data[TieOffset+1] = 0x0F
	data[TieOffset+2] = 0x0F
	data[TieOffset+3] = 0x0F

	// Set notes: C3, D3, E3, F3 (MIDI 60-24=36, 62-24=38, 64-24=40, 65-24=41)
	// Note 36 = 0x24 -> nibbles 02, 04
	data[NotesOffset] = 0x02
	data[NotesOffset+1] = 0x04
	// Note 38 = 0x26 -> nibbles 02, 06
	data[NotesOffset+2] = 0x02
	data[NotesOffset+3] = 0x06
	// Note 40 = 0x28 -> nibbles 02, 08
	data[NotesOffset+4] = 0x02
	data[NotesOffset+5] = 0x08
	// Note 41 = 0x29 -> nibbles 02, 09
	data[NotesOffset+6] = 0x02
	data[NotesOffset+7] = 0x09

	// Set accent on step 2
	data[AccentsOffset+3] = 0x01

	// Set slide on step 3
	data[SlidesOffset+5] = 0x01

	pattern, err := td3.ParseSeq(data)
	if err != nil {
		t.Fatalf("ParseSeq() error = %v", err)
	}

	if len(pattern.Steps) != 4 {
		t.Errorf("ParseSeq() steps = %d, want 4", len(pattern.Steps))
	}

	// Check first step (C3)
	if pattern.Steps[0].Note != 60 {
		t.Errorf("Step 0 note = %d, want 60", pattern.Steps[0].Note)
	}
	if !pattern.Steps[0].Gate {
		t.Error("Step 0 should have gate")
	}

	// Check second step (accent)
	if !pattern.Steps[1].Accent {
		t.Error("Step 1 should have accent")
	}

	// Check third step (slide)
	if !pattern.Steps[2].Slide {
		t.Error("Step 2 should have slide")
	}
}

func TestTD3GenerateSyx(t *testing.T) {
	td3 := NewTD3()

	pattern := &converter.Pattern{
		Name:   "Test",
		Length: 16,
		Steps: []converter.Step{
			{Note: 60, Gate: true, Velocity: 100},
		},
	}

	data, err := td3.GenerateSyx(pattern)
	if err != nil {
		t.Fatalf("GenerateSyx() error = %v", err)
	}

	// Check SysEx structure
	if data[0] != SysExStart {
		t.Errorf("SysEx start = 0x%02X, want 0x%02X", data[0], SysExStart)
	}

	if data[len(data)-1] != SysExEnd {
		t.Errorf("SysEx end = 0x%02X, want 0x%02X", data[len(data)-1], SysExEnd)
	}

	// Check manufacturer ID (Behringer: 00 20 32)
	if data[1] != 0x00 || data[2] != TD3Manufacturer || data[3] != TD3ManufID2 {
		t.Errorf("Manufacturer ID = %02X %02X %02X, want 00 20 32", data[1], data[2], data[3])
	}
}

func TestTD3ParseSyxInvalid(t *testing.T) {
	td3 := NewTD3()

	tests := []struct {
		name string
		data []byte
	}{
		{"empty", []byte{}},
		{"too short", []byte{0xF0, 0xF7}},
		{"no start byte", []byte{0x00, 0x20, 0x32, 0xF7}},
		{"no end byte", []byte{0xF0, 0x00, 0x20, 0x32}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := td3.ParseSyx(tt.data)
			if err == nil {
				t.Error("ParseSyx() expected error for invalid data")
			}
		})
	}
}

func TestTD3RoundTrip(t *testing.T) {
	td3 := NewTD3()

	// Create original pattern
	original := &converter.Pattern{
		Name:   "Test",
		Length: 16,
		Steps:  make([]converter.Step, 16),
	}

	// Set some steps with MIDI notes in valid range (24-127)
	original.Steps[0] = converter.Step{Note: 48, Gate: true, Accent: false, Slide: false, Velocity: 100}  // C2
	original.Steps[1] = converter.Step{Note: 50, Gate: true, Accent: true, Slide: false, Velocity: 127}   // D2
	original.Steps[4] = converter.Step{Note: 52, Gate: true, Accent: false, Slide: true, Velocity: 100}   // E2
	original.Steps[8] = converter.Step{Note: 53, Gate: true, Accent: false, Slide: false, Velocity: 100}  // F2

	// Generate seq data
	seqData, err := td3.GenerateSeq(original)
	if err != nil {
		t.Fatalf("GenerateSeq() error = %v", err)
	}

	// Parse it back
	parsed, err := td3.ParseSeq(seqData)
	if err != nil {
		t.Fatalf("ParseSeq() error = %v", err)
	}

	// Verify key properties preserved
	if parsed.Steps[0].Note != original.Steps[0].Note {
		t.Errorf("Round trip: step 0 note = %d, want %d", parsed.Steps[0].Note, original.Steps[0].Note)
	}
	if parsed.Steps[1].Accent != original.Steps[1].Accent {
		t.Errorf("Round trip: step 1 accent = %v, want %v", parsed.Steps[1].Accent, original.Steps[1].Accent)
	}
	if parsed.Steps[4].Slide != original.Steps[4].Slide {
		t.Errorf("Round trip: step 4 slide = %v, want %v", parsed.Steps[4].Slide, original.Steps[4].Slide)
	}
}
