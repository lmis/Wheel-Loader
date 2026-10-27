// ---------- sound: everything is synthesised, nothing is loaded ----------
const AU = { ctx: null, muted: false };

function audioStart() {
  if (AU.ctx) { if (AU.ctx.state === 'suspended') AU.ctx.resume(); return; }
  const AC = window.AudioContext || window.webkitAudioContext;
  if (!AC) return;
  const C = AU.ctx = new AC();
  const comp = C.createDynamicsCompressor();
  comp.threshold.value = -16; comp.ratio.value = 3; comp.attack.value = 0.01; comp.release.value = 0.2;
  comp.connect(C.destination);
  AU.master = C.createGain(); AU.master.gain.value = AU.muted ? 0 : 0.9; AU.master.connect(comp);
  // a small warm room for the music and the noises
  AU.verb = C.createConvolver();
  const len = Math.floor(C.sampleRate * 2.4), imp = C.createBuffer(2, len, C.sampleRate);
  for (let ch = 0; ch < 2; ch++) {
    const d = imp.getChannelData(ch);
    for (let i = 0; i < len; i++) d[i] = (Math.random() * 2 - 1) * Math.pow(1 - i / len, 3.2);
  }
  AU.verb.buffer = imp;
  const vg = C.createGain(); vg.gain.value = 0.32; AU.verb.connect(vg); vg.connect(AU.master);
  AU.music = C.createGain(); AU.music.gain.value = 0.55; AU.music.connect(AU.master); AU.music.connect(AU.verb);
  AU.amb = C.createGain(); AU.amb.gain.value = 0.7; AU.amb.connect(AU.master);
  AU.fx = C.createGain(); AU.fx.gain.value = 0.85; AU.fx.connect(AU.master);
  const fs = C.createGain(); fs.gain.value = 0.25; AU.fx.connect(fs); fs.connect(AU.verb);
  const nb = C.createBuffer(1, C.sampleRate * 2, C.sampleRate), nd = nb.getChannelData(0);
  for (let i = 0; i < nd.length; i++) nd[i] = Math.random() * 2 - 1;
  AU.noise = nb;
  Music.start();
  Amb.set(Amb.want);
}
function setMuted(m) {
  AU.muted = m;
  if (AU.ctx) AU.master.gain.setTargetAtTime(m ? 0 : 0.9, AU.ctx.currentTime, 0.1);
}

