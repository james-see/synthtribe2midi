# Behringer TD-3 SEQ File Format

This document describes the `.seq` file format used by Behringer's SynthTribe software for the TD-3 synthesizer (TB-303 clone).

## Overview

The TD-3 SEQ format is a binary file format that stores pattern sequences. Each file is exactly **146 bytes** and contains a single pattern with up to 16 steps.

## File Structure

| Offset | Size | Description |
|--------|------|-------------|
| 0x00 | 4 | Magic bytes: `23 98 54 76` |
| 0x04 | 4 | Device name length field |
| 0x08 | 8 | Device name "TD-3" (UTF-16LE) |
| 0x10 | 4 | Version length field |
| 0x14 | 12 | Version string (UTF-16LE, e.g., "1.3.7") |
| 0x20 | 4 | Fill/length field |
| 0x24 | 32 | Note data (16 notes × 2 bytes) |
| 0x44 | 32 | Accent flags (16 × 2 bytes) |
| 0x64 | 32 | Slide flags (16 × 2 bytes) |
| 0x84 | 2 | Triplet flag |
| 0x86 | 2 | Sequence length |
| 0x88 | 2 | Reserved |
| 0x8A | 4 | Tie bitmask |
| 0x8E | 4 | Rest bitmask |

**Total: 146 bytes (0x92)**

## Header (Bytes 0x00 - 0x23)

### Magic Bytes (0x00-0x03)
```
23 98 54 76
```
All TD-3 SEQ files start with these 4 bytes.

### Device Name (0x04-0x0F)
```
00 00 00 08   <- Length (8 bytes for "TD-3" in UTF-16)
00 54 00 44 00 2D 00 33   <- "TD-3" in UTF-16LE
```

### Version (0x10-0x1F)
```
00 00 00 0A   <- Length
00 31 00 2E 00 33 00 2E 00 37 00 00   <- "1.3.7" in UTF-16LE + null
```

### Fill Field (0x20-0x23)
```
00 70 00 00   <- 0x70 = 112 bytes remaining after this field
```

## Note Data (Bytes 0x24 - 0x43)

Each of the 16 steps uses 2 bytes for the note value:

```
[High Nibble] [Low Nibble]
```

**Note Value = (High Nibble × 16) + Low Nibble**

The note value represents pitch where:
- **0** = C0
- **12** = C1
- **24** = C2
- **36** = C3
- etc.

To convert to MIDI note number, add 24 (since TD-3 octave 0 ≈ MIDI octave 2):
```
MIDI Note = Note Value + 24
```

### Example
```
02 04   <- Note value = 2×16 + 4 = 36 → MIDI 60 (C3)
02 07   <- Note value = 2×16 + 7 = 39 → MIDI 63 (D#3)
```

## Accent Flags (Bytes 0x44 - 0x63)

Each step has 2 bytes for accent. The accent flag is in the **second byte**:

```
00 [Flag]   <- Flag: 0x01 = accent ON, 0x00 = accent OFF
```

### Example
```
00 00   <- Step 1: No accent
00 01   <- Step 2: Accent ON
00 00   <- Step 3: No accent
```

## Slide Flags (Bytes 0x64 - 0x83)

Same format as accents. The slide flag is in the **second byte**:

```
00 [Flag]   <- Flag: 0x01 = slide ON, 0x00 = slide OFF
```

Slide (also called "glide") causes the note to smoothly transition to the next note, characteristic of the TB-303 sound.

## Control Fields (Bytes 0x84 - 0x91)

### Triplet Flag (0x84-0x85)
```
00 [Flag]   <- 0x01 = triplet timing, 0x00 = normal
```

### Sequence Length (0x86-0x87)
```
[High] [Low]   <- Length = High×16 + Low
```
Values 1-16 for the number of active steps.

### Reserved (0x88-0x89)
Always `00 00`.

### Tie Bitmask (0x8A-0x8D)
4 bytes representing which steps are "tied" (sustain the previous note):

```
[Nibble1] [Nibble0] [Nibble3] [Nibble2]
```

The bitmask is decoded as:
```
Tie = Nibble0 + (Nibble1 << 4) + (Nibble2 << 8) + (Nibble3 << 12)
```

- **Bit = 1**: New note (play the pitch at this step)
- **Bit = 0**: Tie (sustain the previous note)

### Rest Bitmask (0x8E-0x91)
Same format as tie. Indicates which steps are rests (silent):

- **Bit = 1**: Rest (no sound)
- **Bit = 0**: Play note

---

## FAQ

### Q: What software creates these files?
**A:** Behringer's **SynthTribe** application for macOS/Windows creates and reads `.seq` files when managing TD-3 patterns.

### Q: Can I edit SEQ files in a hex editor?
**A:** Yes! The format is straightforward. Key offsets:
- Notes start at byte 36 (0x24)
- Accents at byte 68 (0x44)
- Slides at byte 100 (0x64)
- Length at byte 134 (0x86)

### Q: How do ties work in the TB-303/TD-3?
**A:** A "tie" connects two steps so the first note sustains through the second step instead of retriggering. In the bitmask:
- Bit = 1 means "new note"
- Bit = 0 means "continue previous note"

### Q: What's the difference between slide and tie?
**A:** 
- **Tie**: Sustains the note without retriggering the envelope
- **Slide**: Glides the pitch smoothly from one note to the next (the classic 303 "squelch")

### Q: Why is the note value different from MIDI?
**A:** The TD-3 uses its own octave numbering starting from 0. Add 24 to convert to standard MIDI note numbers (where middle C = 60).

### Q: Can sequences be longer than 16 steps?
**A:** No, the TD-3 (like the original TB-303) supports a maximum of 16 steps per pattern.

### Q: What's the triplet flag for?
**A:** When enabled, the pattern plays with triplet timing (12 steps per bar instead of 16), giving a swing/shuffle feel.

### Q: Is this format the same as the Crave?
**A:** Similar but not identical. The Crave uses a different note encoding with velocity, gate length, and ratchet fields. See the [CraveSeq project](https://github.com/claziss/CraveSeq) for Crave format details.

---

## Example Hex Dump

A 16-step pattern with some accents and slides:

```
00000000: 2398 5476 0000 0008 0054 0044 002d 0033  #.Tv.....T.D.-.3
00000010: 0000 000a 0031 002e 0033 002e 0037 0000  .....1...3...7..
00000020: 0070 0000 0204 0207 0204 0209 0204 0207  .p..............
00000030: 0204 0200 0204 0204 0204 0204 0204 0204  ................
00000040: 0204 0204 0001 0000 0001 0000 0000 0000  ................
00000050: 0000 0000 0000 0000 0000 0000 0000 0000  ................
00000060: 0000 0000 0000 0001 0000 0000 0001 0000  ................
00000070: 0000 0000 0000 0000 0000 0000 0000 0000  ................
00000080: 0000 0000 0000 0100 0000 ffff 0000 0000  ................
00000090: 0000                                     ..
```

---

## References

- [CraveSeq Project](https://github.com/claziss/CraveSeq) - C library for parsing TD-3/Crave SEQ files
- [Acid-Injector](https://github.com/echolevel/Acid-Injector) - Browser-based MIDI to SEQ converter
- [TB-303 Slide Explanation](https://www.firstpr.com.au/rwi/dfish/303-slide.html)
- [TB-303 Pattern Write Mode](https://www.tinyloops.com/tb303/pattern_write.html)

