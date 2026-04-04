// Galaxy map coordinate system constants and helpers.
//
// The galaxy map image (galaxy-map.jpg) is a 4500×4500 px top-down
// view of the galactic plane (X-Z axes, Sol-centered).
//
// SVG rendering convention (from EDST reference implementation):
//   - viewBox: (-45000, -20000, 90000, 90000)
//   - Image placed at SVG position (-45000, -20000), size 90000×90000
//   - SVG Y-axis is flipped; galactic Z maps directly to SVG Y
//   - A system at galactic (x, z) sits at SVG cx=x, cy=z
//
// All coordinate values are Sol-centered (0,0,0 = Sol).

/** Galactic X coordinate at the left edge of the image (ly). */
export const GALAXY_MAP_ORIGIN_X = -45000;

/** Galactic Z coordinate at the top edge of the image (ly). */
export const GALAXY_MAP_ORIGIN_Z = -20000;

/** Width and height of the image in galactic ly. */
export const GALAXY_MAP_EXTENT = 90000;

/** Image dimensions in pixels. */
export const GALAXY_MAP_IMAGE_SIZE = 4500;

/** Galactic light-years represented by one image pixel. */
export const GALAXY_MAP_LY_PER_PIXEL =
    GALAXY_MAP_EXTENT / GALAXY_MAP_IMAGE_SIZE; // 20

/**
 * SVG viewBox parameters matching the galaxy map coordinate space.
 * Usage: `viewBox="${svgViewBox()}"`
 */
export function galaxyMapViewBox(): string {
    return [
        GALAXY_MAP_ORIGIN_X,
        GALAXY_MAP_ORIGIN_Z,
        GALAXY_MAP_EXTENT,
        GALAXY_MAP_EXTENT,
    ].join(' ');
}

/**
 * Full URL of the galaxy map image given the bundle base URL.
 * @param baseUrl - value of BundlesService.bundleBaseUrl
 */
export function galaxyMapUrl(baseUrl: string): string {
    return `${baseUrl}/galaxy-map.jpg`;
}

/**
 * Convert Sol-centered galactic coordinates to image pixel coordinates.
 * px=0,pz=0 is the top-left corner of the image.
 */
export function galacticToPixel(
    gcX: number,
    gcZ: number,
): { px: number; pz: number } {
    return {
        px: (gcX - GALAXY_MAP_ORIGIN_X) / GALAXY_MAP_LY_PER_PIXEL,
        pz: (gcZ - GALAXY_MAP_ORIGIN_Z) / GALAXY_MAP_LY_PER_PIXEL,
    };
}

/**
 * Convert image pixel coordinates to Sol-centered galactic coordinates.
 * px=0,pz=0 is the top-left corner of the image.
 */
export function pixelToGalactic(
    px: number,
    pz: number,
): { gcX: number; gcZ: number } {
    return {
        gcX: px * GALAXY_MAP_LY_PER_PIXEL + GALAXY_MAP_ORIGIN_X,
        gcZ: pz * GALAXY_MAP_LY_PER_PIXEL + GALAXY_MAP_ORIGIN_Z,
    };
}
