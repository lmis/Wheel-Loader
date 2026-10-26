package main

import (
	"fmt"
	"image/color"
	"math"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

var (
	hudWhite  = rgba(235, 238, 240, 255)
	hudDim    = rgba(170, 176, 180, 255)
	hudPanel  = rgba(24, 28, 34, 205)
	hudAccent = rgba(240, 190, 30, 255)
	hudGreen  = rgba(120, 220, 130, 255)
	hudTrack  = rgba(50, 54, 60, 255)
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

// drawGlobalHUD draws what both players share: the job, money and minimap.
func (g *Game) drawGlobalHUD(screen *ebiten.Image) {
	if g.toddler {
		if g.paused {
			g.drawTextC(screen, "PAUSED  (F2)", g.art.Big, screenW/2, screenH/2-20, hudWhite)
		}
		const hint = "F5: grown-up mode"
		w, _ := text.Measure(hint, g.art.Small, 0)
		g.drawText(screen, hint, g.art.Small, screenW-w-14, screenH-22, rgba(200, 204, 210, 120))
		return
	}
	j := g.jobs.Current
	panel(screen, 16, 16, 360, 96, hudPanel)
	g.drawText(screen, "CURRENT JOB", g.art.Bold, 28, 24, hudAccent)
	if j != nil {
		g.drawText(screen, j.Describe(), g.art.Mid, 28, 46, hudWhite)
		bar(screen, 28, 74, 250, 14, j.Progress(), hudAccent, hudTrack)
		g.drawText(screen, fmt.Sprintf("%.0f / %.0f", math.Max(0, j.Delivered), j.Target), g.art.Small, 286, 73, hudDim)
		g.drawText(screen, fmt.Sprintf("Reward $%d", j.Reward), g.art.Small, 28, 92, hudGreen)
	}
	panel(screen, 16, 120, 190, 54, hudPanel)
	g.drawText(screen, fmt.Sprintf("$ %d", g.jobs.Money), g.art.Big, 28, 126, hudGreen)
	g.drawText(screen, fmt.Sprintf("Level %d   Jobs %d", g.jobs.Level, g.jobs.Jobs), g.art.Small, 28, 156, hudDim)

	if g.jobs.flashT > 0 {
		a := uint8(clamp(g.jobs.flashT/4*255, 0, 255))
		g.drawTextC(screen, g.jobs.flash, g.art.Big, screenW/2, 60, rgba(120, 230, 140, a))
	}
	if g.paused {
		g.drawTextC(screen, "PAUSED  (F2)", g.art.Big, screenW/2, screenH/2-20, hudWhite)
	}
	if !g.coop() {
		g.drawTextC(screen, "Player 2: press N to join", g.art.Small, screenW/2, 8, rgba(200, 204, 210, 150))
	}
	g.drawMinimap(screen)
}

// drawPlayerHUD draws one player's status, prompts and help into their view.
func (g *Game) drawPlayerHUD(dst *ebiten.Image, p *Player, vw, vh float64) {
	if g.toddler && p.td != nil {
		g.drawToddlerHUD(dst, p, vw, vh)
		return
	}
	k := &p.keys
	if g.coop() {
		label := fmt.Sprintf("PLAYER %d", p.ID)
		w, _ := text.Measure(label, g.art.Bold, 0)
		panel(dst, 16, vh-60, w+20, 24, hudPanel)
		g.drawText(dst, label, g.art.Bold, 26, vh-56, p.marker)
	}

	// ---- status panel (bottom-right) ----
	var gauges []gauge
	name := "ON FOOT"
	if v := p.Driving; v != nil {
		name = strings.ToUpper(v.Name())
		gauges = v.Gauges()
	}
	ph := 66 + float64(len(gauges))*20
	sx, sy := vw-300, vh-ph-16
	panel(dst, sx, sy, 284, ph, hudPanel)
	g.drawText(dst, name, g.art.Bold, sx+14, sy+8, hudAccent)
	if v := p.Driving; v != nil {
		spd := v.SpeedVal()
		gear := "N"
		if spd > 5 {
			gear = "D"
		} else if spd < -5 {
			gear = "R"
		}
		g.drawText(dst, fmt.Sprintf("%2.0f km/h   %s", math.Abs(spd)*0.12, gear), g.art.Mid, sx+14, sy+32, hudWhite)
		view := "CHASE"
		if p.fp {
			view = "CAB"
		}
		g.drawText(dst, view, g.art.Small, sx+230, sy+36, hudDim)
	} else {
		spd := p.W.vel.Len()
		state := "standing"
		if spd > kidWalk*1.2 {
			state = "running"
		} else if spd > 4 {
			state = "walking"
		}
		g.drawText(dst, state, g.art.Mid, sx+14, sy+32, hudWhite)
	}
	for i, ga := range gauges {
		y := sy + 60 + float64(i)*20
		g.drawText(dst, ga.Label, g.art.Small, sx+14, y-2, hudDim)
		bar(dst, sx+84, y, 110, 10, ga.Value, rgba(90, 170, 240, 255), hudTrack)
		g.drawText(dst, ga.Text, g.art.Small, sx+202, y-2, rgba(210, 210, 140, 255))
	}

	// ---- prompts and messages ----
	if p.Driving == nil && p.near != nil {
		msg := fmt.Sprintf("[%s]  climb into the %s", keyName(k.enter), p.near.Name())
		w, _ := text.Measure(msg, g.art.Mid, 0)
		panel(dst, vw/2-w/2-14, vh*0.66, w+28, 30, hudPanel)
		g.drawText(dst, msg, g.art.Mid, vw/2-w/2, vh*0.66+5, hudWhite)
	}
	if p.msgT > 0 {
		g.drawTextC(dst, p.msg, g.art.Mid, vw/2, vh*0.58, hudAccent)
	}
	if p.fp && p.Driving == nil {
		c := rgba(255, 255, 255, 180)
		vector.StrokeLine(dst, float32(vw/2-6), float32(vh/2), float32(vw/2+6), float32(vh/2), 1.5, c, true)
		vector.StrokeLine(dst, float32(vw/2), float32(vh/2-6), float32(vw/2), float32(vh/2+6), 1.5, c, true)
	}

	// ---- controls ----
	var lines []string
	if v := p.Driving; v != nil {
		lines = append(lines, strings.ToUpper(v.Name()))
		lines = append(lines, v.ControlHelp()...)
		lines = append(lines, "{E}\tclimb out")
	} else {
		lines = append(lines, "ON FOOT",
			"{T}\twalk forward / back",
			"{S}\tturn",
			"{4}\tstep sideways",
			"{B}\trun",
			"{1}\tlook up / down (first person)",
			"{E}\tclimb into a vehicle")
	}
	lines = append(lines, "{V}\tcab / first-person view")
	if !g.coop() {
		lines = append(lines, "mouse\tclick to look around, Esc frees it")
	}
	lines = append(lines, "GENERAL", "F1\thelp   F2 pause   F3 reset", "F4\tco-op split screen on/off")
	if g.showHelp {
		h := 20 + float64(len(lines))*17.5
		top := vh - h - 50
		if top < 190 {
			top = 190
		}
		panel(dst, 16, top, 330, h, hudPanel)
		y := top + 10
		for _, ln := range lines {
			ln = k.helpText(ln)
			if !strings.Contains(ln, "\t") {
				g.drawText(dst, ln, g.art.Bold, 28, y, hudAccent)
			} else {
				parts := strings.SplitN(ln, "\t", 2)
				g.drawText(dst, parts[0], g.art.Small, 28, y, hudWhite)
				g.drawText(dst, parts[1], g.art.Small, 128, y, hudDim)
			}
			y += 17.5
		}
	} else {
		hint := k.helpText("{E} get in/out   {V} view   F1 help")
		w, _ := text.Measure(hint, g.art.Small, 0)
		panel(dst, 16, vh-30, w+16, 22, rgba(24, 28, 34, 170))
		g.drawText(dst, hint, g.art.Small, 24, vh-28, hudDim)
	}
}

func matColorC(t int) (r, g, b, a uint8) {
	r, g, b = matColor(t)
	return r, g, b, 255
}

// drawMinimap shows the site, zones, machines and players (top-right).
func (g *Game) drawMinimap(screen *ebiten.Image) {
	mw, mh := 220.0, 165.0
	mx, my := float64(screenW)-mw-16, 16.0
	panel(screen, mx-4, my-4, mw+8, mh+8, hudPanel)
	sx := mw / g.world.Width
	sy := mh / g.world.Height
	toMap := func(p Vec) (float32, float32) { return float32(mx + p.X*sx), float32(my + p.Y*sy) }

	vector.DrawFilledRect(screen, float32(mx), float32(my), float32(mw), float32(mh), rgba(70, 58, 42, 255), false)
	for _, z := range g.world.Zones {
		x, y := toMap(z.Pos)
		r, gg, b := matColor(z.Type)
		vector.DrawFilledCircle(screen, x, y, 5, rgba(r, gg, b, 255), true)
	}
	for _, v := range g.vehicles {
		x, y := toMap(v.Center())
		switch m := v.(type) {
		case *Truck:
			c := rgba(90, 140, 200, 255)
			if g.jobs.Current != nil && g.jobs.Current.Truck == m {
				c = hudGreen
			}
			vector.DrawFilledRect(screen, x-3, y-4, 6, 8, c, true)
		case *Loader:
			vector.DrawFilledCircle(screen, x, y, 3.5, m.bodyColor(), true)
		case *Excavator:
			vector.DrawFilledRect(screen, x-3.5, y-3.5, 7, 7, rgba(245, 190, 40, 255), true)
		case *Car:
			vector.DrawFilledCircle(screen, x, y, 2.5, rgba(240, 240, 240, 255), true)
		case *TowerCrane:
			c := rgba(246, 192, 22, 255)
			vector.StrokeLine(screen, x-5, y, x+5, y, 2, c, true)
			vector.StrokeLine(screen, x, y-5, x, y+5, 2, c, true)
			hx, hy := toMap(m.trolleyPos())
			vector.StrokeLine(screen, x, y, hx, hy, 1, c, true)
		}
	}
	for _, p := range g.players {
		pos, h := p.W.Pos, p.W.Yaw
		if p.Driving != nil {
			pos, h = p.Driving.Center(), p.Driving.Facing()
		}
		x, y := toMap(pos)
		tip := func(a, r float64) (float32, float32) {
			return x + float32(math.Cos(h+a)*r), y + float32(math.Sin(h+a)*r)
		}
		x1, y1 := tip(0, 7)
		x2, y2 := tip(2.5, 5)
		x3, y3 := tip(-2.5, 5)
		var path vector.Path
		path.MoveTo(x1, y1)
		path.LineTo(x2, y2)
		path.LineTo(x3, y3)
		path.Close()
		op := &vector.DrawPathOptions{AntiAlias: true}
		op.ColorScale.ScaleWithColor(p.marker)
		vector.FillPath(screen, &path, nil, op)
	}
}
