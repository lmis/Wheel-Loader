package main

import (
	"math"
	"path/filepath"
	"strings"
)

// sceneSpec is one staged screenshot: set the world up, hold the given
// controls for a while, then capture.
type sceneSpec struct {
	name   string
	frames int
	setup  func(g *Game)
	ctl    *Controls // player 1 controls while the scene runs
}

// stageScenes schedules the screenshot collection of every feature; only, if
// non-empty, keeps just the scenes whose name contains it. Scenes run in
// order and build on each other.
func (g *Game) stageScenes(outDir, only string) {
	g.scripted = true
	g.simSpeed = 2
	w := g.world
	p1, p2 := g.players[0], g.players[1]
	walkerCam := func(p *Player, dist, userYaw, pitch float64) {
		p.distWalker, p.cam.UserYaw, p.cam.Pitch = dist, userYaw, pitch
		g.snapCamera(p)
	}
	driveCam := func(p *Player, dist, userYaw, pitch float64) {
		p.distVehicle, p.cam.UserYaw, p.cam.Pitch = dist, userYaw, pitch
		g.snapCamera(p)
	}
	still := func(name string, pos Vec, dist, yaw, pitch float64) sceneSpec {
		return sceneSpec{name: name, frames: 2, setup: func(g *Game) {
			g.tour = []tourStop{{frame: g.frame, pos: pos, zoom: dist, yaw: yaw, pitch: pitch, shot: filepath.Join(outDir, name+".png")}}
		}}
	}

	scenes := []sceneSpec{
		{name: "01_start", frames: 20},
		{name: "01b_animals_riding", frames: 12, setup: func(g *Game) {
			driveCam(p1, 210, 0.7, 0.42)
			driveCam(p2, 210, -0.7, 0.42)
		}},
		{name: "02_scooping", frames: 60, setup: func(g *Game) {
			for _, p := range g.players {
				driveCam(p, p.Driving.CamDist(), 0, 0.6)
			}
			p2.scriptCtl = &Controls{Throttle: 1, Steer: 0.1}
		}, ctl: &Controls{Throttle: 1, Steer: -0.05}},
		{name: "03_loading_the_truck", frames: 34, setup: func(g *Game) {
			tr := w.Trucks[0]
			l := g.loader
			l.Joint = tr.Pos.Add(dir(tr.Angle).Scale(118))
			l.Heading, l.Artic, l.Lift, l.Tilt, l.Load, l.LoadType, l.Speed = tr.Angle+math.Pi, 0, 0.62, 1, 95, MatSand, 0
			p1.td.mode = tdCarry
			driveCam(p1, 380, 0.9, 0.5)
			p2.scriptCtl = &Controls{Throttle: -0.6}
		}},
		{name: "03b_truck_full", frames: 40, setup: func(g *Game) {
			tr := w.Trucks[0]
			tr.Contents = [numMat]float64{}
			tr.Contents[MatSand] = tr.Capacity - 30
			l := g.loader
			l.Lift, l.Tilt, l.Load, l.LoadType = 0.62, 1, 95, MatSand
			p1.td.mode = tdCarry
			driveCam(p1, 330, 1.2, 0.55)
		}},
		{name: "04_walk_over", frames: 46, setup: func(g *Game) {
			for i, p := range g.players {
				g.toddlerPick(p, p.td.own[2-i])
			}
		}},
		{name: "04b_hopping_on", frames: 40},
		{name: "05_cranes", frames: 40, setup: func(g *Game) {
			for _, p := range g.players {
				c, tr := p.td.own[2].(*TowerCrane), p.td.own[1].(*Truck)
				p.td.goTo = nil
				g.put(p, c)
				var box *Cargo
				for _, o := range w.Cargo {
					if o.Kind == cargoBricks && o.Pos.Sub(c.Base).Len() < craneJibLen && !o.held {
						box = o
					}
				}
				box.held, box.truck = true, nil
				c.held, c.heldYaw = box, 0
				d := tr.Pos.Add(V(40, 80)).Sub(c.Base)
				c.Slew, c.slewV, c.Trolley, c.HookLen = d.Angle(), 0, d.Len(), craneJibY-toddlerHookY-box.H-6
				c.hook, c.hookV = c.trolleyPos(), Vec{}
				p.td.mode, p.td.modeT = tdCarry, 0
				driveCam(p, 520, 0.5, 0.5)
				p.scriptCtl = &Controls{Throttle: -0.3, Steer: 0.4}
			}
		}},
		{name: "06_trucks_tipping", frames: 70, setup: func(g *Game) {
			for i, p := range g.players {
				tr := p.td.own[1].(*Truck)
				g.put(p, tr)
				tr.Contents = [numMat]float64{}
				tr.Contents[[]int{MatSand, MatDirt}[i]] = 380
				p.td.mode, p.td.pressed = tdCarry, true
				driveCam(p, 480, 2.3, 0.45)
				p.scriptCtl = nil
			}
		}},
		{name: "07a_rollerblader_zooms_by", frames: 14, setup: func(g *Game) {
			for i, p := range g.players {
				g.put(p, p.td.own[0])
				p.td.goTo = nil
				p.scriptCtl = nil
				driveCam(p, p.Driving.CamDist(), []float64{0.5, -0.5}[i], 0.5)
			}
			sk := &g.wonder.skater
			sk.s, sk.off = 760, 22
			sk.Pos, sk.Yaw = loopAt(sk.s)
		}},
		{name: "07b_helicopter_overhead", frames: 14, setup: func(g *Game) {
			g.wonder.heli.t = 12.5
			for _, p := range g.players {
				driveCam(p, p.Driving.CamDist(), 0, 0.5)
			}
		}},
		{name: "07c_train_in_the_distance", frames: 2, setup: func(g *Game) {
			g.wonder.train.X, g.wonder.train.wait = 1750, 0
			g.tour = []tourStop{{frame: g.frame, pos: V(1450, 330), zoom: 520, yaw: math.Pi/2 + 0.35, pitch: 0.3, look: 25, shot: filepath.Join(outDir, "07c_train_in_the_distance.png")}}
		}},
		{name: "07d_rollerblader_closeup", frames: 2, setup: func(g *Game) {
			sk := &g.wonder.skater
			sk.s = 1500
			at, _ := loopAt(sk.s + 22)
			g.tour = []tourStop{{frame: g.frame, pos: at.Add(V(22, 0)), zoom: 75, yaw: 0.9, pitch: 0.22, look: 18, shot: filepath.Join(outDir, "07d_rollerblader_closeup.png")}}
		}},
		{name: "07e_helicopter_closeup", frames: 2, setup: func(g *Game) {
			g.wonder.heli.t = 9
			pos, _, _ := g.wonder.heli.Pos()
			g.tour = []tourStop{{frame: g.frame, pos: pos.Add(V(20, -16)), zoom: 330, yaw: 2.2, pitch: 0.2, look: heliAlt, shot: filepath.Join(outDir, "07e_helicopter_closeup.png")}}
		}},
		{name: "07_kid_portrait", frames: 20, setup: func(g *Game) {
			g.tour = nil
			g.putOnFoot(p1, V(1600, 1312), math.Pi/2)
			walkerCam(p1, 46, math.Pi+0.35, 0.16)
			p1.cam.lookHeight = 15
		}},
		still("08_workshop_yard", V(2520, 520), 520, 1.2, 0.62),
		still("09_houses_and_trees", V(1250, 170), 640, 1.85, 0.42),
		still("10_site_overview", V(1600, 1250), 1550, 1.1, 0.72),
	}

	f := 5
	for _, sc := range scenes {
		sc := sc
		if only != "" && !strings.Contains(sc.name, only) {
			continue
		}
		start := f
		g.script = append(g.script, scriptStep{at: start, fn: func(g *Game) {
			g.tour = nil
			if sc.setup != nil {
				sc.setup(g)
			}
			g.players[0].scriptCtl = sc.ctl
			if len(g.tour) == 0 {
				g.shots[start+sc.frames] = filepath.Join(outDir, sc.name+".png")
			}
		}})
		f += sc.frames + 3
	}
	g.exitFrame = f + 5
}

// putOnFoot stands p at a spot (leaving any vehicle).
func (g *Game) putOnFoot(p *Player, at Vec, yaw float64) {
	if p.Driving != nil {
		p.Driving.SetOccupied(false)
		p.Driving = nil
	}
	p.W.Pos, p.W.Yaw, p.W.vel = at, yaw, Vec{}
}
