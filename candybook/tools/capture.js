// Captures what every button does, for the button pictures.
// For each page and each of its buttons: wait until the page is quiet, take a
// "before" frame, press the button, take a few frames while the trick plays.
// Writes PNGs to the output folder; crop.py turns them into button pictures.
//   node tools/capture.js <out-dir> [page-name ...]
const { chromium } = require('playwright');
const path = require('path');
const out = process.argv[2] || 'thumbs-raw';
const only = process.argv.slice(3);
const FRAMES = [0.35, 0.7, 1.1, 1.6, 2.3]; // seconds after the press
(async () => {
  const b = await chromium.launch();
  const p = await b.newPage({ viewport: { width: 1280, height: 860 } });
  const errs = []; p.on('pageerror', e => errs.push(e.message));
  await p.goto('file://' + path.resolve(__dirname, '../index.html'));
  await p.waitForTimeout(1500);
  await p.keyboard.press('a');
  const pages = await p.evaluate(() => Book.pages.map(pg => pg.name));
  const canvas = p.locator('#page');
  for (let i = 0; i < pages.length; i++) {
    if (only.length && !only.includes(pages[i])) continue;
    await p.evaluate(i => { Book.turn = null; Book.cur = i; Book.pages[i].shown = 99; Book.pages[i].text = null; }, i);
    const names = await p.evaluate(() => Book.buttons());
    for (let j = 0; j < names.length; j++) {
      // wait for the page to settle: no trick running, nobody walking
      for (let k = 0; k < 100; k++) {
        const quiet = await p.evaluate(() => { const pg = Book.page(); return Object.keys(pg.acts ? pg.acts.a : {}).length === 0 && !pg.kiss && (pg.cast || []).every(c => c.tx === null || c.tx === undefined); });
        if (quiet) break;
        await p.waitForTimeout(100);
      }
      await p.waitForTimeout(300);
      const base = `${out}/${pages[i]}__${names[j]}`;
      await canvas.screenshot({ path: `${base}__0.png` });
      await p.evaluate(j => Book.pressButton(j), j);
      let last = 0;
      for (let f = 0; f < FRAMES.length; f++) {
        await p.waitForTimeout((FRAMES[f] - last) * 1000); last = FRAMES[f];
        await canvas.screenshot({ path: `${base}__${f + 1}.png` });
      }
      console.log('captured', pages[i], names[j]);
    }
  }
  if (errs.length) console.log('page errors:', errs);
  await b.close();
})();
