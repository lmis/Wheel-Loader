// ---------- 13. home at last ----------
class HomeAgainPage extends StoryPage {
  constructor() {
    super({ name: 'home', floor: 700,
      text: 'Mama tiptoes through the door. Léo and Louis hear her and come running. “Mama! Mama!” A big, warm family hug.',
      tricks: [['door', 'byebots', 'chimney'], ['hug', 'jump', 'giggle'], ['kiss', 'twirl', 'shh']],
      who: ['Home', 'Léo & Louis', 'Mama'],
      durs: { door: 5, byebots: 4, chimney: 4, hug: 2.6, jump: 1.4, giggle: 1.8, kiss: 2.4, twirl: 1.6, shh: 2 } });
    this.mama = mama(this.st, { x: 120, y: 660, face: 1, look: { bow: true, bag: true } });
    this.papa = papa(this.st, { x: 940, y: 640, face: -1 });
    this.leo = leo(this.st, { x: 900, y: 650, face: -1, look: { hair: true, pj: '#7c8cf0' } });
    this.louis = louis(this.st, { x: 990, y: 650, face: -1, look: { hair: true, pj: '#ff8cb4' } });
    this.cast = [this.papa, this.mama, this.leo, this.louis];
    this.stars = makeStars(50, 0, W, 110, 320);
    this.mama.walkTo(560, 120);
  }
  act(a) {
    const st = this.st;
    if (a === 'door') sfx('creak', 1);
    if (a === 'byebots') { sfx('beeps', 0.8); st.say('BYE, BOSS!', 640, 200, 50, '#a0ffd2'); }
    if (a === 'chimney') sfx('blow', 0.4, 1.4);
    if (a === 'hug') { this.hug(); }
    if (a === 'jump') { this.leo.do('hop'); setTimeout(() => this.louis.do('hop'), 120); sfx('boing', 0.5, 1.6); st.say('MAMA!', 760, 420, 56, '#ffd25a'); }
    if (a === 'giggle') { this.leo.do('laugh'); this.louis.do('laugh'); sfx('giggle', 1); }
    if (a === 'kiss') { this.mama.do('kiss', 1.6); sfx('smooch', 1); setTimeout(() => sfx('smooch', 1, 1.2), 400); st.hearts(700, 480, 5); }
    if (a === 'twirl') { this.mama.do('spin', 1.4); this.leo.do('spin', 1.4); this.louis.do('spin', 1.4); sfx('whee', 0.7, 1.2); st.sparkles(700, 520, 16, 120); }
    if (a === 'shh') { st.say('SHHH, IT’S LATE!', 560, 400, 46, '#c8e6ff'); sfx('shh', 0.8); this.mama.do('talk', 1.4); }
  }
  hug() {
    this.leo.walkTo(this.mama.x + 110, 200); this.louis.walkTo(this.mama.x + 190, 200);
    sfx('hug', 1); this.st.say('A BIG HUG!', this.mama.x + 120, 380, 56, '#ff8cc8');
    for (let i = 0; i < 4; i++) setTimeout(() => this.st.hearts(this.mama.x + 120, 470, 3), i * 300);
    this.leo.do('cheer', 2); this.louis.do('cheer', 2); this.mama.do('laugh', 2);
  }
  tick(dt, t, A) {
    if (t > 2.6 && t - dt <= 2.6) { this.leo.walkTo(700, 160); this.louis.walkTo(780, 160); this.st.say('MAMA!', 800, 420, 60, '#ffd25a'); sfx('giggle', 1, 1.1); }
    if (t > 4.2 && t - dt <= 4.2) { this.mama.face = 1; this.hug(); }
    if (A.on('chimney') && this.every('c', 0.35)) this.st.add({ kind: 'heart', layer: 0, x: 1060, y: 160, vx: rnd(-20, 20), vy: -60, rot: 0, size: rnd(12, 18), col: '#ffb4cc', life: 2.6 });
    if (this.every('sm', 0.8)) this.st.puff(1060, 170, 12, '#c8c0d8', 0);
    for (const k of [this.leo, this.louis]) if (k.tx === null) k.face = -1;
  }
  draw() {
    const st = this.st, t = st.t, A = this.acts;
    const bg = art('bg-13-home'); if (bg) drawArt(bg);
    if (!bg) {
      rect(0, 0, W, H, grad(0, 0, 0, 560, ['#18204e', '#2e3c7a', '#4a5490']));
      drawStars(this.stars, t);
    }
    drawMoon(180, 180, 46, t, 0.6);
    // the robots fly away into the night
    if (A.on('byebots')) { const k = A.t('byebots') / 4; [0, 1, 2].forEach(i => { const x = lerp(300, 1400, k) + i * 60, y = lerp(260, 120, k) + Math.sin(t * 4 + i) * 10; glow(x, y, 26, '#a0ffd2', 0.6); rrect(x - 12, y - 12, 24, 24, 6, LOOKS[i].col, 2); line(x + 12, y - 4, x + 22 + Math.sin(t * 14) * 4, y - 22, 3, '#aaaabe', 2); }); }
    else if (t < 3) { const k = t / 3; [0, 1, 2].forEach(i => { const x = lerp(200, 900, k) + i * 50, y = lerp(160, 90, k); glow(x, y, 20, '#a0ffd2', 0.5 * (1 - k)); }); }
    // the house: a warm open doorway in the middle
    if (!bg) {
      rect(0, 560, W, H - 560, grad(0, 560, 0, H, ['#4e8a68', '#3c7058']));
      rect(760, 200, 520, 400, '#f5e6c8'); ctx.lineWidth = 3; ctx.strokeStyle = INK; ctx.strokeRect(760, 200, 523, 400);
      poly([730, 210, W + 40, 210, 1050, 70], '#d2505a');
      rrect(1040, 120, 40, 70, 4, '#a05a50');
      const open = A.on('door') ? 1 : 0.85;
      glow(860, 470, 200, '#ffd27a', 0.5 * open);
      rrect(800, 340, 120, 260, 16, '#ffd88a');
      rrect(1120, 330, 110, 100, 10, '#ffe08a'); line(1175, 330, 1175, 430, 4, INK, 0);
      ctx.globalAlpha = 0.3 * open; poly([800, 600, 920, 600, 1000, H, 640, H], '#ffe9a8', 0); ctx.globalAlpha = 1;
      // path
      for (let i = 0; i < 5; i++) ell(700 - i * 130, 630 + i * 10, 40, 10, '#b4aab4', 2.5);
    }
    st.draw(0);
    this.papa.draw();
    this.mama.draw(); this.leo.draw(); this.louis.draw();
    this.finish();
  }
}
