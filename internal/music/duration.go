package music

import (
	"encoding/binary"
	"io"
	"os"
)

var mpegBitrates = [16]int{0, 32, 40, 48, 56, 64, 80, 96, 112, 128, 160, 192, 224, 256, 320, 0}
var mpegSampleRates = [4]int{44100, 48000, 32000, 0}

// Duration returns the approximate duration of an MP3 file in seconds.
// Falls back to 240s on any error.
func Duration(path string) float64 {
	f, err := os.Open(path)
	if err != nil {
		return 240
	}
	defer f.Close()

	offset := skipID3v2(f)
	frameOffset, header, err := findFirstFrame(f, offset)
	if err != nil {
		return estimateFromSize(path)
	}

	bitrateIdx := (header >> 12) & 0x0F
	srIdx := (header >> 10) & 0x03
	channelMode := (header >> 6) & 0x03

	bitrate := mpegBitrates[bitrateIdx] * 1000
	sampleRate := mpegSampleRates[srIdx]
	isStereo := channelMode != 3

	if bitrate == 0 || sampleRate == 0 {
		return estimateFromSize(path)
	}

	// Check Xing/Info VBR header for exact frame count
	sideInfoSize := int64(32)
	if !isStereo {
		sideInfoSize = 17
	}
	xingPos := frameOffset + 4 + sideInfoSize

	if _, err := f.Seek(xingPos, io.SeekStart); err == nil {
		tag := make([]byte, 4)
		if n, _ := f.Read(tag); n == 4 && (string(tag) == "Xing" || string(tag) == "Info") {
			flags := make([]byte, 4)
			if n, _ := f.Read(flags); n == 4 && binary.BigEndian.Uint32(flags)&0x01 != 0 {
				nf := make([]byte, 4)
				if n, _ := f.Read(nf); n == 4 {
					numFrames := binary.BigEndian.Uint32(nf)
					return float64(numFrames) * 1152.0 / float64(sampleRate)
				}
			}
		}
	}

	// CBR estimate: file_size * 8 / bitrate
	stat, err := os.Stat(path)
	if err != nil {
		return 240
	}
	return float64(stat.Size()) * 8.0 / float64(bitrate)
}

func skipID3v2(f *os.File) int64 {
	hdr := make([]byte, 10)
	if n, _ := f.Read(hdr); n < 10 || string(hdr[:3]) != "ID3" {
		return 0
	}
	// Syncsafe integer: 4 bytes × 7 bits each
	size := int64(hdr[6])<<21 | int64(hdr[7])<<14 | int64(hdr[8])<<7 | int64(hdr[9])
	return size + 10
}

func findFirstFrame(f *os.File, offset int64) (int64, uint32, error) {
	if _, err := f.Seek(offset, io.SeekStart); err != nil {
		return 0, 0, err
	}
	buf := make([]byte, 4096)
	pos := offset
	for {
		n, readErr := f.Read(buf)
		for i := 0; i < n-3; i++ {
			if buf[i] != 0xFF || (buf[i+1]&0xE0) != 0xE0 {
				continue
			}
			h := binary.BigEndian.Uint32(buf[i : i+4])
			version := (h >> 19) & 0x03
			layer := (h >> 17) & 0x03
			bIdx := (h >> 12) & 0x0F
			srIdx := (h >> 10) & 0x03
			// MPEG1, Layer3, valid bitrate and sample rate
			if version == 3 && layer == 1 && bIdx > 0 && bIdx < 15 && srIdx < 3 {
				return pos + int64(i), h, nil
			}
		}
		if readErr != nil {
			break
		}
		pos += int64(n - 3)
		if _, err := f.Seek(pos, io.SeekStart); err != nil {
			break
		}
	}
	return 0, 0, io.EOF
}

func estimateFromSize(path string) float64 {
	stat, err := os.Stat(path)
	if err != nil {
		return 240
	}
	// Assume ~128kbps average
	return float64(stat.Size()) * 8.0 / 128000.0
}
