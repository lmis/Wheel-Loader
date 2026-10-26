package main

import (
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	screenW = 1280
	screenH = 720
	fixedDT = 1.0 / 60.0
)

// Game is the top-level Ebitengine game.
type Game struct {
	world    *World
	loader   *Loader // player 1's loader
	vehicles []Vehicle
	crane    *TowerCrane // player 2's crane, by the cargo yard
	crane2   *TowerCrane // player 1's crane, its mirror image
	animals  []*Animal
	players  []*Player
	cam      *Camera // camera of the view being built (players[0] outside rendering)

	ground *ebiten.Image // baked site ground texture
	skirt  *ebiten.Image // repeating dirt for the land beyond the site
	atlas  *ebiten.Image // object-pass texture: white block + material grain
	art    *Art

	paused bool
	frame  int

	simSpeed int     // simulation substeps per rendered frame (default 1)
	scene    Scene   // reused 3D geometry buffer
	renderMS float64 // smoothed CPU time per rendered frame

	// screenshot scheduling
	shots     map[int]string
	exitFrame int // >0: exit when frame reaches it
	done      bool

	tour     []tourStop   // static landmark views for verification
	script   []scriptStep // staged scenes for verification screenshots
	scripted bool
}

type tourStop struct {
	frame int
	pos   Vec
	zoom  float64 // camera distance
	yaw   float64 // 0 = default angle
	pitch float64 // 0 = default angle
	shot  string
}

// scriptStep runs fn once the frame counter reaches at.
type scriptStep struct {
	at   int
	fn   func(g *Game)
	done bool
}

func NewGame() *Game {
	w := NewWorld()
	l := &Loader{Joint: V(1600, 1200), Heading: -math.Pi / 2, Tilt: 0.8, Lift: 0.2}
	g := &Game{
		world:    w,
		loader:   l,
		art:      NewArt(),
		shots:    map[int]string{},
		simSpeed: 1,
	}
	g.ground = buildGround(w)
	g.skirt = buildDirtTile()
	g.atlas = buildAtlas()

	// the fleet
	g.vehicles = append(g.vehicles, l,
		&Loader{Joint: V(2650, 1980), Heading: 0.6, Artic: 0.22, Lift: 0.02, Tilt: 0.5, Color: rgba(226, 150, 30, 255)})
	for _, t := range w.Trucks {
		g.vehicles = append(g.vehicles, t)
	}
	g.vehicles = append(g.vehicles,
		&Excavator{Pos: V(2500, 1820), Heading: 2.4, Slew: 0.35, Boom: 0.55, Stick: 0.45, Bucket: 0.5},
		&Car{Pos: V(2440, 410), Angle: 0.15, Model: "racing/vehicle-truck-yellow"},
		&Car{Pos: V(830, 690), Angle: -0.4, Model: "racing/vehicle-truck-red"},
		&Car{Pos: V(2820, 470), Angle: 1.7, Model: "racing/vehicle-truck-green"},
	)
	g.crane = NewTowerCrane(V(1850, 1250), -0.6)
	g.vehicles = append(g.vehicles, g.crane)
	// loose cargo in reach of the crane
	w.Cargo = append(w.Cargo,
		newCargo(cargoBricks, V(1985, 1170), 0.2),
		newCargo(cargoBricks, V(2025, 1185), 0.25),
		newCargo(cargoBeams, V(2020, 1300), 0.5),
		newCargo(cargoPipe, V(1950, 1360), 1.2),
		newCargo(cargoPipe, V(1985, 1385), 1.25),
		newCargo(cargoCrate, V(2075, 1235), 0),
		newCargo(cargoContainer, V(2110, 1360), 0.9),
	)
	// player 1's crane mirrors the site crane across the road, with its own
	// loads
	const mid = 1600.0
	g.crane2 = NewTowerCrane(V(2*mid-g.crane.Base.X, g.crane.Base.Y), math.Pi-g.crane.Slew)
	g.vehicles = append(g.vehicles, g.crane2)
	for _, c := range w.Cargo[:len(w.Cargo)] {
		w.Cargo = append(w.Cargo, newCargo(c.Kind, V(2*mid-c.Pos.X, c.Pos.Y), math.Pi-c.Yaw))
	}
	// the animals that come along with the kids
	g.animals = []*Animal{
		newAnimal(animalGolden, V(1700, 1270), 2.6, 1),
		newAnimal(animalBernese, V(1495, 1295), 0.4, 2),
		newAnimal(animalChimp, V(1715, 1150), 3.4, 3),
		newAnimal(animalOwl, V(1488, 1150), 0, 4),
	}

	g.players = []*Player{newPlayer(1, l.Joint, l.Heading), newPlayer(2, l.Joint, l.Heading)}
	g.setupLanes()
	g.cam = g.players[0].cam
	return g
}

