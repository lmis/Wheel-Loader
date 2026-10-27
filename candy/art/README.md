# Painted parts

Every character and prop in Candy Factory is a rig of named parts. The game
draws each part itself, unless a picture with the part's name is in this
folder: `bolt-body.png` replaces Bolt's body, `dino.png` replaces the dino,
and so on. The picture is stretched to fill the part's box and then moves,
squashes, bounces and spins with the animation, so no code changes are needed.

Use PNGs with a **transparent background**, at the box's aspect ratio. Twice
the box size (or more) looks sharpest.

| File | Box (w × h) | What it is | Pivot |
| --- | --- | --- | --- |
| `machine.png` | 280 × 230 | The candy machine's front (no dome, no chimney) | bottom centre |
| `lollipop.png` | 144 × 264 | The giant lollipop: a 144 × 144 swirl on top, stick below | centre of the swirl |
| `bolt-body.png` | 92 × 104 | The tall teal robot's body | bottom centre |
| `bolt-head.png` | 88 × 72 | Its head | centre |
| `gumdrop-body.png` | 124 × 110 | The round orange robot's body | bottom centre |
| `gumdrop-head.png` | 92 × 80 | Its head | centre |
| `tinny-body.png` | 74 × 84 | The little purple TV robot's body | bottom centre |
| `tinny-head.png` | 100 × 76 | Its TV head | centre |
| `dino.png` | 190 × 150 | The whole dino, facing right | feet, bottom centre |

A picture is a still: replacing a head loses its blinking, its eyes
following the dino and its chomping mouth, and replacing `dino.png` loses the
dino's wagging tail and open mouth. Bodies, the machine and the lollipop lose
nothing worth missing, so start there.

A prompt that fits the drawn style:

> Children's picture-book illustration of a [cheerful round orange robot
> body with a heart on its tummy], flat pastel colours, soft brown ink
> outline, gentle shading, isolated on a transparent background, front view,
> no text
