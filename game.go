package main

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	screenW = 1280
	screenH = 720
	fixedDT = 1.0 / 60.0
)

// Game is the top-level Ebitengine game.
type Game struct {
	world  *World
	loader *Loader
	cam    *Camera
	jobs   *JobBoard

	ground *ebiten.Image
	art    *Art

	showHelp bool
	paused   bool
	frame    int

	auto     *Autopilot // non-nil in demo/shot mode
	simSpeed int        // simulation substeps per rendered frame (default 1)

	// screenshot scheduling
	shots     map[int]string
	exitFrame int // >0: exit when frame reaches it
	done      bool

	tour []tourStop // static landmark views for verification
}

type tourStop struct {
	frame int
	pos   Vec
	zoom  float64
	shot  string
}

func NewGame() *Game {
	w := NewWorld()
	l := &Loader{
		Joint:   V(1600, 1200),
		Heading: -math.Pi / 2,
		Tilt:    0.8,
		Lift:    0.2,
	}
	cam := NewCamera(screenW, screenH)
	cam.Pos = l.Joint
	cam.Zoom = 1.1
	cam.TgtZoom = 1.1
	g := &Game{
		world:    w,
		loader:   l,
		cam:      cam,
		jobs:     NewJobBoard(w),
		art:      NewArt(),
		shots:    map[int]string{},
		simSpeed: 1,
	}
	g.ground = buildGround(w)
	return g
}

func (g *Game) Layout(outW, outH int) (int, int) { return screenW, screenH }

func (g *Game) Update() error {
	g.frame++

	if g.auto == nil {
		g.handleInput()
	}

	steps := g.simSpeed
	if steps < 1 {
		steps = 1
	}
	for s := 0; s < steps && !g.paused; s++ {
		if g.auto != nil {
			g.auto.Apply(g)
		}
		g.loader.Update(fixedDT, g.world)
		g.world.Update(fixedDT)
		g.jobs.Update(fixedDT)
		g.emitEffects()
	}

	if len(g.tour) > 0 {
		// static landmark views: snap the camera and schedule a capture
		for i := range g.tour {
			ts := &g.tour[i]
			if ts.shot != "" && g.frame >= ts.frame {
				g.cam.Pos = ts.pos
				g.cam.Zoom, g.cam.TgtZoom = ts.zoom, ts.zoom
				g.shots[g.frame] = ts.shot
				ts.shot = ""
				break
			}
		}
	} else {
		// Camera follows the loader, aiming a little ahead.
		lead := dir(g.loader.Heading).Scale(g.loader.Speed * 0.35)
		g.cam.Follow(g.loader.Joint, lead, fixedDT)
	}

	// Camera kick on hard impacts / big loads landing handled via bounceV.
	if math.Abs(g.loader.bounceV) > 60 {
		g.cam.Kick(math.Min(10, math.Abs(g.loader.bounceV)*0.05))
	}

	if g.exitFrame > 0 && g.frame >= g.exitFrame && len(g.shots) == 0 {
		g.done = true
		return ebiten.Termination
	}
	return nil
}

// emitEffects spawns dust and exhaust based on the loader's motion.
func (g *Game) emitEffects() {
	l := g.loader
	w := g.world
	spd := math.Abs(l.Speed)
	if spd > 15 {
		inten := clamp(spd/maxSpeed, 0, 1) * 0.8
		// dust from rear wheels
		ra := l.RearAxle()
		rh := l.RearHeading()
		left := ra.Add(dir(rh + math.Pi/2).Scale(bodyWidth / 2))
		right := ra.Add(dir(rh - math.Pi/2).Scale(bodyWidth / 2))
		w.spawnDriveDust(left, l.Heading, inten)
		w.spawnDriveDust(right, l.Heading, inten)
	}
	// exhaust from the engine stack at the rear body
	stack := l.RearAxle().Add(dir(l.RearHeading()).Scale(10)).Add(dir(l.RearHeading() + math.Pi/2).Scale(10))
	w.spawnExhaust(stack, l.engineRPM)
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.render(screen)
	g.maybeCapture(screen)
}