// osc plays one enveloped oscillator note into dest.
function osc(type, f0, t, dur, vol, dest, o = {}) {
  const C = AU.ctx, n = C.createOscillator(), g = C.createGain();
  n.type = type; n.frequency.setValueAtTime(f0, t);
  if (o.f1) n.frequency.exponentialRampToValueAtTime(Math.max(20, o.f1), t + (o.sweep ?? dur));
  if (o.path) for (const [ft, fv] of o.path) n.frequency.linearRampToValueAtTime(fv, t + ft);
  if (o.vib) {
    const l = C.createOscillator(), lg = C.createGain();
    l.frequency.value = o.vib[0]; lg.gain.value = o.vib[1]; l.connect(lg); lg.connect(n.frequency);
    l.start(t); l.stop(t + dur + 0.2);
  }
  if (o.fm) { // a simple FM modulator for bells and electric piano
    const m = C.createOscillator(), mg = C.createGain();
    m.frequency.value = f0 * o.fm[0]; mg.gain.setValueAtTime(f0 * o.fm[1], t);
    mg.gain.setTargetAtTime(0, t, o.fm[2]); m.connect(mg); mg.connect(n.frequency);
    m.start(t); m.stop(t + dur + 0.2);
  }
  const a = o.a ?? 0.005, hold = o.hold ?? 0;
  g.gain.setValueAtTime(0.0001, t);
  g.gain.linearRampToValueAtTime(vol, t + a);
  if (hold) g.gain.setValueAtTime(vol, t + a + hold);
  g.gain.setTargetAtTime(0.0001, t + a + hold, o.d ?? dur / 4);
  let out = g;
  if (o.lp) { const f = C.createBiquadFilter(); f.type = 'lowpass'; f.frequency.value = o.lp; f.Q.value = o.q ?? 0.7; g.connect(f); out = f; }
  if (o.bp) { const f = C.createBiquadFilter(); f.type = 'bandpass'; f.frequency.value = o.bp; f.Q.value = o.q ?? 2; g.connect(f); out = f; }
  if (o.pan !== undefined && C.createStereoPanner) {
    const p = C.createStereoPanner(); p.pan.setValueAtTime(o.pan, t);
    if (o.pan1 !== undefined) p.pan.linearRampToValueAtTime(o.pan1, t + dur);
    out.connect(p); out = p;
  }
  n.connect(g); out.connect(dest);
  n.start(t); n.stop(t + dur + 0.1);
  return n;
}
// noise plays a burst of filtered noise.
function noise(t, dur, vol, dest, o = {}) {
  const C = AU.ctx, s = C.createBufferSource(), f = C.createBiquadFilter(), g = C.createGain();
  s.buffer = AU.noise; s.loop = true;
  f.type = o.type ?? 'bandpass'; f.frequency.setValueAtTime(o.f0 ?? 1000, t); f.Q.value = o.q ?? 1;
  if (o.f1) f.frequency.exponentialRampToValueAtTime(o.f1, t + dur);
  const a = o.a ?? 0.005;
  g.gain.setValueAtTime(0.0001, t); g.gain.linearRampToValueAtTime(vol, t + a);
  if (o.hold) g.gain.setValueAtTime(vol, t + a + o.hold);
  g.gain.setTargetAtTime(0.0001, t + a + (o.hold ?? 0), o.d ?? dur / 4);
  s.connect(f); f.connect(g); g.connect(dest);
  s.start(t, Math.random() * 1.5); s.stop(t + dur + 0.1);
}
const mtof = m => 440 * Math.pow(2, (m - 69) / 12);

// ---- instruments (used by the music and by some noises) ----
const INST = {
  box(f, t, dur, v, out) { // music box
    [[1, 1, 1.4], [2, 0.32, 0.6], [3, 0.1, 0.3], [4.2, 0.07, 0.2]].forEach(([k, a, d]) => osc('sine', f * k, t, d * 2, v * a * 0.22, out, { d: d / 2.5 }));
  },
  bell(f, t, dur, v, out) { osc('sine', f, t, 1.6, v * 0.2, out, { fm: [3.5, 2.2, 0.25], d: 0.5 }); },
  marimba(f, t, dur, v, out) {
    osc('sine', f, t, 0.6, v * 0.32, out, { d: 0.14 });
    osc('sine', f * 4, t, 0.1, v * 0.06, out, { d: 0.02 });
  },
  epiano(f, t, dur, v, out) {
    osc('sine', f, t, dur + 1.2, v * 0.2, out, { fm: [1, 1.4, 0.35], d: 0.7 });
    osc('sine', f * 2, t, 0.6, v * 0.03, out, { d: 0.2 });
  },
  flute(f, t, dur, v, out) {
    osc('sine', f, t, dur + 0.3, v * 0.16, out, { a: 0.09, hold: dur * 0.75, d: 0.12, vib: [5, f * 0.006] });
    osc('triangle', f * 2, t, dur + 0.3, v * 0.025, out, { a: 0.1, hold: dur * 0.7, d: 0.1 });
    noise(t, 0.12, v * 0.02, out, { f0: f * 2, q: 3, d: 0.05 });
  },
  pluck(f, t, dur, v, out) { osc('triangle', f, t, 0.5, v * 0.09, out, { d: 0.12, lp: 2600 }); },
  harp(f, t, dur, v, out) { osc('triangle', f, t, 1.6, v * 0.08, out, { d: 0.45, lp: 3000 }); },
  pad(f, t, dur, v, out) {
    for (const det of [-7, 7]) osc('sawtooth', f * Math.pow(2, det / 1200), t, dur + 1.4, v * 0.03, out, { a: 0.9, hold: Math.max(0, dur - 0.9), d: 0.6, lp: 750 });
  },
  bass(f, t, dur, v, out) { osc('sine', f, t, dur + 0.4, v * 0.26, out, { a: 0.01, d: dur * 0.35 }); },
  pizz(f, t, dur, v, out) { osc('triangle', f, t, 0.4, v * 0.22, out, { d: 0.09, lp: 900 }); },
  wood(f, t, dur, v, out) { osc('sine', 950, t, 0.05, v * 0.05, out, { d: 0.015 }); },
};

