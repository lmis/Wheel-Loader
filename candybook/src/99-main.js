// ---------- start-up, the frame loop and the keys ----------
(function () {
  const canvas = document.getElementById('page');
  const main = canvas.getContext('2d');
  const offA = document.createElement('canvas'), offB = document.createElement('canvas');
  let scale = 1;
  const reduce = window.matchMedia && window.matchMedia('(prefers-reduced-motion: reduce)').matches;

  function resize() {
    const r = canvas.getBoundingClientRect();
    const dpr = Math.min(window.devicePixelRatio || 1, 2);
    const bw = Math.max(320, Math.min(2200, Math.round(r.width * dpr)));
    canvas.width = bw; canvas.height = Math.round(bw * H / W);
    offA.width = offB.width = canvas.width; offA.height = offB.height = canvas.height;
    scale = canvas.width / W;
  }
  window.addEventListener('resize', resize);
  resize();

  Book.init();
  for (const p of Book.pages) if (reduce) p.reduce = true;

  function prep(g) {
    ctx = g;
    g.setTransform(scale, 0, 0, scale, 0, 0);
    g.lineJoin = 'round'; g.lineCap = 'round';
  }

  let last = performance.now();
  function frame(now) {
    const dt = Math.min(1 / 30, (now - last) / 1000); last = now;
    Book.update(dt);
    if (Book.turn) {
      const tr = Book.turn, lo = Math.min(tr.from, tr.to), hi = Math.max(tr.from, tr.to);
      const k = ease(tr.t / tr.dur), e = tr.dir > 0 ? k : 1 - k;
      prep(offA.getContext('2d')); drawPage(Book.pages[lo]);
      prep(offB.getContext('2d')); drawPage(Book.pages[hi]);
      prep(main); main.fillStyle = '#3a2430'; main.fillRect(0, 0, W, H);
      drawTurn(main, offA, offB, e);
    } else {
      prep(main);
      main.save(); main.translate((Book.nudge || 0) * -40, 0);
      drawPage(Book.page());
      main.restore();
      drawOverlay();
    }
    requestAnimationFrame(frame);
  }

  // page corners and dots, drawn over the page while it rests
  function drawOverlay() {
    const p = Book.page(), t = p.st.t, n = Book.pages.length;
    for (let i = 0; i < n; i++) circ(640 - (n - 1) * 11 + i * 22, 700, i === Book.cur ? 6 : 4, i === Book.cur ? '#fffaf0' : 'rgba(255,250,240,0.5)', i === Book.cur ? 2 : 0);
    const ready = !p.text || p.shown > TEXT_DELAY + 0.5 + p.text.split(' ').length / TEXT_WPS;
    if (Book.cur < n - 1) {
      // a folded-up corner bottom right that wiggles once the text has been read
      const k = ready ? 1 + Math.sin(t * 4) * 0.25 : 0.6, s = 46 * k;
      ctx.save(); ctx.shadowColor = 'rgba(40,20,30,0.35)'; ctx.shadowBlur = 8;
      poly([W, H - s, W - s, H, W - s, H - s], '#f6eedd', 2);
      ctx.restore();
      if (ready) chevron(W - 26, 360, 1, t);
    }
    if (Book.cur > 0) chevron(26, 360, -1, t);
    if (Book.cur === 0) {
      const a = 0.75 + 0.25 * Math.sin(t * 3);
      ctx.globalAlpha = a;
      word(AU.ctx ? 'Press  →  to open the book' : 'Press any key to start', 640, 398, 30, 0, '#fff3a0');
      ctx.globalAlpha = 1;
    }
  }
  function chevron(x, y, dir, t) {
    const o = Math.sin(t * 3) * 4 * dir;
    ctx.globalAlpha = 0.55;
    line(x + o - dir * 8, y - 16, x + o + dir * 6, y, 6, '#fffaf0', 2);
    line(x + o + dir * 6, y, x + o - dir * 8, y + 16, 6, '#fffaf0', 2);
    ctx.globalAlpha = 1;
  }

  // ---- input ----
  function act(k) {
    const first = !AU.ctx;
    audioStart();
    if (first && Book.cur === 0 && k !== 'next') return; // the first press just wakes the sound
    if (k === 'next') Book.go(1);
    else if (k === 'prev') Book.go(-1);
    else Book.press({ space: 0, enter: 1, ctrl: 2 }[k]);
  }
  window.addEventListener('keydown', e => {
    let k = null;
    if (e.key === 'ArrowRight' || e.key === 'PageDown') k = 'next';
    else if (e.key === 'ArrowLeft' || e.key === 'PageUp') k = 'prev';
    else if (e.key === ' ' || e.code === 'Space') k = 'space';
    else if (e.key === 'Enter') k = 'enter';
    else if (e.key === 'Control') k = 'ctrl';
    if (!k) { if (!AU.ctx) audioStart(); return; }
    e.preventDefault();
    if (e.repeat) return;
    act(k);
    flash(k);
  });
  canvas.addEventListener('pointerdown', e => {
    canvas.focus();
    const r = canvas.getBoundingClientRect(), x = (e.clientX - r.left) / r.width;
    const k = x < 0.12 ? 'prev' : x > 0.88 ? 'next' : ['space', 'enter', 'ctrl'][tapN++ % 3];
    act(k); flash(k);
  });
  let tapN = 0;
  document.querySelectorAll('[data-key]').forEach(b => b.addEventListener('click', () => { act(b.dataset.key); flash(b.dataset.key); canvas.focus(); }));
  function flash(k) {
    const b = document.querySelector(`[data-key="${k}"]`);
    if (!b) return;
    b.classList.remove('hit'); void b.offsetWidth; b.classList.add('hit');
  }
  // the page-turn buttons grey out at the ends of the book
  setInterval(() => {
    document.querySelector('[data-key="prev"]').disabled = Book.cur === 0;
    document.querySelector('[data-key="next"]').disabled = Book.cur === Book.pages.length - 1;
  }, 200);
  const mute = document.getElementById('mute');
  mute.addEventListener('click', () => {
    audioStart(); setMuted(!AU.muted);
    mute.setAttribute('aria-pressed', AU.muted ? 'true' : 'false');
    mute.setAttribute('aria-label', AU.muted ? 'Sound off' : 'Sound on');
  });
  requestAnimationFrame(frame);
})();
