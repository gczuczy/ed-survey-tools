#!/usr/bin/env python3
"""
Elite Dangerous Galaxy Map Generator
Outputs: static/galaxy.png, static/galaxy.webp

Coordinate system: Sol-centered ED game units (light years)
  X: galactic east-west  (positive = right in top-down view)
  Z: along galactic plane (positive = toward GC and beyond; up in image)
  Y: galactic height — ignored in this 2D projection

Image centre = Galactic Centre (GC).
"""

import numpy as np
from PIL import Image, ImageDraw, ImageFont
import math, os, sys

# ============================================================
#  COORDINATE SYSTEM
# ============================================================
GC_X = 25.21875        # Galactic Centre, Sol-centered X (ly)
GC_Z = 25899.96875     # Galactic Centre, Sol-centered Z (ly)

GALAXY_R = 50000.0     # approximate galaxy radius from GC (ly)
IMG_HALF  = 56000.0    # image spans ±IMG_HALF ly from GC on each axis

X_MIN = GC_X - IMG_HALF
X_MAX = GC_X + IMG_HALF
Z_MIN = GC_Z - IMG_HALF
Z_MAX = GC_Z + IMG_HALF
SPAN  = IMG_HALF * 2.0   # 112 000 ly per axis

IMG_W = 8192
IMG_H = 8192

PX_PER_LY = IMG_W / SPAN          # pixels per light year

def sol_to_px(sol_x, sol_z):
    """Sol-centered (X,Z) → pixel (px, py).
    Z increases upward → py increases downward."""
    px = (sol_x - X_MIN) / SPAN * IMG_W
    py = (Z_MAX  - sol_z) / SPAN * IMG_H
    return int(round(px)), int(round(py))

def px_to_sol(px, py):
    """Pixel → Sol-centered (X,Z)."""
    return (X_MIN + px / IMG_W * SPAN,
            Z_MAX  - py / IMG_H * SPAN)

# ============================================================
#  REFERENCE POINTS  (sol-centered X, Z)
# ============================================================
REFS = {
    "Sol":                 (0.0,        0.0),
    "Galactic Centre":     (GC_X,       GC_Z),
    "Blia Aod QE-Q e5-0":  (25598.25,  43507.66),
    "LOS":                 (-9509.34,  19820.13),
}

# ============================================================
#  GALAXY BACKGROUND  (generated at BG_RES, upscaled to IMG_W)
# ============================================================
BG_RES = 2048

