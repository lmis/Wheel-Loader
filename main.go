package main

import (
	"flag"
	"fmt"
	"image"
	"image/png"
	"log"
	"math"
	"os"
	"path/filepath"
	"runtime/pprof"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	scenesDir := flag.String("scenes", "", "stage and capture a screenshot of every feature to this dir, then exit")
	only := flag.String("only", "", "with -scenes: capture only scenes whose name contains this")
	tourDir := flag.String("tour", "", "render static views of site landmarks to this dir, then exit")
	frames := flag.Int("frames", 0, "exit after this many frames (0 = run forever)")
	cpuProfile := flag.String("cpuprofile", "", "write a CPU profile to this file")
	flag.Parse()

	if *cpuProfile != "" {
		f, err := os.Create(*cpuProfile)
		if err != nil {
			log.Fatal(err)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal(err)
		}
		defer pprof.StopCPUProfile()
	}

	ebiten.SetWindowSize(screenW, screenH)
	ebiten.SetWindowTitle("Wheel Loader — Construction Site")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetTPS(60)

	g := NewGame()
	if *frames > 0 {
		g.exitFrame = *frames
	}
	if *scenesDir != "" {
		if err := os.MkdirAll(*scenesDir, 0o755); err != nil {
			log.Fatal(err)
		}
		g.stageScenes(*scenesDir, *only)
	}
	if *tourDir != "" {
		if err := os.MkdirAll(*tourDir, 0o755); err != nil {
			log.Fatal(err)
		}
		g.tour = []tourStop{
			{frame: 8, pos: V(640, 620), zoom: 430, shot: filepath.Join(*tourDir, "t1_office.png")},
			{frame: 24, pos: V(1050, 820), zoom: 460, shot: filepath.Join(*tourDir, "t2_barriers.png")},
			{frame: 40, pos: V(2580, 1900), zoom: 470, shot: filepath.Join(*tourDir, "t3_equipment.png")},
			{frame: 56, pos: V(1250, 170), zoom: 640, yaw: 1.85, pitch: 0.42, shot: filepath.Join(*tourDir, "t4_houses.png")},
			{frame: 72, pos: V(2520, 520), zoom: 520, yaw: 1.2, shot: filepath.Join(*tourDir, "t5_workshop.png")},
			{frame: 88, pos: V(2950, 1300), zoom: 620, yaw: math.Pi - 0.3, pitch: 0.38, shot: filepath.Join(*tourDir, "t6_trees.png")},
			{frame: 104, pos: V(1600, 1250), zoom: 1700, yaw: 1.1, pitch: 0.7, shot: filepath.Join(*tourDir, "t7_overview.png")},
			{frame: 120, pos: V(2550, 440), zoom: 300, yaw: -math.Pi/2 + 0.4, pitch: 0.35, shot: filepath.Join(*tourDir, "t8_cabin.png")},
		}
		g.exitFrame = 136
	}

	if err := ebiten.RunGame(g); err != nil && err != ebiten.Termination {
		log.Fatal(err)
	}
	if g.done {
		fmt.Println("finished")
	}
}

func (g *Game) maybeCapture(screen *ebiten.Image) {
	// Draw may skip frame numbers when Update outruns it headlessly, so
	// capture the first draw at or past each scheduled frame.
	target := -1
	name := ""
	for f, n := range g.shots {
		if g.frame >= f && (target == -1 || f < target) {
			target, name = f, n
		}
	}
	if target < 0 {
		return
	}
	delete(g.shots, target)
	w, h := screen.Bounds().Dx(), screen.Bounds().Dy()
	buf := make([]byte, 4*w*h)
	screen.ReadPixels(buf)
	img := &image.RGBA{Pix: buf, Stride: 4 * w, Rect: image.Rect(0, 0, w, h)}
	f, err := os.Create(name)
	if err != nil {
		log.Println("shot:", err)
		return
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		log.Println("shot encode:", err)
		return
	}
	fmt.Printf("captured %s (frame %d)  loader spd=%.1f load=%.0f lift=%.2f tilt=%.2f pos=%.0f,%.0f  tris=%d cpu=%.1fms\n",
		name, g.frame, g.loader.Speed, g.loader.Load, g.loader.Lift, g.loader.Tilt, g.loader.Joint.X, g.loader.Joint.Y,
		len(g.scene.keys), g.renderMS)
}
