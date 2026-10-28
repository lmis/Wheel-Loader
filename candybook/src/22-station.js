// ---------- 2. the dino train ----------
class StationPage extends StoryPage {
  constructor() {
    super({ name: 'station', floor: 700,
      text: 'Mama hops on the dino train. Choo choo! Off it rolls, over the bridge and into the busy city.',
      tricks: [['whistle', 'chug', 'hearts'], ['conductor', 'pigeons', 'clock'], ['wave', 'peek', 'kiss']],
      who: ['Dino train', 'Station', 'Mama'],
      durs: { whistle: 2, chug: 3, hearts: 3, conductor: 2, pigeons: 3, clock: 2.4, wave: 2.4, peek: 2.4, kiss: 2.4 } });
    this.trainX = 1120;
    this.railY = () => art('bg-02-station') ? 508 : 598;
  }
  act(a) {
    const st = this.st;
    if (a === 'whistle') { sfx('choo', 1); st.say('CHOO CHOO!', 1020, 300, 60, WHITE); for (let i = 0; i < 6; i++) st.puff(this.trainX - 210 + rnd(-10, 10), this.railY() - 258, rnd(18, 26), '#ffffff', 2, { vy: rnd(-120, -70) }); }
    if (a === 'chug') sfx('chug', 1);
    if (a === 'hearts') sfx('sparkle', 0.7);
    if (a === 'conductor') { sfx('whistle', 0.6, 1.6); st.say('ALL ABOARD!', 220, 360, 48, '#ffd25a'); }
    if (a === 'pigeons') { sfx('flutter', 1); sfx('coo', 0.6); }
    if (a === 'clock') { [0, 0.5, 1].forEach(d => setTimeout(() => sfx('ding', 0.7, 0.6), d * 1000)); st.say('DONG!', 420, 230, 46, '#ffe0b4'); }
    if (a === 'wave') { sfx('giggle', 0.6, 1.2); st.say('BYE-BYE!', this.trainX - 380, this.railY() - 268, 50, '#d2b4ff'); }
    if (a === 'peek') sfx('squeak', 0.5, 1.2);
    if (a === 'kiss') { sfx('smooch', 1); for (let i = 0; i < 5; i++) st.add({ kind: 'heart', layer: 2, x: this.trainX - 380, y: this.railY() - 158, vx: -rnd(160, 260), vy: rnd(-60, 0), rot: 0, size: 11, col: '#ff6082', life: 2 }); }
  }
  tick(dt, t, A) {
    const st = this.st;
    this.trainX = 1120 + (A.on('chug') ? Math.sin(Math.PI * clamp(A.t('chug') / 3, 0, 1)) * 60 : 0);
    if (this.every('idle', 1.4)) st.puff(this.trainX - 210, this.railY() - 258, 12, '#f0f0f8', 0);
    if (A.on('chug') && this.every('c', 0.25)) st.puff(this.trainX - 210, this.railY() - 258, 14, '#ffffff', 0, { vy: -90 });
    if (A.on('hearts') && this.every('h', 0.3)) st.add({ kind: 'heart', layer: 2, x: this.trainX - 210, y: this.railY() - 268, vx: rnd(-30, 30), vy: -80, rot: 0, size: 14, col: '#ffb4cc', life: 2.4 });
  }
  draw() {
    const st = this.st, t = st.t, A = this.acts;
    const bg = art('bg-02-station'), ry = this.railY();
    if (bg) drawArt(bg);
    else {
    rect(0, 0, W, H, grad(0, 0, 0, 500, ['#8ccdff', '#cfeaff', '#fff1d8']));
    for (let i = 0; i < 4; i++) cloud((i * 340 + t * 8) % (W + 200) - 100, 170 + (i % 2) * 50, 1.1);
    drawCity(t, 470, '#a0b4dc', 0.1, 1);
    // the bridge in the distance, where the train will go
    rect(0, 470, W, 30, '#7ab0d8');
    // station roof, platform, clock
    rect(0, 220, W, 22, '#5a6e9a'); for (let x = 60; x < W; x += 300) line(x, 242, x, 560, 12, '#5a6e9a');
    const clockT = A.on('clock') ? A.t('clock') : 0;
    push(420, 280, Math.sin(clockT * 12) * 4 * (clockT ? 1 : 0));
    circ(0, 0, 38, WHITE, 4); const m = t * 0.4; line(0, 0, Math.sin(m) * 26, -Math.cos(m) * 26, 3, INK, 0); line(0, 0, Math.sin(m / 12) * 16, -Math.cos(m / 12) * 16, 4, INK, 0);
    pop();
    drawRails(590);
    }
    // the train, with Mama at a window
    const peek = A.on('peek') ? Math.sin(Math.PI * clamp(A.t('peek') / 2.4, 0, 1)) : 0;
    const waving = A.on('wave'), kissing = A.on('kiss');
    drawTrainArt(this.trainX + 60, ry, t, { moving: A.on('chug'), cars: 2, window: { car: 0, seat: 1, draw: (x, y) => {
      const head = rigImg(MAMA_RIG, kissing ? 'head-kiss' : waving ? 'head-talk' : 'head-happy');
      if (head && rigReady(MAMA_RIG)) {
        // painted Mama, facing the way the train will go, leaning out to peek
        const arm = rigImg(MAMA_RIG, 'arm');
        if (waving) { push(x - 14, y - 2, -120 + Math.sin(t * 14) * 22, -0.3, 0.3); ctx.drawImage(arm, -arm.naturalWidth * 0.25, -arm.naturalHeight * 0.2); pop(); }
        push(x - peek * 6, y - 26 - peek * 26, -peek * 14, -0.3, 0.3);
        ctx.drawImage(head, -head.naturalWidth / 2, -head.naturalHeight * 0.6);
        pop();
      } else {
        push(x - peek * 6, y - 10 - peek * 26, -peek * 14, -0.55, 0.55);
        drawFace(MAMA_RIG, { t, cols: MAMA_COLS, bow: true, open: waving || kissing ? 0.4 : 0, blink: false });
        pop();
        if (waving) line(x - 18, y + 4, x - 36 + Math.sin(t * 14) * 6, y - 40, 6, MAMA_COLS.body);
      }
    } } });
    // platform in front
    if (!bg) { rect(0, 620, W, H - 620, grad(0, 620, 0, H, ['#d9cfc4', '#c4b8ab'])); rect(0, 616, W, 8, '#ffd25a'); }
    // the conductor robot
    const c = A.on('conductor') ? A.t('conductor') : -1;
    if (robotReady()) { // the painted robot as conductor, cap on, whistle up
      ctx.save(); ctx.filter = ROBOT_TINT[0];
      pic(robotArt('body'), 200, 668, 104); pic(robotArt(c >= 0 ? 'head-talk' : 'head-happy'), 196, 590, 92);
      ctx.restore();
      rrect(166, 498, 64, 14, 4, '#3c3c50'); rrect(176, 482, 44, 20, 6, '#3c3c50');
      if (c >= 0) rrect(236, 560, 18, 8, 3, '#ffd250');
    } else {
    push(200, 660, 0);
    rrect(-24, -80, 48, 70, 10, '#5acdc3'); rrect(-26, -126, 52, 44, 10, '#78dcd2');
    rrect(-30, -136, 60, 14, 4, '#3c3c50'); rrect(-20, -150, 40, 18, 6, '#3c3c50');
    circ(-10, -106, 6, WHITE, 2); circ(10, -106, 6, WHITE, 2); circ(-10, -106, 2.5, INK, 0); circ(10, -106, 2.5, INK, 0);
    line(24, -60, 40 + (c >= 0 ? 6 : 0), c >= 0 ? -100 : -40, 6, '#aaaabe');
    if (c >= 0) rrect(36, -112, 18, 8, 3, '#ffd250');
    line(-12, -10, -12, 4, 6, '#3c3c50'); line(12, -10, 12, 4, 6, '#3c3c50');
    pop();
    }
    // pigeons on the platform
    const pf = A.on('pigeons') ? A.t('pigeons') : -1;
    [380, 450, 620].forEach((x, i) => {
      let bx = x, by = 650, flap = null;
      if (pf >= 0) { const k = clamp(pf / 2.8, 0, 1); bx += Math.sin(k * TAU) * 140 * (i % 2 ? 1 : -1); by -= Math.sin(k * Math.PI) * 200; flap = k < 1 ? pf * 26 + i : null; }
      drawBird(bx, by - (flap === null ? Math.max(0, Math.sin(t * 3 + i)) * 4 : 0), 0.9, flap, '#a0a0c8', i % 2 ? 1 : -1);
    });
    // a bench with a newspaper
    if (!bg) rrect(720, 640, 160, 14, 5, '#c8965a'); if (!bg) { rrect(720, 604, 160, 12, 5, '#c8965a'); line(740, 654, 740, 690, 6, '#5a5a6e'); line(860, 654, 860, 690, 6, '#5a5a6e'); }
    this.finish();
  }
}
