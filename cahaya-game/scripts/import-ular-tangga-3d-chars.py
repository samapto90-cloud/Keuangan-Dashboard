"""Import karakter ular-tangga-3d → PNG transparan (rembg u2netp)."""
from __future__ import annotations

from pathlib import Path

from PIL import Image
from rembg import new_session, remove

SRC = Path(r"C:\Users\DIK_u\OneDrive\Documents\ular-tangga-3d\src\assets\images")
OUT = Path(__file__).resolve().parents[1] / "public" / "raka" / "players"
HEADS = OUT / "heads"

MAP = {
    "char_prabowo_1787395411594.jpg": "prabowo",
    "char_gibran_1787395428148.jpg": "gibran",
    "char_jokowi_1787395440995.jpg": "jokowi",
    "char_anis_1787395455702.jpg": "anies",
    "char_ganjar_1787395471119.jpg": "ganjar",
    "char_sri_1787395485784.jpg": "sri-mulyani",
    "char_khofifah_1787395500769.jpg": "khofifah",
    "char_bahlil_1787395516508.jpg": "bahlil",
}

TARGET_H = 900


def trim(im: Image.Image, pad: int = 12) -> Image.Image:
    bbox = im.split()[-1].getbbox()
    if not bbox:
        return im
    l, t, r, b = bbox
    return im.crop((max(0, l - pad), max(0, t - pad), min(im.width, r + pad), min(im.height, b + pad)))


def make_head(im: Image.Image) -> Image.Image:
    w, h = im.size
    bottom = int(h * 0.52)
    alpha = im.split()[-1]
    bbox = alpha.crop((0, 0, w, bottom)).getbbox()
    if bbox:
        l, t, r, b = bbox
        pad = 8
        crop = im.crop((max(0, l - pad), max(0, t - pad), min(w, r + pad), min(bottom, b + pad)))
    else:
        crop = im.crop((0, 0, w, bottom))
    side = max(crop.width, crop.height, 1)
    canvas = Image.new("RGBA", (side, side), (0, 0, 0, 0))
    canvas.paste(crop, ((side - crop.width) // 2, (side - crop.height) // 2), crop)
    return canvas.resize((256, 256), Image.Resampling.LANCZOS)


def main() -> None:
    OUT.mkdir(parents=True, exist_ok=True)
    HEADS.mkdir(parents=True, exist_ok=True)
    print("loading u2netp session...", flush=True)
    session = new_session("u2netp")
    for fname, role in MAP.items():
        raw = Image.open(SRC / fname).convert("RGBA")
        cut = remove(raw, session=session)
        if not isinstance(cut, Image.Image):
            cut = Image.open(cut).convert("RGBA")  # type: ignore[arg-type]
        else:
            cut = cut.convert("RGBA")
        cut = trim(cut)
        scale = TARGET_H / max(1, cut.height)
        cut = cut.resize((max(1, int(cut.width * scale)), TARGET_H), Image.Resampling.LANCZOS)
        body = OUT / f"{role}.png"
        cut.save(body, "PNG", optimize=True)
        make_head(cut).save(HEADS / f"{role}.png", "PNG", optimize=True)
        print(f"{role}: {cut.size} ({body.stat().st_size // 1024}KB)", flush=True)


if __name__ == "__main__":
    main()
