// ---------- the book: pages, the story text, and turning pages ----------
const LABEL = {
  letters: 'Bouncy letters', cheer: 'Cheer!', familywave: 'Wave hello',
  pancake: 'Flip a pancake', toast: 'Toast pops', kettle: 'Kettle whistles', wave: 'Wave bye-bye', giggle: 'Giggle', spoons: 'Bang the spoons',
  kiss: 'Kisses', bye: 'See you tonight', twirl: 'Twirl',
  whistle: 'Choo choo!', chug: 'Chugga chugga', hearts: 'Heart puffs', conductor: 'All aboard!', pigeons: 'Pigeons fly', clock: 'Station clock', peek: 'Peek out',
  zoom: 'Zoom!', lights: 'Light the windows', blimp: 'Blimp', washer: 'Window washer', birds: 'Birds fly by', copter: 'Helicopter', press: 'Press the button', hello: 'Hello!', wow: 'Wow!',
  copier: 'Paper blizzard', planes: 'Paper planes', coffee: 'Coffee fountain', dance: 'Robot disco', heads: 'Boing heads', burp: 'Robot burps',
  boss: 'Back to work!', tidy: 'Tidy up', dancetoo: 'Dance too',
  ramp: 'Jump the ramp', puddle: 'Splash!', faster: 'Faster, Papa!', spin: 'Spin', wheelie: 'Wheelie', helmet: 'Safety first', papa: '“Papa!”',
  blocks: 'Block tower', paint: 'Paint splat', music: 'Xylophone', friends: 'Dance, friends', teacher: 'Story time', bubbles: 'Bubbles', jump: 'Jump', nap: 'Nap time',
  slides: 'Next slide', chart: 'Chart goes up', cuckoo: 'Cuckoo clock', snooze: 'Bolt snoozes', clap: 'Hooray!', photo: 'Look at the photo', idea: 'Good idea!', thumbs: 'Great job!',
  noodles: 'Noodles', tomato: 'Tomato squirt', fruit: 'Bouncy fruit', slurp: 'Slurp', moustache: 'Noodle moustache', yum: 'Yum yum', juggle: 'Juggle fruit', airplane: 'Spoon airplane', laugh: 'Laugh',
  moon: 'Peeking moon', rain: 'Rain on the window', want: 'We want Mama', window: 'Look outside', teddy: 'Hug teddy', phone: 'Call Mama', cuddle: 'Cuddle', story: 'Read a story',
  board: 'Departure board', thunder: 'Thunder', downpour: 'Pouring rain', cat: 'Kitty', announce: 'Announcement', frog: 'Puddle frog', sigh: 'Oh dear', umbrella: 'Twirl the umbrella',
  boost: 'Rocket boost', shoot: 'Shooting star', fireworks: 'Fireworks', loop: 'Loop the loop', song: 'Robot song', wavemoon: 'Wave at the moon', wheee: 'Wheee!', hold: 'Hold on tight', point: 'There’s home!',
  door: 'Open the door', byebots: 'Bye, robots', chimney: 'Heart smoke', hug: 'Big hug', shh: 'Shhh',
  mobile: 'Spin the mobile', glowstars: 'Starry nightlight', lamp: 'Lights out', yawn: 'Big yawn', sleep: 'Fall asleep', sing: 'Sing the song', blanket: 'Tuck in',
};
const Book = {
  pages: [], cur: 0, turn: null, queued: 0, started: false,
  init() {
    this.pages = [new CoverPage(), new MorningPage(), new StationPage(), new TowerPage(), new OfficePage(), new PapaPage(), new DaycarePage(), new MeetingPage(), new PickupPage(), new DinnerPage(), new WantMamaPage(), new StuckPage(), new RescuePage(), new HomeAgainPage(), new TuckPage()];
    this.pages.forEach(p => { p.shown = 0; });
  },
  page() { return this.pages[this.cur]; },
  // what each of the three buttons does next on this page
  labels() {
    const p = this.page();
    return [0, 1, 2].map(k => {
      const useDino = k === 2 && p.dino && !p.tricks[2];
      const list = useDino ? p.dino.tricks : p.tricks[k];
      const i = useDino ? p.dino.next : p.st.next[k];
      return { who: p.who[k], what: LABEL[list[i % list.length]] || list[i % list.length] };
    });
  },
  go(dir) {
    if (this.turn) { this.queued = dir; return; }
    const to = this.cur + dir;
    if (to < 0 || to >= this.pages.length) { sfx('tick', 0.5, 0.6); this.nudge = 0.3 * dir; return; }
    // a forward turn lifts the current page off; a backward turn lays the earlier page back down
    this.turn = { from: this.cur, to, dir, t: 0, dur: 1.15 };
    this.pages[to].shown = 0;
    sfx('swoosh', 1, dir > 0 ? 1 : 0.85);
    Music.set(this.pages[to].music || this.pages[to].name);
    Amb.set(this.pages[to].name);
  },
  press(k) {
    if (this.turn) return;
    this.page().press(k);
  },
  update(dt) {
    this.nudge = (this.nudge || 0) * Math.pow(0.001, dt);
    const p = this.page();
    p.update(dt); p.shown += dt;
    if (this.turn) {
      const tp = this.pages[this.turn.to];
      tp.update(dt);
      this.turn.t += dt;
      if (this.turn.t >= this.turn.dur) {
        this.cur = this.turn.to; this.turn = null;
        if (this.queued) { const q = this.queued; this.queued = 0; this.go(q); }
      }
    }
    Amb.update(dt);
  },
};

