# Little Dino Comes Home

A living picture book for the web in eleven pages: Little Dino leaves the
candy factory at the end of the day and walks home through the park and the
town, over the river, past the farm, through the woods and up the hill, to
Mama, a bath and bed. Everything on each page moves a little, and the
sound (a "Coming Home" lullaby that changes instrument on every page, the
background sounds of each place, and the noises) is all synthesised with
WebAudio.

| Key | What it does |
| --- | --- |
| ← → | Turn the page (with a page-curl) |
| Space (pink button) | The big thing on the page |
| Enter (blue button) | The friends on the page |
| Ctrl (green button) | Little Dino does a trick |

Each press plays the next trick in that key's list, and the colour buttons
under the book always say which trick comes next. Tapping works too: the
page edges turn pages, the middle cycles through the three keys.

`index.html` is built from `head.html`, `body.html` and `src/*.js` by
`./build.sh` and is a single self-contained file.
