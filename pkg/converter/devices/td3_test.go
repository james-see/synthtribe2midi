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
			{Note: 60, Gate: true, Accent: false, Slide: false, Velocity: 100},
			{Note: 62, Gate: true, Accent: true, Slide: false, Velocity: 127},
			{Note: 64, Gate: true, Accent: false, Slide: true, Velocity: 100},
			{Note: 65, Gate: true, Accent: false, Slide: false, Tie: true, Velocity: 100},
		},
	}

	data, err := td3.GenerateSeq(pattern)
	if err != nil {
		t.Fatalf("GenerateSeq() error = %v", err)
	}

	// Should be 32 bytes (16 steps * 2 bytes)
	if len(data) != 32 {
		t.Errorf("GenerateSeq() data length = %d, want 32", len(data))
	}

	// Check first step
	if data[0] != 60 { // Note
		t.Errorf("Step 0 note = %d, want 60", data[0])
	}
	if data[1] != 0x01 { // Gate only
		t.Errorf("Step 0 attr = 0x%02X, want 0x01", data[1])
	}

	// Check second step (accent)
	if data[2] != 62 {
		t.Errorf("Step 1 note = %d, want 62", data[2])
	}
	if data[3] != 0x03 { // Gate + Accent
		t.Errorf("Step 1 attr = 0x%02X, want 0x03", data[3])
	}

	// Check third step (slide)
	if data[5] != 0x05 { // Gate + Slide
		t.Errorf("Step 2 attr = 0x%02X, want 0x05", data[5])
	}

	// Check fourth step (tie)
	if data[7] != 0x09 { // Gate + Tie
		t.Errorf("Step 3 attr = 0x%02X, want 0x09", data[7])
	}
}

func TestTD3ParseSeq(t *testing.T) {
	td3 := NewTD3()

	// Create test data: 16 steps, each 2 bytes
	data := make([]byte, 32)
	// Step 0: C4 with gate
	data[0] = 60
	data[1] = 0x01
	// Step 1: D4 with gate + accent
	data[2] = 62
	data[3] = 0x03
	// Step 2: E4 with gate + slide
	data[4] = 64
	data[5] = 0x05
	// Step 3: F4 with gate + tie
	data[6] = 65
	data[7] = 0x09

	pattern, err := td3.ParseSeq(data)
	if err != nil {
		t.Fatalf("ParseSeq() error = %v", err)
	}

	if len(pattern.Steps) != 16 {
		t.Errorf("ParseSeq() steps = %d, want 16", len(pattern.Steps))
	}

	// Check first step
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

	// Check fourth step (tie)
	if !pattern.Steps[3].Tie {
		t.Error("Step 3 should have tie")
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
		Steps: make([]converter.Step, 16),
	}

	// Set some steps
	original.Steps[0] = converter.Step{Note: 60, Gate: true, Accent: false, Slide: false, Velocity: 100}
	original.Steps[1] = converter.Step{Note: 62, Gate: true, Accent: true, Slide: false, Velocity: 127}
	original.Steps[4] = converter.Step{Note: 64, Gate: true, Accent: false, Slide: true, Velocity: 100}
	original.Steps[8] = converter.Step{Note: 65, Gate: true, Accent: false, Slide: false, Tie: true, Velocity: 100}

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
	if parsed.Steps[8].Tie != original.Steps[8].Tie {
		t.Errorf("Round trip: step 8 tie = %v, want %v", parsed.Steps[8].Tie, original.Steps[8].Tie)
	}
}

