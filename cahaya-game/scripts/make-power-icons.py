from PIL import Image
from pathlib import Path
import collections

sheet = Path(r"d:\Keuangan-Dashboard\Keuangan-Dashboard\cahaya-game\public\raka\powers-sheet.png")
if not sheet.exists():
    sheet = Path(
        r"C:\Users\DIK_u\.cursor\projects\d-Keuangan-Dashboard-Keuangan-Dashboard\assets"
        r"\c__Users_DIK_u_AppData_Roaming_Cursor_User_workspaceStorage_440b3db4abfb7ac132c8cab3b9ba38ee_images_image-62ecaceb-8170-4e4c-a3c2-4b55aed14695.png"
    )

img = Image.open(sheet).convert("RGBA")
w, h = img.size
cw = w // 3
out_dir = Path(r"d:\Keuangan-Dashboard\Keuangan-Dashboard\cahaya-game\public\raka\powers")
out_dir.mkdir(parents=True, exist_ok=True)
names = ["bomb", "thunder", "superman"]


def is_bg(r, g, b, a):
    if a < 8:
        return True
    avg = (r + g + b) / 3
    md = max(abs(r - g), abs(g - b), abs(r - b))
    if avg >= 235 and md <= 18:
        return True
    if avg >= 220 and md <= 10:
        return True
    return False


for i, name in enumerate(names):
    crop = img.crop((i * cw, 0, (i + 1) * cw, h)).convert("RGBA")
    px = crop.load()
    W, H = crop.size
    q = collections.deque()
    seen = [[False] * H for _ in range(W)]
    for x in range(W):
        q.append((x, 0))
        q.append((x, H - 1))
    for y in range(H):
        q.append((0, y))
        q.append((W - 1, y))
    while q:
        x, y = q.popleft()
        if x < 0 or y < 0 or x >= W or y >= H or seen[x][y]:
            continue
        seen[x][y] = True
        r, g, b, a = px[x, y]
        if not is_bg(r, g, b, a):
            continue
        px[x, y] = (0, 0, 0, 0)
        q.append((x + 1, y))
        q.append((x - 1, y))
        q.append((x, y + 1))
        q.append((x, y - 1))
    for y in range(H):
        for x in range(W):
            r, g, b, a = px[x, y]
            if a and is_bg(r, g, b, a):
                px[x, y] = (0, 0, 0, 0)
            elif a:
                avg = (r + g + b) / 3
                md = max(abs(r - g), abs(g - b), abs(r - b))
                if avg >= 210 and md <= 20:
                    t = (avg - 210) / 45.0
                    na = int(a * max(0.0, 1.0 - t))
                    px[x, y] = (r, g, b, na)
    bbox = crop.getbbox()
    if bbox:
        crop = crop.crop(bbox)
    pad = 8
    canvas = Image.new("RGBA", (crop.width + pad * 2, crop.height + pad * 2), (0, 0, 0, 0))
    canvas.paste(crop, (pad, pad), crop)
    path = out_dir / f"{name}.png"
    canvas.save(path, optimize=True)
    print(name, canvas.size, path.stat().st_size)

print("done")
