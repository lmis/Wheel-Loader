# Wheel Loader — Construction Site

A **3D** construction/quarry game where you drive a large **articulated wheel
loader**, scoop loose material from aggregate piles, and deliver it to dump
trucks for pay. Built in Go with [Ebitengine](https://ebitengine.org/).

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

The loader is the star: it has heavy-machine inertia, true articulated (frame)
steering that bends at the center joint, and an independently controlled bucket
(lift + tilt) that scoops, carries, dumps, pushes, and spills loose material.

## Running

```sh
go run .
```

A window opens at 1280×720. You need a working OpenGL context (on a headless
Linux box, run under `xvfb-run` with `LIBGL_ALWAYS_SOFTWARE=1`).

## Toddler mode (the default)

The game starts in **toddler mode**: two players, split screen. Each player
has their own wheel loader, dump truck and tower crane, a sand or dirt heap,
and a stack of crane loads. There are only four direction keys and one
button, and nobody can climb out by accident.

| | Player 1 | Player 2 |
| --- | --- | --- |
| Drive / steer (crane: slew and trolley) | `W` `A` `S` `D` | arrow keys (or `H` `J` `K` `L`) |
| The button | `Space` | `Enter` |
| Pick loader / truck / crane | `1` `2` `3` | `8` `9` `0` (or numpad `1` `2` `3`) |

Picking a machine makes the kid stop, climb out, walk over (round anything in
the way) and climb in. The machines do the fiddly parts themselves:

- **Loader:** drive into the heap and the bucket fills, then lifts. Drive up
  to the truck and it tips in. The button dumps the bucket anywhere.
- **Dump truck:** the button tips the body all the way, the load pours out,
  and the body comes back down.
- **Crane:** left/right swings the jib, up/down runs the trolley out and in.
  The hook stays at a safe height and its shadow shows where it is. The
  button lowers it to hook onto a load, or to put the load down.

Confetti celebrates every load into a truck, every truck tipped out, and
every crane load set on a truck. Parked full trucks empty themselves.
The machines are speed-limited and stop when no key is held. `F5` switches to
the full game (and back); `go run . -full` starts there.

## The site

In the full game you start in the cab of the wheel loader. Climb out and walk around as a kid
site worker (orange hi-vis vest, yellow hard hat, blonde hair, brown eyes), and
climb into any machine:

| Machine | What it does |
| --- | --- |
| **Wheel loaders** (×2) | Articulated steering, independent arm lift and bucket tilt: scoop, carry, dump, push. |
| **Dump trucks** (×3) | Car-like steering about the rear tandem; heavier when loaded. Raise the body and the load pours out the back as real material. They are the job delivery targets. |
| **Excavator** | Tank-steered tracks, a slewing house with inertia, and a boom / stick / bucket that really digs from the heaps and loads trucks. |
| **Tower crane** | Sit in the cab at the top of the mast: slew the jib, run the trolley, hoist a swinging hook, and lift brick pallets, steel beams, concrete pipes, crates and a container. Loads stack, and ride along on a truck bed. |
| **Pickups** (×3) | Kenney models; quick runabouts. |

A golden retriever, a Bernese mountain dog, a small chimp and a snowy owl (on
its fence post) hang around the starting area. They look at you and get out
of the way of moving machines.

Every machine has a **first-person cab view** (`V`), and on foot you can look
around in first person too.

## Controls

Two players share one keyboard: player 1 uses the left half, player 2 the
right half (vim-style `HJKL`). Player 2 joins by pressing `N` (or `F4`), and
the screen splits left / right.

| Action | Player 1 | Player 2 |
| --- | --- | --- |
| Throttle / reverse (walk on foot) | `W` / `S` | `K` / `J` (or `↑` / `↓`) |
| Steer (turn on foot) | `A` / `D` | `H` / `L` (or `←` / `→`) |
| Brake (run on foot) | `Space` / `LShift` | `RShift` |
| Implement 1: lift arms · dump body · boom · hoist | `R` / `F` | `I` / `,` |
| Implement 2: bucket tilt · stick | `T` / `G` | `O` / `.` |
| Implement 3: excavator bucket · crane hook on / off | `Z` / `X` | `P` / `;` |
| Implement 4: excavator slew (strafe on foot) | `Q` / `E` | `Y` / `U` |
| Get in / out | `C` | `N` |
| Cab / first-person view | `V` | `M` |
| Zoom | `1` / `2` (mouse wheel when solo) | `9` / `0` |

Global keys: `F1` help (lists the current machine's controls, with your
keys), `F2` pause, `F3` reset machines, `F4` co-op on/off, `F5` toddler mode on/off. Solo, click to look
around with the mouse (`Esc` frees it); arrow keys also drive.

## How to play

1. Read the current job (top-left): deliver N units of a material to a truck.
2. Drive to the matching pile. Watch the minimap (top-right).
3. Load a truck with the wheel loader (scoop, lift, tip) or the excavator,
   or drive the truck itself and tip its load elsewhere.
4. You're paid when the job's quota is in the right truck.

You can ignore the jobs entirely and just drive around, scoop dirt, lift
containers, push piles about, and make a mess — that's half the point.

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
| `jobs.go` | Job generation, progress, money / level progression |
| `vehicle.go` | `Vehicle` interface, controls, the loader as a vehicle |
| `truck.go`, `excavator.go`, `crane.go`, `car.go` | The other machines |
| `cargo.go` | Crane loads: falling, stacking, riding on trucks |
| `toddler.go` | Toddler mode: tiny controls, machine assistants, walking over, confetti |
| `walker.go` | The kid on foot |
| `animals.go` | The dogs, the chimp and the owl |
| `player.go`, `keys.go` | Players, getting in and out, cameras, key maps |
| `g3d.go` | Software 3D pipeline: projection, clipping, sorting, mesh builders |
| `camera.go` | Perspective chase/orbit camera |
| `input.go` | Keyboard and mouse → player controls |
| `render.go` | Builds the 3D scene (ground, loader, props, trucks, particles) |
| `art.go` | Procedural ground/dirt textures, grain atlas, fonts, noise |
| `model.go` | Minimal glTF/GLB loader for the embedded Kenney models |
| `hud.go` | HUD, gauges, minimap |
| `autopilot.go` | Hands-free AI that plays the full loop (also test instrumentation) |
| `main.go` | Entry point, demo/screenshot modes |

Simulation, input and rendering are kept separate. The update step is a fixed
1/60 s so physics is stable regardless of frame rate.

## Developer modes

```sh
go run . -demo              # hands-free autopilot plays the job loop
go run . -shots ./shots     # headless: drive a full cycle, save screenshots
go run . -dumpshot ./shots  # headless: render a loaded loader dumping into a truck
go run . -tour ./shots      # headless: static views of site landmarks
go run . -scenes ./shots    # headless: stage and capture every feature (30 shots)
go run . -full              # start in the full game instead of toddler mode
go run . -coop              # start the full game in two-player split screen
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
                # excavator dig-and-load, crane lift onto a truck, animals, co-op,
                # toddler mode: loading, tipping, crane lifts, walking over
go vet ./...
```
