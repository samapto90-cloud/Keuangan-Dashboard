"""Tajamkan sprite pemain + angkat gradasi gelap di kaki tanpa merusak transparansi."""
from __future__ import annotations

from pathlib import Path

from PIL import Image, ImageFilter, ImageEnhance

ROOT = Path(__file__).resolve().parents[1] / "public" / "raka" / "players"
TARGET_H = 1100


def lift_lower_gradient(im: Image.Image) -> Image.Image:
    """Angkat luminance di setengah bawah agar celana/kaki tidak 'hancur' hitam."""
    px = im.load()
    w, h = im.size
    for y in range(h):
        t = y / max(1, h - 1)
        # 0 di kepala → 1 di kaki
        lift = 1.0 + 0.42 * max(0.0, (t - 0.38) / 0.62)
        sat = 1.0 + 0.12 * max(0.0, (t - 0.38) / 0.62)
        for x in range(w):
            r, g, b, a = px[x, y]
            if a < 8:
                continue
            # Pertahankan highlight; angkat midtone/shadow
            nr = min(255, int(r * lift))
            ng = min(255, int(g * lift))
            nb = min(255, int(b * lift))
            # Sedikit saturate di area bawah
            avg = (nr + ng + nb) / 3
            nr = min(255, int(avg + (nr - avg) * sat))
            ng = min(255, int(avg + (ng - avg) * sat))
            nb = min(255, int(avg + (nb - avg) * sat))
            px[x, y] = (nr, ng, nb, a)
    return im


def clean_dark_fringe(im: Image.Image, thr: int = 18) -> Image.Image:
    """Buang fringe gelap semi-transparan di tepi (sering terlihat blur)."""
    px = im.load()
    w, h = im.size
    for y in range(h):
        for x in range(w):
            r, g, b, a = px[x, y]
            if a == 0:
                continue
            if a < 40 and r < thr and g < thr and b < thr:
                px[x, y] = (0, 0, 0, 0)
            elif a < 120 and r < thr and g < thr and b < thr:
                px[x, y] = (r, g, b, max(0, a - 50))
    return im


def enhance_one(src: Path) -> None:
    im = Image.open(src).convert("RGBA")
    im = clean_dark_fringe(im)
    # Upscale dulu agar setelah zoom papan tetap tajam
    if im.height < TARGET_H:
        scale = TARGET_H / im.height
        nw = max(1, int(im.width * scale))
        im = im.resize((nw, TARGET_H), Image.Resampling.LANCZOS)
    im = lift_lower_gradient(im)
    # Unsharp ringan — detail kepala→kaki lebih jelas
    rgb = im.convert("RGB")
    sharp = rgb.filter(ImageFilter.UnsharpMask(radius=1.6, percent=140, threshold=2))
    sharp = ImageEnhance.Contrast(sharp).enhance(1.06)
    sharp = ImageEnhance.Color(sharp).enhance(1.08)
    out = Image.merge("RGBA", (*sharp.split(), im.split()[-1]))
    out.save(src, "PNG", optimize=True)
    print(f"{src.name}: {out.size} ({src.stat().st_size // 1024}KB)")


def main() -> None:
    for path in sorted(ROOT.glob("*.png")):
        enhance_one(path)


if __name__ == "__main__":
    main()
