package branding

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"math"
)

const baseSize = 32

var (
	colorBG     = color.RGBA{R: 0x1a, G: 0x23, B: 0x32, A: 0xff}
	colorAccent = color.RGBA{R: 0x3b, G: 0x82, B: 0xf6, A: 0xff}
	colorChart  = color.RGBA{R: 0x22, G: 0xc5, B: 0x5e, A: 0xff}
)

var statusColors = map[string]color.RGBA{
	"ok":    {R: 0x22, G: 0xc5, B: 0x5e, A: 0xff},
	"error": {R: 0xef, G: 0x44, B: 0x44, A: 0xff},
	"idle":  {R: 0x94, G: 0xa3, B: 0xb8, A: 0xff},
}

func scale(value float64, size int) float64 {
	return value * float64(size) / float64(baseSize)
}

func strokeWidth(size int) int {
	if size <= 20 {
		return 2
	}
	if size <= 32 {
		return 2
	}
	return max(2, int(math.Round(scale(2, size))))
}

// RenderIcon draws the monitor icon at the requested size.
func RenderIcon(size int, status string) image.Image {
	if size >= 64 {
		big := renderIcon(size*2, status)
		dst := image.NewRGBA(image.Rect(0, 0, size, size))
		resizeRGBA(dst, big)
		return dst
	}
	return renderIcon(size, status)
}

func renderIcon(size int, status string) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	stroke := strokeWidth(size)
	radius := max(2, int(math.Round(scale(8, size))))

	fillRoundedRect(img, 0, 0, size-1, size-1, radius, colorBG)

	monitor := makeBounds(
		int(math.Round(scale(5, size))),
		int(math.Round(scale(6, size))),
		int(math.Round(scale(27, size))),
		int(math.Round(scale(21, size))),
	)
	drawRoundedRectOutline(img, monitor, max(1, int(math.Round(scale(2, size)))), stroke, colorAccent)

	standY := int(math.Round(scale(26, size)))
	drawThickLine(img,
		int(math.Round(scale(11, size))), standY,
		int(math.Round(scale(21, size))), standY,
		stroke, colorAccent)
	drawThickLine(img,
		int(math.Round(scale(16, size))), int(math.Round(scale(21, size))),
		int(math.Round(scale(16, size))), standY,
		stroke, colorAccent)

	chart := []point{
		{int(math.Round(scale(9, size))), int(math.Round(scale(16, size)))},
		{int(math.Round(scale(13, size))), int(math.Round(scale(12, size)))},
		{int(math.Round(scale(17, size))), int(math.Round(scale(15, size)))},
		{int(math.Round(scale(23, size))), int(math.Round(scale(9, size)))},
	}
	for i := 1; i < len(chart); i++ {
		drawThickLine(img, chart[i-1].x, chart[i-1].y, chart[i].x, chart[i].y, stroke, colorChart)
	}

	if size >= 32 {
		statusColor := statusColors["idle"]
		if c, ok := statusColors[status]; ok {
			statusColor = c
		}
		dotRadius := max(2, int(math.Round(scale(4, size))))
		cx := int(math.Round(scale(25, size)))
		cy := int(math.Round(scale(25, size)))
		fillCircle(img, cx, cy, dotRadius, statusColor)
		drawCircleOutline(img, cx, cy, dotRadius, max(1, stroke/2), colorBG)
	}

	return img
}

// TrayPNG returns a 16x16 tray icon encoded as PNG.
func TrayPNG(status string) []byte {
	img := RenderIcon(16, status)
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil
	}
	return buf.Bytes()
}

type point struct {
	x, y int
}

type boundsRect struct {
	x0, y0, x1, y1 int
}

func makeBounds(x0, y0, x1, y1 int) boundsRect {
	return boundsRect{x0: x0, y0: y0, x1: x1, y1: y1}
}

func fillRoundedRect(img *image.RGBA, x0, y0, x1, y1, radius int, c color.RGBA) {
	for y := y0; y <= y1; y++ {
		for x := x0; x <= x1; x++ {
			if insideRoundedRect(x, y, x0, y0, x1, y1, radius) {
				img.SetRGBA(x, y, c)
			}
		}
	}
}

