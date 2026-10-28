# Mama Comes Home

A living picture book for the web in fifteen pages. Mama says goodbye to the
twins, Léo and Louis, and takes the dino train to her office at the top of a
tall tower, where the robots are very silly. Papa rollerblades the twins to
daycare in a trolley (helmet on, wheels glowing red) and picks them up at
sunset. At dinner Mama's chair is empty, the twins want Mama, and her train
is late in the rain, until the office robots fly her home to tuck them in.

Everything on each page moves a little, and the sound (a "Coming Home"
lullaby arranged differently for every page, the background sounds of each
place, and the noises) is all synthesised with WebAudio.

| Key | What it does |
| --- | --- |
| ← → | Turn the page (with a page-curl) |
| Space (pink button) | The place: the train, the office, the street... |
| Enter (blue button) | The others on the page: the twins, the robots, the station |
| Ctrl (green button) | Whoever the page is about: Mama, Papa, or the twins |

Each press plays the next trick in that key's list, and the colour buttons
under the book always say which trick comes next. Tapping works too: the
page edges turn pages, the middle cycles through the three buttons.

`index.html` is built from `head.html`, `body.html` and `src/*.js` by
`./build.sh` and is a single self-contained file. `legacy/` holds the pages
of the earlier "Little Dino Comes Home" version, which are not built.
