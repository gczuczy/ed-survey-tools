# Galaxy Map — Coordinate Mapping

Source image: `static/galaxy.png` / `static/galaxy.webp`
Size: 8192 × 8192 px

---

## Coordinate system

All game coordinates are **Sol-centered Elite Dangerous units (light years)**.

| Axis | Direction | Notes |
|------|-----------|-------|
| X    | positive = galactic east (right in image) | |
| Z    | positive = toward GC and beyond (up in image) | |
| Y    | galactic height — **not used in 2D map** | |

Key reference positions (Sol-centered X, Z):

| System | sol_x | sol_z |
|--------|-------|-------|
| Sol | 0.0 | 0.0 |
| Galactic Centre (GC) | 25.21875 | 25899.96875 |
| Blia Aod QE-Q e5-0 | 25 598.25 | 43 507.66 |
| LOS | −9 509.34 | 19 820.13 |

The database stores Sol-centered X, Z coordinates directly.
GC-relative: `gc_x = sol_x − 25.21875`, `gc_z = sol_z − 25899.96875`

---

## Image bounds

The image is centred on the **Galactic Centre**, spanning ±56 000 ly in each axis:

| Constant | Value (Sol-centered ly) |
|----------|------------------------|
| `X_MIN` | −54 974.78125 |
| `X_MAX` | +55 025.21875 |
| `Z_MIN` | −30 100.03125 |  ← bottom edge (low Z)
| `Z_MAX` | +81 899.96875 |  ← top edge   (high Z)
| `SPAN`  | 112 000 ly per axis |
| Scale   | 0.0731 px/ly · 73.1 px per 1 000 ly |

---

## Pixel ↔ coordinate formulas

### Game → pixel

```ts
const X_MIN = 25.21875  - 56000;   // -54974.78125
const Z_MAX = 25899.96875 + 56000;  // +81899.96875
const SPAN  = 112000;
const IMG_W = 8192;
const IMG_H = 8192;

function solToPx(sol_x: number, sol_z: number): [number, number] {
  const px = (sol_x - X_MIN) / SPAN * IMG_W;
  const py = (Z_MAX  - sol_z) / SPAN * IMG_H;   // Z↑ → py↓
  return [Math.round(px), Math.round(py)];
}
```

### Pixel → game

```ts
function pxToSol(px: number, py: number): [number, number] {
  const sol_x = X_MIN + (px / IMG_W) * SPAN;
  const sol_z = Z_MAX  - (py / IMG_H) * SPAN;
  return [sol_x, sol_z];
}
```

### Verified pixel positions

| System | px | py |
|--------|----|----|
| Sol | 4094 | 5990 |
| Galactic Centre | 4096 | 4096 |
| Blia Aod QE-Q e5-0 | 5966 | 2808 |
| LOS | 3399 | 4541 |

---

## UI usage pattern

When rendering the galaxy map as a background image and overlaying data:

1. Load `galaxy.webp` as the background (4.7 MB, quality 88).
2. For each data point `(sol_x, sol_z)` call `solToPx()` to get CSS pixel
   position relative to the image's natural 8192×8192 grid.
3. Scale positions by `displayWidth / 8192` to handle zoom/resize.

```ts
// Example: place an SVG circle at a survey centroid
const [px, py] = solToPx(centroid.sol_x, centroid.sol_z);
const scale    = containerWidth / 8192;
circle.style.left = `${px * scale}px`;
circle.style.top  = `${py * scale}px`;
```

---

## Sector boundaries

The orange sector lines in the image are **approximate** — they are
structured as concentric rings + angular dividers centred on GC, broadly
matching the named regions visible in the in-game Universal Cartographics
galaxy map:

| Ring | GC radius range (ly) | Named regions |
|------|----------------------|---------------|
| Core | 0 – 5 500 | Galactic Centre / Sagittarius A* |
| Inner | 5 500 – 14 000 | Inner SC Arm, Inner Sag-Car Arm, Inner Norma, Empyrean Marches |
| Middle | 14 000 – 27 000 | Scutum-Centaurus, Sag-Carina, Norma, Dryman's Point, Mare Somnia (inner) |
| Outer | 27 000 – 41 000 | Inner Orion Spur (Sol), Mare Somnia (Blia Aod), Formorian Frontier, Khalar, Polaris, Newton's Vault |
| Far outer | 41 000 – 53 000 | Outer SC Arm, Outer Orion-Perseus, Ryker's Hope, Acheron |

If precise sector polygons are needed, regenerate with updated vertex data
in `generate_galaxy.py` → `REGIONS` list.