func insideRoundedRect(x, y, x0, y0, x1, y1, radius int) bool {
	if x < x0 || x > x1 || y < y0 || y > y1 {
		return false
	}
	if radius <= 0 {
		return true
	}
	if dx := x - x0; dx < radius {
		if dy := y - y0; dy < radius && dist2(dx, dy, radius) > radius*radius {
			return false
		}
		if dy := y1 - y; dy < radius && dist2(dx, dy, radius) > radius*radius {
			return false
		}
	}
	if dx := x1 - x; dx < radius {
		if dy := y - y0; dy < radius && dist2(dx, dy, radius) > radius*radius {
			return false
		}
		if dy := y1 - y; dy < radius && dist2(dx, dy, radius) > radius*radius {
			return false
		}
	}
	return true
}

func dist2(dx, dy, radius int) int {
	return (radius-dx)*(radius-dx) + (radius-dy)*(radius-dy)
}

func drawRoundedRectOutline(img *image.RGBA, r boundsRect, cornerRadius, stroke int, c color.RGBA) {
	for t := 0; t < stroke; t++ {
		fillRoundedRect(img, r.x0, r.y0+t, r.x1, r.y1-t, cornerRadius, colorBG)
	}
	inner := makeBounds(r.x0+stroke, r.y0+stroke, r.x1-stroke, r.y1-stroke)
	fillRoundedRect(img, inner.x0, inner.y0, inner.x1, inner.y1, max(0, cornerRadius-stroke), colorBG)
	for y := r.y0; y <= r.y1; y++ {
		for x := r.x0; x <= r.x1; x++ {
			onEdge := false
			if x >= r.x0 && x < r.x0+stroke {
				onEdge = insideRoundedRect(x, y, r.x0, r.y0, r.x1, r.y1, cornerRadius)
			} else if x > r.x1-stroke && x <= r.x1 {
				onEdge = insideRoundedRect(x, y, r.x0, r.y0, r.x1, r.y1, cornerRadius)
			} else if y >= r.y0 && y < r.y0+stroke {
				onEdge = insideRoundedRect(x, y, r.x0, r.y0, r.x1, r.y1, cornerRadius)
			} else if y > r.y1-stroke && y <= r.y1 {
				onEdge = insideRoundedRect(x, y, r.x0, r.y0, r.x1, r.y1, cornerRadius)
			}
			if onEdge {
				img.SetRGBA(x, y, c)
			}
		}
	}
}

func drawThickLine(img *image.RGBA, x0, y0, x1, y1, width int, c color.RGBA) {
	dx := abs(x1 - x0)
	dy := -abs(y1 - y0)
	sx := -1
	if x0 < x1 {
		sx = 1
	}
	sy := -1
	if y0 < y1 {
		sy = 1
	}
	err := dx + dy
	for {
		for oy := -width / 2; oy <= width/2; oy++ {
			for ox := -width / 2; ox <= width/2; ox++ {
				set(img, x0+ox, y0+oy, c)
			}
		}
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 >= dy {
			err += dy
			x0 += sx
		}
		if e2 <= dx {
			err += dx
			y0 += sy
		}
	}
}

func fillCircle(img *image.RGBA, cx, cy, radius int, c color.RGBA) {
	for y := cy - radius; y <= cy+radius; y++ {
		for x := cx - radius; x <= cx+radius; x++ {
			dx, dy := x-cx, y-cy
			if dx*dx+dy*dy <= radius*radius {
				img.SetRGBA(x, y, c)
			}
		}
	}
}

func drawCircleOutline(img *image.RGBA, cx, cy, radius, width int, c color.RGBA) {
	for y := cy - radius - width; y <= cy+radius+width; y++ {
		for x := cx - radius - width; x <= cx+radius+width; x++ {
			dx, dy := x-cx, y-cy
			d2 := dx*dx + dy*dy
			outer := (radius + width) * (radius + width)
			inner := max(0, radius-width) * max(0, radius-width)
			if d2 <= outer && d2 >= inner {
				img.SetRGBA(x, y, c)
			}
		}
	}
}

func set(img *image.RGBA, x, y int, c color.RGBA) {
	if x < 0 || y < 0 || x >= img.Bounds().Dx() || y >= img.Bounds().Dy() {
		return
	}
	img.SetRGBA(x, y, c)
}

func resizeRGBA(dst *image.RGBA, src image.Image) {
	srcB := src.Bounds()
	dstB := dst.Bounds()
	for y := 0; y < dstB.Dy(); y++ {
		sy := srcB.Min.Y + y*srcB.Dy()/dstB.Dy()
		for x := 0; x < dstB.Dx(); x++ {
			sx := srcB.Min.X + x*srcB.Dx()/dstB.Dx()
			dst.Set(x, y, src.At(sx, sy))
		}
	}
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
