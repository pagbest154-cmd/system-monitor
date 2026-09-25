//go:build windows

package main

import (
	"strings"

	"github.com/lxn/walk"
)

var (
	colorBg      = walk.RGB(15, 20, 25)
	colorSurface = walk.RGB(26, 35, 50)
	colorInput   = walk.RGB(36, 48, 68)
	colorText    = walk.RGB(232, 237, 244)
	colorMuted   = walk.RGB(139, 156, 179)
	colorAccent  = walk.RGB(59, 130, 246)
	colorOk      = walk.RGB(34, 197, 94)
	colorWarn    = walk.RGB(245, 158, 11)
)

func solidBrush(c walk.Color) *walk.SolidColorBrush {
	b, err := walk.NewSolidColorBrush(c)
	if err != nil {
		b, _ = walk.NewSolidColorBrush(walk.RGB(30, 30, 30))
	}
	return b
}

func styleLabel(l *walk.Label, muted bool) {
	if l == nil {
		return
	}
	_ = l.SetTextColor(walk.RGB(232, 237, 244))
	if muted {
		_ = l.SetTextColor(colorMuted)
	}
	font, _ := walk.NewFont("Segoe UI", 9, walk.FontWeightNormal)
	_ = l.SetFont(font)
}

func styleHeading(l *walk.Label) {
	if l == nil {
		return
	}
	_ = l.SetTextColor(colorText)
	font, _ := walk.NewFont("Segoe UI", 14, walk.FontWeightBold)
	_ = l.SetFont(font)
}

func styleLineEdit(le *walk.LineEdit) {
	if le == nil {
		return
	}
	_ = le.SetBackground(solidBrush(colorInput))
	_ = le.SetTextColor(colorText)
	font, _ := walk.NewFont("Segoe UI", 10, walk.FontWeightNormal)
	_ = le.SetFont(font)
}

func styleNumberEdit(ne *walk.NumberEdit) {
	if ne == nil {
		return
	}
	_ = ne.SetBackground(solidBrush(colorInput))
	_ = ne.SetTextColor(colorText)
	font, _ := walk.NewFont("Segoe UI", 10, walk.FontWeightNormal)
	_ = ne.SetFont(font)
}

func stylePushButton(pb *walk.PushButton, primary bool) {
	if pb == nil {
		return
	}
	font, _ := walk.NewFont("Segoe UI", 9, walk.FontWeightBold)
	_ = pb.SetFont(font)
	if primary {
		_ = pb.SetTextColor(colorText)
	}
}

func notifyStatusColor(text string) walk.Color {
	if strings.Contains(text, "подключены") {
		return colorOk
	}
	if strings.Contains(text, "выкл") || strings.Contains(text, "нет token") {
		return colorMuted
	}
	return colorWarn
}
