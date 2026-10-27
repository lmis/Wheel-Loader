# Candy Factory

A living storybook page for small kids: a whimsical candy factory where
three silly robots munch sweets off a conveyor belt and a little green dino
hops about. Everything on the page moves a bit, all the time, and three keys
make things happen.

```sh
./run.sh candy        # from the repository root
go run ./candy        # or by hand
```

| Key | What happens |
| --- | --- |
| Space | The candy machine does a trick: a candy fountain, a giant lollipop springing out of its dome, a bubblegum bubble that pops into confetti |
| Enter | The robots do a trick: a disco dance, heads boinging up on springs, rainbow burps |
| Ctrl | The dino does a trick: a big flip jump (THUMP!), a RAWR that makes the robots jump, catching a falling candy, riding the conveyor belt |

Each press plays the next trick in the list. Pressing while a trick is
playing gives a little sparkle and queues the next one. Esc does nothing,
so little hands can't quit by accident.

## Things that move by themselves

- Candies squeeze down the wobbly pipe and drop out of the machine's chute,
  ride the belt and tumble into the basket at the end.
- The robots grab passing candies, chomp them (crumbs!), and do a happy
  hop. Their eyes follow the dino.
- The dino hops about in little runs, looks around, and gobbles any candy it
  finds on the floor.
- The machine breathes, its gears turn, its dial eyes wobble, its lights
  chase and its chimney puffs.
- Clouds drift past the windows, the sun turns, the windmill spins, a bird
  flies by, the bunting flutters, the lamps swing, the clock ticks, and a
  mouse peeks out of its hole now and then.

The sounds are made up in code (no sound files).

## Painted art

The characters are rigs of named parts, and any part can be replaced by a
painted (or AI-made) picture: see [art/README.md](art/README.md).

## Files

| File | What's in it |
| --- | --- |
| `main.go` | Window, game loop, keys, the paper texture, screenshot mode |
| `draw.go` | Inked storybook shapes, transforms, words |
| `parts.go` | Named parts that a PNG in `art/` can replace |
| `room.go` | Wallpaper, windows, bunting, lamps, clock, mouse, floor |
| `machine.go` | The candy machine and its tricks |
| `belt.go` | The conveyor belt, the candy pipe and the basket |
| `robots.go` | The three robots, eating and their tricks |
| `dino.go` | The dino, hopping and its tricks |
| `fx.go` | Flying candies, confetti, sparkles, hearts, notes, bubbles |
| `sound.go` | The synthesised sounds |

`go run ./candy -shots shots` presses every trick on a script and saves
screenshots (headless; under `xvfb-run` with `LIBGL_ALWAYS_SOFTWARE=1` on a
box without a screen).
