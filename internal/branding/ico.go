package branding

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/draw"
	"io"
	"os"
	"sort"
)

// SaveICO writes a multi-size Windows .ico (BMP payloads for systray/LoadImage compatibility).
func SaveICO(path string, sizes []int, status string) error {
	data, err := EncodeICO(sizes, status)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// EncodeICO returns ICO bytes with BMP-encoded images (required by Windows LoadImage/systray).
func EncodeICO(sizes []int, status string) ([]byte, error) {
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

	var buf bytes.Buffer
	if err := writeICO(&buf, ordered, images); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// TrayICO returns a systray-compatible ICO for the given status (16 + 32 px).
func TrayICO(status string) []byte {
	data, err := EncodeICO([]int{32, 16}, status)
	if err != nil {
		return nil
	}
	return data
}

func writeICO(w io.Writer, sizes []int, images []*image.RGBA) error {
	if len(images) == 0 {
		return io.ErrUnexpectedEOF
	}

	payloads := make([][]byte, len(images))
	for i, img := range images {
		payloads[i] = encodeICOBitmap(img)
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

func encodeICOBitmap(img *image.RGBA) []byte {
	size := img.Bounds().Dx()
	rowBytes := ((size + 31) / 32) * 4
	maskSize := rowBytes * size
	xorSize := size * size * 4

	buf := make([]byte, 40+xorSize+maskSize)
	binary.LittleEndian.PutUint32(buf[0:], 40)
	binary.LittleEndian.PutUint32(buf[4:], uint32(size))
	binary.LittleEndian.PutUint32(buf[8:], uint32(size*2))
	binary.LittleEndian.PutUint16(buf[12:], 1)
	binary.LittleEndian.PutUint16(buf[14:], 32)
	binary.LittleEndian.PutUint32(buf[20:], uint32(xorSize+maskSize))

	offset := 40
	for y := size - 1; y >= 0; y-- {
		for x := 0; x < size; x++ {
			r, g, b, a := img.RGBAAt(x, y).RGBA()
			buf[offset] = byte(b >> 8)
			buf[offset+1] = byte(g >> 8)
			buf[offset+2] = byte(r >> 8)
			buf[offset+3] = byte(a >> 8)
			offset += 4
		}
	}
	// AND mask: zero = opaque for 32-bit icons with alpha channel.
	return buf
}
