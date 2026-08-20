"""Split snake sheet by connected components (black bg)."""
from pathlib import Path

from PIL import Image, ImageEnhance
import colorsys

SRC = Path(
    r"C:\Users\DIK_u\.cursor\projects\d-Keuangan-Dashboard-Keuangan-Dashboard\assets"
    r"\c__Users_DIK_u_AppData_Roaming_Cursor_User_workspaceStorage_440b3db4abfb7ac132c8cab3b9ba38ee_images"
    r"_image-f1caf621-6133-4377-8f68-4120db8f4a5e.png"
)
OUT = Path(__file__).resolve().parents[1] / "public" / "raka" / "snakes"


def is_fg(r: int, g: int, b: int) -> bool:
    return (r + g + b) > 70


def flood_bbox(mask: list[list[bool]], w: int, h: int, sx: int, sy: int, seen: set[tuple[int, int]]):
    stack = [(sx, sy)]
    seen.add((sx, sy))
    minx = maxx = sx
    miny = maxy = sy
    count = 0
    while stack:
        x, y = stack.pop()
        count += 1
        if x < minx:
            minx = x
        if x > maxx:
            maxx = x
        if y < miny:
            miny = y
        if y > maxy:
            maxy = y
        for nx, ny in ((x - 1, y), (x + 1, y), (x, y - 1), (x, y + 1), (x - 1, y - 1), (x + 1, y - 1), (x - 1, y + 1), (x + 1, y + 1)):
            if 0 <= nx < w and 0 <= ny < h and mask[ny][nx] and (nx, ny) not in seen:
                seen.add((nx, ny))
                stack.append((nx, ny))
    return minx, miny, maxx, maxy, count


def remove_black(img: Image.Image, thresh: int = 28) -> Image.Image:
    px = img.load()
    w, h = img.size
    for y in range(h):
        for x in range(w):
            r, g, b, a = px[x, y]
            if r <= thresh and g <= thresh and b <= thresh:
                px[x, y] = (0, 0, 0, 0)
            elif r <= 40 and g <= 40 and b <= 40:
                avg = (r + g + b) / 3
                alpha = max(0, min(255, int(avg / 40 * 255)))
                px[x, y] = (r, g, b, min(a, alpha))
    return img


def trim(img: Image.Image, pad: int = 6) -> Image.Image:
    bbox = img.split()[-1].getbbox()
    if not bbox:
        return img
    l, t, r, b = bbox
    return img.crop((max(0, l - pad), max(0, t - pad), min(img.width, r + pad), min(img.height, b + pad)))


def dominant_hue(img: Image.Image) -> float:
    px = img.load()
    w, h = img.size
    acc = 0.0
    n = 0
    for y in range(0, h, 4):
        for x in range(0, w, 4):
            r, g, b, a = px[x, y]
            if a < 40:
                continue
            hh, s, v = colorsys.rgb_to_hsv(r / 255, g / 255, b / 255)
            if s > 0.25 and v > 0.2:
                acc += hh
                n += 1
    return acc / n if n else 0


def name_from_hue(h: float) -> str:
    # map HSV hue to labels
    if 0.08 <= h < 0.22:
        return "emas"
    if 0.22 <= h < 0.45:
        return "hijau"
    if 0.45 <= h < 0.58:
        return "toska"
    if 0.58 <= h < 0.72:
        return "biru"
    if 0.72 <= h < 0.92:
        return "ungu"
    return "merah"


def hue_shift(img: Image.Image, shift: float) -> Image.Image:
    src = img.convert("RGBA")
    out = Image.new("RGBA", src.size)
    sp = src.load()
    op = out.load()
    w, h = src.size
    for y in range(h):
        for x in range(w):
            r, g, b, a = sp[x, y]
            if a < 8:
                op[x, y] = (0, 0, 0, 0)
                continue
            hh, s, v = colorsys.rgb_to_hsv(r / 255, g / 255, b / 255)
            hh = (hh + shift) % 1.0
            rr, gg, bb = colorsys.hsv_to_rgb(hh, min(1.0, s * 1.05), v)
            op[x, y] = (int(rr * 255), int(gg * 255), int(bb * 255), a)
    return out


def main() -> None:
    OUT.mkdir(parents=True, exist_ok=True)
    for old in OUT.glob("*.png"):
        old.unlink()

    im = Image.open(SRC).convert("RGBA")
    w, h = im.size
    px = im.load()
    mask = [[is_fg(*px[x, y][:3]) for x in range(w)] for y in range(h)]
    seen: set[tuple[int, int]] = set()
    blobs = []
    for y in range(h):
        for x in range(w):
            if mask[y][x] and (x, y) not in seen:
                bbox = flood_bbox(mask, w, h, x, y, seen)
                if bbox[4] > 2500:
                    blobs.append(bbox)

    # largest 4 blobs
    blobs.sort(key=lambda b: b[4], reverse=True)
    blobs = blobs[:4]
    print("blobs", [(b[0], b[1], b[2], b[3], b[4]) for b in blobs])

    bases: list[Image.Image] = []
    used_names: set[str] = set()
    for i, (minx, miny, maxx, maxy, _c) in enumerate(blobs):
        pad = 4
        crop = im.crop((max(0, minx - pad), max(0, miny - pad), min(w, maxx + 1 + pad), min(h, maxy + 1 + pad)))
        crop = remove_black(crop)
        crop = trim(crop)
        th = 480
        scale = th / crop.height
        crop = crop.resize((max(1, int(crop.width * scale)), th), Image.Resampling.LANCZOS)
        hue = dominant_hue(crop)
        name = name_from_hue(hue)
        if name in used_names:
            name = f"{name}-{i}"
        used_names.add(name)
        path = OUT / f"{name}.png"
        crop.save(path, "PNG", optimize=True)
        print("saved", name, crop.size, f"hue={hue:.3f}")
        bases.append(crop)

    # Ensure we have canonical 4 names; rename by hue order if needed
    # Create 4 extra variants for 8 snakes
    variant_specs = [
        ("oranye", 0, 0.06),
        ("emas", 1, 0.10),
        ("toska", 2, 0.15),
        ("coklat", 3, 0.20),
    ]
    for name, idx, shift in variant_specs:
        src = bases[idx % len(bases)]
        # if file already exists from base naming, pick next shift
        out_name = name
        n = 0
        while (OUT / f"{out_name}.png").exists():
            n += 1
            out_name = f"{name}{n}" if n else name
            if n > 3:
                out_name = f"ular-{idx}-{n}"
                break
        v = ImageEnhance.Color(hue_shift(src, shift)).enhance(1.1)
        v.save(OUT / f"{out_name}.png", "PNG", optimize=True)
        print("variant", out_name, v.size)

    files = sorted(OUT.glob("*.png"))
    print("total", len(files), [f.name for f in files])


if __name__ == "__main__":
    main()
