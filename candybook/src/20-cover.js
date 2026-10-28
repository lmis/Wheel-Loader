// ---------- the cover ----------
class CoverPage extends StoryPage {
  constructor() {
    super({ name: 'cover', text: null, floor: 660, tricks: [['letters'], ['cheer'], ['familywave']],
      who: ['Title', 'Léo & Louis', 'Mama & Papa'], durs: { letters: 1.6, cheer: 1.6, familywave: 2 } });
    this.mama = mama(this.st, { x: 520, y: 650, face: 1 });
    this.papa = papa(this.st, { x: 790, y: 650, face: -1, look: { helmet: true, skates: true } });
    this.leo = leo(this.st, { x: 610, y: 660, face: 1 });
    this.louis = louis(this.st, { x: 690, y: 660, face: -1 });
    this.cast = [this.mama, this.papa, this.leo, this.louis];
    this.stars = makeStars(40, 0, W, 20, 300);
  }
  act(a) {
    if (a === 'letters') { sfx('boing', 0.7, 1.3); this.st.sparkles(640, 200, 16, 300); }
    if (a === 'cheer') { this.leo.do('cheer'); this.louis.do('cheer'); this.leo.jump(); setTimeout(() => this.louis.jump(), 150); sfx('giggle', 1); }
    if (a === 'familywave') { this.mama.do('wave', 2); this.papa.do('wave', 2); sfx('hug', 0.8); this.st.hearts(655, 470, 4); }
  }
  draw() {
    const t = this.st.t, A = this.acts;
    rect(0, 0, W, H, grad(0, 0, 0, H, ['#2f3a7c', '#8a6fc0', '#ff9fb0', '#ffd08a']));
    drawStars(this.stars, t, 0.8);
    drawCity(t, 560, '#6a5a9c', 0.5, 1.3);
    // the office tower on the left, the house on the right
    rect(80, 140, 140, 420, '#5a6aa8'); ctx.lineWidth = 3; ctx.strokeStyle = INK; ctx.strokeRect(80, 140, 140, 420);
    for (let r = 0; r < 13; r++) for (let c = 0; c < 4; c++) rect(92 + c * 32, 154 + r * 31, 22, 20, (r + c) % 3 ? 'rgba(255,224,140,0.8)' : '#7c8cc8');
    ctx.beginPath(); ctx.moveTo(0, 580); ctx.quadraticCurveTo(640, 520, W, 580); ctx.lineTo(W, H); ctx.lineTo(0, H); ctx.closePath(); paint('#6faa7a', 3);
    rect(1010, 430, 150, 120, '#f5e6c8'); ctx.strokeStyle = INK; ctx.strokeRect(1010, 430, 150, 120); poly([995, 434, 1175, 434, 1085, 350], '#d2505a');
    rrect(1030, 460, 34, 34, 5, '#ffe08a'); glow(1047, 477, 50, '#ffd27a', 0.5); rrect(1090, 480, 40, 70, 18, '#5aa06e');
    const title = ['Mama', 'Comes Home'];
    title.forEach((ln, r) => {
      ctx.font = '700 110px Fredoka, "Baloo 2", system-ui, sans-serif';
      const total = ctx.measureText(ln).width; let x = 640 - total / 2;
      [...ln].forEach((ch, i) => {
        const w = ctx.measureText(ch).width, n = r * 5 + i;
        let jump = 0;
        if (A.on('letters')) { const k = A.t('letters') - n * 0.05; if (k > 0 && k < 0.6) jump = Math.sin(k / 0.6 * Math.PI) * 50; }
        if (ch !== ' ') word(ch, x + w / 2, 130 + r * 116 + Math.sin(t * 2 + n * 0.5) * 5 - jump, 110, Math.sin(t * 1.5 + n) * 4, CANDY[n % 7]);
        x += w;
      });
    });
    word('A dino family bedtime story', 640, 345, 30, 0, '#fffaf0');
    this.cast.forEach(c => c.draw());
    this.finish();
  }
}
