package main

import (
	"os"
	"path/filepath"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// The rig: every character and prop is built from named parts, each drawn in
// its own local frame. A part is drawn by code, unless candy/art/<name>.png
// exists: then that picture is drawn instead, centred on the part's pivot and
// scaled to the part's box. So painted (or AI-made) art can replace any part
// without touching the animation.

type sprite struct{ tex rl.Texture2D }

var sprites = map[string]sprite{}

func loadSprites(dir string) {
	files, _ := filepath.Glob(filepath.Join(dir, "*.png"))
	for _, f := range files {
		if _, err := os.Stat(f); err != nil {
			continue
		}
		t := rl.LoadTexture(f)
		rl.SetTextureFilter(t, rl.FilterBilinear)
		sprites[strings.TrimSuffix(filepath.Base(f), ".png")] = sprite{t}
	}
}

// part draws the named part filling the box (x, y, w, h) in the current
// frame, either from its picture or with the drawing function.
func part(name string, x, y, w, h float64, draw func()) {
	if s, ok := sprites[name]; ok {
		rl.DrawTexturePro(s.tex, rl.NewRectangle(0, 0, float32(s.tex.Width), float32(s.tex.Height)),
			rl.NewRectangle(float32(x), float32(y), float32(w), float32(h)), rl.Vector2{}, 0, rl.White)
		return
	}
	draw()
}
