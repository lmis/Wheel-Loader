// ---------- 12. the robots fly Mama home ----------
class RescuePage extends StoryPage {
  constructor() {
    super({ name: 'rescue', floor: 720,
      text: 'Then, whirr! Whizz! Here come the robots! “We’ll fly you home, boss!” And up they go, over the sleepy city.',
      tricks: [['boost', 'shoot', 'fireworks'], ['loop', 'song', 'wavemoon'], ['wheee', 'hold', 'point']],
      who: ['Night sky', 'Robots', 'Mama'],
      durs: { boost: 2.6, shoot: 2.4, fireworks: 3.4, loop: 2.4, song: 3, wavemoon: 2.6, wheee: 2, hold: 2, point: 2.6 } });
    this.crew = { gag: null, gt: 0 }; this.machine = { lollipop: () => 0 }; this.belt = { nearest: () => null };
    this.robots = [new Robot(this, 0, 0), new Robot(this, 1, 0), new Robot(this, 2, 0)];
    this.mama = mama(this.st, { x: 0, y: 0, face: 1, noShadow: true, look: { bow: true, bag: true, cheer: true } });
    this.dino = this.mama; this.cast = [this.mama];
    this.stars = makeStars(90, 0, W, 110, 500);
    this.scroll = 0; this.fw = [];
  }
  act(a) {
    const st = this.st;
    if (a === 'boost') { sfx('zoom', 1, 0.8); st.say('ZOOOM!', 520, 260, 60, '#ffb45a'); }
    if (a === 'shoot') { sfx('shimmer', 1); st.say('MAKE A WISH!', 900, 200, 44, '#fff3a0'); }
    if (a === 'fireworks') for (let i = 0; i < 4; i++) setTimeout(() => { const x = rnd(150, 1150), y = rnd(250, 420); this.st.confetti(x, y, 40, 300); this.st.sparkles(x, y, 12, 60); sfx('pop', 0.8, rnd(0.5, 0.9)); }, i * 600);
    if (a === 'loop') sfx('whee', 0.8, 1);
    if (a === 'song') { sfx('beeps', 1); st.notes(560, 300, 3); }
    if (a === 'wavemoon') { sfx('giggle', 0.6, 1.4); st.say('HELLO, MOON!', 1000, 300, 44, '#fff3a0'); }
    if (a === 'wheee') { this.mama.do('laugh', 2); sfx('whee', 1, 1.2); st.say('WHEEE!', 620, 560, 60, '#d2b4ff'); }
    if (a === 'hold') { this.mama.do('talk', 1.6); st.say('HOLD ON TIGHT!', 620, 560, 46, '#d2b4ff'); sfx('squeak', 0.5); }
    if (a === 'point') { this.mama.do('talk', 2.2); st.say('THERE’S HOME!', 900, 560, 50, '#ffe08a'); sfx('sparkle', 0.8); }
  }
  pos() {
    const t = this.st.t, A = this.acts;
    let x = 560 + Math.sin(t * 0.6) * 60, y = 300 + Math.sin(t * 1.1) * 24, rot = 0;
    if (A.on('loop')) { const k = A.t('loop') / 2.4, a = ease(k) * TAU; x += Math.sin(a) * 160; y += (1 - Math.cos(a)) * -110; rot = a * 57.3; }
    if (A.on('boost')) x += Math.sin(Math.PI * A.t('boost') / 2.6) * 200;
    return [x, y, rot];
  }
  tick(dt, t, A) {
    this.scroll += dt * (A.on('boost') ? 320 : 120);
    this.robots.forEach(r => r.update(dt));
    const [x, y] = this.pos();
    if (this.every('jet', 0.05)) for (const dx of [-130, 0, 130]) this.st.add({ kind: 'trail', layer: 0, x: x + dx, y: y + 20, vx: -120, vy: 60, size: rnd(5, 9), col: '#ffa03c', life: 0.5 });
    if (A.on('wavemoon')) this.robots.forEach(r => r.kick(0.1));
  }
  draw() {
    const st = this.st, t = st.t, A = this.acts;
    const bg = art('bg-12-night-sky'); if (bg) drawArtScroll(bg, this.scroll * 0.3);
    if (!bg) {
      rect(0, 0, W, H, grad(0, 0, 0, H, ['#0f1440', '#232c6a', '#3a3a7c']));
      drawStars(this.stars, t);
    }
    if (!bg) drawMoon(1080, 190, 62, t, A.on('wavemoon') ? 1 : 0.2);
    if (A.on('shoot')) { const k = A.t('shoot') / 1.4; if (k < 1) { const x = lerp(100, 900, k), y = lerp(140, 280, k); line(x - 150, y - 26, x, y, 3, 'rgba(255,246,200,0.6)', 0); star(x, y, 12, t * 6, 5, '#fff6c8', 0); glow(x, y, 40, '#fff3a0', 0.6); } }
    // the city sliding past below, and home far away on the right
    const s = this.scroll;
    if (!bg) {
      push(-(s * 0.3) % 600, 0); for (let i = 0; i < 3; i++) { push(i * 600, 0); drawCity(t, 720, '#1c2250', 0.55, 5); pop(); } pop();
    }
    if (!bg) for (let i = 0; i < 3; i++) cloud(((i * 520 - s * 0.6) % (W + 400) + W + 400) % (W + 400) - 200, 420 + i * 60, 1.6, '#3c4282', 0.8);
    if (!bg) glow(1220, 640, 60, '#ffd27a', 0.7); if (!bg) rect(1200, 620, 40, 30, '#3a2a40'); poly([1194, 622, 1246, 622, 1220, 600], '#5a3048', 0); bg || rect(1214, 630, 10, 10, '#ffd27a');
    st.draw(0);
    // the robots, flying in a row, with Mama holding on underneath
    const [x, y, rot] = this.pos();
    push(x, y, rot);
    this.robots.forEach((r, i) => {
      r.x = (i - 1) * 130;
      push(0, -ROBOT_BASE + Math.sin(t * 3 + i) * 8, 0);
      // jets under the robots
      const f = 18 + Math.sin(t * 40 + i) * 6 + (A.on('boost') ? 20 : 0);
      if (!bg) poly([r.x - 14, ROBOT_BASE, r.x + 14, ROBOT_BASE, r.x, ROBOT_BASE + f + 20], '#ffd25a', 0); if (!bg) poly([r.x - 8, ROBOT_BASE, r.x + 8, ROBOT_BASE, r.x, ROBOT_BASE + f], '#fff6be', 0);
      r.drawBody(); r.drawArms();
      if (A.on('wavemoon') && !bg) line(r.x + 40, ROBOT_BASE - 80, r.x + 60 + Math.sin(t * 14) * 8, ROBOT_BASE - 130, 8, '#aaaabe');
      pop();
    });
    // Mama hangs on to Gumdrop's hands, swinging
    this.mama.x = 0; this.mama.y = 200 + Math.sin(t * 1.6) * 6;
    if (!bg) { line(-30, 30, -20, 200 - 140, 6, MAMA_COLS.body); line(30, 30, 20, 200 - 140, 6, MAMA_COLS.body); }
    this.mama.rot = Math.sin(t * 1.6) * 8;
    this.mama.draw();
    pop();
    if (A.on('song') && this.every('n', 0.3)) st.notes(x + rnd(-150, 150), y - 120, 1);
    st.draw(1); st.draw(2); st.drawWords();
  }
}
