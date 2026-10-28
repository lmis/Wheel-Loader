// ---------- 10. "We want Mama" ----------
class WantMamaPage extends StoryPage {
  constructor() {
    super({ name: 'wantmama', floor: 690,
      text: 'Pyjamas on, teeth brushed. But where is Mama? “We want Mama,” say Léo and Louis. It is getting very, very late.',
      tricks: [['clock', 'moon', 'rain'], ['want', 'window', 'teddy'], ['phone', 'cuddle', 'story']],
      who: ['Night', 'Léo & Louis', 'Papa'],
      durs: { clock: 2.4, moon: 3, rain: 6, want: 2.6, window: 3, teddy: 2.4, phone: 4, cuddle: 2.6, story: 3 } });
    this.leo = leo(this.st, { x: 360, y: 640, face: -1, look: { hair: true, pj: '#7c8cf0' } });
    this.louis = louis(this.st, { x: 470, y: 640, face: -1, look: { hair: true, pj: '#ff8cb4' } });
    this.papa = papa(this.st, { x: 860, y: 650, face: -1 });
    this.cast = [this.leo, this.louis, this.papa];
    this.stars = makeStars(16, 120, 320, 160, 330);
  }
  act(a) {
    const st = this.st;
    if (a === 'clock') { sfx('cuckoo', 1); st.say('SO LATE!', 1100, 220, 46, '#ffe0b4'); }
    if (a === 'moon') sfx('sparkle', 0.5, 0.8);
    if (a === 'rain') sfx('raindrops', 0.8);
    if (a === 'want') { this.leo.do('sad', 2.6); this.louis.do('sad', 2.6); sfx('sniffle', 1); st.say('WE WANT MAMA!', 420, 380, 50, '#a8c8ff'); }
    if (a === 'window') { this.leo.do('reach', 3); this.louis.do('reach', 3); sfx('squeak', 0.3, 1.4); }
    if (a === 'teddy') { this.leo.do('laugh', 2); this.louis.do('laugh', 2); st.hearts(415, 520, 3); sfx('hug', 0.5, 1.3); }
    if (a === 'phone') { sfx('ring', 1); this.papa.do('talk', 3.8); }
    if (a === 'cuddle') { this.papa.walkTo(560, 200); sfx('hug', 0.8); }
    if (a === 'story') { this.papa.do('talk', 3); st.say('ONCE UPON A TIME...', 800, 380, 40, '#ffe9b4'); sfx('swoosh', 0.3, 1.6); }
  }
  tick(dt, t, A) {
    const st = this.st;
    if (this.at('phone', 1.2)) { st.say('I’M COMING, MY LOVES!', 760, 300, 42, '#d2b4ff'); this.leo.do('cheer', 1.6); this.louis.do('cheer', 1.6); }
    if (this.at('cuddle', 1.2)) { st.hearts(470, 480, 6); this.leo.do('laugh'); this.louis.do('laugh'); }
    if (this.at('cuddle', 2.4)) this.papa.walkTo(860, 200);
    if (A.on('cuddle') && this.papa.tx === null && A.t('cuddle') > 2.4) this.papa.face = -1;
    if (A.on('want') && this.every('tear', 0.4)) for (const k of [this.leo, this.louis]) { const [hx, hy] = k.head(); st.add({ kind: 'drop', layer: 2, x: hx - k.face * 6, y: hy + 6, vx: 0, vy: 20, size: 3, col: '#9ad8ff', life: 0.7, g: 300 }); }
  }
  draw() {
    const st = this.st, t = st.t, A = this.acts;
    drawRoom('#5a5a9a', '#6e6aa8', '#8c6a5a', '#74584a', 610);
    ctx.globalAlpha = 0.18; for (let x = 40; x < W; x += 100) for (let y = 160; y < 560; y += 100) star(x + (y % 200 ? 50 : 0), y, 7, 0, 5, '#fff3c8', 0); ctx.globalAlpha = 1;
    // the window the twins look out of
    const moonK = A.on('moon') ? Math.sin(Math.PI * clamp(A.t('moon') / 3, 0, 1)) : 0;
    const raining = A.on('rain') ? Math.min(1, A.t('rain')) * Math.min(1, (6 - A.t('rain')) * 2) : 0.15;
    drawWindow(100, 150, 240, 230, '#1e2456', () => {
      drawStars(this.stars, t);
      drawMoon(280, 220 + (1 - moonK) * 30, 34, t, moonK);
      cloud(220 + Math.sin(t * 0.3) * 30, 250 - moonK * 40, 1, '#4a4a7a');
      if (raining) { ctx.save(); ctx.globalAlpha = raining; drawRain(t, 1.2); ctx.restore(); }
      // fog where noses press on the glass
      if (A.on('window')) { ctx.globalAlpha = 0.35; circ(300, 330, 22, WHITE, 0); circ(250, 334, 20, WHITE, 0); ctx.globalAlpha = 1; }
    });
    // the big clock, ticking late
    const sw = Math.sin(t * 2.2) * 10, ck = A.on('clock') ? A.t('clock') : -1;
    push(1100, 260, ck >= 0 ? Math.sin(ck * 14) * 4 : 0);
    rrect(-60, -80, 120, 300, 20, '#8c5a3c'); circ(0, -10, 44, WHITE, 3);
    for (let i = 0; i < 12; i++) { const a = i / 12 * TAU; circ(Math.cos(a) * 34, -10 + Math.sin(a) * 34, 2.5, INK, 0); }
    line(0, -10, Math.sin(-0.4) * 24, -10 - Math.cos(-0.4) * 24, 5, INK, 0); line(0, -10, 0, 22, 3, INK, 0);
    push(0, 50, sw); line(0, 0, 0, 110, 3, '#ffd250', 0); circ(0, 112, 16, '#ffd250', 3); pop();
    if (ck >= 0) drawBird(0, -70 - Math.sin(Math.PI * clamp(ck / 2.4, 0, 1)) * 30, 0.7, null, '#ffd25a', -1);
    pop();
    // the sofa
    rrect(660, 520, 380, 110, 30, '#8c9ce0'); rrect(640, 470, 60, 160, 26, '#7a8ad2'); rrect(1000, 470, 60, 160, 26, '#7a8ad2');
    // teddies
    const teddy = (x, y) => { circ(x, y - 30, 16, '#c88c5a'); circ(x - 12, y - 44, 6, '#c88c5a'); circ(x + 12, y - 44, 6, '#c88c5a'); ell(x, y, 18, 20, '#c88c5a'); circ(x - 5, y - 32, 2, INK, 0); circ(x + 5, y - 32, 2, INK, 0); };
    teddy(300, 660);
    this.leo.draw(); this.louis.draw(); this.papa.draw();
    if (A.on('phone')) { // Papa's phone, with Mama on the screen
      const [hx, hy] = this.papa.head();
      rrect(hx - 64, hy + 10, 30, 50, 6, '#3c3c50');
      const k = clamp((A.t('phone') - 0.8) * 3, 0, 1) * clamp((4 - A.t('phone')) * 3, 0, 1);
      if (k > 0) { push(hx - 120, hy - 120, 0, k, k); rrect(-70, -60, 140, 110, 20, '#fff6ea'); push(0, 10, 0, 0.7, 0.7); drawDinoHead({ t, cols: MAMA_COLS, bow: true, open: 0.3 + 0.3 * Math.sin(t * 20) }); pop(); poly([30, 50, 60, 50, 60, 80], '#fff6ea', 0); pop(); }
    }
    if (A.on('story')) { const [hx, hy] = this.papa.head(); poly([hx - 90, hy + 50, hx - 40, hy + 40, hx - 40, hy + 80, hx - 90, hy + 86], '#fffaf0'); poly([hx - 40, hy + 40, hx + 10, hy + 50, hx + 10, hy + 86, hx - 40, hy + 80], '#fffaf0'); }
    this.finish();
  }
}
