package branding

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/draw"
	"image/png"
	"io"
	"os"
	"sort"
)

// SaveICO writes a multi-size Windows .ico (PNG payloads, largest first).
func SaveICO(path string, sizes []int, status string) error {
	ordered := append([]int(nil), sizes...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i] > ordered[j] })

	images := make([]*image.RGBA, 0, len(ordered))
	for _, size := range ordered {
		img := RenderIcon(size, status)
		if rgba, ok := img.(*image.RGBA); ok {
			images = append(images, rgba)
			continue
		}
		b := img.Bounds()
		rgba := image.NewRGBA(b)
		draw.Draw(rgba, b, img, b.Min, draw.Src)
		images = append(images, rgba)
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return writeICO(f, ordered, images)
}

func writeICO(w io.Writer, sizes []int, images []*image.RGBA) error {
	if len(images) == 0 {
		return io.ErrUnexpectedEOF
	}

	payloads := make([][]byte, len(images))
	for i, img := range images {
		var buf bytes.Buffer
		if err := png.Encode(&buf, img); err != nil {
			return err
		}
		payloads[i] = buf.Bytes()
	}

	if err := binary.Write(w, binary.LittleEndian, uint16(0)); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, uint16(1)); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, uint16(len(images))); err != nil {
		return err
	}

	offset := 6 + 16*len(images)
	for i, size := range sizes {
		width := byte(size)
		height := byte(size)
		if size >= 256 {
			width, height = 0, 0
		}
		entry := make([]byte, 16)
		entry[0] = width
		entry[1] = height
		entry[4] = 1
		entry[6] = 32
		binary.LittleEndian.PutUint32(entry[8:], uint32(len(payloads[i])))
		binary.LittleEndian.PutUint32(entry[12:], uint32(offset))
		if _, err := w.Write(entry); err != nil {
			return err
		}
		offset += len(payloads[i])
	}

	for _, payload := range payloads {
		if _, err := w.Write(payload); err != nil {
			return err
		}
	}
	return nil
}
