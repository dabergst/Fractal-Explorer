// Daniel Bergström
// dabergst@kth.se

package palette

import "image/color"

type Palette []color.RGBA

var (
	OrangeRed    = color.RGBA{0xFF, 0x45, 0x00, 0xFF}
	Orange       = color.RGBA{0xFF, 0xA5, 0x00, 0xFF}
	Gold         = color.RGBA{0xFF, 0xD7, 0x00, 0xFF}
	DarkYellow   = color.RGBA{0xEE, 0xEE, 0x9E, 0xFF}
	DarkGreen    = color.RGBA{0x44, 0x88, 0x44, 0xFF}
	PaleGreyBlue = color.RGBA{0x49, 0x93, 0xDD, 0xFF}
	Purple       = color.RGBA{0xA0, 0x20, 0xF0, 0xFF}
	Cyan         = color.RGBA{0x00, 0xFF, 0xFF, 0xFF}
	Red          = color.RGBA{0xFF, 0x00, 0x00, 0xFF}
	Green        = color.RGBA{0x00, 0xFF, 0x00, 0xFF}
	Blue         = color.RGBA{0x00, 0x00, 0xFF, 0xFF}
	White        = color.RGBA{0xFF, 0xFF, 0xFF, 0xFF}
	Black        = color.RGBA{0x00, 0x00, 0x00, 0xFF}
	MistyRose    = color.RGBA{0xFF, 0xE4, 0xE1, 0xFF}
)

// CyclicPalette generates a cyclic gradient palette of the given colors.
// Each two consequential colors are interpolated.
func CyclicPalette(colors []color.RGBA) Palette {	
	palette := make([]color.RGBA, 256 * len(colors))		
	pIdx := 0
	j := len(colors) - 1
	for i, _ := range colors {
		diff := differens(colors[j], colors[i])
		step := float64(1) / float64(diff)
		f := float64(0)		
		for f <= 1.0 {
			c := interpolateColors(colors[j], colors[i], f)
			palette[pIdx] = c
			f += step
			pIdx++
		}
		j = i
	}
	return palette[0:pIdx]
}

func Mean(colors []color.RGBA) color.RGBA {
	var r, g, b, a int
	for _, color := range colors {
		r += int(color.R)
		g += int(color.G)
		b += int(color.B)
		a += int(color.A)
	}
	r /= len(colors)
	g /= len(colors)
	b /= len(colors)
	a /= len(colors)

	return color.RGBA{uint8(r), uint8(g), uint8(b), uint8(a)}
}

func differens(c1, c2 color.RGBA) uint8 {
	maxDiff := abs(c1.R, c2.R)
	maxDiff = max(maxDiff, abs(c1.G, c2.G))
	maxDiff = max(maxDiff, abs(c1.B, c2.B))
	maxDiff = max(maxDiff, abs(c1.A, c2.A))
	return maxDiff
}

func max(x, y uint8) uint8 {
	if x > y {
		return x
	}
	return y
}

func abs(x, y uint8) uint8 {
	if x > y {
		x, y = y, x
	}
	return y - x
}

// Interpolates colors.
// 0 <= f <= 1
func interpolateColors(c1, c2 color.RGBA, f float64) color.RGBA {
	return color.RGBA{
		uint8(float64(c1.R)*(1.0-f) + float64(c2.R)*f),
		uint8(float64(c1.G)*(1.0-f) + float64(c2.G)*f),
		uint8(float64(c1.B)*(1.0-f) + float64(c2.B)*f),
		uint8(float64(c1.A)*(1.0-f) + float64(c2.A)*f)}
}
