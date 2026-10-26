package main

import (
	"image/color"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var (
	hudWhite = rgba(235, 238, 240, 255)
	hudPanel = rgba(24, 28, 34, 205)
)

// canvas is where 2D drawing lands: an offset into the current target.
type canvas struct{ x, y float64 }

func rlc(c color.RGBA) rl.Color { return rl.NewColor(c.R, c.G, c.B, c.A) }

func (g *Game) drawText(dst canvas, s string, f *Face, x, y float64, c color.RGBA) {
	rl.DrawTextEx(f.font, s, rl.NewVector2(float32(dst.x+x), float32(dst.y+y)), f.size, 0, rlc(c))
}

func textWidth(s string, f *Face) float64 {
	return float64(rl.MeasureTextEx(f.font, s, f.size, 0).X)
}

func (g *Game) drawTextC(dst canvas, s string, f *Face, cx, y float64, c color.RGBA) {
	g.drawText(dst, s, f, cx-textWidth(s, f)/2, y, c)
}

func panel(dst canvas, x, y, w, h float64, c color.RGBA) {
	rl.DrawRectangleRec(rl.NewRectangle(float32(dst.x+x), float32(dst.y+y), float32(w), float32(h)), rlc(c))
}

func circle(dst canvas, x, y, r float64, c color.RGBA) {
	rl.DrawCircleV(rl.NewVector2(float32(dst.x+x), float32(dst.y+y)), float32(r), rlc(c))
}

// drawGlobalHUD draws over both views: the pause notice.
func (g *Game) drawGlobalHUD(screen canvas) {
	if g.paused {
		g.drawTextC(screen, "PAUSED  (F2)", g.art.Big, screenW/2, screenH/2-20, hudWhite)
	}
}
