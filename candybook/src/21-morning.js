// ---------- 1. morning: Mama says goodbye ----------
class MorningPage extends StoryPage {
  constructor() {
    super({ name: 'morning', floor: 680,
      text: 'Good morning! The sun peeks into the dino house. Mama has to go to work. “Bye-bye, Léo! Bye-bye, Louis!” she says, and kisses their fuzzy blonde heads.',
      tricks: [['pancake', 'toast', 'kettle'], ['wave', 'giggle', 'spoons'], ['kiss', 'bye', 'twirl']],
      who: ['Kitchen', 'Léo & Louis', 'Mama'],
      durs: { pancake: 2.2, toast: 2.4, kettle: 3, wave: 2, giggle: 1.8, spoons: 1.8, kiss: 4.6, bye: 2.4, twirl: 1.6 } });
    this.mama = mama(this.st, { x: 250, y: 660, face: 1, look: { bow: true, bag: true } });
    this.papa = papa(this.st, { x: 1010, y: 664, face: -1 });
    this.leo = leo(this.st, { x: 530, y: 540, face: 1 });
    this.louis = louis(this.st, { x: 700, y: 540, face: -1 });
    this.cast = [this.mama, this.papa, this.leo, this.louis];
    this.toast = [];
  }
  act(a) {
    const st = this.st;
    if (a === 'pancake') { this.papa.do('reach', 2); sfx('whee', 0.4, 1.4); }
    if (a === 'toast') { sfx('pop', 1, 0.8); this.toast = [{ x: 1150, y: 470, vx: -60, vy: -520 }, { x: 1170, y: 470, vx: -20, vy: -600 }]; }
    if (a === 'kettle') sfx('whistle', 0.8, 1.3);
    if (a === 'wave') { this.leo.do('wave', 2); this.louis.do('wave', 2); st.say('BYE, MAMA!', 615, 380, 50, '#8cd2ff'); sfx('giggle', 0.8, 1.2); }
    if (a === 'giggle') { this.leo.do('laugh'); this.louis.do('laugh'); this.leo.jump(300); this.louis.jump(300); sfx('giggle', 1); }
    if (a === 'spoons') { this.leo.do('cheer', 1.6); this.louis.do('cheer', 1.6); st.say('BANG BANG!', 615, 380, 48, '#ffd25a'); for (let i = 0; i < 4; i++) setTimeout(() => sfx('tick', 0.8, 0.6 + i * 0.1), i * 180); }
    if (a === 'kiss') { this.mama.walkTo(450, 220); }
    if (a === 'bye') { this.mama.face = 1; this.mama.do('wave', 2.2); st.say('SEE YOU TONIGHT!', 300, 360, 46, '#d2b4ff'); sfx('hug', 0.6); }
    if (a === 'twirl') { this.mama.do('spin', 1.4); sfx('whee', 0.6, 1.3); st.sparkles(this.mama.x, 560, 10, 60); }
  }
  tick(dt, t, A) {
    const st = this.st;
    if (this.at('kiss', 1.2)) { this.mama.do('kiss', 1.4); sfx('smooch', 1); st.hearts(560, 420, 3); st.hearts(690, 420, 3); this.leo.do('laugh'); this.louis.do('laugh'); }
    if (this.at('kiss', 3.0)) this.mama.walkTo(250, 220);
    if (A.on('kiss') && this.mama.tx === null && A.t('kiss') > 3.2) this.mama.face = 1;
    if (this.at('pancake', 1.6)) { sfx('tick', 0.6, 0.5); st.say('FLIP!', 1010, 380, 52, '#ffd25a'); }
    if (this.at('kettle', 0.3)) st.say('WHEEEE!', 1100, 360, 44, WHITE);
    if (A.on('kettle') && this.every('k', 0.1)) st.puff(1150 + rnd(-6, 6), 500, 8, '#ffffff', 2, { vy: -90 });
    if (this.every('steam', 0.9)) st.puff(570 + rnd(-30, 30), 500, 6, '#ffffff', 0, { vy: -30 });
    for (const b of this.toast) { b.vy += 1400 * dt; b.x += b.vx * dt; b.y = Math.min(b.y + b.vy * dt, 470); if (b.y >= 470 && b.vy > 0) { b.vx = 0; b.vy = 0; } }
  }
  draw() {
    const st = this.st, t = st.t, A = this.acts;
    drawRoom('#ffe9c8', '#ffd8b0', '#d69664', '#ba7850');
    // wallpaper hearts
    ctx.globalAlpha = 0.25; for (let x = 40; x < W; x += 110) for (let y = 160; y < 560; y += 110) heart(x + (y % 220 ? 55 : 0), y, 8, '#ff9ab4', 0); ctx.globalAlpha = 1;
    // window with the sun coming up
    const sunY = 330 - Math.min(120, t * 4);
    drawWindow(330, 170, 230, 190, grad(0, 170, 0, 360, ['#9ad0ff', '#ffe0b4']), () => {
      glow(450, sunY, 120, '#fff0a0', 0.7); circ(450, sunY, 40, '#ffdc5a', 3);
      cloud(380 + (t * 6) % 260, 230, 0.7);
    });
    // the front door, and the coat hooks
    rrect(70, 330, 130, 270, 16, '#5aa06e'); circ(180, 470, 7, '#ffd250'); rrect(100, 360, 70, 60, 30, '#cdeeff');
    // stove, pan and kettle; toaster on the counter
    rrect(940, 470, 280, 130, 8, '#f0ebe1'); rect(940, 470, 280, 14, '#c8c8d7');
    rrect(1120, 430, 64, 44, 10, '#e6505a'); rect(1132, 426, 14, 6, INK); rect(1156, 426, 14, 6, INK);
    rrect(1110, 478 - 0, 60, 0.1, 0, INK, 0);
    // the kettle
    ell(1150, 500 - 40, 0, 0, null, 0);
    for (const b of this.toast) rrect(b.x - 12, b.y - 30, 24, 28, 6, '#e6b46e');
    // Papa at the stove, and the pancake going up
    this.papa.draw({ reach: this.papa.is('reach') });
    const pk = A.on('pancake') ? A.t('pancake') : -1;
    const [px, py] = [this.papa.x - 74, 470];
    line(this.papa.x - 58, 548, px + 14, py + 6, 6, '#5a5a6e'); ell(px, py + 6, 30, 8, '#3c3c50');
    const ph = pk >= 0 && pk < 1.6 ? Math.sin(pk / 1.6 * Math.PI) * 230 : 0;
    push(px, py - ph, pk >= 0 && pk < 1.6 ? pk / 1.6 * 360 : 0); ell(0, 0, 24, 7, '#f0c070'); pop();
    // kettle on the back of the stove
    ell(1050, 455, 30, 22, '#8ad0ec'); line(1072, 450, 1094, 436, 6, '#8ad0ec'); rrect(1040, 428, 20, 8, 4, INK, 0);
    // the twins in their chairs behind the table
    for (const k of [this.leo, this.louis]) { line(k.x - 40, 600, k.x - 40, 470, 8, '#c8965a'); line(k.x + 40, 600, k.x + 40, 470, 8, '#c8965a'); }
    this.leo.draw(); this.louis.draw();
    // table with breakfast
    rrect(400, 520, 460, 30, 10, '#c8965a');
    line(430, 550, 430, 650, 12, '#a0703c'); line(830, 550, 830, 650, 12, '#a0703c');
    ell(570, 516, 34, 9, WHITE); ell(570, 510, 22, 6, '#ffb45a', 0); ell(690, 516, 34, 9, WHITE); ell(690, 510, 22, 6, '#ffb45a', 0);
    rrect(625, 490, 22, 28, 5, '#fff6dc'); rect(628, 498, 16, 16, '#ffa03c');
    this.mama.draw();
    this.finish();
  }
}