def generate_galaxy_bg():
    """Procedural Milky Way background."""
    rng = np.random.default_rng(1701)
    t   = np.linspace(0.5/BG_RES, 1.0 - 0.5/BG_RES, BG_RES, dtype=np.float32)
    PX, PY = np.meshgrid(t, t)

    XGC = (X_MIN + PX * SPAN).astype(np.float32) - GC_X
    ZGC = (Z_MAX  - PY * SPAN).astype(np.float32) - GC_Z
    R     = np.hypot(XGC, ZGC)
    THETA = np.arctan2(XGC, ZGC)   # CW from GC-north (+ZGC direction)

    # -- disk (exponential scale radius 14 kly) --
    disk = np.exp(-R / 14000.0)

    # -- central bulge --
    bulge = 3.0 * np.exp(-0.5 * (R / 2200.0)**2)

    # -- bar (elongated along XGC, i.e. galactic east-west) --
    bar = 1.2 * np.exp(-(XGC / 5500.0)**2 - (ZGC / 1400.0)**4)

    # -- 4 logarithmic spiral arms (pitch ~13 deg) --
    pitch = np.tan(np.radians(13.0))
    r0    = 4000.0
    arm_w = 0.20
    arms  = np.zeros_like(R)
    for ph in (0.0, math.pi/2, math.pi, 3*math.pi/2):
        ideal = np.log(np.maximum(R, r0) / r0) / pitch + ph
        dth   = (THETA - ideal + math.pi) % (2*math.pi) - math.pi
        a     = np.exp(-0.5*(dth/arm_w)**2) * np.exp(-R/42000.0)
        a[R < r0] = 0
        arms += a

    galaxy = np.clip(disk + bulge + bar + 0.65*arms, 0, None)

    # soft galaxy edge
    edge   = np.clip((GALAXY_R - R) / 5000.0, 0, 1)
    galaxy *= edge
    galaxy /= galaxy.max()

    # -- colour: dark bg → blue-grey disk → orange bulge → white core --
    b6k = np.exp(-0.5 * (R / 6000.0)**2)
    b7h = np.exp(-0.5 * (R /  700.0)**2)

    Rc = np.clip(galaxy*0.28 + b6k*0.60 + b7h*0.95, 0, 1)
    Gc = np.clip(galaxy*0.30 + b6k*0.35 + b7h*0.90, 0, 1)
    Bc = np.clip(galaxy*0.50 + b6k*0.05 + b7h*0.75, 0, 1)

    # sparse star field
    stars = rng.random((BG_RES, BG_RES), dtype=np.float32) ** 80 * 0.5
    Rc = np.clip(Rc + stars*0.5, 0, 1) ** (1/2.2)
    Gc = np.clip(Gc + stars*0.5, 0, 1) ** (1/2.2)
    Bc = np.clip(Bc + stars*0.6, 0, 1) ** (1/2.2)

    rgb = np.stack([
        (Rc*255).astype(np.uint8),
        (Gc*255).astype(np.uint8),
        (Bc*255).astype(np.uint8),
    ], axis=2)
    img = Image.fromarray(rgb, "RGB")
    return img.resize((IMG_W, IMG_H), Image.LANCZOS)

# ============================================================
#  SECTOR BOUNDARY GEOMETRY
# ============================================================
D = math.pi / 180.0

def _arc(r, t0, t1, n=48):
    """List of (sol_x, sol_z) along a GC-centred arc.
    Theta in degrees, CW from GC-north."""
    return [
        (GC_X + r*math.sin(a*D), GC_Z + r*math.cos(a*D))
        for a in np.linspace(t0, t1, n+1)
    ]

def ring_poly(r_in, r_out, t0, t1, n=48):
    """Annular sector polygon (sol-centred coords)."""
    outer = _arc(r_out, t0, t1, n)
    inner = list(reversed(_arc(r_in, t0, t1, n)))
    return outer + inner

def full_ring(r_in, r_out, n=128):
    """Full annulus (360 deg)."""
    return ring_poly(r_in, r_out, 0, 360, n)

def poly_to_px(poly):
    """Convert list of (sol_x, sol_z) → list of (px, py)."""
    return [sol_to_px(x, z) for x, z in poly]

# ---- Region definitions -------------------------------------------
# Angles are GC-centred, CW from GC-north (+ZGC = high Z direction).
# Sol is at ~180 deg, r ~25 900 ly from GC.
# Blia Aod QE-Q e5-0 is at ~55 deg, r ~31 050 ly from GC.
# LOS is at ~238 deg, r ~11 300 ly from GC.
#
# Boundaries are APPROXIMATE — derived from in-game screenshot analysis
# and the known galactic structure of Elite Dangerous.
# -----------------------------------------------------------------------

