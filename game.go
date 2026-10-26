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
	world  *World
	loader *Loader
	cam    *Camera
	jobs   *JobBoard

	ground *ebiten.Image // baked site ground texture
	skirt  *ebiten.Image // repeating dirt for the land beyond the site
	atlas  *ebiten.Image // object-pass texture: white block + material grain
	art    *Art

	showHelp bool
	paused   bool
	frame    int

	auto     *Autopilot // non-nil in demo/shot mode
	simSpeed int        // simulation substeps per rendered frame (default 1)
	scene    Scene      // reused 3D geometry buffer
	renderMS float64    // smoothed CPU time per rendered frame

	// screenshot scheduling
	shots     map[int]string
	exitFrame int // >0: exit when frame reaches it
	done      bool

	tour []tourStop // static landmark views for verification
}

type tourStop struct {
	frame int
	pos   Vec
	zoom  float64 // camera distance
	yaw   float64 // 0 = default angle
	pitch float64 // 0 = default angle
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
	cam.Target = l.Joint
	cam.followYaw = l.Heading
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
	g.skirt = buildDirtTile()
	g.atlas = buildAtlas()
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
				g.cam.Target = ts.pos
				g.cam.Dist, g.cam.TgtDist = ts.zoom, ts.zoom
				g.cam.Yaw, g.cam.Pitch = 0.85, 0.62
				if ts.yaw != 0 {
					g.cam.Yaw = ts.yaw
				}
				if ts.pitch != 0 {
					g.cam.Pitch = ts.pitch
				}
				g.shots[g.frame] = ts.shot
				ts.shot = ""
				break
			}
		}
	} else {
		// Chase camera follows the loader and swings behind its heading.
		g.cam.Follow(g.loader.Joint, g.loader.Heading, fixedDT)
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
	// exhaust stack on the engine hood (matches buildLoader3D)
	stack := l.Joint.Sub(dir(l.RearHeading()).Scale(36)).Add(dir(l.RearHeading() + math.Pi/2).Scale(7))
	w.spawnExhaust(stack, l.engineRPM)
}

func (g *Game) Draw(screen *ebiten.Image) {
	start := time.Now()
	g.render(screen)
	// smoothed CPU cost of building and submitting the frame (diagnostics)
	g.renderMS = lerp(g.renderMS, float64(time.Since(start).Microseconds())/1000, 0.2)
	g.maybeCapture(screen)
}
