"""Recrop Midnight Soft mockups into clean hero + sidebar-foot assets (no baked UI/text)."""
from pathlib import Path
from PIL import Image

ASSETS = Path(r"C:\Users\DIK_u\.cursor\projects\d-Keuangan-Dashboard-Keuangan-Dashboard\assets")
OUT = Path(r"d:\Keuangan-Dashboard\Keuangan-Dashboard\go-app\assets\midnight")
OUT.mkdir(parents=True, exist_ok=True)

darks = sorted(ASSETS.glob("*images_fdb5a731*.jpg"), key=lambda p: p.stat().st_size, reverse=True)
lights = sorted(ASSETS.glob("*images_9750bffe*.jpg"), key=lambda p: p.stat().st_size, reverse=True)
if not darks or not lights:
    raise SystemExit("Mockup source images not found in assets/")


def crop_hero(im: Image.Image, mode: str) -> Image.Image:
    w, h = im.size
    sx, ty, by = int(w * 0.175), int(h * 0.078), int(h * 0.188)
    band = im.crop((sx, ty, w - int(w * 0.012), by))
    bw, bh = band.size
    # Center skyline only — skip left greeting + right baked quote
    left, right = int(bw * 0.28), int(bw * 0.68)
    hero = band.crop((left, int(bh * 0.08), right, bh - int(bh * 0.05)))
    return hero.resize((hero.width * 3, hero.height * 3), Image.Resampling.LANCZOS)


def crop_foot(im: Image.Image, mode: str) -> Image.Image:
    w, h = im.size
    sw = int(w * 0.168)
    # Decorative art only (waves / Barelang sketch) — no Collapse / motto text
    top, bottom = int(h * 0.855), int(h * 0.905)
    foot = im.crop((int(sw * 0.06), top, int(sw * 0.94), bottom))
    return foot.resize((foot.width * 4, foot.height * 4), Image.Resampling.LANCZOS)


for mode, path in (("dark", darks[0]), ("light", lights[0])):
    im = Image.open(path).convert("RGB")
    print(f"{mode} source {im.size}")
    hero = crop_hero(im, mode)
    foot = crop_foot(im, mode)
    hero.save(OUT / f"hero-{mode}.jpg", quality=93, optimize=True)
    foot.save(OUT / f"sidebar-foot-{mode}.jpg", quality=92, optimize=True)
    print(f"  hero {hero.size}  foot {foot.size}")