// drawPage paints one page (scene and its story text) onto the current ctx.
function drawPage(p) {
  ctx.save();
  const st = p.st;
  if (st.shake) ctx.translate(Math.sin(st.t * 61) * st.shake * 9, Math.cos(st.t * 47) * st.shake * 7);
  p.draw();
  ctx.restore();
  if (p.text) drawStoryText(p);
}

// The story text sits on a strip of paper at the top and appears slowly,
// word by word, the way it would be read aloud.
function drawStoryText(p) {
  ctx.font = '400 31px Andika, "Trebuchet MS", system-ui, sans-serif';
  const lines = wrapText(p.text, 1020), lh = 42, h = lines.length * lh + 30;
  ctx.save();
  ctx.shadowColor = 'rgba(40,20,30,0.25)'; ctx.shadowBlur = 12; ctx.shadowOffsetY = 4;
  rrect(110, 12, W - 220, h, 18, 'rgba(255,250,238,0.94)', 0);
  ctx.restore();
  rrect(110, 12, W - 220, h, 18, null, 2);
  ctx.textAlign = 'left'; ctx.textBaseline = 'middle';
  const shown = (p.shown - 0.9) * 2.4; // words per second
  let n = 0;
  lines.forEach((ln, i) => {
    let x = 640 - ctx.measureText(ln).width / 2;
    const y = 12 + 15 + lh / 2 + i * lh;
    for (const w of ln.split(' ')) {
      const a = clamp(shown - n, 0, 1);
      ctx.fillStyle = `rgba(70,40,47,${a})`;
      ctx.fillText(w, x, y + (1 - a) * 6);
      x += ctx.measureText(w + ' ').width;
      n++;
    }
  });
}

// ---- the page curl ----
// The bottom-right corner of the lifting page is pulled across to the left.
// The fold is the line halfway between where the corner was and where it is;
// beyond the fold we see the back of the page, reflected across that line.
function drawTurn(main, A, B, e) {
  const g = main;
  if (e <= 0.002) { g.drawImage(A, 0, 0, W, H); return; }
  if (e >= 0.998) { g.drawImage(B, 0, 0, W, H); return; }
  const C = { x: W, y: H }, P = { x: W - 2.02 * W * e, y: H - Math.sin(Math.PI * e) * H * 0.22 };
  const Mx = (C.x + P.x) / 2, My = (C.y + P.y) / 2;
  let nx = C.x - P.x, ny = C.y - P.y; const len = Math.hypot(nx, ny); nx /= len; ny /= len;
  const half = (side) => { // the half-plane on one side of the fold
    const ax = -ny, ay = nx, big = 4000;
    g.beginPath();
    g.moveTo(Mx + ax * big, My + ay * big); g.lineTo(Mx - ax * big, My - ay * big);
    g.lineTo(Mx - ax * big + nx * big * side, My - ay * big + ny * big * side);
    g.lineTo(Mx + ax * big + nx * big * side, My + ay * big + ny * big * side);
    g.closePath();
  };
  // the page underneath, darkened near the fold
  g.drawImage(B, 0, 0, W, H);
  g.save(); half(1); g.clip();
  const sh = g.createLinearGradient(Mx, My, Mx + nx * 140, My + ny * 140);
  sh.addColorStop(0, `rgba(40,20,30,${0.45 * Math.sin(Math.PI * e)})`); sh.addColorStop(1, 'rgba(40,20,30,0)');
  g.fillStyle = sh; g.fillRect(0, 0, W, H);
  g.restore();
  // the part of the lifting page that has not folded yet
  g.save(); half(-1); g.clip(); g.drawImage(A, 0, 0, W, H);
  // the flap's shadow on it
  const s2 = g.createLinearGradient(Mx, My, Mx - nx * 60, My - ny * 60);
  s2.addColorStop(0, 'rgba(40,20,30,0.25)'); s2.addColorStop(1, 'rgba(40,20,30,0)');
  g.fillStyle = s2; g.fillRect(0, 0, W, H);
  g.restore();
  // the back of the page, folded over
  g.save();
  const k = 2 * (Mx * nx + My * ny);
  g.transform(1 - 2 * nx * nx, -2 * nx * ny, -2 * nx * ny, 1 - 2 * ny * ny, k * nx, k * ny);
  g.beginPath(); g.rect(0, 0, W, H); g.clip();
  half(1); g.clip();
  g.fillStyle = '#f6eedd'; g.fillRect(0, 0, W, H);
  g.globalAlpha = 0.14; g.drawImage(A, 0, 0, W, H); g.globalAlpha = 1;
  const cg = g.createLinearGradient(Mx, My, Mx + nx * len * 0.5, My + ny * len * 0.5);
  cg.addColorStop(0, 'rgba(60,35,40,0.35)'); cg.addColorStop(0.25, 'rgba(255,255,255,0.15)'); cg.addColorStop(1, 'rgba(60,35,40,0.12)');
  g.fillStyle = cg; g.fillRect(0, 0, W, H);
  g.restore();
}
