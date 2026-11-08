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

// drawText draws s into the output frame. See-through text goes through a
// layer: the letters are drawn opaque onto it first and the whole layer is
// then blended once, so where neighbouring letters overlap it is no darker
// than anywhere else.
func (g *Game) drawText(dst canvas, s string, f *Face, x, y float64, c color.RGBA) {
	at := rl.NewVector2(float32(dst.x+x), float32(dst.y+y))
	if c.A == 255 || g.layer.ID == 0 {
		rl.DrawTextEx(f.font, s, at, f.size, 0, rlc(c))
		return
	}
	opaque := c
	opaque.A = 255
	g.inLayer(func() { rl.DrawTextEx(f.font, s, at, f.size, 0, rlc(opaque)) }, c.A)
}

// GL blend factors for drawing into a layer.
const (
	glOne              = 1
	glSrcAlpha         = 0x0302
	glOneMinusSrcAlpha = 0x0303
	glFuncAdd          = 0x8006
)

// inLayer runs draw onto the empty layer target, then blends the layer onto
// the output frame at the given opacity. Inside the layer the colour is
// blended as usual but coverage builds up as in ordinary compositing, so
// the layer ends up premultiplied and is drawn back that way.
func (g *Game) inLayer(draw func(), alpha uint8) {
	rl.EndTextureMode()
	rl.BeginTextureMode(g.layer)
	rl.ClearBackground(rl.Blank)
	rl.SetBlendFactorsSeparate(glSrcAlpha, glOneMinusSrcAlpha, glOne, glOneMinusSrcAlpha, glFuncAdd, glFuncAdd)
	rl.BeginBlendMode(rl.BlendCustomSeparate)
	draw()
	rl.EndBlendMode()
	rl.EndTextureMode()
	rl.BeginTextureMode(g.out)
	rl.BeginBlendMode(rl.BlendAlphaPremultiply)
	src := rl.NewRectangle(0, 0, float32(g.layer.Texture.Width), -float32(g.layer.Texture.Height))
	dst := rl.NewRectangle(0, 0, screenW, screenH)
	rl.DrawTexturePro(g.layer.Texture, src, dst, rl.Vector2{}, 0, rl.NewColor(alpha, alpha, alpha, alpha))
	rl.EndBlendMode()
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

// drawGlobalHUD draws over both views (nothing for now: no pause, no menus).
func (g *Game) drawGlobalHUD(screen canvas) {}
