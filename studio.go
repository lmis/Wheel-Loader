package main

import (
	"fmt"
	"image"
	"image/png"
	"log"
	"math"
	"os"
	"path/filepath"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// The character studio: a case-study area that renders each character in
// each of its animation poses from four angles, one contact sheet per
// character, for checking and polishing the models (go run . -studio dir).

const (
	studioCellW, studioCellH = 400, 300
	studioCols               = 4
)

// studioSpot is a clear patch of concrete in the equipment yard; only the
// ground and the subject are drawn there.
var studioSpot = V(2700, 2000)

// studioPose is one row of a sheet: how to pose the subject, and how far
// back and how high the camera looks.
type studioPose struct {
	name       string
	dist, look float64
	draw       func(g *Game, s *Scene)
}

type studioSheet struct {
	name  string
	yaw   float64 // the subject's heading
	poses []studioPose
}

var studioAngles = []struct {
	name string
	off  float64 // camera yaw relative to the subject's heading
}{{"front", 0}, {"three-quarter", 0.75}, {"side", math.Pi / 2}, {"back", math.Pi}}

func (g *Game) studioSheets() []studioSheet {
	const yaw = 0.0
	kid := func(moving, phase float64) func(g *Game, s *Scene) {
		return func(g *Game, s *Scene) {
			w := &Walker{Pos: studioSpot, Yaw: yaw, moving: moving, phase: phase}
			w.Render(g, s)
		}
	}
	animal := func(kind int, set func(a *Animal)) func(g *Game, s *Scene) {
		return func(g *Game, s *Scene) {
			a := newAnimal(kind, g.players[0], 0, 1)
			a.Pos, a.Yaw, a.blink = studioSpot, yaw, -1
			set(a)
			a.Render(g, s)
		}
	}
	dog := func(kind int) []studioPose {
		return []studioPose{
			{"standing", 46, 13, animal(kind, func(a *Animal) {})},
			{"trot, step 1", 46, 13, animal(kind, func(a *Animal) { a.speed, a.phase = 30, 0.4 })},
			{"trot, step 2", 46, 13, animal(kind, func(a *Animal) { a.speed, a.phase = 30, 1.95 })},
			{"happy hop", 46, 18, animal(kind, func(a *Animal) { a.bounce, a.t = 9, 0.6 })},
		}
	}
	skater := func(phase, lean float64, st *skaterStyle) func(g *Game, s *Scene) {
		return func(g *Game, s *Scene) {
			k := &Skater{Pos: studioSpot, Yaw: yaw, phase: phase, lean: lean, speed: skaterSpeed, style: st}
			g.renderSkater(s, k)
		}
	}
	return []studioSheet{
		{"kid", yaw, []studioPose{
			{"standing", 30, 11, kid(0, 0)},
			{"walk, step 1", 30, 11, kid(1, 0.9)},
			{"walk, step 2", 30, 11, kid(1, 0.9+math.Pi)},
			{"running", 30, 11, kid(1.6, 1.3)},
		}},
		{"golden_retriever", yaw, dog(animalGolden)},
		{"bernese", yaw, dog(animalBernese)},
		{"chimp", yaw, []studioPose{
			{"standing", 40, 13, animal(animalChimp, func(a *Animal) { a.t = 3 })},
			{"walk, step 1", 40, 13, animal(animalChimp, func(a *Animal) { a.speed, a.phase, a.t = 20, 0.4, 3 })},
			{"walk, step 2", 40, 13, animal(animalChimp, func(a *Animal) { a.speed, a.phase, a.t = 20, 1.95, 3 })},
			{"scratching its head", 40, 15, animal(animalChimp, func(a *Animal) { a.t = 1.75 })},
		}},
		{"snowy_owl", yaw, []studioPose{
			{"perched", 36, 12, animal(animalOwl, func(a *Animal) {})},
			{"blinking", 36, 12, animal(animalOwl, func(a *Animal) { a.blink = 0.1 })},
			{"happy hop", 36, 18, animal(animalOwl, func(a *Animal) { a.bounce, a.t = 9, 0.2 })},
			{"flying", 78, 34, animal(animalOwl, func(a *Animal) { a.flying, a.Y, a.t = true, 20, 0.17 })},
		}},
		{"rollerblader", yaw, []studioPose{
			{"gliding", 50, 18, skater(0, 0, &skaterTeal)},
			{"left push", 50, 18, skater(math.Pi/2, 0, &skaterTeal)},
			{"right push", 50, 18, skater(3*math.Pi/2, 0, &skaterTeal)},
			{"leaning into a turn", 50, 18, skater(0.6, 0.4, &skaterTeal)},
		}},
		{"rollerblader_pink", yaw, []studioPose{
			{"gliding", 50, 18, skater(0, 0, &skaterPink)},
			{"left push", 50, 18, skater(math.Pi/2, 0, &skaterPink)},
			{"right push", 50, 18, skater(3*math.Pi/2, 0, &skaterPink)},
			{"leaning into a turn", 50, 18, skater(0.6, 0.4, &skaterPink)},
		}},
	}
}

// renderStudio draws every sheet and writes them to dir.
func (g *Game) renderStudio(dir string) {
	for _, sh := range g.studioSheets() {
		sheet := rl.LoadRenderTexture(studioCols*studioCellW, int32(len(sh.poses)*studioCellH))
		type cell struct {
			rt   rl.RenderTexture2D
			x, y float64
			text string
		}
		var cells []cell
		for r, pose := range sh.poses {
			for c, ang := range studioAngles {
				cam := NewCamera(studioCellW, studioCellH)
				cam.Target = studioSpot
				cam.Dist, cam.TgtDist = pose.dist, pose.dist
				cam.Yaw, cam.Pitch, cam.lookHeight = sh.yaw+ang.off, 0.14, pose.look
				rt := g.renderStudioCell(cam, pose.draw)
				// each cell needs its own copy, as views of one size share a target
				cp := rl.LoadRenderTexture(studioCellW, studioCellH)
				rl.BeginTextureMode(cp)
				blit(rt, 0, 0, studioCellW, studioCellH)
				rl.EndTextureMode()
				cells = append(cells, cell{cp, float64(c * studioCellW), float64(r * studioCellH), fmt.Sprintf("%s — %s", pose.name, ang.name)})
			}
		}
		rl.BeginTextureMode(sheet)
		rl.ClearBackground(rlc(rgba(40, 44, 50, 255)))
		for _, c := range cells {
			blit(c.rt, c.x, c.y, studioCellW, studioCellH)
			g.drawText(canvas{c.x, c.y}, c.text, g.art.Bold, 10, 8, hudWhite)
		}
		rl.EndTextureMode()
		savePNG(readTarget(sheet), filepath.Join(dir, "studio_"+sh.name+".png"))
		for _, c := range cells {
			rl.UnloadRenderTexture(c.rt)
		}
		rl.UnloadRenderTexture(sheet)
	}
}

// renderStudioCell builds a scene of just the ground and one subject.
func (g *Game) renderStudioCell(cam *Camera, subject func(g *Game, s *Scene)) rl.RenderTexture2D {
	g.cam = cam
	cam.ScreenW, cam.ScreenH = studioCellW, studioCellH
	cam.setup()
	s := &g.scene
	s.reset()
	s.eye = cam.eye
	g.buildGround3D(s)
	subject(g, s)
	return g.drawScene(s, cam, studioCellW, studioCellH, 0)
}

func savePNG(img *image.RGBA, name string) {
	f, err := os.Create(name)
	if err != nil {
		log.Println("studio:", err)
		return
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		log.Println("studio:", err)
		return
	}
	fmt.Println("wrote", name)
}