// ---- the theme: "Coming Home", a lullaby in 3/4 ----
const THEME = {
  melody: [[64, 2], [67, 1], [72, 2], [71, 1], [69, 2], [67, 1], [64, 3],
    [65, 2], [69, 1], [74, 2], [72, 1], [71, 1], [72, 1], [74, 1], [67, 3],
    [64, 2], [67, 1], [72, 2], [76, 1], [74, 2], [72, 1], [69, 3],
    [65, 2], [69, 1], [67, 2], [64, 1], [62, 2], [71, 1], [72, 3]],
  // one chord per bar: root and the two notes above it
  chords: ['C', 'Em', 'F', 'C', 'F', 'Dm', 'G', 'G', 'C', 'Am', 'Dm', 'Am', 'F', 'C', 'G', 'C'],
};
const CHORD = { C: [48, 52, 55], Em: [52, 55, 59], F: [53, 57, 60], Dm: [50, 53, 57], G: [43, 47, 50], Am: [45, 48, 52] };
// Each page plays the theme its own way.
const ARRANGE = {
  cover: { bpm: 76, lead: 'box', oct: 12, pad: 0.8, arp: 'harp', bass: 0 },
  factory: { bpm: 104, lead: 'marimba', oct: 0, pad: 0, arp: 'pluck', bass: 'pizz', perc: 1 },
  town: { bpm: 86, lead: 'epiano', oct: 0, pad: 0.8, arp: 0, bass: 'bass' },
  river: { bpm: 74, lead: 'flute', oct: 12, pad: 0.7, arp: 'harp', bass: 'bass' },
  hill: { bpm: 66, lead: 'box', oct: 12, pad: 1, arp: 0, bass: 'bass' },
  home: { bpm: 70, lead: 'epiano', oct: 0, pad: 0.9, arp: 'harp', bass: 'bass' },
  bed: { bpm: 56, lead: 'box', oct: 12, pad: 0.6, arp: 0, bass: 0, soft: 0.7 },
};

const Music = {
  want: 'cover', arr: ARRANGE.cover, step: 0, nextT: 0, timer: null, events: null,
  start() {
    // melody notes indexed by their starting eighth
    this.events = {};
    let b = 0;
    for (const [m, d] of THEME.melody) { this.events[b * 2] = [m, d]; b += d; }
    this.nextT = AU.ctx.currentTime + 0.15;
    this.timer = setInterval(() => this.tick(), 40);
  },
  set(name) { this.want = name; },
  tick() {
    const C = AU.ctx;
    while (this.nextT < C.currentTime + 0.25) {
      const i = this.step % 96; // 16 bars of 3 beats, in eighths
      if (i % 6 === 0) this.arr = ARRANGE[this.want] || ARRANGE.cover; // change instruments on the bar
      const a = this.arr, beat = 60 / a.bpm, t = this.nextT, v = a.soft ?? 1;
      const bar = Math.floor(i / 6), ch = CHORD[THEME.chords[bar]], out = AU.music;
      if (i % 6 === 0) {
        if (a.pad) for (const m of ch) INST.pad(mtof(m + 12), t, beat * 3, a.pad * v, out);
        if (a.bass) INST[a.bass](mtof(ch[0] - (ch[0] > 47 ? 12 : 0)), t, beat * 2, v, out);
      }
      if (a.bass === 'pizz' && i % 6 === 2) INST.pizz(mtof(ch[2] - 12), t, beat, v * 0.7, out);
      if (a.bass === 'pizz' && i % 6 === 4) INST.pizz(mtof(ch[1] - 12), t, beat, v * 0.7, out);
      if (a.arp) INST[a.arp](mtof(ch[[0, 1, 2, 1, 2, 1][i % 6]] + 24), t, beat / 2, v * (i % 2 ? 0.6 : 0.9), out);
      if (a.perc && i % 2 === 0) INST.wood(0, t, 0, v * (i % 6 === 0 ? 1.4 : 0.8), out);
      const e = this.events[i];
      if (e) INST[a.lead](mtof(e[0] + a.oct), t, e[1] * beat, v, out);
      this.nextT += beat / 2;
      this.step++;
    }
  },
};

