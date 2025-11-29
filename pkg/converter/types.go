// Package converter provides conversion between MIDI and Behringer SynthTribe formats
package converter

// Step represents a single step in a pattern
type Step struct {
	Note     uint8 // MIDI note number (0-127)
	Accent   bool  // Accent flag
	Slide    bool  // Slide/glide flag
	Gate     bool  // Note on/off
	Tie      bool  // Tie to next step
	Velocity uint8 // Velocity (0-127)
}

// Pattern represents a sequence pattern
type Pattern struct {
	Name     string
	Steps    []Step
	Length   int    // Number of steps (typically 16)
	Tempo    float64
	DeviceID uint8
}

// ConversionResult holds the result of a conversion
type ConversionResult struct {
	Data     []byte
	Filename string
	Format   string
	Error    error
}

// Device interface for device-specific format handling
type Device interface {
	Name() string
	ID() uint8
	ParseSeq(data []byte) (*Pattern, error)
	GenerateSeq(pattern *Pattern) ([]byte, error)
	ParseSyx(data []byte) (*Pattern, error)
	GenerateSyx(pattern *Pattern) ([]byte, error)
}

// Converter handles format conversions
type Converter struct {
	device Device
}

// New creates a new Converter with the specified device
func New(device Device) *Converter {
	return &Converter{device: device}
}

// GetDevice returns the current device
func (c *Converter) GetDevice() Device {
	return c.device
}

// SetDevice sets the device for conversion
func (c *Converter) SetDevice(device Device) {
	c.device = device
}

