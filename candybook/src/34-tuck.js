// ---------- 14. tucked in: goodnight ----------
class TuckPage extends StoryPage {
  constructor() {
    super({ name: 'tuck', floor: 720,
      text: 'Mama tucks Léo and Louis in, snug and warm. She sings their song and kisses them goodnight. Goodnight, Léo. Goodnight, Louis.',
      tricks: [['mobile', 'glowstars', 'lamp'], ['yawn', 'giggle', 'sleep'], ['sing', 'kiss', 'blanket']],
      who: ['Bedroom', 'Léo & Louis', 'Mama'],
      durs: { mobile: 3.5, glowstars: 5, lamp: 6, yawn: 2.4, giggle: 2, sleep: 5, sing: 6, kiss: 2.4, blanket: 2 } });
    this.mama = mama(this.st, { x: 640, y: 640, face: 1, look: { bow: true } });
    this.papa = papa(this.st, { x: 1150, y: 640, face: -1 });
    this.cast = [this.mama, this.papa];
    this.stars = makeStars(18, 140, 340, 160, 360);
    this.ceilingStars = makeStars(40, 20, W - 20, 120, 380);
    this.blinkA = 2; this.blinkB = 3;
  }
  act(a) {
    const st = this.st;
    if (a === 'mobile') sfx('chime', 1);
    if (a === 'glowstars') sfx('shimmer', 0.7);
    if (a === 'lamp') sfx('click', 1, 0.6);
    if (a === 'yawn') { sfx('yawn', 0.8, 1.5); setTimeout(() => sfx('yawn', 0.8, 1.3), 300); st.say('YAWN!', 640, 380, 50, '#cfe0ff'); }
    if (a === 'giggle') { sfx('giggle', 1, 1.2); st.hearts(470, 470, 2); st.hearts(810, 470, 2); }
    if (a === 'sleep') sfx('snore', 0.6);
    if (a === 'sing') { this.mama.do('talk', 5.8); sfx('lullaby', 1); }
    if (a === 'kiss') { this.mama.do('kiss', 2); sfx('smooch', 1); setTimeout(() => sfx('smooch', 1, 1.2), 500); st.hearts(470, 460, 3); setTimeout(() => this.st.hearts(810, 460, 3), 500); }
    if (a === 'blanket') { sfx('swoosh', 0.4, 1.4); st.say('SNUG AS A BUG!', 640, 380, 46, '#ffe9b4'); }
  }
  sleepy() { return clamp((this.st.t - 6) / 40, 0, 0.8); }
  tick(dt, t, A) {
    this.blinkA -= dt; if (this.blinkA < -0.2) this.blinkA = rnd(2.5, 5);
    this.blinkB -= dt; if (this.blinkB < -0.2) this.blinkB = rnd(2.5, 5);
    if (A.on('sing') && this.every('n', 0.5)) this.st.notes(this.mama.x + 40, 430, 1);
    if (A.on('sleep') && this.every('z', 1)) { this.st.zzz(500, 430); this.st.zzz(840, 430); }
    if (!A.on('sleep') && this.sleepy() > 0.6 && Math.random() < dt * 0.3) this.st.zzz(pick([500, 840]), 430);
    if (A.on('mobile') && Math.random() < dt * 6) this.st.sparkles(640, 260, 1, 80, '#fff3a0');
  }
  drawBed(x, cols, pj, blink, flip) {
    const A = this.acts, t = this.st.t;
    const g = A.on('giggle') ? A.t('giggle') : -1, wig = g >= 0 ? Math.sin(g * 24 + x) * 5 : 0;
    const yawn = A.on('yawn') ? A.t('yawn') : -1, open = yawn > 0.1 && yawn < 1.6 ? Math.sin((yawn - 0.1) / 1.5 * Math.PI) : g >= 0 ? 0.5 : 0;
    const sleepy = A.on('sleep') ? 1 : this.sleepy();
    rrect(x - 130, 420, 26, 200, 12, '#c8965a');
    ell(x - 70, 470, 60, 26, '#fffaf0');
    push(x - 66 + wig * 0.3, 446, -10 + wig, 0.75, 0.75);
    drawDinoHead({ t, cols, hair: true, open, blink, sleepy: open > 0.3 ? 0 : sleepy });
    pop();
    const tuck = A.on('blanket') ? Math.sin(Math.PI * clamp(A.t('blanket') / 2, 0, 1)) * 10 : 0;
    push(0, Math.sin(t * 1.4 + x) * 1.5 - tuck);
    ctx.beginPath(); ctx.moveTo(x - 90, 600); ctx.lineTo(x - 90, 490); ctx.quadraticCurveTo(x, 470, x + 120, 486); ctx.lineTo(x + 120, 600); ctx.closePath(); paint(pj, 3);
    ctx.save(); ctx.clip(); for (let i = 0; i < 6; i++) for (let j = 0; j < 4; j++) if ((i + j) % 2) rect(x - 90 + i * 36, 480 + j * 30, 36, 30, shade(pj, 0.4)); ctx.restore();
    rrect(x - 94, 486, 218, 12, 6, '#fffaf0', 2.5);
    pop();
    rect(x - 130, 598, 270, 14, '#a0703c');
  }
  draw() {
    const st = this.st, t = st.t, A = this.acts;
    const dark = A.on('lamp') ? ease(A.t('lamp') / 0.4) * (1 - ease((A.t('lamp') - 5.4) / 0.5)) : 0;
    drawRoom('#5a5aa0', '#7a74b4', '#b47c5a', '#9a6448');
    ctx.globalAlpha = 0.22; for (let x = 30; x < W; x += 90) for (let y = 150; y < 560; y += 90) star(x + (y / 90 % 2) * 45, y, 7, 0, 5, '#fff3c8', 0); ctx.globalAlpha = 1;
    drawWindow(120, 150, 210, 210, '#1e2456', () => { drawStars(this.stars, t); drawMoon(270, 220, 34, t, 0); });
    // the mobile over the beds
    const spin = A.on('mobile') ? t * 2.4 : t * 0.35;
    line(640, 0, 640, 150, 2, INK, 0); rrect(530, 148, 220, 8, 4, '#c8965a', 2.5);
    for (let i = 0; i < 5; i++) { const a = spin + i / 5 * TAU, px = 640 + Math.cos(a) * 100, len = 60 + (i % 2) * 30; line(px, 152, px, 152 + len, 1.5, INK, 0); push(px, 160 + len, 0, 0.8 + Math.sin(a) * 0.2, 0.8 + Math.sin(a) * 0.2); drawCandy(0, 0, Math.sin(t + i) * 0.3, 18, CANDY[i], i % 4); pop(); }
    // the robot nightlight, a little Bolt
    rrect(1000, 480, 90, 16, 5, '#c8965a'); glow(1045, 450, 80, '#8cffc8', 0.3 + dark * 0.3);
    rrect(1030, 440, 30, 40, 6, '#5acdc3', 2.5); circ(1039, 455, 3, WHITE, 0); circ(1051, 455, 3, WHITE, 0);
    // two little beds, Léo on the left, Louis on the right
    this.drawBed(470, LEO_COLS, '#7c8cf0', this.blinkA < 0, false);
    this.drawBed(810, LOUIS_COLS, '#ff8cb4', this.blinkB < 0, true);
    // Papa leaning in the doorway, Mama between the beds
    rrect(1100, 300, 140, 310, 12, '#ffd88a'); glow(1170, 450, 140, '#ffd27a', 0.3);
    this.papa.draw();
    this.mama.draw();
    if (dark > 0) {
      ctx.globalAlpha = dark * 0.55; rect(0, 0, W, H, '#0c0e2a'); ctx.globalAlpha = dark;
      for (const s of this.ceilingStars) { glow(s.x, s.y, 14, '#c8ffb4', 0.5); star(s.x, s.y, 6, 0.3, 5, '#d8ffc8', 0); }
      ctx.globalAlpha = 1;
    }
    if (A.on('glowstars')) {
      const k = A.t('glowstars'), a = clamp(k / 0.5, 0, 1) * clamp((5 - k) / 0.6, 0, 1);
      for (let i = 0; i < 16; i++) { const ang = i / 16 * TAU + k * 0.5, r = 260 + (i % 3) * 120; ctx.globalAlpha = a * 0.8; star(1045 + Math.cos(ang) * r, 420 + Math.sin(ang) * r * 0.6, 12, ang, 5, '#fff3a0', 0); }
      ctx.globalAlpha = 1;
    }
    st.draw(0); st.draw(1); st.draw(2);
    if (t > 16) word('The End', 1130, 680, 34 * Math.min(1, (t - 16)), -4, '#ffe9b4');
    st.drawWords();
  }
}
