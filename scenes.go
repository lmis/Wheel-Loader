package main

import (
	"math"
	"path/filepath"
)

// sceneSpec is one staged screenshot: set the world up, hold the given
// controls for a while, then capture.
type sceneSpec struct {
	name   string
	frames int
	setup  func(g *Game)
	ctl    *Controls // player 1 controls while the scene runs
}

// stageScenes schedules the screenshot collection of every feature.
func (g *Game) stageScenes(outDir string) {
	g.scripted = true
	g.simSpeed = 2
	w := g.world
	p1 := func() *Player { return g.players[0] }
	ex := g.findExcavator()
	cr := g.crane
	byKind := func(kind int) *Cargo {
		for _, c := range w.Cargo {
			if c.Kind == kind {
				return c
			}
		}
		return nil
	}
	walkerCam := func(p *Player, dist, userYaw, pitch float64) {
		p.distWalker, p.cam.UserYaw, p.cam.Pitch = dist, userYaw, pitch
		g.snapCamera(p)
	}
	driveCam := func(p *Player, dist, userYaw, pitch float64) {
		p.distVehicle, p.cam.UserYaw, p.cam.Pitch = dist, userYaw, pitch
		g.snapCamera(p)
	}
	homeAnimals := func() {
		for _, a := range g.animals {
			a.Pos, a.walking, a.timer, a.speed = a.home, false, 30, 0
		}
	}
	still := func(name string, pos Vec, dist, yaw, pitch float64) sceneSpec {
		return sceneSpec{name: name, frames: 2, setup: func(g *Game) {
			g.tour = []tourStop{{frame: g.frame, pos: pos, zoom: dist, yaw: yaw, pitch: pitch, shot: filepath.Join(outDir, name+".png")}}
		}}
	}

	scenes := []sceneSpec{
		{name: "01_kid_portrait", frames: 20, setup: func(g *Game) {
			p := p1()
			g.putOnFoot(p, V(1600, 1312), math.Pi/2)
			p.fp = false
			homeAnimals()
			walkerCam(p, 46, math.Pi+0.35, 0.16)
			p.cam.lookHeight = 15
		}},
		{name: "02_kid_and_loader", frames: 30, setup: func(g *Game) {
			p := p1()
			g.putOnFoot(p, V(1555, 1310), -1.15)
			walkerCam(p, 150, 0.5, 0.36)
		}, ctl: &Controls{Throttle: 0.6}},
		{name: "03_animals_group", frames: 20, setup: func(g *Game) {
			p := p1()
			homeAnimals()
			g.putOnFoot(p, V(1600, 1350), -math.Pi/2)
			walkerCam(p, 230, 0, 0.42)
		}},
		still("04_golden_retriever", V(1700, 1270), 80, 2.2, 0.32),
		still("05_bernese", V(1495, 1295), 85, 1.0, 0.32),
		still("06_chimp", V(1715, 1150), 65, 1.9, 0.3),
		still("07_snowy_owl", V(1488, 1150), 60, 1.2, 0.34),
		{name: "08_loader_digging", frames: 70, setup: func(g *Game) {
			l := g.loader
			l.Joint, l.Heading, l.Artic, l.Lift, l.Tilt, l.Load, l.Speed = V(905, 1060), -2.24, 0, 0, 0.3, 0, 0
			g.tour = nil
			g.put(p1(), l)
			p1().fp = false
			driveCam(p1(), 420, 0.8, 0.5)
		}, ctl: &Controls{Throttle: 0.5, A1: -1, A2: 1}},
		{name: "09_loader_cab_view", frames: 30, setup: func(g *Game) {
			l := g.loader
			l.Joint, l.Heading, l.Artic, l.Lift, l.Tilt, l.Load, l.Speed = V(1600, 1450), -math.Pi/2, 0.25, 0.12, 0.9, 80, 0
			l.LoadType = MatDirt
			p1().fp, p1().headYaw, p1().headPitch = true, 0, -0.08
		}, ctl: &Controls{Throttle: 0.3, Steer: 0.4}},
		{name: "10_truck_tipping", frames: 60, setup: func(g *Game) {
			tr := w.Trucks[1]
			tr.Pos, tr.Angle, tr.Speed, tr.BedTilt = V(1250, 1600), 0, 0, 0.55
			tr.Contents = [numMat]float64{}
			tr.Contents[MatGravel] = 520
			g.put(p1(), tr)
			p1().fp = false
			driveCam(p1(), 320, math.Pi/2+0.35, 0.3)
		}, ctl: &Controls{Brake: true, A1: 1}},
		{name: "11_truck_cab_view", frames: 40, setup: func(g *Game) {
			tr := w.Trucks[0]
			tr.Pos, tr.Angle, tr.Speed, tr.BedTilt = V(900, 300), 0, 120, 0
			g.put(p1(), tr)
			p1().fp, p1().headYaw, p1().headPitch = true, 0, 0
		}, ctl: &Controls{Throttle: 0.6}},
		{name: "12_excavator_digging", frames: 70, setup: func(g *Game) {
			ex.Pos, ex.Heading, ex.Slew = V(2255, 1520), 0, 0
			ex.Boom, ex.Stick, ex.Bucket, ex.Load = 0.5, 0.3, 0.15, 0
			tc := w.Trucks[2]
			tc.Pos, tc.Angle, tc.Speed, tc.BedTilt = V(2235, 1660), 0, 0, 0
			tc.Contents = [numMat]float64{}
			tc.Contents[MatSand] = 260
			g.put(p1(), ex)
			p1().fp = false
			driveCam(p1(), 400, 1.1, 0.42)
		}, ctl: &Controls{A1: -0.4, A2: -1, A3: 1}},
		{name: "13_excavator_cab_view", frames: 40, setup: func(g *Game) {
			ex.Boom, ex.Stick, ex.Bucket = 0.75, 0.2, 0.9
			p1().fp, p1().headYaw, p1().headPitch = true, 0.35, -0.15
		}, ctl: &Controls{A4: 0.6}},
		{name: "14_crane_lifting", frames: 80, setup: func(g *Game) {
			c := byKind(cargoContainer)
			g.hookUp(c)
			cr.HookLen -= 120
			g.put(p1(), cr)
			p1().fp = false
			driveCam(p1(), 560, 0.9, 0.45)
		}, ctl: &Controls{Steer: 0.8}},
		{name: "15_crane_cab_view", frames: 40, setup: func(g *Game) {
			p1().fp, p1().headYaw, p1().headPitch = true, 0, 0
		}, ctl: &Controls{Throttle: -0.5}},
		{name: "16_cargo_on_truck", frames: 70, setup: func(g *Game) {
			cr.held.held, cr.held = false, nil
			tr := w.Trucks[0]
			tr.Pos, tr.Angle, tr.Speed, tr.BedTilt = V(2150, 1080), 0.2, 0, 0
			tr.Contents = [numMat]float64{}
			bed := tr.bedRect()
			for i, kind := range []int{cargoBricks, cargoCrate} {
				c := byKind(kind)
				c.held, c.truck = false, nil
				c.Pos = bed.Center.Add(dir(tr.Angle).Scale(float64(i)*30 - 15))
				c.Y, c.vy, c.Yaw = 70, 0, tr.Angle
			}
			g.put(p1(), tr)
			p1().fp = false
			driveCam(p1(), 300, 0.7, 0.42)
		}, ctl: &Controls{}},
		{name: "17_pickup", frames: 50, setup: func(g *Game) {
			car := g.findCar("Yellow pickup")
			car.Pos, car.Angle, car.Speed = V(2300, 2100), math.Pi, 200
			g.put(p1(), car)
			p1().fp = false
			driveCam(p1(), 330, 0.5, 0.35)
		}, ctl: &Controls{Throttle: 1, Steer: 0.15}},
		{name: "18_first_person_on_foot", frames: 30, setup: func(g *Game) {
			p := p1()
			g.putOnFoot(p, V(1715, 1420), -0.82)
			p.fp, p.W.Pitch = true, 0.42
			g.snapCamera(p)
		}, ctl: &Controls{Throttle: 0.5}},
		{name: "19_coop_split_screen", frames: 50, setup: func(g *Game) {
			homeAnimals()
			l := g.loader
			l.Joint, l.Heading, l.Artic, l.Lift, l.Tilt, l.Load, l.Speed = V(1600, 1180), -math.Pi/2, 0, 0.2, 0.8, 0, 0
			g.put(p1(), l)
			p1().fp = false
			g.setCoop(true)
			p2 := g.players[1]
			g.putOnFoot(p2, V(1640, 1350), -math.Pi/2-0.3)
			driveCam(p1(), 480, 0.3, 0.55)
			walkerCam(p2, 160, 0, 0.36)
			p2.scriptCtl = &Controls{Throttle: 0.4, Steer: -0.2}
		}, ctl: &Controls{Throttle: 0.5, Steer: -0.4}},
		{name: "20_coop_both_driving", frames: 50, setup: func(g *Game) {
			p2 := g.players[1]
			tr := w.Trucks[1]
			tr.Pos, tr.Angle, tr.Speed, tr.BedTilt = V(1600, 1820), math.Pi, 0, 0
			g.put(p2, tr)
			driveCam(p2, 420, 0.4, 0.5)
			p2.scriptCtl = &Controls{Throttle: 0.5, Steer: 0.3}
		}, ctl: &Controls{Throttle: 0.6}},
		{name: "21_loader_dumping_into_truck", frames: 24, setup: func(g *Game) {
			g.setCoop(false)
			tr := w.Trucks[2]
			tr.Pos, tr.Angle, tr.Speed, tr.BedTilt = V(2500, 300), 0, 0, 0
			tr.Contents = [numMat]float64{}
			tr.Contents[MatSand] = 120
			l := g.loader
			l.Joint = tr.Pos.Add(dir(tr.Angle).Scale(118))
			l.Heading, l.Artic, l.Lift, l.Tilt, l.Load, l.LoadType, l.Speed = math.Pi, 0, 0.85, 0.7, 95, MatSand, 0
			g.put(p1(), l)
			p1().fp = false
			driveCam(p1(), 300, 0.9, 0.42)
		}, ctl: &Controls{Brake: true, A2: -1}},
		still("22_workshop_yard", V(2520, 520), 520, 1.2, 0.62),
		still("23_houses_and_trees", V(1250, 170), 640, 1.85, 0.42),
		{name: "24_site_overview", frames: 2, setup: func(g *Game) {
			g.setCoop(false)
			g.tour = []tourStop{{frame: g.frame, pos: V(1850, 1300), zoom: 1550, yaw: 1.1, pitch: 0.72, shot: filepath.Join(outDir, "24_site_overview.png")}}
		}},
	}

	f := 5
	for _, sc := range scenes {
		sc := sc
		start := f
		g.script = append(g.script, scriptStep{at: start, fn: func(g *Game) {
			g.tour = nil
			sc.setup(g)
			g.players[0].scriptCtl = sc.ctl
			if len(g.tour) == 0 {
				g.shots[start+sc.frames] = filepath.Join(outDir, sc.name+".png")
			}
		}})
		f += sc.frames + 3
	}
	g.exitFrame = f + 5
}

