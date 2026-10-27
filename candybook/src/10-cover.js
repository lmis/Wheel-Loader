// ---------- the cover ----------
class CoverPage {
  constructor() {
    this.name = 'cover'; this.text = null;
    this.st = new Stage(); this.st.floor = 640;
    this.acts = new Acts({ letters: 1.6, rain: 3 });
    this.stars = makeStars(50, 0, W, 20, 300);
    this.floaters = Array.from({ length: 14 }, (_, i) => ({ x: rnd(40, W - 40), y: rnd(80, 600), ph: rnd(0, 6), col: CANDY[i % 7], kind: i % 4, s: rnd(12, 20) }));
    this.dino = new Dino(this.st, { x: 640, ground: x => 640 - 30 * Math.cos((x - 640) / 500), minX: 420, maxX: 860, tricks: ['jump', 'spin', 'roar', 'wave'],
      host: { waveWord: 'HELLO!', roarAt: 900 } });
  }
  press(k) {
    const st = this.st;
    if (k === 2) return this.dino.press();
    if (k === 0) { this.acts.run('letters'); sfx('boing', 0.7, 1.3); st.sparkles(640, 200, 16, 300); }
    else { this.acts.run('rain'); sfx('shimmer', 0.6); }
  }
  update(dt) {
    const st = this.st, A = this.acts;
    st.update(dt); A.update(dt);
    if (A.on('rain') && A.t('rain') < 2 && Math.random() < dt * 14) st.candy(rnd(80, W - 80), -20, rnd(-40, 40), 0);
    this.dino.update(dt);
  }
  draw() {
    const st = this.st, t = st.t, A = this.acts;
    rect(0, 0, W, H, grad(0, 0, 0, H, ['#2f3a7c', '#8a6fc0', '#ff9fb0', '#ffd08a']));
    drawStars(this.stars, t, 0.8);
    // hills: the factory on the left, home on the right, the path between
    ctx.beginPath(); ctx.moveTo(0, 560); ctx.quadraticCurveTo(320, 470, 640, 560); ctx.quadraticCurveTo(960, 470, W, 520); ctx.lineTo(W, H); ctx.lineTo(0, H); ctx.closePath(); paint('#5f9a7e', 3);
    ctx.beginPath(); ctx.moveTo(0, H); for (let x = 0; x <= W; x += 20) ctx.lineTo(x, 640 - 30 * Math.cos((x - 640) / 500) + 6); ctx.lineTo(W, H); ctx.closePath(); paint('#79b67e', 3);
    // the factory
    rect(120, 400, 150, 110, '#ff9bb4'); ctx.lineWidth = 3; ctx.strokeStyle = INK; ctx.strokeRect(120, 400, 150, 110);
    for (let i = 0; i < 3; i++) poly([120 + i * 50, 400, 170 + i * 50, 400, 170 + i * 50, 370], '#ff6e96');
    rrect(140, 320, 24, 82, 4, '#aa6ec8');
    if (Math.floor(t * 0.8) !== Math.floor((t - 1 / 60) * 0.8)) st.puff(152, 318, 10, '#fff0f5', 0);
    // home
    rect(1000, 420, 130, 90, '#f5e6c8'); ctx.strokeRect(1000, 420, 130, 90); poly([985, 424, 1145, 424, 1065, 350], '#d2505a');
    rrect(1020, 446, 30, 30, 5, '#ffe08a'); glow(1035, 461, 50, '#ffd27a', 0.5); rrect(1072, 460, 34, 50, 14, '#5aa06e');
    for (const f of this.floaters) drawCandy(f.x + Math.sin(t * 0.7 + f.ph) * 10, f.y + Math.sin(t + f.ph) * 12, Math.sin(t * 0.5 + f.ph) * 0.4, f.s, f.col, f.kind);
    // the title, each letter bobbing; Space makes them jump
    const title = ['Little Dino', 'Comes Home'];
    title.forEach((line1, r) => {
      ctx.font = '700 104px Fredoka, "Baloo 2", system-ui, sans-serif';
      const total = ctx.measureText(line1).width;
      let x = 640 - total / 2;
      [...line1].forEach((ch, i) => {
        const w = ctx.measureText(ch).width, n = r * 11 + i;
        let jump = 0;
        if (A.on('letters')) { const k = A.t('letters') - n * 0.05; if (k > 0 && k < 0.6) jump = Math.sin(k / 0.6 * Math.PI) * 50; }
        if (ch !== ' ') word(ch, x + w / 2, 150 + r * 112 + Math.sin(t * 2 + n * 0.5) * 5 - jump, 104, Math.sin(t * 1.5 + n) * 4, CANDY[n % 7]);
        x += w;
      });
    });
    word('A Candy Factory bedtime story', 640, 360, 30, 0, '#fffaf0');
    st.draw(0); st.draw(1);
    this.dino.draw();
    st.draw(2);
    st.drawWords();
  }
}
