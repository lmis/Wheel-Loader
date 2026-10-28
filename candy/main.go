// Candy Factory: a living storybook page. Silly robots munch candy off a
// conveyor belt, a little dino hops about, and everything wiggles. Three keys
// make things happen: Space works the candy machine, Enter sets the robots
// off, Ctrl is the dino.
package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"log"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	"golang.org/x/image/font/gofont/gobold"
)

const (
	W      = 1280
	H      = 720
	dt     = 1.0 / 60
	floorY = 600.0 // where the wall meets the floor
	feetY  = 684.0 // where the dino stands
)

type Game struct {
	t       float64
	rng     *rand.Rand
	font    rl.Font
	paper   rl.Texture2D
	sfx     *Sounds
	shake   float64
	disco   float64
	next    [3]int   // the next gag for each key
	presses []string // scripted key presses this frame (screenshot mode)

	machine Machine
	crew    Crew
	robots  []*Robot
	dino    Dino
	belt    Belt
	room    Room
	bits    []*Bit
	words   []*Word
}

func NewGame(seed int64) *Game {
	g := &Game{rng: rand.New(rand.NewSource(seed))}
	g.robots = []*Robot{newRobot(0, 245), newRobot(1, 965), newRobot(2, 1112)}
	for i, r := range g.robots {
		r.blink = 1 + float64(i)*1.3
		r.cool = 1.5 + float64(i)
	}
	g.machine.gag, g.crew.gag = -1, -1
	g.dino = Dino{X: 520, Y: feetY, face: 1, rest: 1}
	g.room.init(g.rng)
	g.belt.init()
	return g
}

func (g *Game) rnd(lo, hi float64) float64 { return lo + g.rng.Float64()*(hi-lo) }

func (g *Game) Update() {
	g.t += dt
	g.input()
	g.room.update(g)
	g.machine.update(g)
	g.belt.update(g)
	for _, r := range g.robots {
		r.update(g)
	}
	g.dino.update(g)
	g.updateBits()
	g.updateWords()
	g.shake = math.Max(0, g.shake-dt*2.5)
	g.disco = math.Max(0, g.disco-dt)
}

func pressed(keys ...int32) bool {
	for _, k := range keys {
		if rl.IsKeyPressed(k) {
			return true
		}
	}
	return false
}

func (g *Game) input() {
	space := pressed(rl.KeySpace)
	enter := pressed(rl.KeyEnter, rl.KeyKpEnter)
	ctrl := pressed(rl.KeyLeftControl, rl.KeyRightControl)
	for _, p := range g.presses {
		switch p {
		case "space":
			space = true
		case "enter":
			enter = true
		case "ctrl":
			ctrl = true
		}
	}
	g.presses = g.presses[:0]
	if space {
		g.machine.press(g)
	}
	if enter {
		g.robotsPress()
	}
	if ctrl {
		g.dino.press(g)
	}
}

// Draw paints the whole page, back to front.
func (g *Game) Draw() {
	rl.ClearBackground(c3(250, 232, 210))
	sx := math.Sin(g.t*61) * g.shake * 9
	sy := math.Cos(g.t*47) * g.shake * 7
	push(sx, sy, 0, 1, 1)
	g.room.drawBack(g)
	g.machine.drawBack(g)
	for _, r := range g.robots {
		r.drawBody(g)
	}
	g.machine.draw(g)
	g.drawBits(layerBack)
	g.belt.draw(g)
	for _, r := range g.robots {
		r.drawArms(g)
	}
	g.drawBits(layerMid)
	g.dino.draw(g)
	g.drawBits(layerFront)
	g.room.drawLights(g)
	g.drawWords()
	pop()
	rl.DrawTexturePro(g.paper, rl.NewRectangle(0, 0, float32(g.paper.Width), float32(g.paper.Height)),
		rl.NewRectangle(0, 0, W, H), rl.Vector2{}, 0, rl.White)
}

// makePaper is the page itself: a warm grain and a soft vignette laid over
// the picture so it reads like a printed book.
func makePaper(rng *rand.Rand) rl.Texture2D {
	const pw, ph = 640, 360
	img := image.NewRGBA(image.Rect(0, 0, pw, ph))
	for y := 0; y < ph; y++ {
		for x := 0; x < pw; x++ {
			dx, dy := (float64(x)/pw-0.5)*1.6, (float64(y)/ph-0.5)*1.9
			v := math.Max(0, math.Hypot(dx, dy)-0.55) * 150
			n := rng.Float64()
			a := v + 10*n
			if n > 0.5 {
				img.Set(x, y, color.RGBA{uint8(a * 0.55), uint8(a * 0.35), uint8(a * 0.3), uint8(a)})
			} else {
				img.Set(x, y, color.RGBA{uint8(a * 0.25), uint8(a * 0.15), uint8(a * 0.12), uint8(a)})
			}
		}
	}
	t := rl.LoadTextureFromImage(rl.NewImageFromImage(img))
	rl.SetTextureFilter(t, rl.FilterBilinear)
	return t
}

