package main

import (
	"fmt"
	"image"
	"image/png"
	"log"
	"math"
	"os"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
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
	skater := func(phase, lean float64) func(g *Game, s *Scene) {
		return func(g *Game, s *Scene) {
			k := &Skater{Pos: studioSpot, Yaw: yaw, phase: phase, lean: lean, speed: skaterSpeed}
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
			{"gliding", 50, 18, skater(0, 0)},
			{"left push", 50, 18, skater(math.Pi/2, 0)},
			{"right push", 50, 18, skater(3*math.Pi/2, 0)},
			{"leaning into a turn", 50, 18, skater(0.6, 0.4)},
		}},
	}
}

// renderStudio draws every sheet and writes them to dir.
func (g *Game) renderStudio(dir string) {
	for _, sh := range g.studioSheets() {
		img := ebiten.NewImage(studioCols*studioCellW, len(sh.poses)*studioCellH)
		img.Fill(rgba(40, 44, 50, 255))
		cell := ebiten.NewImage(studioCellW, studioCellH)
		for r, pose := range sh.poses {
			for c, ang := range studioAngles {
				cam := NewCamera(studioCellW, studioCellH)
				cam.Target = studioSpot
				cam.Dist, cam.TgtDist = pose.dist, pose.dist
				cam.Yaw, cam.Pitch, cam.lookHeight = sh.yaw+ang.off, 0.14, pose.look
				cell.Clear()
				g.renderStudioCell(cell, cam, pose.draw)
				g.drawText(cell, fmt.Sprintf("%s — %s", pose.name, ang.name), g.art.Bold, 10, 8, hudWhite)
				op := &ebiten.DrawImageOptions{}
				op.GeoM.Translate(float64(c*studioCellW), float64(r*studioCellH))
				img.DrawImage(cell, op)
			}
		}
		savePNG(img, filepath.Join(dir, "studio_"+sh.name+".png"))
	}
}

// renderStudioCell builds a scene of just the ground and one subject.
func (g *Game) renderStudioCell(dst *ebiten.Image, cam *Camera, subject func(g *Game, s *Scene)) {
	g.cam = cam
	b := dst.Bounds()
	cam.ScreenW, cam.ScreenH = float64(b.Dx()), float64(b.Dy())
	cam.setup()
	s := &g.scene
	s.reset()
	s.eye = cam.eye
	g.buildGround3D(s)
	subject(g, s)
	s.Flush(dst, cam, g.ground, g.skirt, g.atlas, g.art.white)
}

func savePNG(img *ebiten.Image, name string) {
	w, h := img.Bounds().Dx(), img.Bounds().Dy()
	buf := make([]byte, 4*w*h)
	img.ReadPixels(buf)
	out := &image.RGBA{Pix: buf, Stride: 4 * w, Rect: image.Rect(0, 0, w, h)}
	f, err := os.Create(name)
	if err != nil {
		log.Println("studio:", err)
		return
	}
	defer f.Close()
	if err := png.Encode(f, out); err != nil {
		log.Println("studio:", err)
		return
	}
	fmt.Println("wrote", name)
}
