// ---------- 4. the silly office ----------
// The robots are the same ones from the factory code (Bolt, Gumdrop,
// Tinny); the page hands them what they look at and the shows they do.
class OfficePage extends StoryPage {
  constructor(o = {}) {
    super(Object.assign({ name: 'office', floor: 700,
      text: 'Her office is full of robots, and the robots are very, very silly. “Back to work, everyone!” says Mama.',
      tricks: [['copier', 'planes', 'coffee'], ['dance', 'heads', 'burp'], ['boss', 'tidy', 'dancetoo']],
      who: ['Office', 'Robots', 'Mama'],
      durs: { copier: 3, planes: 4, coffee: 3.4, boss: 2, tidy: 2.4, dancetoo: 4.4 } }, o));
    this.crew = { gag: null, gt: 0, pending: false, notes: 0 };
    this.machine = { lollipop: () => 0 };
    this.belt = { nearest: () => null };
    this.robots = [new Robot(this, 0, 300), new Robot(this, 1, 640), new Robot(this, 2, 980)];
    this.mama = mama(this.st, { x: 640, y: 690, face: 1, look: { bow: true } });
    this.dino = this.mama;
    this.cast = [this.mama];
    this.planes = []; this.disco = 0;
  }
  crewDo(g) { this.crew.gag = g; this.crew.gt = 0; if (g === 'dance') { this.disco = 4.4; sfx('tune', 0.9); this.st.say('ROBOT DISCO!', 640, 200, 60, '#82dcff'); } }
  act(a, k) {
    const st = this.st;
    if (k === 1) { sfx('bloop', 1); this.robots.forEach(r => r.kick(0.6)); this.crewDo(a); return; }
    if (a === 'copier') { sfx('copier', 1); st.say('WHIRR WHIRR!', 980, 260, 46, WHITE); }
    if (a === 'planes') { sfx('whee', 0.5, 1.4); for (let i = 0; i < 5; i++) this.planes.push({ x: 300, y: 380, vx: rnd(220, 420), vy: rnd(-160, 40), t: i * -0.25, rot: 0 }); }
    if (a === 'coffee') { sfx('glug', 1); st.say('GLUG GLUG!', 1160, 300, 46, '#c8966e'); }
    if (a === 'boss') {
      this.mama.do('talk', 1.6); this.crew.gag = null; this.disco = 0;
      st.say('BACK TO WORK!', 640, 300, 58, '#d2b4ff'); sfx('whistle', 0.5, 1.8);
      this.robots.forEach(r => { r.surprise = 0.6; r.kick(1.2); });
    }
    if (a === 'tidy') { this.mama.do('reach', 2); sfx('sparkle', 0.8); st.say('ALL TIDY!', 640, 320, 50, '#fff3a0'); st.sparkles(640, 500, 20, 200); this.planes = []; }
    if (a === 'dancetoo') { this.crewDo('dance'); this.mama.do('spin', 1.2); setTimeout(() => this.mama.do('laugh', 2.4), 1300); }
  }
  tick(dt, t, A) {
    const st = this.st;
    this.disco = Math.max(0, this.disco - dt);
    FactoryPage.prototype.updateCrew.call(this, dt);
    this.robots.forEach(r => r.update(dt));
    // paper flies out of the copier
    if (A.on('copier') && this.every('p', 0.06)) st.add({ kind: 'confetti', layer: 2, x: 1060, y: 420, vx: rnd(-320, -60), vy: rnd(-380, -160), rot: rnd(0, 6), vr: rnd(-8, 8), size: rnd(16, 22), col: '#fffcf2', life: rnd(2.4, 3.4), g: 300 });
    if (A.on('coffee') && this.every('c', 0.08)) st.add({ kind: 'drop', layer: 2, x: 1160, y: 410, vx: rnd(-80, 80), vy: rnd(-260, -120), size: rnd(4, 7), col: '#8c5a3c', life: 1, g: 800 });
    if (A.on('coffee') && this.every('cp', 0.3)) st.puff(1160, 380, 10, '#ffffff', 0);
    // Bolt types away, and now and then Gumdrop throws a paper plane
    if (this.every('type', 0.12) && Math.random() < 0.6) sfx('tick', 0.06, rnd(1.4, 2));
    if (this.every('amb', 6) && Math.random() < 0.6) this.planes.push({ x: 640, y: 360, vx: rnd(-300, -150), vy: -120, t: 0, rot: 0 });
    for (const p of this.planes) { p.t += dt; if (p.t < 0) continue; p.vy += 120 * dt; p.x += p.vx * dt; p.y += p.vy * dt; p.rot = Math.atan2(p.vy, p.vx); }
    this.planes = this.planes.filter(p => p.x > -60 && p.x < W + 60 && p.y < 720);
  }
  drawPlane(p) {
    push(p.x, p.y, p.rot * 57.3);
    poly([24, 0, -18, -12, -10, 0], WHITE, 2); poly([24, 0, -18, 12, -10, 0], '#e6e6f0', 2);
    pop();
  }
  drawBack(t) {
    const bg = this.name === 'office' && art('bg-04-office');
    if (bg) { drawArt(bg); return; }
    drawRoom('#dfe6f5', '#c8d2ec', '#9aa4c8', '#8690b8', 620);
    // windows onto the city, high up
    for (const x of [120, 470, 820]) drawWindow(x, 120, 300, 200, grad(0, 120, 0, 320, ['#8ccdff', '#e0f2ff']), () => { cloud(x + 80 + (t * 6) % 200, 170, 0.8); drawCity(t, 330, '#a8b8e0', 0.1, (x / 100) | 0); });
    // a plant, a clock, a water cooler
    rrect(30, 520, 60, 80, 10, '#e6905a'); for (let i = 0; i < 5; i++) { const a = -Math.PI / 2 + (i - 2) * 0.4 + Math.sin(t + i) * 0.05; line(60, 520, 60 + Math.cos(a) * 70, 520 + Math.sin(a) * 70, 8, '#5ab46e'); }
  }
  draw() {
    const st = this.st, t = st.t, A = this.acts;
    this.drawBack(t);
    const painted = art(this.name === 'office' ? 'bg-04-office' : 'bg-07-meeting');
    // the coffee machine
    if (!painted) rrect(1120, 380, 90, 140, 10, '#5a5a6e'); if (!painted) { rrect(1140, 400, 50, 30, 6, '#8ce6b4'); ell(1165, 500, 18, 6, '#8c5a3c'); }
    if (this.disco > 0 && !painted) for (let i = 0; i < 8; i++) { ctx.globalAlpha = 0.22; circ(640 + Math.cos(t * 1.5 + i) * 500, 300 + Math.sin(t * 2 + i * 2) * 160, 40, CANDY[i % 7], 0); ctx.globalAlpha = 1; }
    this.robots.forEach(r => r.drawBody());
    // desks, a computer, and the copier
    if (!painted) for (const x of [300, 640, 980]) {
      rrect(x - 130, 500, 260, 26, 8, '#c8965a'); line(x - 110, 526, x - 110, 620, 10, '#a0703c'); line(x + 110, 526, x + 110, 620, 10, '#a0703c');
    }
    if (!painted) rrect(380, 440, 80, 60, 8, '#3c3c50'); if (!painted) { rrect(388, 448, 64, 42, 4, '#8cd2ff'); for (let i = 0; i < 4; i++) rect(394, 454 + i * 9, 20 + ((t * 30 + i * 13) % 30), 4, '#ffffff'); }
    if (!painted) { rrect(1010, 420, 110, 80, 10, '#e6e6f0'); rect(1020, 412, 90, 10, '#c8c8d7'); }
    if (A.on('copier') && !painted) glow(1060, 430, 60, '#a0ffd2', 0.4 + 0.3 * Math.sin(t * 20));
    if (!painted) for (let i = 0; i < 6; i++) rrect(520 + (i % 3) * 30, 480 - Math.floor(i / 3) * 6, 26, 18, 2, WHITE, 1.5);
    this.robots.forEach(r => r.drawArms());
    if (!painted) for (const p of this.planes) if (p.t >= 0) this.drawPlane(p);
    st.draw(0); st.draw(1);
    this.mama.draw({ reach: this.mama.is('reach') });
    st.draw(2); st.drawWords();
  }
}
