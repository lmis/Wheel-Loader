# Little Dino Comes Home

A living picture book for the web: Little Dino leaves the candy factory at
the end of the day and walks home through town, over the river and up the
hill, to Mama and bed. Everything on each page moves a little, and the
sound (a "Coming Home" lullaby that changes instrument on every page, the
background sounds of each place, and the noises) is all synthesised with
WebAudio.

| Key | What it does |
| --- | --- |
| ← → | Turn the page (with a page-curl) |
| Space | The big thing on the page |
| Enter | The friends on the page |
| Ctrl | Little Dino does a trick |

Each press plays the next trick in that key's list. Tapping works too: the
page edges turn pages, the middle cycles through the three keys.

`index.html` is built from `head.html`, `body.html` and `src/*.js` by
`./build.sh` and is a single self-contained file.
