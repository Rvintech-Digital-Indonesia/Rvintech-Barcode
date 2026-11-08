// Package barcode generates QR Code, Code 39, and Code 128 symbols without
// third-party barcode dependencies.
package symbol

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"io"
	"math"
)

// Shape selects how QR modules are drawn. The matrix data is identical for
// every shape; only the raster/vector rendering changes.
type Shape string

const (
	ShapeSquare    Shape = "square"
	ShapeCircle    Shape = "circle"
	ShapeRounded   Shape = "rounded"
	ShapeConnected Shape = "connected"
	ShapeBars      Shape = "bars"
	ShapeMixed     Shape = "mixed"
)

// Options controls rendering. Zero values select black on white, square modules,
// and a four-module quiet zone. Scale is the pixel width of one module in PNG.
type ErrorCorrection string

const (
	ErrorCorrectionL ErrorCorrection = "L"
	ErrorCorrectionM ErrorCorrection = "M"
	ErrorCorrectionQ ErrorCorrection = "Q"
	ErrorCorrectionH ErrorCorrection = "H"
)

type Options struct {
	ErrorCorrection ErrorCorrection
	Scale           int
	QuietZone       int
	Foreground      color.Color
	Background      color.Color
	Shape           Shape
	Checksum        bool
	Logo            image.Image
}

// Code is an immutable symbol matrix safe for concurrent rendering.
type Code struct {
	modules   [][]bool
	protected [][]bool
	width     int
	height    int
	options   Options
}

var (
	ErrEmptyInput  = errors.New("barcode: input must not be empty")
	ErrUnsupported = errors.New("barcode: input or option is not supported")
)

func NormalizeOptions(o Options) (Options, error) {
	if o.Scale <= 0 {
		o.Scale = 8
	}
	if o.QuietZone < 0 {
		o.QuietZone = 0
	} else if o.QuietZone == 0 {
		o.QuietZone = 4
	}
	if o.Foreground == nil {
		o.Foreground = color.Black
	}
	if o.Background == nil {
		o.Background = color.White
	}
	if o.ErrorCorrection == "" {
		o.ErrorCorrection = ErrorCorrectionL
	}
	if o.Shape == "" {
		o.Shape = ShapeSquare
	}
	return o, ValidateOptions(o)
}
func ValidateOptions(o Options) error {
	switch o.ErrorCorrection {
	case ErrorCorrectionL, ErrorCorrectionM, ErrorCorrectionQ, ErrorCorrectionH:
	default:
		return fmt.Errorf("barcode: invalid error correction level %q", o.ErrorCorrection)
	}
	if o.Scale > 128 {
		return errors.New("barcode: scale must not exceed 128 pixels per module")
	}
	if o.QuietZone > 64 {
		return errors.New("barcode: quiet zone must not exceed 64 modules")
	}
	switch o.Shape {
	case ShapeSquare, ShapeCircle, ShapeRounded, ShapeConnected, ShapeBars, ShapeMixed:
	default:
		return fmt.Errorf("barcode: invalid shape %q", o.Shape)
	}
	_, _, _, fa := o.Foreground.RGBA()
	_, _, _, ba := o.Background.RGBA()
	if fa != 0xFFFF || ba != 0xFFFF {
		return errors.New("barcode: foreground and background colors must be fully opaque")
	}
	fg, bg := luminance(o.Foreground), luminance(o.Background)
	if bg-fg < .55 {
		return errors.New("barcode: foreground must be darker than background by at least 0.55 in linear luminance")
	}
	return nil
}
func luminance(c color.Color) float64 {
	r, g, b, _ := c.RGBA()
	f := func(v uint32) float64 {
		x := float64(v) / 65535
		if x <= .04045 {
			return x / 12.92
		}
		return math.Pow((x+.055)/1.055, 2.4)
	}
	return .2126*f(r) + .7152*f(g) + .0722*f(b)
}

// New creates an immutable symbol from a rectangular module matrix.
func New(modules [][]bool, options Options) *Code {
	h := len(modules)
	w := 0
	if h > 0 {
		w = len(modules[0])
	}
	copied := make([][]bool, h)
	for y := range modules {
		copied[y] = append([]bool(nil), modules[y]...)
	}
	return &Code{modules: copied, width: w, height: h, options: options}
}