func (g *Game) Layout(outW, outH int) (int, int) { return screenW, screenH }

func (g *Game) frameTime() float64 { return float64(g.frame) * fixedDT }

func (g *Game) Update() error {
	// in staged screenshot runs, hold everything still until a due shot has
	// been captured (headless, Update can run ahead of Draw)
	if g.scripted {
		for f := range g.shots {
			if f <= g.frame {
				return nil
			}
		}
	}
	g.frame++
	for i := range g.script {
		if st := &g.script[i]; !st.done && g.frame >= st.at {
			st.done = true
			st.fn(g)
		}
	}

	if g.scripted {
		for _, p := range g.players {
			p.ctl = Controls{}
			if p.scriptCtl != nil {
				p.ctl = *p.scriptCtl
			}
		}
	} else {
		g.handleInput()
	}

	steps := g.simSpeed
	if steps < 1 {
		steps = 1
	}
	for s := 0; s < steps && !g.paused; s++ {
		g.step(fixedDT)
	}

	if len(g.tour) > 0 {
		// static landmark views: snap the camera and schedule a capture
		for i := range g.tour {
			ts := &g.tour[i]
			if ts.shot != "" && g.frame >= ts.frame {
				c := g.players[0].cam
				c.Target = ts.pos
				c.Dist, c.TgtDist = ts.zoom, ts.zoom
				c.Yaw, c.Pitch = 0.85, 0.62
				if ts.yaw != 0 {
					c.Yaw = ts.yaw
				}
				if ts.pitch != 0 {
					c.Pitch = ts.pitch
				}
				g.shots[g.frame] = ts.shot
				ts.shot = ""
				break
			}
		}
	} else {
		for _, p := range g.players {
			g.updatePlayerCamera(p, fixedDT)
		}
	}
	g.cam = g.players[0].cam

	if g.exitFrame > 0 && g.frame >= g.exitFrame && len(g.shots) == 0 {
		g.done = true
		return ebiten.Termination
	}
	return nil
}

// step advances the whole simulation by one fixed timestep.
func (g *Game) step(dt float64) {
	w := g.world
	g.toddlerStep(dt)
	for _, v := range g.vehicles {
		if p := g.driverOf(v); p != nil {
			v.Drive(&p.ctl)
		} else {
			v.Drive(nil)
		}
	}
	for _, p := range g.players {
		g.toddlerAssist(p, dt)
	}
	g.toddlerTrucks(dt)
	g.rebuildObstacles()
	for _, v := range g.vehicles {
		v.Update(dt, w)
	}
	for _, p := range g.players {
		if p.Driving == nil {
			p.W.steer(p.ctl)
			p.W.Update(dt, w)
		}
	}
	for _, c := range w.Cargo {
		c.Update(dt, w)
	}
	for _, a := range g.animals {
		var look *Vec
		best := math.Inf(1)
		for _, p := range g.players {
			pos := p.W.Pos
			if p.Driving != nil {
				pos = p.Driving.Center()
			}
			if d := pos.Sub(a.Pos).Len(); d < best {
				best, look = d, &pos
			}
		}
		a.Update(dt, w, look, g.vehicles)
	}
	w.Update(dt)
	for _, v := range g.vehicles {
		v.Effects(w)
	}
}

// rebuildObstacles registers every moving collider with the world.
func (g *Game) rebuildObstacles() {
	w := g.world
	obs := w.obstacles[:0]
	for _, v := range g.vehicles {
		obs = append(obs, v.Colliders()...)
	}
	for _, p := range g.players {
		if p.Driving == nil {
			obs = append(obs, p.W.Colliders()...)
		}
	}
	for _, c := range w.Cargo {
		if o, ok := c.collider(); ok {
			obs = append(obs, o)
		}
	}
	for _, a := range g.animals {
		obs = append(obs, a.collider())
	}
	w.obstacles = obs
}

func (g *Game) Draw(screen *ebiten.Image) {
	start := time.Now()
	g.render(screen)
	// smoothed CPU cost of building and submitting the frame (diagnostics)
	g.renderMS = lerp(g.renderMS, float64(time.Since(start).Microseconds())/1000, 0.2)
	g.maybeCapture(screen)
}
