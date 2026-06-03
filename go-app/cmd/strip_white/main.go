package main

import (
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"os"
	"path/filepath"
)

func main() {
	dir := filepath.Join("..", "..", "assets", "op-runners")
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		panic(err)
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		ext := filepath.Ext(e.Name())
		if ext != ".png" && ext != ".jpg" && ext != ".jpeg" {
			continue
		}
		path := filepath.Join(dir, e.Name())
		f, err := os.Open(path)
		if err != nil {
			panic(err)
		}
		img, _, err := image.Decode(f)
		f.Close()
		if err != nil {
			panic(err)
		}
		b := img.Bounds()
		out := image.NewRGBA(b)
		for y := b.Min.Y; y < b.Max.Y; y++ {
			for x := b.Min.X; x < b.Max.X; x++ {
				c := img.At(x, y)
				r, g, bl, a := c.RGBA()
				r8, g8, b8, a8 := uint8(r>>8), uint8(g>>8), uint8(bl>>8), uint8(a>>8)
				if a8 == 0 {
					out.SetRGBA(x, y, color.RGBA{0, 0, 0, 0})
					continue
				}
				if r8 > 235 && g8 > 235 && b8 > 235 {
					out.SetRGBA(x, y, color.RGBA{0, 0, 0, 0})
					continue
				}
				if abs(int(r8)-int(g8)) < 12 && abs(int(g8)-int(b8)) < 12 && r8 > 195 {
					out.SetRGBA(x, y, color.RGBA{0, 0, 0, 0})
					continue
				}
				out.SetRGBA(x, y, color.RGBA{r8, g8, b8, a8})
			}
		}
		outPath := path
		if ext != ".png" {
			outPath = path[:len(path)-len(ext)] + ".png"
			if outPath != path {
				_ = os.Remove(path)
			}
		}
		wf, err := os.Create(outPath)
		if err != nil {
			panic(err)
		}
		if err := png.Encode(wf, out); err != nil {
			panic(err)
		}
		wf.Close()
		println("processed", e.Name(), "->", filepath.Base(outPath))
	}
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
