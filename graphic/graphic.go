// Daniel Bergström
// dabergst@kth.se

package graphic

import (
	"image"
	"image/color"
)

type Pixel struct {
	color.RGBA
	image.Point
}

func Pix(color color.RGBA, x, y int) Pixel {
	return Pixel{color, image.Point{x, y}}
}

type Box struct {
	Width, Height int
}
