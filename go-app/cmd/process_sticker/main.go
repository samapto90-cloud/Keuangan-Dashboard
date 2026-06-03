package main

import (
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"math"
	"os"
	"path/filepath"
)

func main() {
	outDir := filepath.Join("..", "..", "assets", "naruto-runners")
	if len(os.Args) > 1 {
		outDir = os.Args[1]
	}
	_ = os.MkdirAll(outDir, 0o755)

	type job struct{ src, dst string }
	jobs := []job{}
	if len(os.Args) > 2 {
		for i := 2; i+1 < len(os.Args); i += 2 {
			jobs = append(jobs, job{src: os.Args[i], dst: os.Args[i+1]})
		}
	}

	for _, j := range jobs {
		if err := processFile(j.src, filepath.Join(outDir, j.dst)); err != nil {
			panic(err)
		}
		println("ok", j.dst)
	}
}

func processFile(src, dst string) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	img, _, err := image.Decode(f)
	f.Close()
	if err != nil {
		return err
	}
	b := img.Bounds()
	rgba := image.NewRGBA(b)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			c := img.At(x, y)
			r, g, bl, a := c.RGBA()
			r8, g8, b8, a8 := uint8(r>>8), uint8(g>>8), uint8(bl>>8), uint8(a>>8)
			if a8 == 0 {
				rgba.SetRGBA(x, y, color.RGBA{0, 0, 0, 0})
				continue
			}
			rgba.SetRGBA(x, y, color.RGBA{r8, g8, b8, a8})
		}
	}

	removeEdgeBackground(rgba)
	removeScribbleOverlay(rgba)
	keepLargestComponent(rgba)
	cropped := cropTransparent(rgba, 4)

	wf, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer wf.Close()
	return png.Encode(wf, cropped)
}

// removeEdgeBackground flood-fills grey/white sticker background from image borders.
func removeEdgeBackground(img *image.RGBA) {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w == 0 || h == 0 {
		return
	}
	visited := make([]bool, w*h)
	idx := func(x, y int) int { return (y-b.Min.Y)*w + (x - b.Min.X) }

	queue := [][2]int{}
	add := func(x, y int) {
		if x < b.Min.X || x >= b.Max.X || y < b.Min.Y || y >= b.Max.Y {
			return
		}
		i := idx(x, y)
		if visited[i] {
			return
		}
		r, g, bl, a := rgbaAt(img, x, y)
		if !isRemovableBg(r, g, bl, a) {
			return
		}
		visited[i] = true
		queue = append(queue, [2]int{x, y})
	}

	for x := b.Min.X; x < b.Max.X; x++ {
		add(x, b.Min.Y)
		add(x, b.Max.Y-1)
	}
	for y := b.Min.Y; y < b.Max.Y; y++ {
		add(b.Min.X, y)
		add(b.Max.X-1, y)
	}

	for len(queue) > 0 {
		p := queue[0]
		queue = queue[1:]
		img.SetRGBA(p[0], p[1], color.RGBA{0, 0, 0, 0})
		add(p[0]+1, p[1])
		add(p[0]-1, p[1])
		add(p[0], p[1]+1)
		add(p[0], p[1]-1)
	}
}

func isRemovableBg(r, g, b, a uint8) bool {
	if a < 20 {
		return true
	}
	if r > 232 && g > 232 && b > 232 {
		return true
	}
	drg := abs(int(r) - int(g))
	dgb := abs(int(g) - int(b))
	drb := abs(int(r) - int(b))
	if drg <= 18 && dgb <= 18 && drb <= 18 {
		avg := (int(r) + int(g) + int(b)) / 3
		if avg >= 95 && avg <= 215 {
			return true
		}
	}
	return false
}

func removeScribbleOverlay(img *image.RGBA) {
	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, bl, a := rgbaAt(img, x, y)
			if a == 0 {
				continue
			}
			if a < 180 {
				img.SetRGBA(x, y, color.RGBA{0, 0, 0, 0})
				continue
			}
			drg := abs(int(r) - int(g))
			dgb := abs(int(g) - int(bl))
			drb := abs(int(r) - int(bl))
			if drg <= 10 && dgb <= 10 && drb <= 10 {
				avg := (int(r) + int(g) + int(bl)) / 3
				if avg >= 150 && avg <= 210 && a < 240 {
					img.SetRGBA(x, y, color.RGBA{0, 0, 0, 0})
				}
			}
		}
	}
}

func keepLargestComponent(img *image.RGBA) {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w == 0 || h == 0 {
		return
	}
	idx := func(x, y int) int { return (y-b.Min.Y)*w + (x - b.Min.X) }
	labels := make([]int, w*h)
	labelSizes := map[int]int{}
	nextLabel := 1

	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			i := idx(x, y)
			if labels[i] != 0 {
				continue
			}
			_, _, _, a := rgbaAt(img, x, y)
			if a < 30 {
				continue
			}
			label := nextLabel
			nextLabel++
			size := 0
			stack := [][2]int{{x, y}}
			for len(stack) > 0 {
				p := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				pi := idx(p[0], p[1])
				if labels[pi] != 0 {
					continue
				}
				_, _, _, pa := rgbaAt(img, p[0], p[1])
				if pa < 30 {
					continue
				}
				labels[pi] = label
				size++
				for _, n := range [][2]int{{p[0] + 1, p[1]}, {p[0] - 1, p[1]}, {p[0], p[1] + 1}, {p[0], p[1] - 1}} {
					nx, ny := n[0], n[1]
					if nx < b.Min.X || nx >= b.Max.X || ny < b.Min.Y || ny >= b.Max.Y {
						continue
					}
					stack = append(stack, [2]int{nx, ny})
				}
			}
			labelSizes[label] = size
		}
	}

	bestLabel, bestSize := 0, 0
	for l, s := range labelSizes {
		if s > bestSize {
			bestSize = s
			bestLabel = l
		}
	}
	if bestLabel == 0 {
		return
	}
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			i := idx(x, y)
			if labels[i] != bestLabel {
				img.SetRGBA(x, y, color.RGBA{0, 0, 0, 0})
			}
		}
	}
}

func rgbaAt(img *image.RGBA, x, y int) (uint8, uint8, uint8, uint8) {
	c := img.RGBAAt(x, y)
	return c.R, c.G, c.B, c.A
}

func cropTransparent(img *image.RGBA, pad int) image.Image {
	b := img.Bounds()
	minX, minY, maxX, maxY := b.Max.X, b.Max.Y, b.Min.X, b.Min.Y
	found := false
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if img.RGBAAt(x, y).A > 20 {
				found = true
				if x < minX {
					minX = x
				}
				if y < minY {
					minY = y
				}
				if x > maxX {
					maxX = x
				}
				if y > maxY {
					maxY = y
				}
			}
		}
	}
	if !found {
		return img
	}
	minX = intMax(b.Min.X, minX-pad)
	minY = intMax(b.Min.Y, minY-pad)
	maxX = intMin(b.Max.X-1, maxX+pad)
	maxY = intMin(b.Max.Y-1, maxY+pad)
	rect := image.Rect(0, 0, maxX-minX+1, maxY-minY+1)
	out := image.NewRGBA(rect)
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			out.Set(x-minX, y-minY, img.At(x, y))
		}
	}
	return out
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func intMax(a, b int) int {
	return int(math.Max(float64(a), float64(b)))
}

func intMin(a, b int) int {
	return int(math.Min(float64(a), float64(b)))
}