REGIONS = [
    # ---- Galactic core ------------------------------------------------
    {"name": "Galactic Centre / Sagittarius A*",
     "label_angle":   0, "label_r": 3200,
     "polys": [full_ring(0, 5500, 96)]},

    # ---- Inner ring  5 500 – 14 000 ly --------------------------------
    {"name": "Inner Scutum-Centaurus Arm",
     "label_angle": 145, "label_r": 9500,
     "polys": [ring_poly(5500, 14000, 120, 175)]},

    {"name": "Inner Sagittarius-Carina Arm",
     "label_angle": 195, "label_r": 9500,
     "polys": [ring_poly(5500, 14000, 175, 230)]},

    {"name": "Inner Norma Arm",
     "label_angle": 250, "label_r": 9500,
     "polys": [ring_poly(5500, 14000, 230, 285)]},

    {"name": "Empyrean Marches",
     "label_angle": 310, "label_r": 9500,
     "polys": [ring_poly(5500, 14000, 285, 355)]},

    {"name": "Inner Orion-GC Bridge",
     "label_angle":  40, "label_r": 9500,
     "polys": [ring_poly(5500, 14000, 355, 120)]},

    # ---- Middle ring  14 000 – 27 000 ly ------------------------------
    {"name": "Scutum-Centaurus Arm",
     "label_angle": 130, "label_r": 20000,
     "polys": [ring_poly(14000, 27000, 108, 162)]},

    {"name": "Sagittarius-Carina Arm",
     "label_angle": 182, "label_r": 20000,
     "polys": [ring_poly(14000, 27000, 162, 215)]},

    {"name": "Norma Arm",
     "label_angle": 232, "label_r": 20000,
     "polys": [ring_poly(14000, 27000, 215, 268)]},

    {"name": "Outer Empyrean Marches",
     "label_angle": 295, "label_r": 20000,
     "polys": [ring_poly(14000, 27000, 268, 328)]},

    {"name": "Dryman's Point",
     "label_angle": 355, "label_r": 20000,
     "polys": [ring_poly(14000, 27000, 328,  18)]},

    {"name": "Mare Somnia (inner)",
     "label_angle":  55, "label_r": 20000,
     "polys": [ring_poly(14000, 27000,  18,  88)]},

    {"name": "Formorian Frontier (inner)",
     "label_angle": 100, "label_r": 20000,
     "polys": [ring_poly(14000, 27000,  88, 108)]},

    # ---- Outer ring  27 000 – 41 000 ly  (Sol & Blia Aod band) -------
    # Sol at 180 deg, r ~25 900 → just inside the inner edge of this ring
    # Blia Aod at ~55 deg, r ~31 050 → solidly in this ring
    {"name": "Newton's Vault",
     "label_angle": 118, "label_r": 33500,
     "polys": [ring_poly(27000, 41000, 100, 150)]},

    {"name": "The Veils",
     "label_angle": 155, "label_r": 33500,
     "polys": [ring_poly(27000, 41000, 150, 172)]},

    {"name": "Inner Orion Spur",       # Sol is here
     "label_angle": 180, "label_r": 33500,
     "polys": [ring_poly(27000, 41000, 172, 202)]},

    {"name": "Hawking's Gap",
     "label_angle": 215, "label_r": 33500,
     "polys": [ring_poly(27000, 41000, 202, 255)]},

    {"name": "Khalar",
     "label_angle": 275, "label_r": 33500,
     "polys": [ring_poly(27000, 41000, 255, 318)]},

    {"name": "Polaris",
     "label_angle": 340, "label_r": 33500,
     "polys": [ring_poly(27000, 41000, 318,   2)]},

    {"name": "Acheron (inner)",
     "label_angle":  22, "label_r": 33500,
     "polys": [ring_poly(27000, 41000,   2,  42)]},

    {"name": "Mare Somnia",            # Blia Aod QE-Q e5-0 is here
     "label_angle":  62, "label_r": 33500,
     "polys": [ring_poly(27000, 41000,  42,  82)]},

    {"name": "Formorian Frontier",
     "label_angle":  92, "label_r": 33500,
     "polys": [ring_poly(27000, 41000,  82, 100)]},

    # ---- Far outer ring  41 000 – 53 000 ly ---------------------------
    {"name": "Outer Scutum-Centaurus Arm",
     "label_angle": 118, "label_r": 46500,
     "polys": [ring_poly(41000, 53000,  95, 158)]},

    {"name": "Outer Orion-Perseus Conflux",
     "label_angle": 178, "label_r": 46500,
     "polys": [ring_poly(41000, 53000, 158, 220)]},

    {"name": "Ryker's Hope",
     "label_angle": 240, "label_r": 46500,
     "polys": [ring_poly(41000, 53000, 220, 270)]},

    {"name": "Hawking's Gap (outer)",
     "label_angle": 292, "label_r": 46500,
     "polys": [ring_poly(41000, 53000, 270, 318)]},

    {"name": "Polaris / Far North",
     "label_angle": 340, "label_r": 46500,
     "polys": [ring_poly(41000, 53000, 318, 360)]},

    {"name": "Acheron",
     "label_angle":  18, "label_r": 46500,
     "polys": [ring_poly(41000, 53000,   0,  45)]},

    {"name": "Formorian Frontier (outer)",
     "label_angle":  72, "label_r": 46500,
     "polys": [ring_poly(41000, 53000,  45,  95)]},
]

