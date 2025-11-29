package converter

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"

	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/smf"
)

// MIDIConverter handles MIDI file parsing and generation
type MIDIConverter struct {
	ticksPerQuarter uint16
	tempo           float64
}

// NewMIDIConverter creates a new MIDI converter
func NewMIDIConverter() *MIDIConverter {
	return &MIDIConverter{
		ticksPerQuarter: 480,
		tempo:           120.0,
	}
}

// ParseMIDIFile reads a MIDI file and extracts pattern data
func (m *MIDIConverter) ParseMIDIFile(filename string) (*Pattern, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read MIDI file: %w", err)
	}
	return m.ParseMIDI(data)
}

// ParseMIDI parses MIDI data and extracts pattern data
func (m *MIDIConverter) ParseMIDI(data []byte) (*Pattern, error) {
	reader := bytes.NewReader(data)

	s, err := smf.ReadFrom(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to parse MIDI: %w", err)
	}

	// Get ticks per quarter note from time format
	if mt, ok := s.TimeFormat.(smf.MetricTicks); ok {
		m.ticksPerQuarter = mt.Resolution()
	}

	pattern := &Pattern{
		Name:   "MIDI Pattern",
		Steps:  make([]Step, 0, 16),
		Length: 16,
		Tempo:  m.tempo,
	}

	// Calculate ticks per step (assuming 16th notes in a 4/4 bar)
	ticksPerStep := int64(m.ticksPerQuarter) / 4

	// Track note events
	type noteEvent struct {
		tick     int64
		note     uint8
		velocity uint8
		on       bool
	}

	var events []noteEvent
	var currentTick int64

	// Process all tracks
	for _, track := range s.Tracks {
		currentTick = 0
		for _, ev := range track {
			currentTick += int64(ev.Delta)

			msg := ev.Message

			// Check for tempo meta message (FF 51 03 ...)
			if len(msg) >= 6 && msg[0] == 0xFF && msg[1] == 0x51 && msg[2] == 0x03 {
				microsecondsPerBeat := uint32(msg[3])<<16 | uint32(msg[4])<<8 | uint32(msg[5])
				if microsecondsPerBeat > 0 {
					m.tempo = 60000000.0 / float64(microsecondsPerBeat)
					pattern.Tempo = m.tempo
				}
			}

			// Handle note on/off using direct byte parsing
			// Note On: 0x9n nn vv (status, note, velocity)
			// Note Off: 0x8n nn vv (status, note, velocity)
			if len(msg) >= 3 {
				status := msg[0]
				noteNum := msg[1]
				velocity := msg[2]

				// Note On (0x90-0x9F)
				if status >= 0x90 && status <= 0x9F && velocity > 0 {
					events = append(events, noteEvent{
						tick:     currentTick,
						note:     noteNum,
						velocity: velocity,
						on:       true,
					})
				}
				// Note Off (0x80-0x8F) or Note On with velocity 0
				if (status >= 0x80 && status <= 0x8F) || (status >= 0x90 && status <= 0x9F && velocity == 0) {
					events = append(events, noteEvent{
						tick:     currentTick,
						note:     noteNum,
						velocity: 0,
						on:       false,
					})
				}
			}
		}
	}

	// Quantize events to steps
	steps := make([]Step, 16)
	for i := range steps {
		steps[i] = Step{Note: 0, Gate: false}
	}

	// Process note on events
	for _, ev := range events {
		if !ev.on {
			continue
		}

		stepIndex := int(ev.tick / ticksPerStep)
		if stepIndex >= 16 {
			stepIndex = stepIndex % 16
		}

		steps[stepIndex].Note = ev.note
		steps[stepIndex].Gate = true
		steps[stepIndex].Velocity = ev.velocity
		steps[stepIndex].Accent = ev.velocity > 100
	}

	// Detect slides and ties by looking at consecutive notes
	for i := 0; i < 15; i++ {
		if steps[i].Gate && steps[i+1].Gate {
			// If notes are adjacent and the second is the same or close, it might be a slide
			noteDiff := int(steps[i+1].Note) - int(steps[i].Note)
			if noteDiff >= -2 && noteDiff <= 2 && noteDiff != 0 {
				steps[i].Slide = true
			}
			// If same note, it's a tie
			if steps[i].Note == steps[i+1].Note {
				steps[i].Tie = true
			}
		}
	}

	pattern.Steps = steps
	return pattern, nil
}

// GenerateMIDI creates MIDI data from a Pattern
func (m *MIDIConverter) GenerateMIDI(pattern *Pattern) ([]byte, error) {
	if pattern == nil {
		return nil, errors.New("nil pattern")
	}

	if pattern.Tempo <= 0 {
		pattern.Tempo = 120.0
	}

	// Create SMF with one track
	s := smf.New()
	s.TimeFormat = smf.MetricTicks(m.ticksPerQuarter)

	var track smf.Track

	// Add tempo meta event
	microsecondsPerBeat := uint32(60000000.0 / pattern.Tempo)
	tempoData := smf.Message([]byte{
		0xFF, 0x51, 0x03,
		byte(microsecondsPerBeat >> 16),
		byte(microsecondsPerBeat >> 8),
		byte(microsecondsPerBeat),
	})
	track.Add(0, tempoData)

	// Add time signature (4/4)
	timeSigData := smf.Message([]byte{0xFF, 0x58, 0x04, 0x04, 0x02, 0x18, 0x08})
	track.Add(0, timeSigData)

	// Calculate ticks per step (16th notes)
	ticksPerStep := uint32(m.ticksPerQuarter) / 4

	// Default note length (slightly less than full step for non-tied notes)
	defaultNoteLength := ticksPerStep - 10
	if defaultNoteLength > ticksPerStep {
		defaultNoteLength = ticksPerStep - 1
	}

	channel := uint8(0)
	var lastTick uint32

	for i, step := range pattern.Steps {
		if !step.Gate {
			continue
		}

		stepTick := uint32(i) * ticksPerStep
		delta := stepTick - lastTick

		// Note on
		velocity := step.Velocity
		if velocity == 0 {
			velocity = 100
		}
		if step.Accent {
			velocity = 127
		}

		noteOn := midi.NoteOn(channel, step.Note, velocity)
		track.Add(delta, noteOn)
		lastTick = stepTick

		// Calculate note duration
		noteDuration := defaultNoteLength
		if step.Tie && i < len(pattern.Steps)-1 {
			// Extend note to next step
			noteDuration = ticksPerStep
		}
		if step.Slide && i < len(pattern.Steps)-1 {
			// For slides, extend slightly past the next note
			noteDuration = ticksPerStep + 10
		}

		// Note off
		noteOff := midi.NoteOff(channel, step.Note)
		track.Add(noteDuration, noteOff)
		lastTick = stepTick + noteDuration
	}

	// Add end of track
	track.Close(0)

	s.Add(track)

	// Write to buffer
	var buf bytes.Buffer
	_, err := s.WriteTo(&buf)
	if err != nil {
		return nil, fmt.Errorf("failed to write MIDI: %w", err)
	}

	return buf.Bytes(), nil
}

// WriteMIDIFile writes MIDI data to a file
func (m *MIDIConverter) WriteMIDIFile(pattern *Pattern, filename string) error {
	data, err := m.GenerateMIDI(pattern)
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}

// Ensure io.Reader is used (for interface compliance)
var _ io.Reader = (*bytes.Reader)(nil)
