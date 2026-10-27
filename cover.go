package main

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// coverTitle is the name on the cover art.
const coverTitle = "Léo & Louis onsite"

// stageCover sets up one still for the cover art: both kids out front with
// their animal friends, the loaders behind them, a skater gliding past on
// each side, a tower crane off to one side, the trucks and excavators at
// work in the back, the train in the distance and the helicopter overhead.
// The simulation is then held still so everything stays exactly where it
// was put.
func (g *Game) stageCover(path string) {
	g.scripted = true
	g.script = append(g.script, scriptStep{at: 3, fn: func(g *Game) {
		g.frozen, g.title = true, coverTitle
		g.world.Particles = nil
		// just the one crane, off to the right
		var vs []Vehicle
		for _, v := range g.vehicles {
			if v != g.crane2 {
				vs = append(vs, v)
			}
		}
		g.vehicles = vs
		g.crane.Slew = math.Pi * 0.78 // jib reaching out over the site

		// the kids stand out front, facing the camera, their loaders behind
		facing := math.Pi / 2
		for i, p := range g.players {
			if p.Driving != nil {
				p.Driving.SetOccupied(false)
			}
			p.Driving, p.td.goTo = nil, nil
			x := 1600 + float64(2*i-1)*34
			p.W.Pos, p.W.Yaw = V(x, 1450), facing-float64(2*i-1)*0.25
			p.W.moving, p.W.phase = 0, 0
		}
		// the animals gather round their kid: the owl flying between them, the chimp at
		// player 1's side, the dogs flanking player 2
		for _, a := range g.animals {
			k := g.players[0].W.Pos
			if a.owner == g.players[1] {
				k = g.players[1].W.Pos
			}
			a.ride, a.hopping, a.speed, a.bounce = nil, false, 0, 0
			switch a.Kind {
			case animalOwl:
				a.flying, a.Pos, a.Y, a.Yaw, a.t = true, V(1600, 1440), 40, facing, 0.17
			case animalChimp:
				a.Pos, a.Y, a.Yaw, a.t = k.Add(V(-24, 14)), 0, facing+0.4, 3
			case animalGolden:
				a.Pos, a.Y, a.Yaw = k.Add(V(28, 12)), 0, facing-0.5
			case animalBernese:
				a.Pos, a.Y, a.Yaw = k.Add(V(70, -26)), 0, facing-1.1
			}
		}
		// the skaters glide past on either side
		sk := &g.wonder.skaters
		sk[0].Pos, sk[0].Yaw, sk[0].phase, sk[0].lean, sk[0].look = V(1500, 1478), math.Pi-0.25, math.Pi/2, 0.15, 0
		sk[1].Pos, sk[1].Yaw, sk[1].phase, sk[1].lean, sk[1].look = V(1705, 1484), 0.3, 3*math.Pi/2, -0.2, 0
		for i := range sk {
			sk[i].speed = skaterSpeed
		}
		// the helicopter overhead, the train in the distance
		g.wonder.heli.t = 12.9
		g.wonder.train.X, g.wonder.train.wait = 1950, 0

		g.tour = []tourStop{{frame: g.frame, pos: V(1600, 1370), zoom: 250, yaw: math.Pi/2 + 0.06, pitch: 0.2, look: 34, shot: path}}
	}})
	g.exitFrame = 8
}

// drawTitle writes the cover title across the bottom of the frame (the
// helicopter has the sky) in chunky loader yellow with a dark outline.
func (g *Game) drawTitle(title string) {
	f := g.art.Title
	x := (screenW - textWidth(title, f)) / 2
	y := screenH - float64(f.size) - 40
	ink := rgba(46, 30, 16, 255)
	for a := 0.0; a < 2*math.Pi; a += math.Pi / 8 {
		g.drawText(canvas{}, title, f, x+math.Cos(a)*5, y+math.Sin(a)*5+3, ink)
	}
	g.drawText(canvas{}, title, f, x, y, colLoader)
	// a lighter band across the top of the letters
	rl.BeginScissorMode(0, int32(y+10), screenW, int32(f.size*0.32))
	g.drawText(canvas{}, title, f, x, y, rgba(255, 222, 96, 255))
	rl.EndScissorMode()
}
