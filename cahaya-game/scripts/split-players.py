"""Split roster art into 4 transparent PNG player sprites."""
from pathlib import Path

from PIL import Image

SRC = Path(
    r"C:\Users\DIK_u\.cursor\projects\d-Keuangan-Dashboard-Keuangan-Dashboard\assets"
    r"\c__Users_DIK_u_AppData_Roaming_Cursor_User_workspaceStorage_440b3db4abfb7ac132c8cab3b9ba38ee_images"
    r"_gambar_pemain_ular_tangga-3ac0d6c1-a42e-4be2-8c00-ca3bbe4d8d54.png"
)
OUT = Path(__file__).resolve().parents[1] / "public" / "raka" / "players"
NAMES = ["prabowo", "ganjar", "anies", "sri-mulyani"]


def remove_white(img: Image.Image, thresh: int = 245) -> Image.Image:
    px = img.load()
    ww, hh = img.size
    for y in range(hh):
        for x in range(ww):
            r, g, b, _a = px[x, y]
            if r >= thresh and g >= thresh and b >= thresh:
                px[x, y] = (255, 255, 255, 0)
            elif r >= 230 and g >= 230 and b >= 230:
                avg = (r + g + b) / 3
                alpha = max(0, min(255, int((255 - avg) / (255 - 230) * 255)))
                px[x, y] = (r, g, b, alpha)
    return img


def trim(img: Image.Image, pad: int = 10) -> Image.Image:
    bbox = img.split()[-1].getbbox()
    if not bbox:
        return img
    l, t, r, b = bbox
    return img.crop(
        (
            max(0, l - pad),
            max(0, t - pad),
            min(img.width, r + pad),
            min(img.height, b + pad),
        )
    )


def main() -> None:
    OUT.mkdir(parents=True, exist_ok=True)
    im = Image.open(SRC).convert("RGBA")
    w, h = im.size
    slot = w // 4
    edge = 6
    for i, name in enumerate(NAMES):
        left = i * slot + (edge if i else 0)
        right = (i + 1) * slot - (edge if i < 3 else 0)
        crop = remove_white(im.crop((left, 0, right, h)))
        crop = trim(crop, pad=10)
        target_h = 520
        scale = target_h / crop.height
        nw = max(1, int(crop.width * scale))
        crop = crop.resize((nw, target_h), Image.Resampling.LANCZOS)
        path = OUT / f"{name}.png"
        crop.save(path, "PNG", optimize=True)
        print(f"{name}: {crop.size} -> {path} ({path.stat().st_size} bytes)")


if __name__ == "__main__":
    main()
