# Galaxy Map

Read this skill when implementing any feature that renders or interacts with
the galaxy map image.

## Utility file

`frontend/src/app/utils/galaxy-map.ts` — pure TypeScript, no Angular
dependencies. Import individual symbols as needed.

## Coordinate system

- Top-down view of the galactic plane (X-Z axes), Sol-centered (0,0,0 = Sol)
- Image: `galaxy-map.jpg`, 4500×4500 px
- Galactic extent: 90000×90000 ly
- Top-left of image = galactic (-45000, -20000) in (x, z)
- Scale: 20 ly/pixel (`GALAXY_MAP_LY_PER_PIXEL`)
- All inputs/outputs are Sol-centered; GC-relative conversion is the
  caller's responsibility if needed

## SVG rendering convention

- SVG canvas Y-axis is flipped (`svg.transform({ flip: 'y' })`)
- A system at galactic (x, z) is placed at SVG `cx=x, cy=z` directly
- Image is also flipped on Y (`galaxy.flip('y')`) to display correctly
- `galaxyMapViewBox()` returns the correct `viewBox` string:
  `"-45000 -20000 90000 90000"`

## Image URL

```typescript
import { galaxyMapUrl } from '../utils/galaxy-map';
// bundlesService.bundleBaseUrl is fetched from /api/config at startup
const url = galaxyMapUrl(bundlesService.bundleBaseUrl!);
```

## Available exports

| Symbol | Type | Value / purpose |
|---|---|---|
| `GALAXY_MAP_ORIGIN_X` | `number` | `-45000` — left edge (ly) |
| `GALAXY_MAP_ORIGIN_Z` | `number` | `-20000` — top edge (ly) |
| `GALAXY_MAP_EXTENT` | `number` | `90000` — ly per axis |
| `GALAXY_MAP_IMAGE_SIZE` | `number` | `4500` — px |
| `GALAXY_MAP_LY_PER_PIXEL` | `number` | `20` |
| `galaxyMapViewBox()` | `() => string` | SVG viewBox attribute value |
| `galaxyMapUrl(baseUrl)` | `(string) => string` | Full image URL |
| `galacticToPixel(gcX, gcZ)` | `→ { px, pz }` | Galactic → pixel |
| `pixelToGalactic(px, pz)` | `→ { gcX, gcZ }` | Pixel → galactic |
