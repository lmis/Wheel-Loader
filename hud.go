package main

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
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

func (g *Game) drawHUD(screen *ebiten.Image) {
	white := rgba(235, 238, 240, 255)
	dim := rgba(170, 176, 180, 255)
	panelBg := rgba(24, 28, 34, 205)
	accent := rgba(240, 190, 30, 255)

	// ---- Job panel (top-left) ----
	j := g.jobs.Current
	panel(screen, 16, 16, 360, 96, panelBg)
	g.drawText(screen, "CURRENT JOB", g.art.Bold, 28, 24, accent)
	if j != nil {
		g.drawText(screen, j.Describe(), g.art.Mid, 28, 46, white)
		bar(screen, 28, 74, 250, 14, j.Progress(), accent, rgba(50, 54, 60, 255))
		g.drawText(screen, fmt.Sprintf("%.0f / %.0f", j.Delivered, j.Target), g.art.Small, 286, 73, dim)
		g.drawText(screen, fmt.Sprintf("Reward $%d", j.Reward), g.art.Small, 28, 92, rgba(120, 220, 130, 255))
	}

	// ---- Money / level (top-left, below job) ----
	panel(screen, 16, 120, 190, 54, panelBg)
	g.drawText(screen, fmt.Sprintf("$ %d", g.jobs.Money), g.art.Big, 28, 126, rgba(120, 220, 130, 255))
	g.drawText(screen, fmt.Sprintf("Level %d   Jobs %d", g.jobs.Level, g.jobs.Jobs), g.art.Small, 28, 156, dim)

	// ---- Loader status (bottom-right) ----
	sx, sy := 980.0, 560.0
	panel(screen, sx, sy, 284, 144, panelBg)
	l := g.loader
	// speed + gear
	spdKmh := math.Abs(l.Speed) * 0.12
	gear := "N"
	if l.Speed > 5 {
		gear = "D"
	} else if l.Speed < -5 {
		gear = "R"
	}
	g.drawText(screen, fmt.Sprintf("%2.0f", spdKmh), g.art.Big, sx+16, sy+12, white)
	g.drawText(screen, "km/h", g.art.Small, sx+54, sy+30, dim)
	g.drawText(screen, gear, g.art.Big, sx+120, sy+12, accent)
	// RPM bar
	g.drawText(screen, "RPM", g.art.Small, sx+160, sy+14, dim)
	bar(screen, sx+160, sy+30, 110, 8, l.engineRPM, rgba(230, 120, 40, 255), rgba(50, 54, 60, 255))

	// bucket load
	g.drawText(screen, "BUCKET", g.art.Bold, sx+16, sy+48, accent)
	loadCol := rgba(matColorC(l.LoadType))
	g.drawText(screen, fmt.Sprintf("%s  %.0f%%", loadName(l), l.LoadFrac()*100), g.art.Small, sx+86, sy+50, white)
	bar(screen, sx+16, sy+70, 252, 12, l.LoadFrac(), loadCol, rgba(50, 54, 60, 255))

	// lift & tilt gauges
	g.drawText(screen, "LIFT", g.art.Small, sx+16, sy+90, dim)
	bar(screen, sx+54, sy+92, 90, 10, l.Lift, rgba(90, 170, 240, 255), rgba(50, 54, 60, 255))
	g.drawText(screen, "TILT", g.art.Small, sx+16, sy+108, dim)
	bar(screen, sx+54, sy+110, 90, 10, l.Tilt, rgba(90, 170, 240, 255), rgba(50, 54, 60, 255))
	// dig/carry labels
	hint := "DUMP"
	if l.Tilt > 0.6 {
		hint = "CARRY"
	} else if l.Tilt > 0.35 {
		hint = "LEVEL"
	}
	g.drawText(screen, hint, g.art.Small, sx+150, sy+108, rgba(200, 200, 120, 255))
	grd := "RAISED"
	if l.Lift < 0.28 {
		grd = "GROUND"
	}
	g.drawText(screen, grd, g.art.Small, sx+150, sy+90, rgba(200, 200, 120, 255))

	// ---- Controls hint (bottom-left) ----
	if g.showHelp {
		panel(screen, 16, 432, 340, 272, panelBg)
		lines := []string{
			"DRIVE",
			"  W / Up        accelerate",
			"  S / Down      reverse",
			"  Space         brake",
			"STEER (articulated)",
			"  A / D         bend left / right",
			"BUCKET",
			"  R / F         raise / lower arms",
			"  T / G         curl / dump bucket",
			"CAMERA",
			"  wheel / + -   zoom",
			"  Q / E         orbit view",
			"  [ / ]         tilt view",
			"GENERAL",
			"  Tab  help    P  pause    Bksp reset",
		}
		y := 442.0
		for _, s := range lines {
			c := white
			if len(s) > 0 && s[0] != ' ' {
				c = accent
			}
			g.drawText(screen, s, g.art.Small, 28, y, c)
			y += 17.5
		}
	} else {
		panel(screen, 16, 684, 300, 22, rgba(24, 28, 34, 170))
		g.drawText(screen, "W/S drive  A/D steer  R/F lift  T/G bucket  Tab=help", g.art.Small, 24, 686, dim)
	}

	// ---- flash message ----
	if g.jobs.flashT > 0 {
		a := uint8(clamp(g.jobs.flashT/4*255, 0, 255))
		g.drawTextC(screen, g.jobs.flash, g.art.Big, screenW/2, 60, rgba(120, 230, 140, a))
	}
	if g.paused {
		g.drawTextC(screen, "PAUSED", g.art.Big, screenW/2, screenH/2-20, white)
	}

	g.drawMinimap(screen)
}