// ---- background sounds for each place ----
const Amb = {
  want: 'cover', beds: [], clock: 0, timers: {},
  set(name) {
    this.want = name;
    if (!AU.ctx) return;
    const C = AU.ctx, now = C.currentTime;
    for (const b of this.beds) { b.g.gain.setTargetAtTime(0, now, 0.4); b.nodes.forEach(n => n.stop(now + 2)); }
    this.beds = [];
    const bed = (vol, filt, f, q = 0.7, lfo = 0) => {
      const s = C.createBufferSource(), fl = C.createBiquadFilter(), g = C.createGain(), nodes = [s];
      s.buffer = AU.noise; s.loop = true; fl.type = filt; fl.frequency.value = f; fl.Q.value = q;
      g.gain.setValueAtTime(0, now); g.gain.setTargetAtTime(vol, now + 0.3, 0.8);
      s.connect(fl); fl.connect(g);
      if (lfo) { // a slow swell, like wind or water
        const l = C.createOscillator(), lg = C.createGain(), sw = C.createGain();
        l.frequency.value = lfo; lg.gain.value = 0.5; sw.gain.value = 0.6;
        l.connect(lg); lg.connect(sw.gain); fl.connect(sw); sw.connect(g); fl.disconnect(g);
        l.start(); nodes.push(l);
      }
      g.connect(AU.amb); s.start(0, Math.random()); this.beds.push({ g, nodes });
    };
    const hum = (f, vol) => {
      const o = C.createOscillator(), fl = C.createBiquadFilter(), g = C.createGain();
      o.type = 'sawtooth'; o.frequency.value = f; fl.type = 'lowpass'; fl.frequency.value = 160;
      g.gain.setValueAtTime(0, now); g.gain.setTargetAtTime(vol, now + 0.3, 0.8);
      o.connect(fl); fl.connect(g); g.connect(AU.amb); o.start(); this.beds.push({ g, nodes: [o] });
    };
    if (name === 'factory') { hum(55, 0.05); bed(0.012, 'lowpass', 300); }
    if (name === 'town') bed(0.03, 'lowpass', 260, 0.7, 0.07);
    if (name === 'river') { bed(0.09, 'bandpass', 520, 0.6, 0.23); bed(0.012, 'highpass', 4000, 0.7, 0.4); }
    if (name === 'hill') bed(0.1, 'lowpass', 420, 0.8, 0.09);
    if (name === 'home') bed(0.05, 'lowpass', 380, 0.8, 0.08);
  },
  // little sounds now and then, called every frame
  update(dt) {
    if (!AU.ctx || AU.ctx.state !== 'running') return;
    const T = this.timers, n = this.want, t = AU.ctx.currentTime + 0.02, out = AU.amb;
    const every = (k, lo, hi, f) => { T[k] = (T[k] ?? rnd(0, hi)) - dt; if (T[k] <= 0) { T[k] = rnd(lo, hi); f(); } };
    const cricket = () => { for (let i = 0; i < 3; i++) osc('sine', 4300 + rnd(-100, 100), t + i * 0.06, 0.03, 0.018, out, { d: 0.01, pan: rnd(-0.8, 0.8) }); };
    if (n === 'factory') {
      every('chuff', 0.8, 0.8, () => { noise(t, 0.16, 0.08, out, { f0: 320, q: 1.2, d: 0.05 }); osc('square', 1700, t + 0.4, 0.03, 0.012, out, { d: 0.01 }); });
    }
    if (n === 'town' || n === 'cover') every('bird', 1.5, 4.5, () => sfx('tweet', 0.35, rnd(0.85, 1.2), rnd(-0.9, 0.9), out));
    if (n === 'river') { every('drip', 0.6, 2.4, () => osc('sine', rnd(500, 900), t, 0.08, 0.04, out, { f1: rnd(1000, 1600), sweep: 0.06, d: 0.02 })); every('frog', 6, 12, () => sfx('ribbit', 0.25, rnd(0.9, 1.1), 0.6, out)); every('crk', 0.5, 1.6, cricket); }
    if (n === 'hill') { every('crk', 0.3, 1.1, cricket); every('owl', 9, 16, () => sfx('hoo', 0.18, 1, -0.7, out)); }
    if (n === 'home') every('crk', 0.4, 1.4, cricket);
    if (n === 'bed') every('tick', 0.5, 0.5, () => { this.clock ^= 1; osc('sine', this.clock ? 2100 : 1600, t, 0.03, 0.03, out, { d: 0.008 }); });
  },
};

