// Daniel Bergström
// dabergst@kth.se

package fractal

import (
	"image"
	"image/color"
	"math"
	"saph/graphic"
	"saph/graphic/palette"
	"sync"
)

type Fractal struct {
	xMin, yMin    float64
	xMax, yMax    float64
	isMandelbrot  bool
	juliaConstant complex128
	progress
	sync.Mutex
}

func NewMandelbrot() *Fractal {
	fr := new(Fractal)
	fr.xMin, fr.yMin = -2.5, -1.5
	fr.xMax, fr.yMax = 1.0, 1.5
	fr.isMandelbrot = true
	fr.juliaConstant = complex(0, 0)
	fr.isFinished = true
	return fr
}

func NewJulia(c complex128) *Fractal {
	fr := new(Fractal)
	fr.xMin, fr.yMin = -1.7, -1.0
	fr.xMax, fr.yMax = 1.7, 1.0
	fr.isMandelbrot = false
	fr.juliaConstant = c
	fr.isFinished = true
	return fr
}

func (fr *Fractal) IsMandelbrot() bool        { return fr.isMandelbrot }
func (fr *Fractal) JuliaConstant() complex128 { return fr.juliaConstant }

func (fr *Fractal) Magnify(imageSize, magnifySize graphic.Box, magnifyPoint image.Point) {
	fr.Lock()
	defer fr.Unlock()

	colScaler := fr.colScalerGenerator(imageSize.Width)
	rowScaler := fr.rowScalerGenerator(imageSize.Height)
	fr.xMin = colScaler(magnifyPoint.X - magnifySize.Width/2)
	fr.yMin = rowScaler(magnifyPoint.Y - magnifySize.Height/2)
	fr.xMax = colScaler(magnifyPoint.X + magnifySize.Width/2)
	fr.yMax = rowScaler(magnifyPoint.Y + magnifySize.Height/2)
}

func (fr *Fractal) DeMagnify(ratio float64) {
	fr.Lock()
	defer fr.Unlock()

	fr.xMin *= ratio
	fr.yMin *= ratio
	fr.xMax *= ratio
	fr.yMax *= ratio
}

func (fr *Fractal) Render(imageSize graphic.Box, maxIterations int, bailoutRadius float64, normalize bool, sampleRatio int, setColor color.RGBA, palette palette.Palette, colorFrequency float64) chan graphic.Pixel {
	fr.Lock()
	pixChan := make(chan graphic.Pixel, imageSize.Height)
	fr.newRequest(imageSize.Height * imageSize.Width)
	go func() {
		defer close(pixChan)
		defer fr.Unlock()

		rs := &renderSettings{imageSize, maxIterations, bailoutRadius, normalize, sampleRatio, setColor, palette, colorFrequency}		

		if rs.sampleRatio > 1 {
			fr.renderOverSampled(rs, pixChan)
		} else {
			fr.renderStandardSampled(rs, pixChan)
		}
	}()
	return pixChan
}

func (fr *Fractal) renderStandardSampled(rs *renderSettings, pixChan chan graphic.Pixel) {
	colScaler := fr.colScalerGenerator(rs.Width)
	rowScaler := fr.rowScalerGenerator(rs.Height)
	wg := new(sync.WaitGroup)
	for row := 0; row < rs.Height; row++ {
		wg.Add(1)
		go func(row int) {
			for col := 0; col < rs.Width; col++ {
				x, y := colScaler(col), rowScaler(row)
				color := fr.renderPoint(complex(x, y), rs)
				pixChan <- graphic.Pix(color, col, row)
				fr.elementFinished()
			}
			wg.Done()
		}(row)
	}
	wg.Wait()
}

func (fr *Fractal) renderOverSampled(rs *renderSettings, pixChan chan graphic.Pixel) {
	colScaler := fr.colScalerGenerator(rs.Width)
	rowScaler := fr.rowScalerGenerator(rs.Height)
	pixelWidth := (fr.xMax - fr.xMin) / float64(rs.Width)
	pixelHeight := (fr.yMax - fr.yMin) / float64(rs.Height)
	wg := new(sync.WaitGroup)
	for row := 0; row < rs.Height; row++ {
		wg.Add(1)
		go func(row int) {
			for col := 0; col < rs.Width; col++ {
				x, y := colScaler(col), rowScaler(row)
				color := fr.renderPixel(complex(x, y), pixelWidth, pixelHeight, rs)
				pixChan <- graphic.Pix(color, col, row)
				fr.elementFinished()
			}
			wg.Done()
		}(row)
	}
	wg.Wait()
}

func (fr *Fractal) renderPixel(point complex128, pixelWidth, pixelHeight float64, rs *renderSettings) color.RGBA {
	colors := make([]color.RGBA, rs.sampleRatio*rs.sampleRatio)
	for i := 0; i < rs.sampleRatio; i++ {
		x := real(point) + pixelWidth*float64(i)/float64(rs.sampleRatio) - pixelWidth/2
		for j := 0; j < rs.sampleRatio; j++ {
			y := imag(point) + pixelHeight*float64(j)/float64(rs.sampleRatio) - pixelHeight/2
			colors[i*rs.sampleRatio+j] = fr.renderPoint(complex(x, y), rs)
		}
	}

	return palette.Mean(colors)
}

func (fr *Fractal) renderPoint(point complex128, rs *renderSettings) color.RGBA {
	var z, c complex128
	if fr.isMandelbrot {
		z, c = complex(0, 0), point
	} else {
		z, c = point, fr.juliaConstant
	}
	reSqr := real(z) * real(z)
	imSqr := imag(z) * imag(z)
	var n uint64
	for n = 0; n < uint64(rs.maxIterations) && reSqr+imSqr < rs.bailoutRadius; n++ {
		z = complex(reSqr-imSqr, 2*real(z)*imag(z)) + c
		reSqr = real(z) * real(z)
		imSqr = imag(z) * imag(z)
	}
	if n == uint64(rs.maxIterations) {
		return rs.setColor
	}

	if rs.normalize {
		v := float64(n)
		zAbs := math.Sqrt(reSqr + imSqr)
		corr := math.Log10(math.Log10(zAbs)) / math.Log10(2.0)
		v -= corr
		v *= rs.colorFrequency
		n = uint64(v)
	} else {
		n *= uint64(rs.colorFrequency)
	}

	return rs.palette[n%uint64(len(rs.palette))]
}

func (fr *Fractal) colScalerGenerator(width int) func(col int) float64 {
	xRange := fr.xMax - fr.xMin
	xScale := xRange / float64(width)
	xOffset := fr.xMin

	return func(col int) float64 {
		x := float64(col)
		x *= xScale
		x += xOffset
		return x
	}
}

func (fr *Fractal) rowScalerGenerator(height int) func(row int) float64 {
	yRange := fr.yMax - fr.yMin
	yScale := yRange / float64(height)
	yOffset := fr.yMin

	return func(row int) float64 {
		y := float64(row)
		y *= yScale
		y += yOffset
		return y
	}
}

type renderSettings struct {
	graphic.Box
	maxIterations  int
	bailoutRadius  float64
	normalize      bool
	sampleRatio    int
	setColor       color.RGBA
	palette        palette.Palette
	colorFrequency float64
}
