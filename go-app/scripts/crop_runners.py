"""Remove backgrounds and crop runner PNGs to character silhouette."""
from PIL import Image, ImageFilter
import os
import numpy as np
from collections import deque

BASE = os.path.join(os.path.dirname(__file__), '..', 'assets')
DIRS = ['op-runners', 'frozen-runners', 'doraemon-runners', 'naruto-smp-runners', 'ds-kas-runners']
MAX_H = 520
PAD = 16


def sample_bg_color(arr):
    h, w = arr.shape[:2]
    pts = [
        arr[0, 0, :3], arr[0, w - 1, :3], arr[h - 1, 0, :3], arr[h - 1, w - 1, :3],
        arr[0, w // 2, :3], arr[h - 1, w // 2, :3], arr[h // 2, 0, :3], arr[h // 2, w - 1, :3],
    ]
    return np.median(np.array(pts, dtype=np.float32), axis=0)


def is_background(r, g, b, bg, tol=42):
    dr = abs(float(r) - bg[0])
    dg = abs(float(g) - bg[1])
    db = abs(float(b) - bg[2])
    if max(dr, dg, db) <= tol:
        return True
    brightness = (float(r) + float(g) + float(b)) / 3.0
    spread = max(r, g, b) - min(r, g, b)
    if brightness >= 248 and spread < 24:
        return True
    if brightness >= 232 and spread < 20:
        return True
    if brightness >= 215 and spread < 18:
        return True
    # Off-white / cream studio backdrop
    if brightness >= 200 and spread < 28 and abs(dr - dg) < 12 and abs(dg - db) < 12:
        return True
    return False


def flood_remove_background(arr, tol=42):
    h, w = arr.shape[:2]
    bg = sample_bg_color(arr)
    visited = np.zeros((h, w), dtype=bool)
    q = deque()

    def try_seed(x, y):
        if 0 <= x < w and 0 <= y < h and not visited[y, x]:
            r, g, b, a = arr[y, x]
            if a > 10 and is_background(r, g, b, bg, tol):
                q.append((x, y))

    for x in range(w):
        try_seed(x, 0)
        try_seed(x, h - 1)
    for y in range(h):
        try_seed(0, y)
        try_seed(w - 1, y)

    while q:
        x, y = q.popleft()
        if visited[y, x]:
            continue
        r, g, b, a = arr[y, x]
        if a <= 10 or not is_background(r, g, b, bg, tol):
            continue
        visited[y, x] = True
        arr[y, x, 3] = 0
        for nx, ny in ((x + 1, y), (x - 1, y), (x, y + 1), (x, y - 1)):
            if 0 <= nx < w and 0 <= ny < h and not visited[ny, nx]:
                q.append((nx, ny))

    return arr


def refine_alpha_edges(arr):
    h, w = arr.shape[:2]
    alpha = arr[:, :, 3].astype(np.float32)
    for y in range(1, h - 1):
        for x in range(1, w - 1):
            if alpha[y, x] <= 0:
                continue
            r, g, b = arr[y, x, :3]
            brightness = (int(r) + int(g) + int(b)) / 3.0
            spread = max(r, g, b) - min(r, g, b)
            nbr = alpha[y - 1:y + 2, x - 1:x + 2]
            transparent_n = int(np.sum(nbr == 0))
            if brightness >= 245 and spread < 22:
                arr[y, x, 3] = 0
                continue
            if transparent_n >= 4 and brightness >= 190 and spread < 32:
                arr[y, x, 3] = 0
                continue
            if transparent_n >= 2 and brightness >= 215 and spread < 28:
                arr[y, x, 3] = max(0, int(min(alpha[y, x], (255 - brightness) * 5)))
            elif transparent_n >= 1 and brightness >= 230 and spread < 24:
                arr[y, x, 3] = max(0, int(min(alpha[y, x], (255 - brightness) * 4)))
    return arr


def despeckle(arr, min_neighbors=2):
    h, w = arr.shape[:2]
    alpha = arr[:, :, 3].copy()
    for y in range(1, h - 1):
        for x in range(1, w - 1):
            if alpha[y, x] <= 0:
                continue
            nbr = alpha[y - 1:y + 2, x - 1:x + 2]
            if int(np.sum(nbr > 128)) <= min_neighbors:
                arr[y, x, 3] = 0
    return arr


def process(path):
    img = Image.open(path).convert('RGBA')
    arr = np.array(img, dtype=np.uint8)

    arr = flood_remove_background(arr, tol=48)
    arr = flood_remove_background(arr, tol=36)
    arr = refine_alpha_edges(arr)
    arr = despeckle(arr)

    img = Image.fromarray(arr, 'RGBA')
    img = img.filter(ImageFilter.SMOOTH_MORE)

    arr = np.array(img, dtype=np.uint8)
    arr = refine_alpha_edges(arr)
    img = Image.fromarray(arr, 'RGBA')

    bbox = img.getbbox()
    if not bbox:
        print(f'  skip (empty): {path}')
        return False

    w, h = img.size
    left = max(0, bbox[0] - PAD)
    top = max(0, bbox[1] - PAD)
    right = min(w, bbox[2] + PAD)
    bottom = min(h, bbox[3] + PAD)
    img = img.crop((left, top, right, bottom))

    if img.height > MAX_H:
        ratio = MAX_H / img.height
        img = img.resize((max(1, int(img.width * ratio)), MAX_H), Image.LANCZOS)

    img.save(path, 'PNG', optimize=True)
    return True


def main():
    total = 0
    for d in DIRS:
        folder = os.path.join(BASE, d)
        if not os.path.isdir(folder):
            continue
        for name in sorted(os.listdir(folder)):
            if not name.lower().endswith('.png'):
                continue
            p = os.path.join(folder, name)
            if process(p):
                total += 1
                print(f'ok {d}/{name}')
    print(f'done: {total} files')


if __name__ == '__main__':
    main()
