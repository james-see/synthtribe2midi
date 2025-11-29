package converter

import (
	"testing"
)

func TestDetectFormat(t *testing.T) {
	tests := []struct {
		filename string
		expected Format
	}{
		{"test.mid", FormatMIDI},
		{"test.midi", FormatMIDI},
		{"test.seq", FormatSeq},
		{"test.syx", FormatSyx},
		{"test.txt", FormatUnknown},
		{"test", FormatUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			result := DetectFormat(tt.filename)
			if result != tt.expected {
				t.Errorf("DetectFormat(%q) = %v, want %v", tt.filename, result, tt.expected)
			}
		})
	}
}

func TestDetectFormatFromContent(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected Format
	}{
		{"MIDI file", []byte("MThd\x00\x00\x00\x06"), FormatMIDI},
		{"SysEx message", []byte{0xF0, 0x00, 0x20, 0x32, 0x00, 0xF7}, FormatSyx},
		{"Short data", []byte{0x00, 0x01}, FormatUnknown},
		{"SEQ data (assumed)", []byte{0x3C, 0x01, 0x3E, 0x02, 0x40, 0x03}, FormatSeq},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectFormatFromContent(tt.data)
			if result != tt.expected {
				t.Errorf("DetectFormatFromContent() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// mockDevice implements Device interface for testing
type mockDevice struct{}

func (m *mockDevice) Name() string { return "Mock Device" }
func (m *mockDevice) ID() uint8   { return 0 }
func (m *mockDevice) ParseSeq(data []byte) (*Pattern, error) {
	return &Pattern{Name: "Mock"}, nil
}
func (m *mockDevice) GenerateSeq(pattern *Pattern) ([]byte, error) {
	return []byte{0x00}, nil
}
func (m *mockDevice) ParseSyx(data []byte) (*Pattern, error) {
	return &Pattern{Name: "Mock"}, nil
}
func (m *mockDevice) GenerateSyx(pattern *Pattern) ([]byte, error) {
	return []byte{0xF0, 0xF7}, nil
}

func TestConverterNew(t *testing.T) {
	device := &mockDevice{}
	conv := New(device)

	if conv == nil {
		t.Fatal("New() returned nil")
	}

	if conv.GetDevice() != device {
		t.Error("GetDevice() did not return the expected device")
	}
}

func TestConverterSetDevice(t *testing.T) {
	device1 := &mockDevice{}
	device2 := &mockDevice{}

	conv := New(device1)
	if conv.GetDevice() != device1 {
		t.Error("GetDevice() should return device1")
	}

	conv.SetDevice(device2)
	if conv.GetDevice() != device2 {
		t.Error("GetDevice() should return device2 after SetDevice")
	}
}

func TestPatternCreation(t *testing.T) {
	pattern := &Pattern{
		Name:   "Test Pattern",
		Length: 16,
		Tempo:  120.0,
		Steps: []Step{
			{Note: 60, Gate: true, Accent: false, Slide: false, Velocity: 100},
			{Note: 0, Gate: false},
			{Note: 62, Gate: true, Accent: true, Slide: false, Velocity: 127},
			{Note: 64, Gate: true, Accent: false, Slide: true, Velocity: 100},
		},
	}

	if pattern.Name != "Test Pattern" {
		t.Errorf("Pattern name = %q, want %q", pattern.Name, "Test Pattern")
	}

	if len(pattern.Steps) != 4 {
		t.Errorf("Pattern steps = %d, want %d", len(pattern.Steps), 4)
	}

	if pattern.Steps[0].Note != 60 {
		t.Errorf("Step 0 note = %d, want %d", pattern.Steps[0].Note, 60)
	}

	if !pattern.Steps[2].Accent {
		t.Error("Step 2 should have accent")
	}

	if !pattern.Steps[3].Slide {
		t.Error("Step 3 should have slide")
	}
}

func TestGetSupportedConversions(t *testing.T) {
	conversions := GetSupportedConversions()

	if len(conversions) != 6 {
		t.Errorf("GetSupportedConversions() returned %d conversions, want 6", len(conversions))
	}

	expected := []string{
		"midi -> seq",
		"midi -> syx",
		"seq -> midi",
		"seq -> syx",
		"syx -> midi",
		"syx -> seq",
	}

	for i, exp := range expected {
		if conversions[i] != exp {
			t.Errorf("conversions[%d] = %q, want %q", i, conversions[i], exp)
		}
	}
}

func TestStepDefaults(t *testing.T) {
	step := Step{}

	if step.Note != 0 {
		t.Errorf("Default Note = %d, want 0", step.Note)
	}
	if step.Gate {
		t.Error("Default Gate should be false")
	}
	if step.Accent {
		t.Error("Default Accent should be false")
	}
	if step.Slide {
		t.Error("Default Slide should be false")
	}
	if step.Tie {
		t.Error("Default Tie should be false")
	}
	if step.Velocity != 0 {
		t.Errorf("Default Velocity = %d, want 0", step.Velocity)
	}
}
