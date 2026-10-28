// ---------- 3. up the tall tower ----------
class TowerPage extends StoryPage {
  constructor() {
    super({ name: 'tower', floor: 720,
      text: 'Mama works at the very top of a big, tall tower. Up, up, up goes the glass lift. Ding!',
      tricks: [['zoom', 'lights', 'blimp'], ['washer', 'birds', 'copter'], ['press', 'hello', 'wow']],
      who: ['Tower', 'City friends', 'Mama'],
      durs: { zoom: 3, lights: 3.4, blimp: 9, washer: 3, birds: 3, copter: 6, press: 1.4, hello: 2.2, wow: 1.8 } });
  }
  liftY() {
    const base = lerp(620, 170, ease(this.st.t / 9)), A = this.acts;
    if (!A.on('zoom')) return base;
    const k = A.t('zoom') / 3;
    return lerp(base, 620, Math.sin(Math.PI * clamp(k * 1.2, 0, 1)) ** 2);
  }
  act(a) {
    const st = this.st;
    if (a === 'zoom') sfx('whee', 0.8, 0.9);
    if (a === 'lights') sfx('chime', 0.6);
    if (a === 'blimp') sfx('engine', 0.4, 1.6);
    if (a === 'washer') { sfx('squeak', 0.7, 1.4); st.bubbles(450, 360, 8); }
    if (a === 'birds') { sfx('tweet', 0.6); sfx('flutter', 0.6); }
    if (a === 'copter') sfx('copter', 1);
    if (a === 'press') { sfx('ding', 1, 1.2); st.say('DING!', 960, this.liftY() - 80, 52, '#ffd25a'); }
    if (a === 'hello') { sfx('giggle', 0.6, 1.2); st.say('HELLO, CITY!', 1060, this.liftY() - 90, 44, '#d2b4ff'); }
    if (a === 'wow') { sfx('sparkle', 0.8); st.say('WOW!', 1000, this.liftY() - 90, 58, '#fff3a0'); st.sparkles(940, this.liftY(), 10, 60); }
  }
  tick(dt, t, A) {
    if (this.at('zoom', 1.4)) { sfx('whee', 0.8, 1.2); this.st.say('WHOOSH!', 1030, 540, 50, '#8cd2ff'); }
    if (this.st.t > 9 && this.st.t - dt <= 9) sfx('ding', 0.6, 1.2);
  }
  draw() {
    const st = this.st, t = st.t, A = this.acts;
    const bg = art('bg-03-tower'), tx = 560, tw = 320;
    if (bg) drawArt(bg);
    else {
      rect(0, 0, W, H, grad(0, 0, 0, H, ['#6cb8f5', '#bfe3ff', '#fff3dc']));
      for (let i = 0; i < 5; i++) cloud((i * 290 + t * 10) % (W + 200) - 100, 160 + (i % 3) * 70, 1 + (i % 2) * 0.4);
      drawCity(t, 720, '#9aa8d0', 0.15, 2);
      // the tower: glass from top to bottom

      rect(tx, 70, tw, 650, '#6e86c8'); ctx.lineWidth = 4; ctx.strokeStyle = INK; ctx.strokeRect(tx, 70, tw, 660);
      const lit = A.on('lights') ? A.t('lights') : -1;
      for (let r = 0; r < 16; r++) for (let c = 0; c < 6; c++) {
        const x = tx + 14 + c * 51, y = 90 + r * 40;
        const on = lit >= 0 && Math.abs((16 - r) * 0.12 - lit) < 0.25 + (c === 2 || c === 3 ? 0.2 : 0);
        rect(x, y, 40, 30, on ? '#ffe08a' : 'rgba(200,230,255,0.55)');
        if (!on) rect(x + 4, y + 3, 8, 20, 'rgba(255,255,255,0.4)');
      }
      rect(tx - 10, 60, tw + 20, 16, '#4a5a90');
    }
    if (A.on('blimp')) { const x = -200 + A.t('blimp') * 190, y = 230 + Math.sin(t) * 8; ell(x, y, 110, 40, '#ff9ac8'); rrect(x - 30, y + 34, 60, 22, 6, '#fff6ea'); poly([x - 110, y, x - 140, y - 30, x - 140, y + 30], '#ff78b4'); word('HELLO!', x, y, 30, 0, WHITE); }
    // the window washer robot on its little platform
    const wy = 330 + Math.sin(t * 0.5) * 30, ws = A.on('washer') ? Math.sin(t * 12) * 16 : Math.sin(t * 2) * 6;
    line(tx + 40, 70, tx + 40, wy, 2, INK, 0); line(tx + 140, 70, tx + 140, wy, 2, INK, 0);
    rrect(tx + 30, wy, 120, 12, 4, '#ffd250');
    rrect(tx + 72, wy - 50, 34, 46, 8, '#ffa555'); circ(tx + 89, wy - 66, 16, '#ffa555'); circ(tx + 89, wy - 68, 6, WHITE, 2); circ(tx + 89, wy - 68, 2.5, INK, 0);
    line(tx + 104, wy - 36, tx + 124 + ws, wy - 70, 5, '#aaaabe'); rect(tx + 116 + ws, wy - 84, 24, 6, '#5a5a6e');
    // the lift shaft and the glass lift with Mama inside
    const lx = 900, ly = this.liftY();
    rect(lx - 6, 70, 12, 650, '#4a5a90'); line(lx + 70, 70, lx + 70, ly - 80, 2, INK, 0);
    rrect(lx + 10, ly - 80, 120, 110, 12, 'rgba(220,240,255,0.6)'); ctx.lineWidth = 4; ctx.strokeStyle = INK; ctx.beginPath(); ctx.roundRect(lx + 10, ly - 80, 120, 110, 12); ctx.stroke();
    push(lx + 70, ly + 26, 0, -0.55, 0.55);
    drawDino({ t, cols: MAMA_COLS, bow: true, bag: true, open: A.on('hello') || A.on('wow') ? 0.6 : 0, wave: A.on('hello') ? Math.sin(t * 14) : null, reach: A.on('press'), cheer: A.on('wow'), blink: false, arm: 0 });
    pop();
    rect(lx + 10, ly + 30, 120, 10, '#4a5a90');
    // birds and the little helicopter
    if (A.on('birds')) { const k = A.t('birds') / 3; for (let i = 0; i < 4; i++) drawBird(W + 40 - k * (W + 200) + i * 50, 200 + i * 20 + Math.sin(t * 4 + i) * 10, 0.8, t * 24 + i, '#6482d2', -1); }
    if (A.on('copter')) {
      const k = A.t('copter') / 6, x = lerp(-100, W + 100, k), y = 300 + Math.sin(k * TAU) * 80;
      ell(x, y, 46, 26, '#e6505a'); rrect(x - 90, y - 6, 60, 10, 4, '#e6505a'); ell(x + 18, y - 4, 16, 12, '#cdeeff', 2);
      push(x, y - 30, 0, Math.cos(t * 40), 1); line(-60, 0, 60, 0, 5, INK, 0); pop(); line(x, y - 26, x, y - 30, 4, INK, 0);
    }
    this.finish();
  }
}
