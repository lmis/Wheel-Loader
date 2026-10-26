package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

var (
	hudWhite = rgba(235, 238, 240, 255)
	hudPanel = rgba(24, 28, 34, 205)
	hudTrack = rgba(50, 54, 60, 255)
)

func (g *Game) drawText(dst *ebiten.Image, s string, face text.Face, x, y float64, c color.RGBA) {
	op := &text.DrawOptions{}
	op.GeoM.Translate(x, y)
	op.ColorScale.ScaleWithColor(c)
	text.Draw(dst, s, face, op)
}

func (g *Game) drawTextC(dst *ebiten.Image, s string, face text.Face, cx, y float64, c color.RGBA) {
	w, _ := text.Measure(s, face, 0)
	g.drawText(dst, s, face, cx-w/2, y, c)
}

func panel(dst *ebiten.Image, x, y, w, h float64, c color.RGBA) {
	vector.DrawFilledRect(dst, float32(x), float32(y), float32(w), float32(h), c, true)
}

func bar(dst *ebiten.Image, x, y, w, h, frac float64, fg, bg color.RGBA) {
	vector.DrawFilledRect(dst, float32(x), float32(y), float32(w), float32(h), bg, true)
	vector.DrawFilledRect(dst, float32(x), float32(y), float32(w*clamp(frac, 0, 1)), float32(h), fg, true)
}

// drawGlobalHUD draws over both views: the pause notice.
func (g *Game) drawGlobalHUD(screen *ebiten.Image) {
	if g.paused {
		g.drawTextC(screen, "PAUSED  (F2)", g.art.Big, screenW/2, screenH/2-20, hudWhite)
	}
}
