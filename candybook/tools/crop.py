"""Turns captured frames into button pictures.

For each button, picks the frame that differs most from the "before" frame,
crops a 16:9 box around what changed (at least a quarter of the page wide),
and saves it as art/btn-<page>-<trick>.webp.
    python3 tools/crop.py <raw-dir>
"""
import sys, glob, os, collections
import numpy as np
from PIL import Image, ImageFilter

raw = sys.argv[1] if len(sys.argv) > 1 else 'thumbs-raw'
art = os.path.join(os.path.dirname(__file__), '..', 'art')
groups = collections.defaultdict(dict)
for f in glob.glob(os.path.join(raw, '*__*__*.png')):
    page, trick, n = os.path.basename(f)[:-4].split('__')
    groups[(page, trick)][int(n)] = f

for (page, trick), frames in sorted(groups.items()):
    before = np.asarray(Image.open(frames[0]).convert('L'), dtype=np.float32)
    best, best_score, best_mask = None, -1, None
    for n in sorted(k for k in frames if k):
        im = Image.open(frames[n]).convert('RGB')
        d = np.abs(np.asarray(im.convert('L'), dtype=np.float32) - before)
        d = np.asarray(Image.fromarray(d.astype(np.uint8)).filter(ImageFilter.BoxBlur(6)), dtype=np.float32)
        mask = d > 18
        score = mask.sum()
        if score > best_score: best, best_score, best_mask = im, score, mask
    W, H = best.size
    if best_score < W * H * 0.002:          # hardly anything changed: show the whole page
        box = (0, 0, W, H)
    else:
        ys, xs = np.nonzero(best_mask)
        x0, x1 = np.percentile(xs, [3, 97]); y0, y1 = np.percentile(ys, [3, 97])
        cx, cy = (x0 + x1) / 2, (y0 + y1) / 2
        w = max(x1 - x0, (y1 - y0) * 16 / 9, W * 0.28) * 1.25
        w = min(w, W); h = w * 9 / 16
        x = min(max(cx - w / 2, 0), W - w); y = min(max(cy - h / 2, 0), H - h)
        box = (int(x), int(y), int(x + w), int(y + h))
    thumb = best.crop(box).resize((240, 135), Image.LANCZOS)
    out = os.path.join(art, f'btn-{page}-{trick}.webp')
    thumb.save(out, 'WEBP', quality=80)
    print(f'{page:10s} {trick:12s} -> {os.path.basename(out)}  box {box}')
