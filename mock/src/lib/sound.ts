let ctx: AudioContext | null = null;
let enabled = false;

export function isSoundEnabled(): boolean {
  return enabled;
}

export function setSoundEnabled(value: boolean) {
  enabled = value;
  if (enabled && !ctx && typeof window !== 'undefined') {
    const Ctor =
      window.AudioContext ||
      (window as unknown as { webkitAudioContext: typeof AudioContext })
        .webkitAudioContext;
    if (Ctor) ctx = new Ctor();
  }
  if (enabled && ctx && ctx.state === 'suspended') {
    void ctx.resume();
  }
}

function tone(freq: number, start: number, dur: number, gainPeak = 0.18) {
  if (!ctx) return;
  const osc = ctx.createOscillator();
  const gain = ctx.createGain();
  osc.type = 'triangle';
  osc.frequency.setValueAtTime(freq, start);
  gain.gain.setValueAtTime(0, start);
  gain.gain.linearRampToValueAtTime(gainPeak, start + 0.012);
  gain.gain.exponentialRampToValueAtTime(0.0001, start + dur);
  osc.connect(gain).connect(ctx.destination);
  osc.start(start);
  osc.stop(start + dur);
}

export function playWinChime() {
  if (!enabled || !ctx) return;
  if (ctx.state === 'suspended') void ctx.resume();
  const t = ctx.currentTime;
  // ピンッ (high click)
  tone(1760, t, 0.06, 0.12);
  // チャリン (B5 → E6 sweep)
  tone(987.77, t + 0.05, 0.18, 0.18);
  tone(1318.51, t + 0.15, 0.32, 0.16);
  // tail shimmer
  tone(2637.02, t + 0.2, 0.4, 0.06);
}