// NewProtected stores function-module locations so QR rendering keeps finder,
// timing, alignment, and information patterns square.
func NewProtected(modules, protected [][]bool, options Options) *Code {
	c := New(modules, options)
	c.protected = make([][]bool, len(protected))
	for y := range protected {
		c.protected[y] = append([]bool(nil), protected[y]...)
	}
	return c
}

// Width returns the number of data modules, excluding quiet zone.
func (c *Code) Width() int { return c.width }

// Height returns the number of data modules, excluding quiet zone.
func (c *Code) Height() int { return c.height }

// PNG writes a deterministic PNG image to w.
func (c *Code) PNG(w io.Writer) error {
	if c == nil {
		return errors.New("barcode: nil code")
	}
	o := c.options
	size := (c.width + 2*o.QuietZone) * o.Scale
	h := (c.height + 2*o.QuietZone) * o.Scale
	if int64(size)*int64(h) > 32_000_000 {
		return errors.New("barcode: PNG dimensions exceed 32 million pixels")
	}
	img := image.NewRGBA(image.Rect(0, 0, size, h))
	draw.Draw(img, img.Bounds(), &image.Uniform{C: o.Background}, image.Point{}, draw.Src)
	for y, row := range c.modules {
		for x, on := range row {
			if on {
				if len(c.protected) > 0 && c.protected[y][x] {
					standard := o
					standard.Shape = ShapeSquare
					drawModule(img, x, y, standard)
				} else {
					drawModule(img, x, y, o)
				}
			}
		}
	}
	if o.Logo != nil {
		if err := drawLogo(img, o.Logo, o); err != nil {
			return err
		}
	}
	return png.Encode(w, img)
}

// PNGBytes returns the PNG representation of the code.
func (c *Code) PNGBytes() ([]byte, error) {
	var b bytes.Buffer
	err := c.PNG(&b)
	return b.Bytes(), err
}

