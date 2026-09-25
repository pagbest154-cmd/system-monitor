//go:build windows

package main

import (
	"strings"

	"github.com/lxn/walk"
)

var (
	colorBg     = walk.RGB(15, 20, 25)
	colorInput  = walk.RGB(36, 48, 68)
	colorText   = walk.RGB(232, 237, 244)
	colorMuted  = walk.RGB(139, 156, 179)
	colorOk     = walk.RGB(34, 197, 94)
	colorWarn   = walk.RGB(245, 158, 11)
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
	if muted {
		l.SetTextColor(colorMuted)
	} else {
		l.SetTextColor(colorText)
	}
	font, _ := walk.NewFont("Segoe UI", 9, 0)
	l.SetFont(font)
}

func styleHeading(l *walk.Label) {
	if l == nil {
		return
	}
	l.SetTextColor(colorText)
	font, _ := walk.NewFont("Segoe UI", 14, walk.FontBold)
	l.SetFont(font)
}

func styleLineEdit(le *walk.LineEdit) {
	if le == nil {
		return
	}
	le.SetBackground(solidBrush(colorInput))
	le.SetTextColor(colorText)
	font, _ := walk.NewFont("Segoe UI", 10, 0)
	le.SetFont(font)
}

func styleNumberEdit(ne *walk.NumberEdit) {
	if ne == nil {
		return
	}
	ne.SetBackground(solidBrush(colorInput))
	ne.SetTextColor(colorText)
	font, _ := walk.NewFont("Segoe UI", 10, 0)
	ne.SetFont(font)
}

func stylePushButton(pb *walk.PushButton, primary bool) {
	if pb == nil {
		return
	}
	font, _ := walk.NewFont("Segoe UI", 9, walk.FontBold)
	pb.SetFont(font)
	if primary {
		pb.SetTextColor(colorText)
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