func main() {
	shots := flag.String("shots", "", "headless: press the keys on a script and save screenshots to this dir, then exit")
	art := flag.String("art", "candy/art", "folder of PNG pictures that replace drawn parts (see candy/art/README.md)")
	flag.Parse()

	headless := *shots != ""
	flags := uint32(rl.FlagWindowResizable | rl.FlagVsyncHint | rl.FlagMsaa4xHint)
	if headless {
		flags = rl.FlagWindowHidden
	}
	rl.SetConfigFlags(flags)
	rl.SetTraceLogLevel(rl.LogWarning)
	rl.InitWindow(W, H, "Candy Factory")
	defer rl.CloseWindow()
	rl.SetExitKey(0) // Esc does not quit; little hands find it
	rl.DisableBackfaceCulling()

	g := NewGame(time.Now().UnixNano())
	if headless {
		g = NewGame(7)
	}
	runes := make([]rune, 0, 96)
	for r := rune(32); r < 127; r++ {
		runes = append(runes, r)
	}
	g.font = rl.LoadFontFromMemory(".ttf", gobold.TTF, 96, runes)
	rl.SetTextureFilter(g.font.Texture, rl.FilterBilinear)
	g.paper = makePaper(g.rng)
	loadSprites(*art)
	if !headless {
		g.sfx = loadSounds()
		defer g.sfx.close()
	}
	target := rl.LoadRenderTexture(W, H)
	rl.SetTextureFilter(target.Texture, rl.FilterBilinear)

	script := map[int][]string{}
	if headless {
		if err := os.MkdirAll(*shots, 0o755); err != nil {
			log.Fatal(err)
		}
		script = shotScript
	}

	last, acc := time.Now(), 0.0
	frame := 0
	for !rl.WindowShouldClose() {
		steps := 1
		if !headless {
			now := time.Now()
			acc = math.Min(acc+now.Sub(last).Seconds(), 4*dt)
			last = now
			steps = int(acc / dt)
			acc -= float64(steps) * dt
		}
		for i := 0; i < steps; i++ {
			if keys, ok := script[frame]; ok {
				for _, k := range keys {
					if k != "shot" && k != "quit" {
						g.presses = append(g.presses, k)
					}
				}
			}
			g.Update()
			frame++
		}

		rl.BeginTextureMode(target)
		g.Draw()
		rl.EndTextureMode()

		if headless {
			for _, k := range script[frame-1] {
				switch k {
				case "shot":
					im := rl.LoadImageFromTexture(target.Texture)
					rl.ImageFlipVertical(im)
					name := filepath.Join(*shots, fmt.Sprintf("f%04d.png", frame-1))
					rl.ExportImage(*im, name)
					rl.UnloadImage(im)
					fmt.Println("saved", name)
				case "quit":
					return
				}
			}
		}

		// letterbox the page into the window
		sw, sh := float64(rl.GetScreenWidth()), float64(rl.GetScreenHeight())
		s := math.Min(sw/W, sh/H)
		dw, dh := W*s, H*s
		rl.BeginDrawing()
		rl.ClearBackground(c3(40, 26, 30))
		rl.DrawTexturePro(target.Texture, rl.NewRectangle(0, 0, W, -H),
			rl.NewRectangle(float32((sw-dw)/2), float32((sh-dh)/2), float32(dw), float32(dh)), rl.Vector2{}, 0, rl.White)
		rl.EndDrawing()
	}
}

// shotScript presses every gag in turn and photographs each one mid-way.
var shotScript = func() map[int][]string {
	s := map[int][]string{}
	add := func(f int, k ...string) { s[f] = append(s[f], k...) }
	add(90, "shot")
	f := 120
	for i := 0; i < 4; i++ {
		for _, k := range []string{"space", "enter", "ctrl"} {
			add(f, k)
			add(f+40, "shot")
			add(f+100, "shot")
			f += 300
		}
	}
	add(f, "shot", "quit")
	return s
}()
