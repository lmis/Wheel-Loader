// ---------- 9. dinner: tomatoes, fruit and noodles ----------
class DinnerPage extends StoryPage {
  constructor() {
    super({ name: 'dinner', floor: 690,
      text: 'For dinner there are tomatoes, fruit and noodles. Slurp, slurp, slurp! But Mama’s chair is still empty.',
      tricks: [['noodles', 'tomato', 'fruit'], ['slurp', 'moustache', 'yum'], ['juggle', 'airplane', 'laugh']],
      who: ['Dinner', 'Léo & Louis', 'Papa'],
      durs: { noodles: 2.6, tomato: 2, fruit: 2.4, slurp: 2.4, moustache: 3.4, yum: 1.8, juggle: 3.4, airplane: 2.6, laugh: 1.8 } });
    this.leo = leo(this.st, { x: 470, y: 528, face: 1, scale: 0.7 });
    this.louis = louis(this.st, { x: 640, y: 528, face: 1, scale: 0.7 });
    this.papa = papa(this.st, { x: 1000, y: 600, face: -1 });
    this.cast = [this.leo, this.louis, this.papa];
  }
  act(a) {
    const st = this.st;
    if (a === 'noodles' || a === 'slurp') { this.leo.do('eat', 2.2); this.louis.do('eat', 2.2); sfx('slurp', 1); st.say('SLUUURP!', 560, 360, 56, '#ffd25a'); }
    if (a === 'tomato') { sfx('splat', 1); st.say('SPLAT!', 900, 380, 58, '#ff5050'); this.papa.do('surprise', 1); }
    if (a === 'fruit') sfx('boing', 0.6, 1.4);
    if (a === 'moustache') { this.louis.do('laugh', 3); this.leo.do('laugh', 3); sfx('giggle', 1); st.say('HA HA HA!', 560, 380, 52, '#8cd2ff'); }
    if (a === 'yum') { this.leo.do('cheer'); this.louis.do('cheer'); sfx('yum', 1); st.say('YUM YUM!', 560, 380, 56, '#ff8cc8'); }
    if (a === 'juggle') { this.papa.do('cheer', 3.2); sfx('tune', 0.4, 1.4); }
    if (a === 'airplane') { st.say('HERE COMES THE AIRPLANE!', 700, 330, 38, '#d2b4ff'); sfx('whee', 0.6, 1); }
    if (a === 'laugh') { this.papa.do('laugh'); this.leo.do('laugh'); this.louis.do('laugh'); sfx('giggle', 1, 0.8); }
  }
  tick(dt, t, A) {
    const st = this.st;
    if (this.at('tomato', 0.5)) st.drops(1000, 470, 16, 280, '#ff5050');
    if (this.at('airplane', 1.6)) { this.leo.do('eat', 1.2); sfx('chomp', 0.8); st.hearts(470, 440, 2); }
    if (this.at('noodles', 1.8)) st.hearts(560, 420, 2);
  }
  draw() {
    const st = this.st, t = st.t, A = this.acts;
    const bg = art('bg-09-dinner'); if (bg) drawArt(bg);
    if (!bg) {
      drawRoom('#ffe2c0', '#ffcfa8', '#c88c64', '#b07850');
      drawWindow(120, 150, 240, 180, grad(0, 150, 0, 330, ['#5a5aa0', '#c88cb4']), () => { drawStars(this.stars || (this.stars = makeStars(14, 130, 350, 160, 250)), t); });
      // a warm lamp over the table
      line(640, 0, 640, 120, 3, INK, 0); poly([590, 160, 690, 160, 660, 120, 620, 120], '#ff8c6e'); glow(640, 200, 300, '#ffd27a', 0.35);
    }
    // Mama's empty chair, with her place set
    if (!bg) { line(820, 600, 820, 430, 10, '#a0703c'); rrect(790, 400, 60, 50, 10, '#b48ce6');
    heart(820, 425, 9, '#ffb4d2', 2); }
    // the twins in their chairs, Papa across
    if (!bg) for (const k of [this.leo, this.louis]) { line(k.x - 40, 620, k.x - 40, 440, 8, '#c8965a'); line(k.x + 40, 620, k.x + 40, 440, 8, '#c8965a'); }
    this.leo.draw(); this.louis.draw();
    // noodle strands going up into the twins' mouths
    const nd = A.on('noodles') || A.on('slurp') ? (A.t('noodles') ?? A.t('slurp')) : -1;
    for (const k of [this.leo, this.louis]) {
      const [mx, my] = k.mouth();
      if (nd >= 0 && nd < 2) { const len = Math.sin(Math.PI * nd / 2); for (let i = -1; i <= 1; i++) curve(k.x + 30 + i * 6, 512, mx - 10 + i * 4, lerp(512, my, 0.5) + 10, lerp(k.x + 30, mx, len), lerp(512, my + 6, len), 3, '#ffe08a', 1.5); }
    }
    if (A.on('moustache')) { const [mx, my] = this.louis.mouth(); curve(mx - 26, my - 2, mx - 10, my + 10, mx + 8, my - 2, 4, '#ffe08a', 1.5); curve(mx - 6, my - 2, mx + 10, my + 10, mx + 26, my - 2, 4, '#ffe08a', 1.5); }
    this.papa.draw();
    if (A.on('tomato') && A.t('tomato') > 0.5) for (let i = 0; i < 4; i++) circ(this.papa.x - 20 + i * 12, 470 + (i % 2) * 10, 5, '#ff5050', 0);
    // the table and dinner
    if (!bg) { rrect(360, 520, 640, 30, 10, '#e6b46e');
    line(400, 550, 400, 650, 12, '#c8965a'); line(960, 550, 960, 650, 12, '#c8965a'); }
    for (const x of [500, 670]) { ell(x, 518, 44, 12, WHITE); ell(x, 510, 34, 8, '#ffe08a', 0); for (let i = 0; i < 4; i++) curve(x - 20 + i * 12, 508, x - 14 + i * 12, 500, x - 8 + i * 12, 508, 2, '#f0c860', 0); }
    // a bowl of cherry tomatoes
    ell(800, 516, 40, 12, '#8ad0ec');
    for (let i = 0; i < 6; i++) { const tx = 774 + (i % 3) * 22, ty = 500 - Math.floor(i / 3) * 14; circ(tx, ty, 10, '#ff5050', 2); line(tx - 2, ty - 10, tx + 3, ty - 13, 2, '#5ab46e', 0); }
    // the fruit bowl, fruit hopping out when Space says so
    const fr = A.on('fruit') ? A.t('fruit') : -1;
    const fruits = [['#ffe050', 'banana'], ['#ff7878', 'apple'], ['#a078e6', 'grape'], ['#ffa03c', 'orange']];
    fruits.forEach(([c, k], i) => {
      let x = 900 + (i - 1.5) * 22, y = 496 - (i % 2) * 10;
      if (fr >= 0 && fr < 2) y -= Math.sin(Math.PI * clamp((fr - i * 0.15) / 1.4, 0, 1)) * 140;
      if (A.on('juggle')) { const a = A.t('juggle') * 5 + i * TAU / 4; x = this.papa.x - 10 + Math.cos(a) * 40; y = 350 - Math.abs(Math.sin(a)) * 90; }
      if (k === 'banana') curve(x - 14, y, x, y + 12, x + 14, y - 4, 8, c, 2.5);
      else if (k === 'grape') for (let j = 0; j < 4; j++) circ(x - 6 + (j % 2) * 10, y - 6 + Math.floor(j / 2) * 9, 6, c, 2);
      else circ(x, y, 12, c, 2.5);
    });
    ell(900, 516, 50, 12, '#ffb4d2');
    if (A.on('airplane')) { const k = clamp(A.t('airplane') / 1.6, 0, 1), [mx, my] = this.leo.mouth(); const x = lerp(this.papa.x - 60, mx, k), y = lerp(440, my, k) - Math.sin(k * Math.PI) * 120; push(x, y, -20); line(-20, 0, 20, 0, 5, '#c8c8d7'); ell(22, 0, 10, 6, '#c8c8d7'); ell(22, -2, 6, 4, '#ffe08a', 0); pop(); }
    this.finish();
  }
}
