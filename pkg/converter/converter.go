package converter

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Format represents a file format
type Format string

const (
	FormatMIDI    Format = "midi"
	FormatSeq     Format = "seq"
	FormatSyx     Format = "syx"
	FormatUnknown Format = "unknown"
)

// DetectFormat detects the format of a file based on extension and content
func DetectFormat(filename string) Format {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".mid", ".midi":
		return FormatMIDI
	case ".seq":
		return FormatSeq
	case ".syx":
		return FormatSyx
	default:
		return FormatUnknown
	}
}

// DetectFormatFromContent detects format from file content
func DetectFormatFromContent(data []byte) Format {
	if len(data) < 4 {
		return FormatUnknown
	}

	// Check for MIDI file signature "MThd"
	if string(data[:4]) == "MThd" {
		return FormatMIDI
	}

	// Check for SysEx (starts with F0)
	if data[0] == SysExStart {
		return FormatSyx
	}

	// Assume .seq format for other binary data
	return FormatSeq
}

// ConvertFile converts a file from one format to another
func (c *Converter) ConvertFile(inputPath, outputPath string) error {
	inputFormat := DetectFormat(inputPath)
	outputFormat := DetectFormat(outputPath)

	if inputFormat == FormatUnknown {
		// Try to detect from content
		data, err := os.ReadFile(inputPath)
		if err != nil {
			return fmt.Errorf("failed to read input file: %w", err)
		}
		inputFormat = DetectFormatFromContent(data)
	}

	if outputFormat == FormatUnknown {
		return errors.New("cannot determine output format from filename")
	}

	// Read input
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	// Convert based on input/output formats
	var outputData []byte
	
	switch {
	case inputFormat == FormatMIDI && outputFormat == FormatSeq:
		outputData, err = c.MIDIToSeq(data)
	case inputFormat == FormatMIDI && outputFormat == FormatSyx:
		outputData, err = c.MIDIToSyx(data)
	case inputFormat == FormatSeq && outputFormat == FormatMIDI:
		outputData, err = c.SeqToMIDI(data)
	case inputFormat == FormatSeq && outputFormat == FormatSyx:
		outputData, err = c.SeqToSyx(data)
	case inputFormat == FormatSyx && outputFormat == FormatMIDI:
		outputData, err = c.SyxToMIDI(data)
	case inputFormat == FormatSyx && outputFormat == FormatSeq:
		outputData, err = c.SyxToSeq(data)
	default:
		return fmt.Errorf("unsupported conversion: %s to %s", inputFormat, outputFormat)
	}

	if err != nil {
		return fmt.Errorf("conversion failed: %w", err)
	}

	// Write output
	if err := os.WriteFile(outputPath, outputData, 0644); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	return nil
}

// MIDIToSeq converts MIDI data to .seq format
func (c *Converter) MIDIToSeq(midiData []byte) ([]byte, error) {
	midiConv := NewMIDIConverter()
	pattern, err := midiConv.ParseMIDI(midiData)
	if err != nil {
		return nil, err
	}
	return c.device.GenerateSeq(pattern)
}

// MIDIToSyx converts MIDI data to .syx format
func (c *Converter) MIDIToSyx(midiData []byte) ([]byte, error) {
	midiConv := NewMIDIConverter()
	pattern, err := midiConv.ParseMIDI(midiData)
	if err != nil {
		return nil, err
	}
	return c.device.GenerateSyx(pattern)
}

// SeqToMIDI converts .seq data to MIDI format
func (c *Converter) SeqToMIDI(seqData []byte) ([]byte, error) {
	pattern, err := c.device.ParseSeq(seqData)
	if err != nil {
		return nil, err
	}
	midiConv := NewMIDIConverter()
	return midiConv.GenerateMIDI(pattern)
}

// SeqToSyx converts .seq data to .syx format
func (c *Converter) SeqToSyx(seqData []byte) ([]byte, error) {
	pattern, err := c.device.ParseSeq(seqData)
	if err != nil {
		return nil, err
	}
	return c.device.GenerateSyx(pattern)
}

// SyxToMIDI converts .syx data to MIDI format
func (c *Converter) SyxToMIDI(syxData []byte) ([]byte, error) {
	pattern, err := c.device.ParseSyx(syxData)
	if err != nil {
		return nil, err
	}
	midiConv := NewMIDIConverter()
	return midiConv.GenerateMIDI(pattern)
}

// SyxToSeq converts .syx data to .seq format
func (c *Converter) SyxToSeq(syxData []byte) ([]byte, error) {
	pattern, err := c.device.ParseSyx(syxData)
	if err != nil {
		return nil, err
	}
	return c.device.GenerateSeq(pattern)
}

// GetSupportedConversions returns a list of supported conversion paths
func GetSupportedConversions() []string {
	return []string{
		"midi -> seq",
		"midi -> syx",
		"seq -> midi",
		"seq -> syx",
		"syx -> midi",
		"syx -> seq",
	}
}

