# Wheel Loader — Construction Site

A **3D** construction/quarry game where you drive a large **articulated wheel
loader**, scoop loose material from aggregate piles, and deliver it to dump
trucks for pay. Built in Go with [Ebitengine](https://ebitengine.org/).

The world is rendered with a small **software 3D pipeline** written on top of
Ebitengine's 2D triangle rasterizer (`g3d.go`): a perspective chase camera,
real 3D box/cylinder geometry, flat directional-light shading, back-face
culling, near-plane clipping, back-to-front sorting, and a textured ground
plane. Everything — the loader, trucks, buildings, aggregate heaps — is actual
3D geometry you orbit around.

The loader is the star: it has heavy-machine inertia, true articulated (frame)
steering that bends at the center joint, and an independently controlled bucket
(lift + tilt) that scoops, carries, dumps, pushes, and spills loose material.

## Running

```sh
go run .
```

A window opens at 1280×720. You need a working OpenGL context (on a headless
Linux box, run under `xvfb-run` with `LIBGL_ALWAYS_SOFTWARE=1`).

## Controls

| Action | Keys |
| --- | --- |
| Accelerate / reverse | `W` / `S` (or `↑` / `↓`) |
| Brake | `Space` |
| Articulated steer (bend left / right) | `A` / `D` (or `←` / `→`) |
| Raise / lower loader arms | `R` / `F` |
| Curl / dump bucket | `T` / `G` |
| Zoom camera | mouse wheel or `+` / `-` |
| Orbit camera | `Q` / `E` |
| Tilt camera | `[` / `]` |
| Help · Pause · Reset loader | `Tab` · `P` · `Backspace` |

Press `Tab` in-game for the full control list.

## How to play

1. Read the current job (top-left): deliver N units of a material to a truck.
2. Drive to the matching pile. Watch the minimap (top-right).
3. Lower the arms (`F`) and level/drop the bucket edge (`G`).
4. Drive into the pile while curling the bucket (`T`) to scoop.
5. Raise the arms (`R`) and keep the bucket curled to carry without spilling.
6. Drive to the job's dump truck and position the bucket over its bed.
7. Tip the bucket forward (`G`) to unload. You're paid when the quota is met.

You can ignore the jobs entirely and just drive around, scoop dirt, push piles
about, and make a mess — that's half the point.

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
| `g3d.go` | Software 3D pipeline: projection, clipping, sorting, mesh builders |
| `camera.go` | Perspective chase/orbit camera |
| `input.go` | Keyboard → loader + camera intents |
| `render.go` | Builds the 3D scene (ground, loader, props, trucks, particles) |
| `art.go` | Procedural ground texture + fonts + value noise |
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
```

## Tests

```sh
go test ./...   # inertia, articulated steering, scooping, dumping, full cycle
go vet ./...
```