func matColorC(t int) (r, g, b, a uint8) {
	r, g, b = matColor(t)
	return r, g, b, 255
}

func loadName(l *Loader) string {
	if l.Load < 1 {
		return "empty"
	}
	return matNames[l.LoadType]
}

// drawMinimap shows the whole site, zones, trucks and the loader (top-right).
func (g *Game) drawMinimap(screen *ebiten.Image) {
	mw, mh := 220.0, 165.0
	mx, my := float64(screenW)-mw-16, 16.0
	panel(screen, mx-4, my-4, mw+8, mh+8, rgba(24, 28, 34, 205))
	sx := mw / g.world.Width
	sy := mh / g.world.Height
	toMap := func(p Vec) (float64, float64) { return mx + p.X*sx, my + p.Y*sy }

	// ground
	vector.DrawFilledRect(screen, float32(mx), float32(my), float32(mw), float32(mh), rgba(70, 58, 42, 255), false)
	// zones
	for _, z := range g.world.Zones {
		x, y := toMap(z.Pos)
		r, gg, b := matColor(z.Type)
		vector.DrawFilledCircle(screen, float32(x), float32(y), 5, rgba(r, gg, b, 255), true)
	}
	// trucks
	for _, t := range g.world.Trucks {
		x, y := toMap(t.Pos)
		c := rgba(90, 140, 200, 255)
		if g.jobs.Current != nil && g.jobs.Current.Truck == t {
			c = rgba(120, 230, 140, 255)
		}
		vector.DrawFilledRect(screen, float32(x-3), float32(y-4), 6, 8, c, true)
	}
	// loader
	lx, ly := toMap(g.loader.Joint)
	vector.DrawFilledCircle(screen, float32(lx), float32(ly), 4, rgba(240, 190, 30, 255), true)
	// heading tick
	h := g.loader.Heading
	vector.StrokeLine(screen, float32(lx), float32(ly), float32(lx+math.Cos(h)*8), float32(ly+math.Sin(h)*8), 2, rgba(255, 255, 255, 255), true)
	// camera view box
	mn, mx2 := g.cam.VisibleBounds()
	vx, vy := toMap(mn)
	vector.StrokeRect(screen, float32(vx), float32(vy), float32((mx2.X-mn.X)*sx), float32((mx2.Y-mn.Y)*sy), 1, rgba(255, 255, 255, 120), false)
}