// put seats p in v, taking them out of whatever they were in.
func (g *Game) put(p *Player, v Vehicle) {
	if p.Driving == v {
		return
	}
	if p.Driving != nil {
		p.Driving.SetOccupied(false)
		p.Driving = nil
	}
	if q := g.driverOf(v); q != nil {
		q.Driving = nil
	}
	g.enter(p, v)
}

// putOnFoot stands p at a spot (leaving any vehicle).
func (g *Game) putOnFoot(p *Player, at Vec, yaw float64) {
	if p.Driving != nil {
		p.Driving.SetOccupied(false)
		p.Driving = nil
	}
	p.W.Pos, p.W.Yaw, p.W.vel, p.W.Pitch = at, yaw, Vec{}, 0
}

// snapCamera moves p's chase camera straight to its resting position.
func (g *Game) snapCamera(p *Player) {
	c := p.cam
	switch v := p.Driving; {
	case p.fp:
		return
	case v != nil:
		tgt, h, head := v.Center(), 40.0, v.Facing()
		if ct, ok := v.(interface {
			CamTarget() (Vec, float64, float64)
		}); ok {
			tgt, h, head = ct.CamTarget()
		}
		c.Target, c.followYaw, c.lookHeight = tgt, head, h
		c.Dist, c.TgtDist = p.distVehicle, p.distVehicle
	default:
		c.Target, c.followYaw, c.lookHeight = p.W.Pos, p.W.Yaw, 14
		c.Dist, c.TgtDist = p.distWalker, p.distWalker
	}
}

// hookUp positions the crane hook over an item and attaches it.
func (g *Game) hookUp(item *Cargo) {
	c := g.crane
	off := item.Pos.Sub(c.Base)
	c.Slew, c.Trolley = off.Angle(), clamp(off.Len(), 70, craneJibLen-14)
	c.hook, c.hookV = c.trolleyPos(), Vec{}
	c.HookLen = craneJibY - 6 - (item.Y + item.H) - 4
	c.attach(g.world)
}

func (g *Game) findExcavator() *Excavator {
	for _, v := range g.vehicles {
		if e, ok := v.(*Excavator); ok {
			return e
		}
	}
	return nil
}

func (g *Game) findCar(label string) *Car {
	for _, v := range g.vehicles {
		if c, ok := v.(*Car); ok && c.Label == label {
			return c
		}
	}
	return nil
}