// ---- the noises ----
const SFX = {
  click: (t, v, p, o) => osc('sine', 1400 * p, t, 0.06, 0.25 * v, o, { d: 0.02 }),
  tick: (t, v, p, o) => osc('sine', 2200 * p, t, 0.03, 0.2 * v, o, { d: 0.01 }),
  boing: (t, v, p, o) => osc('sine', 150 * p, t, 0.8, 0.5 * v, o, { f1: 430 * p, sweep: 0.45, vib: [16, 45], d: 0.25 }),
  pop: (t, v, p, o) => { osc('sine', 900 * p, t, 0.12, 0.45 * v, o, { f1: 250 * p, d: 0.04 }); noise(t, 0.05, 0.2 * v, o, { f0: 2000, d: 0.015 }); },
  bloop: (t, v, p, o) => osc('sine', 300 * p, t, 0.2, 0.35 * v, o, { f1: 800 * p, d: 0.06 }),
  squeak: (t, v, p, o) => osc('sine', 1200 * p, t, 0.16, 0.22 * v, o, { path: [[0.08, 1900 * p], [0.16, 1300 * p]], d: 0.05 }),
  hop: (t, v, p, o) => osc('sine', 260 * p, t, 0.1, 0.2 * v, o, { f1: 560 * p, d: 0.035 }),
  slide: (t, v, p, o) => osc('sine', 300 * p, t, 1.1, 0.22 * v, o, { f1: 1300 * p, sweep: 1, a: 0.05, hold: 0.8, d: 0.08, vib: [6, 18] }),
  whistle: (t, v, p, o) => osc('sine', 1600 * p, t, 0.9, 0.16 * v, o, { f1: 600 * p, a: 0.02, hold: 0.7, d: 0.05 }),
  whee: (t, v, p, o) => osc('sine', 500 * p, t, 0.8, 0.25 * v, o, { path: [[0.3, 1150 * p], [0.75, 650 * p]], hold: 0.6, d: 0.08, vib: [7, 20] }),
  windup: (t, v, p, o) => osc('triangle', 110 * p, t, 0.6, 0.25 * v, o, { f1: 440 * p, a: 0.45, d: 0.05, vib: [28, 30] }),
  blow: (t, v, p, o) => noise(t, 2.6, 0.16 * v, o, { f0: 350, f1: 900, q: 1.5, a: 1.6, hold: 0.8, d: 0.08 }),
  chomp: (t, v, p, o) => { noise(t, 0.07, 0.4 * v, o, { f0: 1300 * p, q: 1, d: 0.02 }); noise(t + 0.05, 0.05, 0.25 * v, o, { f0: 900 * p, d: 0.015 }); },
  thump: (t, v, p, o) => { osc('sine', 95 * p, t, 0.5, 0.9 * v, o, { f1: 38, d: 0.14 }); noise(t, 0.2, 0.3 * v, o, { type: 'lowpass', f0: 400, d: 0.05 }); },
  burp: (t, v, p, o) => osc('sawtooth', 100 * p, t, 0.55, 0.45 * v, o, { f1: 62 * p, a: 0.02, hold: 0.3, d: 0.06, vib: [13, 12], lp: 650 }),
  rawr: (t, v, p, o) => { osc('sawtooth', 340 * p, t, 0.85, 0.32 * v, o, { f1: 190 * p, a: 0.05, hold: 0.5, d: 0.1, vib: [32, 22], lp: 1600 }); noise(t, 0.7, 0.08 * v, o, { f0: 900, a: 0.05, hold: 0.4, d: 0.1 }); },
  giggle: (t, v, p, o) => [880, 988, 880, 1047, 880, 1175].forEach((f, i) => osc('sine', f * p, t + i * 0.085, 0.08, 0.16 * v, o, { d: 0.03 })),
  yum: (t, v, p, o) => [659, 880, 1319].forEach((f, i) => INST.box(f * p, t + i * 0.12, 0.2, v * 1.6, o)),
  tune: (t, v, p, o) => [72, 76, 79, 76, 72, 76, 79, 84, 79, 76, 79, 72, 74, 76, 72].forEach((m, i) => INST.marimba(mtof(m) * p, t + i * 0.24, 0.2, v * 1.2, o)),
  ding: (t, v, p, o) => INST.bell(1568 * p, t, 1, v * 1.4, o),
  beep: (t, v, p, o) => { for (let i = 0; i < 2; i++) osc('square', 520 * p, t + i * 0.22, 0.16, 0.12 * v, o, { a: 0.01, hold: 0.12, d: 0.02, lp: 1800 }); },
  engine: (t, v, p, o) => osc('sawtooth', 58 * p, t, 3.2, 0.2 * v, o, { a: 0.8, hold: 1.4, d: 0.5, lp: 220, vib: [9, 4], pan: -0.9, pan1: 0.9 }),
  meow: (t, v, p, o) => osc('sawtooth', 480 * p, t, 0.7, 0.22 * v, o, { path: [[0.2, 900 * p], [0.65, 430 * p]], a: 0.04, hold: 0.5, d: 0.06, bp: 1200, q: 1.6 }),
  woof: (t, v, p, o) => { for (let i = 0; i < 2; i++) osc('sawtooth', 230 * p, t + i * 0.28, 0.2, 0.35 * v, o, { f1: 120 * p, a: 0.01, d: 0.06, lp: 800 }); },
  tweet: (t, v, p, pan, o) => { o = o || AU.fx; const n = 2 + irnd(3); for (let i = 0; i < n; i++) osc('sine', 2600 * p, t + i * 0.1, 0.08, 0.1 * v, o, { f1: 3900 * p, sweep: 0.06, d: 0.02, pan: pan ?? 0 }); },
  flutter: (t, v, p, o) => { for (let i = 0; i < 9; i++) noise(t + i * 0.05, 0.03, 0.12 * v, o, { f0: 1500, q: 0.8, d: 0.01 }); },
  quack: (t, v, p, o) => { for (let i = 0; i < 2; i++) osc('sawtooth', 400 * p, t + i * 0.24, 0.18, 0.28 * v, o, { f1: 290 * p, a: 0.01, hold: 0.08, d: 0.03, bp: 950, q: 3 }); },
  ribbit: (t, v, p, pan, o) => { o = o || AU.fx; for (let i = 0; i < 3; i++) osc('square', 150 * p, t + i * 0.07, 0.05, 0.22 * v, o, { a: 0.005, d: 0.015, lp: 800, pan: pan ?? 0 }); },
  splash: (t, v, p, o) => { noise(t, 0.7, 0.45 * v, o, { type: 'lowpass', f0: 3500, f1: 400, a: 0.01, d: 0.2 }); for (let i = 0; i < 4; i++) osc('sine', rnd(600, 1000), t + 0.2 + i * 0.12, 0.06, 0.08 * v, o, { f1: 1600, sweep: 0.05, d: 0.02 }); },
  blub: (t, v, p, o) => { for (let i = 0; i < 3; i++) osc('sine', 300 * p, t + i * 0.13, 0.1, 0.28 * v, o, { f1: 750 * p, sweep: 0.08, d: 0.03 }); },
  hoo: (t, v, p, pan, o) => { o = o || AU.fx; for (let i = 0; i < 2; i++) osc('sine', 410 * p, t + i * 0.6, 0.5, 0.3 * v, o, { f1: 370 * p, a: 0.09, hold: 0.25, d: 0.08, vib: [5, 4], pan: pan ?? 0 }); },
  baa: (t, v, p, o) => osc('sawtooth', 380 * p, t, 0.65, 0.22 * v, o, { a: 0.03, hold: 0.45, d: 0.06, vib: [8, 28], bp: 1100, q: 1.4 }),
  shimmer: (t, v, p, o) => { for (let i = 0; i < 10; i++) INST.bell(2600 * Math.pow(0.93, i) * p, t + i * 0.06, 0.6, v * 0.5, o); },
  creak: (t, v, p, o) => osc('sawtooth', 90 * p, t, 0.8, 0.16 * v, o, { f1: 150 * p, a: 0.05, hold: 0.6, d: 0.05, vib: [24, 26], lp: 650 }),
  smooch: (t, v, p, o) => { osc('sine', 600 * p, t, 0.08, 0.3 * v, o, { f1: 1900 * p, d: 0.02 }); noise(t + 0.05, 0.03, 0.1 * v, o, { f0: 3000, d: 0.01 }); },
  chime: (t, v, p, o) => [84, 88, 91, 96, 91, 88].forEach((m, i) => INST.box(mtof(m) * p, t + i * 0.16, 0.3, v * 2, o)),
  yawn: (t, v, p, o) => osc('sawtooth', 420 * p, t, 1.3, 0.16 * v, o, { path: [[0.4, 520 * p], [1.2, 240 * p]], a: 0.25, hold: 0.8, d: 0.12, vib: [5, 12], lp: 850 }),
  swoosh: (t, v, p, o) => { noise(t, 0.55, 0.3 * v, o, { f0: 500, f1: 2600, q: 0.8, a: 0.18, d: 0.1 }); noise(t + 0.5, 0.1, 0.12 * v, o, { type: 'lowpass', f0: 700, d: 0.03 }); },
  sparkle: (t, v, p, o) => [2093, 2637, 3136].forEach((f, i) => INST.bell(f * p, t + i * 0.07, 0.4, v * 0.5, o)),
  snore: (t, v, p, o) => noise(t, 1.2, 0.08 * v, o, { f0: 300, q: 2, a: 0.5, hold: 0.3, d: 0.15 }),
  lampOn: (t, v, p, o) => { osc('square', 120, t, 0.03, 0.05 * v, o, { d: 0.01, lp: 2000 }); INST.bell(1047 * p, t + 0.02, 1, v * 0.8, o); },
  hug: (t, v, p, o) => [60, 64, 67, 72].forEach((m, i) => INST.epiano(mtof(m + 12) * p, t + i * 0.09, 0.6, v * 1.2, o)),
};
function sfx(name, v = 1, p = 1, pan, out) {
  if (!AU.ctx || AU.ctx.state !== 'running' || !SFX[name]) return;
  const f = SFX[name];
  if (f.length >= 5) f(AU.ctx.currentTime + 0.01, v, p, pan, out || AU.fx);
  else f(AU.ctx.currentTime + 0.01, v, p, out || AU.fx);
}
