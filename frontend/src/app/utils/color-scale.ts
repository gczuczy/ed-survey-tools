// Color scale utilities for density visualization.
// Provides Viridis/Plasma palettes, log normalization for rho values,
// and radius mapping for bowling-pin segments.

// ── Types ──────────────────────────────────────────────────────────────────

/** [t, [r, g, b]] where t ∈ [0,1], r/g/b ∈ [0,255] */
export type ColorStop = [number, [number, number, number]];

// ── Palette data ───────────────────────────────────────────────────────────

export const VIRIDIS: ColorStop[] = [
  [0.000, [ 68,   1,  84]],
  [0.125, [ 72,  40, 120]],
  [0.250, [ 62,  83, 160]],
  [0.375, [ 49, 123, 133]],
  [0.500, [ 33, 165, 133]],
  [0.625, [ 94, 201,  98]],
  [0.750, [183, 222,  41]],
  [0.875, [237, 248,  32]],
  [1.000, [253, 231,  37]],
];

export const PLASMA: ColorStop[] = [
  [0.000, [ 13,   8, 135]],
  [0.250, [126,   3, 168]],
  [0.500, [204,  71, 120]],
  [0.750, [248, 149,  64]],
  [1.000, [240, 249,  33]],
];

// ── Interpolation ──────────────────────────────────────────────────────────

/**
 * Interpolate a colour palette at position t ∈ [0,1].
 * Returns [r, g, b] each ∈ [0,255].
 */
export function colorScale(
  stops:  ColorStop[],
  t:      number,
): [number, number, number] {
  const clamped = Math.max(0, Math.min(1, t));
  if (stops.length === 0) return [0, 0, 0];
  if (clamped <= stops[0][0]) return [...stops[0][1]];
  const last = stops[stops.length - 1];
  if (clamped >= last[0]) return [...last[1]];
  for (let i = 1; i < stops.length; i++) {
    const [t0, c0] = stops[i - 1];
    const [t1, c1] = stops[i];
    if (clamped <= t1) {
      const f = (clamped - t0) / (t1 - t0);
      return [
        Math.round(c0[0] + f * (c1[0] - c0[0])),
        Math.round(c0[1] + f * (c1[1] - c0[1])),
        Math.round(c0[2] + f * (c1[2] - c0[2])),
      ];
    }
  }
  return [...last[1]];
}

// ── Rho normalization ──────────────────────────────────────────────────────

/**
 * Map a rho value to t ∈ [0,1] using log1p normalization.
 *
 * scaleMin: effective minimum rho (dataMin or viewMin, non-zero)
 * scaleMax: effective upper bound (dataMax or viewMax, user-adjustable)
 *
 * rho=0 → t≈0 (visible stub).  rho≥scaleMax → t=1 (saturated).
 */
export function rhoLogT(
  rho:      number,
  scaleMin: number,
  scaleMax: number,
): number {
  if (scaleMax <= 0 || scaleMin <= 0) return 0;
  const floor = scaleMin / 10;
  const num   = Math.log1p(Math.max(rho, 0) / floor);
  const denom = Math.log1p(scaleMax / floor);
  return Math.max(0, Math.min(1, num / denom));
}

// ── Radius mapping ─────────────────────────────────────────────────────────

/** Minimum pin-segment radius (ly) — always visible, even for rho=0 gaps. */
export const MIN_RADIUS = 30;

/** Maximum pin-segment radius (ly) — corresponds to t=1. */
export const MAX_RADIUS = 500;

/** Map t ∈ [0,1] to a cylinder radius (ly). */
export function tToRadius(t: number): number {
  return MIN_RADIUS + (MAX_RADIUS - MIN_RADIUS) * Math.max(0, Math.min(1, t));
}

// ── Three.js colour helper ─────────────────────────────────────────────────

/**
 * Return a Three.js-compatible hex colour integer from rho.
 * Import THREE.Color and pass the result to `new THREE.Color(hexInt)`.
 */
export function rhoToHex(
  rho:      number,
  scaleMin: number,
  scaleMax: number,
  palette:  ColorStop[] = VIRIDIS,
): number {
  const [r, g, b] = colorScale(
    palette, rhoLogT(rho, scaleMin, scaleMax),
  );
  return (r << 16) | (g << 8) | b;
}