func drawModule(img *image.RGBA, x, y int, o Options) {
	s := o.Scale
	ox := (x + o.QuietZone) * s
	oy := (y + o.QuietZone) * s
	shape := o.Shape
	if shape == ShapeMixed && (x+y)%2 == 0 {
		shape = ShapeCircle
	}
	switch shape {
	case ShapeCircle:
		r := float64(s) * .48
		cx, cy := float64(ox)+float64(s)/2, float64(oy)+float64(s)/2
		for yy := oy; yy < oy+s; yy++ {
			for xx := ox; xx < ox+s; xx++ {
				dx := float64(xx) + .5 - cx
				dy := float64(yy) + .5 - cy
				if dx*dx+dy*dy <= r*r {
					img.Set(xx, yy, o.Foreground)
				}
			}
		}
	case ShapeRounded, ShapeMixed, ShapeConnected:
		rad := s / 4
		if s < 4 {
			rad = 0
		}
		for yy := oy; yy < oy+s; yy++ {
			for xx := ox; xx < ox+s; xx++ {
				dx := minInt(xx-ox, s-1-(xx-ox))
				dy := minInt(yy-oy, s-1-(yy-oy))
				if dx >= rad || dy >= rad {
					img.Set(xx, yy, o.Foreground)
				} else {
					a := float64(rad - dx)
					b := float64(rad - dy)
					if a*a+b*b <= float64(rad*rad) {
						img.Set(xx, yy, o.Foreground)
					}
				}
			}
		}
	case ShapeBars:
		margin := s / 6
		for yy := oy; yy < oy+s; yy++ {
			for xx := ox + margin; xx < ox+s-margin; xx++ {
				img.Set(xx, yy, o.Foreground)
			}
		}
	default:
		for yy := oy; yy < oy+s; yy++ {
			for xx := ox; xx < ox+s; xx++ {
				img.Set(xx, yy, o.Foreground)
			}
		}
	}
}
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func drawLogo(dst *image.RGBA, logo image.Image, o Options) error {
	// Keep the logo small so it obscures as few data modules as possible.
	side := minInt(dst.Bounds().Dx(), dst.Bounds().Dy()) * 6 / 100
	if side < 1 {
		side = 1
	}
	if side < 1 {
		return errors.New("barcode: image canvas is too small for logo")
	}
	src := logo.Bounds()
	if src.Empty() {
		return errors.New("barcode: empty logo image")
	}
	scale := math.Min(float64(side)/float64(src.Dx()), float64(side)/float64(src.Dy()))
	w, h := maxInt(1, int(float64(src.Dx())*scale)), maxInt(1, int(float64(src.Dy())*scale))
	x, y := (dst.Bounds().Dx()-w)/2, (dst.Bounds().Dy()-h)/2
	// Nearest-neighbor scaling is deterministic and keeps the package dependency-free.
	for yy := 0; yy < h; yy++ {
		for xx := 0; xx < w; xx++ {
			sx := src.Bounds().Min.X + xx*src.Bounds().Dx()/w
			sy := src.Bounds().Min.Y + yy*src.Bounds().Dy()/h
			dst.Set(x+xx, y+yy, logo.At(sx, sy))
		}
	}
	return nil
}
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// SVG writes a self-contained SVG representation of the symbol.
func (c *Code) SVG(w io.Writer) error {
	if c == nil {
		return errors.New("barcode: nil code")
	}
	o := c.options
	width, height := c.width+2*o.QuietZone, c.height+2*o.QuietZone
	fg, bg := cssColor(o.Foreground), cssColor(o.Background)
	if _, err := fmt.Fprintf(w, "<svg xmlns=\"http://www.w3.org/2000/svg\" viewBox=\"0 0 %d %d\" shape-rendering=\"crispEdges\"><rect width=\"100%%\" height=\"100%%\" fill=\"%s\"/><g fill=\"%s\">", width, height, bg, fg); err != nil {
		return err
	}
	// Each dark module is emitted as a vector primitive for sharp scaling.
	for y, row := range c.modules {
		for x, on := range row {
			if !on {
				continue
			}
			px, py := x+o.QuietZone, y+o.QuietZone
			shape := o.Shape
			if len(c.protected) > 0 && c.protected[y][x] {
				shape = ShapeSquare
			}
			if shape == ShapeMixed && (x+y)%2 == 0 {
				shape = ShapeCircle
			}
			var err error
			switch shape {
			case ShapeCircle:
				_, err = fmt.Fprintf(w, "<circle cx=\"%d.5\" cy=\"%d.5\" r=\"0.46\"/>", px, py)
			case ShapeRounded, ShapeConnected:
				_, err = fmt.Fprintf(w, "<rect x=\"%d\" y=\"%d\" width=\"1\" height=\"1\" rx=\"0.22\"/>", px, py)
			case ShapeBars:
				_, err = fmt.Fprintf(w, "<rect x=\"%.3f\" y=\"%d\" width=\"0.667\" height=\"1\"/>", float64(px)+1.0/6, py)
			default:
				_, err = fmt.Fprintf(w, "<rect x=\"%d\" y=\"%d\" width=\"1\" height=\"1\"/>", px, py)
			}
			if err != nil {
				return err
			}
		}
	}
	if _, err := io.WriteString(w, "</g>"); err != nil {
		return err
	}
	if o.Logo != nil {
		var b bytes.Buffer
		if err := png.Encode(&b, o.Logo); err != nil {
			return err
		}
		box := math.Max(1, float64(minInt(width, height))*.06)
		src := o.Logo.Bounds()
		scale := math.Min(box/float64(src.Dx()), box/float64(src.Dy()))
		lw, lh := float64(src.Dx())*scale, float64(src.Dy())*scale
		x, y := (float64(width)-lw)/2, (float64(height)-lh)/2
		if _, err := fmt.Fprintf(w, "<image x=\"%.3f\" y=\"%.3f\" width=\"%.3f\" height=\"%.3f\" href=\"data:image/png;base64,%s\"/>", x, y, lw, lh, encodeBase64(b.Bytes())); err != nil {
			return err
		}
	}
	_, err := io.WriteString(w, "</svg>")
	return err
}
func (c *Code) SVGBytes() ([]byte, error) {
	var b bytes.Buffer
	err := c.SVG(&b)
	return b.Bytes(), err
}
func cssColor(c color.Color) string {
	r, g, b, _ := c.RGBA()
	return fmt.Sprintf("#%02x%02x%02x", r>>8, g>>8, b>>8)
}
func encodeBase64(src []byte) string { return base64.StdEncoding.EncodeToString(src) }

// LoadLogo decodes a PNG, JPEG, or GIF image using registered standard-library codecs.
func LoadLogo(r io.Reader) (image.Image, error) {
	img, _, err := image.Decode(r)
	if err != nil {
		return nil, fmt.Errorf("barcode: decode logo: %w", err)
	}
	return img, nil
}
