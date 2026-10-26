# Wheel Loader — Construction Site

A **3D** construction-site game for two small kids on one keyboard. Each kid
has their own **articulated wheel loader**, dump truck and tower crane, and
plays with four direction keys and one button. Built in Go with
[Ebitengine](https://ebitengine.org/).

The world is rendered with a small **software 3D pipeline** written on top of
Ebitengine's 2D triangle rasterizer (`g3d.go`): a perspective chase camera,
smooth (per-vertex) shading with sun + sky/ground ambient light, distance fog
and a sky gradient, back-face culling, near-plane clipping, and depth-sorted
drawing in a single batched call. Everything — the loader, trucks, buildings,
aggregate heaps — is actual 3D geometry you orbit around.

- **Aggregate heaps** are a continuous heightfield mesh built from the
  material grid each frame, smoothed, shaded and textured with a seamless
  procedural grain, so digging and dumping visibly reshape them.
- **Machines** are lofted from chamfered cross-sections: tapered hood,
  counterweight and grille, cab on the rear frame, lift arms with cylinders
  and Z-bar linkage, a U-shaped toothed bucket, and lugged tyres that turn.
- **Scenery** mixes procedural props with free **Kenney CC0 models** (houses,
  a workshop garage, pickups, trees) loaded by a small built-in glTF/GLB
  reader (`model.go`).

Objects are sorted per triangle (painter's algorithm), since Ebitengine has no
depth buffer. Large triangles are split, horizontal surfaces sort by their far
edge, and surface-mounted details get a small bias, which removes nearly all
overlap errors. A faint seam can still show at a few very close angles.

Under the simple controls the machines are properly simulated: the loader
has heavy-machine inertia, articulated (frame) steering that bends at the
centre joint, and a bucket (lift + tilt) that scoops, carries, dumps, pushes
and spills loose material; the helpers in `toddler.go` work the bucket, the
dump body and the crane hook for the kids.

## Running

```sh
go run .
```

A window opens at 1280×720. You need a working OpenGL context (on a headless
Linux box, run under `xvfb-run` with `LIBGL_ALWAYS_SOFTWARE=1`).

## How to play

The screen is split in two: player 1 on the left, player 2 on the right.
Each kid starts in their loader, with a sand or dirt heap and their dump
truck right in front of them, and their crane and a stack of loads nearby.

| | Player 1 | Player 2 |
| --- | --- | --- |
| Drive / steer (crane: swing and trolley) | `W` `A` `S` `D` | arrow keys (or `H` `J` `K` `L`) |
| The button | `Space` | `Enter` |
| Pick loader / truck / crane | `1` `2` `3` | `8` `9` `0` (or numpad `1` `2` `3`) |

`F2` pauses. Picking a machine makes the kid stop, climb out, walk over
(round anything in the way) and climb in; there is no other way out. The
machines do the fiddly parts themselves:

- **Loader:** drive into the heap and the bucket fills, then lifts. Drive up
  to the truck and it tips in. The button dumps the bucket anywhere.
- **Dump truck:** the button tips the body all the way, the load pours out,
  and the body comes back down.
- **Crane:** left/right swings the jib, up/down runs the trolley out and in.
  The hook stays at a safe height and its shadow shows where it is. The
  button lowers it to hook onto a load, or to put the load down.

Each truck takes about five bucket loads before it is full; it stays full
until its kid drives it somewhere and tips it out (and the new heap can be
scooped up again). The machines are speed-limited and stop when no key is
held.

The kids are site workers in orange hi-vis vests and yellow hard hats, with
blonde hair and brown eyes. Each kid has two animal friends who ride on
their machine: a snowy owl and a small chimp for player 1, a golden
retriever and a Bernese mountain dog for player 2. When the kid switches
machines, the animals hop down, run along beside them (the owl flies), and
hop up onto the next machine: on the cab roof and the engine hood of the
loader, on the truck's cab roof, and on the crane's footing. An excavator, a
third dump truck and three Kenney pickups stand parked around the site.

## A little wonder

Things happen around the site just to look at:

- **The animals** bounce up and down on the spot every couple of seconds
  (the owl flaps its wings).
- **A train** of a red locomotive and five coloured carriages runs along a
  line in the distance, between the north road and the houses, every so
  often. It stops and waits if anything is standing on the track.
- **A red helicopter** flies low over the middle of the site now and then,
  rotors spinning, its shadow sweeping across the ground.
- **A rollerblader** laps the roads at speed, right down the road between
  the two kids: dark helmet, long brown hair streaming behind, blue
  glasses, and red rollerblades whose wheels glow and leave a red trail. He
  swerves round anything in his way.

They are in `wonder.go`.

## Material simulation

Loose material is a coarse height grid (`material.go`). Scooping removes amount
from cells under the bucket edge; dumping deposits it (into a truck bed if one
is underneath, otherwise onto the ground where it slumps); pushing bulldozes
material between cells; a badly tilted bucket spills its load. Buckets can be
partially filled and loads carry their material type.

## Project layout

| File | Responsibility |
| --- | --- |
| `game.go` | Main loop, fixed-timestep update, effect spawning |
| `loader.go` | Articulated loader kinematics, dynamics, bucket interaction |
| `material.go` | Loose-material height grid (scoop / dump / push / slump) |
| `world.go` | Site layout, collision, trucks, particles |
| `vehicle.go` | `Vehicle` interface, controls, the loader as a vehicle |
| `truck.go`, `excavator.go`, `crane.go`, `car.go` | The other machines |
| `cargo.go` | Crane loads: falling, stacking, riding on trucks |
| `shapes.go` | Smooth shapes: ellipsoids, capsules, rounded lofts, spots |
| `studio.go` | Character studio: contact sheets of every pose from four angles |
| `wonder.go` | The train, the helicopter and the rollerblader |
| `toddler.go` | Toddler mode: tiny controls, machine assistants, walking over |
| `walker.go` | The kid on foot |
| `animals.go` | The dogs, the chimp and the owl: riding, following, hopping |
| `player.go` | Players, getting in and out, cameras |
| `g3d.go` | Software 3D pipeline: projection, clipping, sorting, mesh builders |
| `camera.go` | Perspective chase/orbit camera |
| `input.go` | Keyboard → player controls |
| `render.go` | Builds the 3D scene (ground, loader, props, trucks, particles) |
| `art.go` | Procedural ground/dirt textures, grain atlas, fonts, noise |
| `model.go` | Minimal glTF/GLB loader for the embedded Kenney models |
| `hud.go` | HUD drawing helpers, pause notice |
| `main.go` | Entry point, screenshot modes |

Simulation, input and rendering are kept separate. The update step is a fixed
1/60 s so physics is stable regardless of frame rate.

## Developer modes

```sh
go run . -scenes ./shots    # headless: stage and capture every feature
go run . -scenes ./shots -only crane  # just the scenes whose name matches
go run . -tour ./shots      # headless: static views of site landmarks
go run . -studio ./shots    # headless: contact sheets of every character in every pose
go run . -cpuprofile cpu.prof  # write a CPU profile (combine with any mode)
```

Screenshot modes print the triangle count and smoothed CPU time per frame.

## Assets and credits

All graphics are procedural except the models in `assets/kenney/`, which are by
[Kenney](https://kenney.nl) under **CC0 1.0** (public domain), taken from the
[Starter Kit City Builder](https://github.com/KenneyNL/Starter-Kit-City-Builder)
and [Starter Kit Racing](https://github.com/KenneyNL/Starter-Kit-Racing)
repositories. They are embedded into the binary. See `assets/kenney/LICENSE.txt`.

## Tests

```sh
go test ./...   # driving, scooping, dumping, entering/exiting, truck tipping,
                # excavator dig-and-load, crane lift onto a truck,
                # toddler mode: loading, tipping, crane lifts, walking over,
                # the animals riding and coming along
go vet ./...
```
