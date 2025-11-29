# synthtribe2midi

[![CI](https://github.com/james-see/synthtribe2midi/actions/workflows/ci.yml/badge.svg)](https://github.com/james-see/synthtribe2midi/actions/workflows/ci.yml)
[![Release](https://github.com/james-see/synthtribe2midi/actions/workflows/release.yml/badge.svg)](https://github.com/james-see/synthtribe2midi/actions/workflows/release.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/james-see/synthtribe2midi)](https://goreportcard.com/report/github.com/james-see/synthtribe2midi)

Convert between MIDI and Behringer SynthTribe formats (.seq/.syx).

![TUI Demo](docs/tui-demo.gif)

## Features

- **Bidirectional conversion**: MIDI ↔ .seq ↔ .syx
- **TD-3 support**: Full support for Behringer TD-3 (TB-303 clone) patterns
- **Multiple interfaces**: CLI, TUI, and REST API server
- **Cross-platform**: macOS, Linux, Windows (amd64/arm64)

## Installation

### Homebrew (macOS/Linux)

```bash
brew install james-see/tap/synthtribe2midi
```

### Download Binary

Download the latest release from the [releases page](https://github.com/james-see/synthtribe2midi/releases).

### Build from Source

```bash
go install github.com/james-see/synthtribe2midi/cmd/synthtribe2midi@latest
```

## Usage

### Command Line

```bash
# Auto-detect format and convert
synthtribe2midi convert pattern.mid -o pattern.seq

# Explicit conversions
synthtribe2midi midi2seq pattern.mid -o pattern.seq
synthtribe2midi seq2midi pattern.seq -o pattern.mid
synthtribe2midi midi2syx pattern.mid -o pattern.syx
synthtribe2midi syx2midi pattern.syx -o pattern.mid

# Launch interactive TUI
synthtribe2midi tui

# Start API server
synthtribe2midi serve --port 8080
```

### Interactive TUI

Launch the terminal UI for a guided conversion experience:

```bash
synthtribe2midi tui
```

Features:
- File browser with format filtering
- Conversion progress visualization
- Acid-inspired color scheme

### REST API

Start the server:

```bash
synthtribe2midi serve --port 8080
```

Endpoints:

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/convert/midi2seq` | Convert MIDI to .seq |
| POST | `/api/v1/convert/seq2midi` | Convert .seq to MIDI |
| POST | `/api/v1/convert/midi2syx` | Convert MIDI to .syx |
| POST | `/api/v1/convert/syx2midi` | Convert .syx to MIDI |
| POST | `/api/v1/convert/seq2syx` | Convert .seq to .syx |
| POST | `/api/v1/convert/syx2seq` | Convert .syx to .seq |
| GET | `/api/v1/health` | Health check |
| GET | `/api/v1/formats` | List supported formats |
| GET | `/api/v1/devices` | List supported devices |

Example:

```bash
curl -X POST http://localhost:8080/api/v1/convert/midi2seq \
  -F "file=@pattern.mid" \
  -o pattern.seq
```

Swagger documentation available at `http://localhost:8080/swagger/index.html`

### As a Go Library

```go
package main

import (
    "os"
    "github.com/james-see/synthtribe2midi/pkg/converter"
    "github.com/james-see/synthtribe2midi/pkg/converter/devices"
)

func main() {
    // Create converter with TD-3 device
    device := devices.NewTD3()
    conv := converter.New(device)
    
    // Read MIDI file
    midiData, _ := os.ReadFile("pattern.mid")
    
    // Convert to .seq
    seqData, _ := conv.MIDIToSeq(midiData)
    
    // Write output
    os.WriteFile("pattern.seq", seqData, 0644)
}
```

## Supported Devices

- **Behringer TD-3** (TB-303 clone) - Full support
- More devices planned (PRO-VS MINI, VICTOR)

## Format Reference

For detailed format documentation, see **[TD-3 SEQ Format Specification](docs/TD3_SEQ_FORMAT.md)**.

### .seq Format (TD-3)

SynthTribe's 146-byte binary pattern format:
- **Header** (36 bytes): Magic bytes, device name "TD-3", version
- **Notes** (32 bytes): 16 steps × 2 bytes (nibble-encoded pitch)
- **Accents** (32 bytes): 16 steps × 2 bytes (flag in second byte)
- **Slides** (32 bytes): 16 steps × 2 bytes (flag in second byte)
- **Control** (14 bytes): Triplet flag, length, tie/rest bitmasks

Note values are offset by 24 from MIDI (TD-3 octave 0 = MIDI note 24).

### .syx Format

Standard SysEx format with Behringer manufacturer ID (00 20 32).

## Development

```bash
# Clone
git clone https://github.com/james-see/synthtribe2midi
cd synthtribe2midi

# Build
go build ./cmd/synthtribe2midi

# Test
go test ./...

# Run TUI
go run ./cmd/synthtribe2midi tui
```

## License

MIT License - see [LICENSE](LICENSE)

## Credits

- Inspired by [Acid-Injector](https://github.com/echolevel/Acid-Injector) by echolevel
- Uses [gomidi](https://gitlab.com/gomidi/midi) for MIDI parsing
- TUI powered by [Bubble Tea](https://github.com/charmbracelet/bubbletea)

