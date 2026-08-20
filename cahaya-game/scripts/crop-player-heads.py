"""Crop head busts from full-body player sprites for board tokens."""
from pathlib import Path

from PIL import Image

ROOT = Path(__file__).resolve().parents[1] / "public" / "raka" / "players"
OUT = ROOT / "heads"
NAMES = ["prabowo", "ganjar", "anies", "sri-mulyani"]


def main() -> None:
    OUT.mkdir(parents=True, exist_ok=True)
    for name in NAMES:
        src = ROOT / f"{name}.png"
        im = Image.open(src).convert("RGBA")
        w, h = im.size
        # Focus on oversized head: top ~48% height, centered width
        top = 0
        bottom = int(h * 0.48)
        # tighten horizontal to non-transparent
        alpha = im.split()[-1]
        bbox = alpha.crop((0, top, w, bottom)).getbbox()
        if bbox:
            l, t, r, b = bbox
            pad = 8
            l = max(0, l - pad)
            t = max(0, t - pad)
            r = min(w, r + pad)
            b = min(bottom, b + pad)
            crop = im.crop((l, top + t, r, top + b))
        else:
            crop = im.crop((0, top, w, bottom))

        # square canvas, head centered, slight upscale feel
        side = max(crop.width, crop.height)
        canvas = Image.new("RGBA", (side, side), (0, 0, 0, 0))
        ox = (side - crop.width) // 2
        oy = (side - crop.height) // 2
        canvas.paste(crop, (ox, oy), crop)
        out = canvas.resize((256, 256), Image.Resampling.LANCZOS)
        path = OUT / f"{name}.png"
        out.save(path, "PNG", optimize=True)
        print(name, out.size, path.stat().st_size)


if __name__ == "__main__":
    main()