# ============================================================
#  DRAWING
# ============================================================
BORDER_COL  = (220, 130,  30, 210)   # orange, semi-transparent
LABEL_COL   = (220, 155,  60, 180)   # dim orange for region labels
BORDER_W    = 4                       # px at 8192 resolution
GRID_COL    = ( 60,  60,  80, 100)   # faint grid lines

REF_COLOURS = {
    "Sol":                (80, 200, 255),
    "Galactic Centre":    (255, 220, 60),
    "Blia Aod QE-Q e5-0": (255, 80, 80),
    "LOS":                (120, 255, 120),
}
DOT_R = 22    # reference dot radius (px)

def draw_sectors(img):
    """Draw all region boundary polygons onto img (RGBA)."""
    overlay = Image.new("RGBA", img.size, (0, 0, 0, 0))
    draw    = ImageDraw.Draw(overlay)

    for region in REGIONS:
        for poly in region["polys"]:
            pts = poly_to_px(poly)
            # filled area (very faint)
            draw.polygon(pts, fill=(220, 140, 30, 12))
            # border lines
            draw.polygon(pts, outline=BORDER_COL, width=BORDER_W)

    img = Image.alpha_composite(img.convert("RGBA"), overlay)
    return img

def draw_labels(img, font_size=28):
    """Draw region name labels."""
    overlay = Image.new("RGBA", img.size, (0, 0, 0, 0))
    draw    = ImageDraw.Draw(overlay)

    try:
        font = ImageFont.truetype(
            "/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf", font_size)
        font_sm = ImageFont.truetype(
            "/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf",
            int(font_size * 0.75))
    except Exception:
        font = font_sm = ImageFont.load_default()

    for region in REGIONS:
        la = region["label_angle"]
        lr = region["label_r"]
        lx = GC_X + lr * math.sin(la * D)
        lz = GC_Z + lr * math.cos(la * D)
        px, py = sol_to_px(lx, lz)

        name = region["name"]
        # word-wrap long names at ~20 chars
        words  = name.split()
        lines  = []
        cur    = ""
        for w in words:
            test = (cur + " " + w).strip()
            if len(test) > 18 and cur:
                lines.append(cur)
                cur = w
            else:
                cur = test
        if cur:
            lines.append(cur)

        line_h = font_size + 4
        y0     = py - (len(lines) * line_h) // 2
        for i, line in enumerate(lines):
            bbox = draw.textbbox((0, 0), line, font=font)
            tw   = bbox[2] - bbox[0]
            draw.text((px - tw//2, y0 + i*line_h), line,
                      font=font, fill=LABEL_COL)

    img = Image.alpha_composite(img.convert("RGBA"), overlay)
    return img

def draw_refs(img):
    """Draw reference system markers and labels."""
    overlay = Image.new("RGBA", img.size, (0, 0, 0, 0))
    draw    = ImageDraw.Draw(overlay)

    try:
        font = ImageFont.truetype(
            "/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf", 34)
    except Exception:
        font = ImageFont.load_default()

    for name, (sx, sz) in REFS.items():
        px, py = sol_to_px(sx, sz)
        col    = REF_COLOURS.get(name, (255, 255, 255))
        r      = DOT_R

        # glow ring
        draw.ellipse((px-r-6, py-r-6, px+r+6, py+r+6),
                     outline=(*col, 80), width=6)
        # filled dot
        draw.ellipse((px-r, py-r, px+r, py+r),
                     fill=(*col, 230), outline=(255,255,255,180), width=3)

        # cross-hair lines
        L = r + 18
        draw.line((px-L, py, px-r-2, py), fill=(*col, 200), width=3)
        draw.line((px+r+2, py, px+L, py),   fill=(*col, 200), width=3)
        draw.line((px, py-L, px, py-r-2), fill=(*col, 200), width=3)
        draw.line((px, py+r+2, px, py+L),   fill=(*col, 200), width=3)

        # label
        label = name
        bbox  = draw.textbbox((0, 0), label, font=font)
        tw    = bbox[2] - bbox[0]
        draw.text((px - tw//2, py + r + 12), label,
                  font=font, fill=(*col, 230))

    img = Image.alpha_composite(img.convert("RGBA"), overlay)
    return img

def draw_grid(img, step_ly=10000):
    """Faint coordinate grid lines at every step_ly."""
    overlay = Image.new("RGBA", img.size, (0, 0, 0, 0))
    draw    = ImageDraw.Draw(overlay)

    # vertical lines (constant sol_x)
    x = math.ceil(X_MIN / step_ly) * step_ly
    while x <= X_MAX:
        px, _ = sol_to_px(x, 0)
        draw.line((px, 0, px, IMG_H), fill=GRID_COL, width=1)
        x += step_ly

    # horizontal lines (constant sol_z)
    z = math.ceil(Z_MIN / step_ly) * step_ly
    while z <= Z_MAX:
        _, py = sol_to_px(0, z)
        draw.line((0, py, IMG_W, py), fill=GRID_COL, width=1)
        z += step_ly

    img = Image.alpha_composite(img.convert("RGBA"), overlay)
    return img

def draw_galaxy_border(img):
    """Draw outer galaxy boundary circle."""
    overlay = Image.new("RGBA", img.size, (0, 0, 0, 0))
    draw    = ImageDraw.Draw(overlay)
    pts     = poly_to_px(_arc(GALAXY_R, 0, 360, 256))
    draw.polygon(pts, outline=(180, 100, 20, 160), width=6)
    img = Image.alpha_composite(img.convert("RGBA"), overlay)
    return img

# ============================================================
#  MAIN
# ============================================================
def main():
    out_dir = os.path.join(os.path.dirname(__file__), "static")
    os.makedirs(out_dir, exist_ok=True)
    png_path  = os.path.join(out_dir, "galaxy.png")
    webp_path = os.path.join(out_dir, "galaxy.webp")

    print("Generating galaxy background…")
    img = generate_galaxy_bg()
    print(f"  background: {img.size}")

    print("Drawing coordinate grid…")
    img = draw_grid(img, step_ly=10000)

    print("Drawing galaxy boundary…")
    img = draw_galaxy_border(img)

    print("Drawing sector boundaries…")
    img = draw_sectors(img)

    print("Drawing region labels…")
    img = draw_labels(img, font_size=36)

    print("Drawing reference points…")
    img = draw_refs(img)

    # flatten RGBA → RGB for saving
    bg  = Image.new("RGB", img.size, (0, 0, 0))
    bg.paste(img, mask=img.split()[3])
    img = bg

    print(f"Saving PNG → {png_path}")
    img.save(png_path, "PNG", optimize=False)

    print(f"Saving WebP → {webp_path}")
    img.save(webp_path, "WEBP", quality=88, method=6)

    print("Done.")
    pxs = sol_to_px(*REFS["Sol"])
    pxg = sol_to_px(*REFS["Galactic Centre"])
    pxb = sol_to_px(*REFS["Blia Aod QE-Q e5-0"])
    pxl = sol_to_px(*REFS["LOS"])
    print(f"  Sol px:        {pxs}")
    print(f"  GC  px:        {pxg}")
    print(f"  Blia Aod px:   {pxb}")
    print(f"  LOS px:        {pxl}")
    print(f"  Scale: {PX_PER_LY:.4f} px/ly  |  1000 ly = {PX_PER_LY*1000:.1f} px")


if __name__ == "__main__":
    main()
