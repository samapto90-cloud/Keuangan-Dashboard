"""Split Gundam sprite sheet (4x3) into transparent PNG icons."""
from pathlib import Path
from PIL import Image

ROOT = Path(__file__).resolve().parents[1]
SRC = ROOT / "assets" / "gundam-sheet.png"
OUT = ROOT / "go-app" / "assets" / "gundam-icons"

NAMES = [
    "01-rx78", "02-wing-zero", "03-exia",
    "04-barbatos", "05-sazabi", "06-wing-custom",
    "07-unicorn", "08-red-frame", "09-full-armor",
    "10-zz", "11-freedom", "12-banshee",
]

COLS, ROWS = 3, 4
BLACK_THRESH = 28


def is_bg(r, g, b, a):
    if a < 10:
        return True
    return r <= BLACK_THRESH and g <= BLACK_THRESH and b <= BLACK_THRESH


def trim_alpha(img: Image.Image) -> Image.Image:
    bbox = img.getbbox()
    if not bbox:
        return img
    return img.crop(bbox)


def main():
    alt = list(ROOT.glob("assets/*gundam*.png"))
    src = SRC if SRC.exists() else (alt[0] if alt else None)
    if not src:
        raise SystemExit("Gundam sheet not found")
    OUT.mkdir(parents=True, exist_ok=True)
    sheet = Image.open(src).convert("RGBA")
    w, h = sheet.size
    cw, ch = w // COLS, h // ROWS
    idx = 0
    for row in range(ROWS):
        for col in range(COLS):
            if idx >= len(NAMES):
                break
            box = (col * cw, row * ch, (col + 1) * cw, (row + 1) * ch)
            cell = sheet.crop(box)
            px = cell.load()
            for y in range(cell.height):
                for x in range(cell.width):
                    r, g, b, a = px[x, y]
                    if is_bg(r, g, b, a):
                        px[x, y] = (0, 0, 0, 0)
            cell = trim_alpha(cell)
            dest = OUT / f"{NAMES[idx]}.png"
            cell.save(dest, "PNG", optimize=True)
            print(f"Wrote {dest.name} ({cell.width}x{cell.height})")
            idx += 1
    print(f"Done: {idx} icons -> {OUT}")


if __name__ == "__main__":
    main()
