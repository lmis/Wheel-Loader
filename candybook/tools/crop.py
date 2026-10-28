"""Turns captured frames into button pictures.

For each button, picks the frame where the trick shows most (the book reports
where its new hearts, words and moving characters are), crops a 16:9 box
around them, and saves it as art/btn-<page>-<trick>.webp. Falls back to what
changed between frames when the book reports nothing.
    python3 tools/crop.py <raw-dir>
"""
import sys, glob, os, collections, json
import numpy as np
from PIL import Image, ImageFilter

raw = sys.argv[1] if len(sys.argv) > 1 else 'thumbs-raw'
art = os.path.join(os.path.dirname(__file__), '..', 'art')
groups = collections.defaultdict(dict)
for f in glob.glob(os.path.join(raw, '*__*__*.png')):
    page, trick, n = os.path.basename(f)[:-4].split('__')
    groups[(page, trick)][int(n)] = f

PAGE_W, PAGE_H = 1280, 720
for (page, trick), frames in sorted(groups.items()):
    # 1) the book's own idea of where the action is
    best_pts, best_n = [], None
    for n in sorted(k for k in frames if k):
        jf = frames[n][:-4] + '.json'
        d = json.load(open(jf)) if os.path.exists(jf) else {}
        pts = d.get('pts', []) if isinstance(d, dict) else d   # words alone don't show the trick
        if len(pts) > len(best_pts): best_pts, best_n = pts, n
    if best_pts:
        im = Image.open(frames[best_n]).convert('RGB'); W, H = im.size; k = W / PAGE_W
        xs = [x * k for x, y in best_pts]; ys = [y * k for x, y in best_pts]
        x0, x1 = np.percentile(xs, [2, 98]); y0, y1 = np.percentile(ys, [2, 98])
        cx, cy = (x0 + x1) / 2, (y0 + y1) / 2
        w = min(max(x1 - x0, (y1 - y0) * 16 / 9, W * 0.3) * 1.3, W); h = w * 9 / 16
        x = min(max(cx - w / 2, 0), W - w); y = min(max(cy - h / 2, 0), H - h)
        box = (int(x), int(y), int(x + w), int(y + h))
        im.crop(box).resize((240, 135), Image.LANCZOS).save(os.path.join(art, f'btn-{page}-{trick}.webp'), 'WEBP', quality=80)
        print(f'{page:10s} {trick:12s} focus {box}')
        continue
    # 2) otherwise: what changed between frames
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
