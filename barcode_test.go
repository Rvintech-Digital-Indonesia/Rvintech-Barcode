package barcode

import (
	"bytes"
	"encoding/xml"
	"image"
	"image/color"
	"image/png"
	"strings"
	"testing"
)

func TestEncodeFormatsAndExport(t *testing.T) {
	cases := []struct {
		name          string
		encode        func() (*Code, error)
		width, height int
	}{
		{"qr", func() (*Code, error) { return EncodeQR("HELLO WORLD", Options{}) }, 21, 21},
		{"code39", func() (*Code, error) { return EncodeCode39("RVT-39", Options{Checksum: true}) }, 0, 40},
		{"code128", func() (*Code, error) { return EncodeCode128("RVT123456", Options{}) }, 0, 40},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c, err := tc.encode()
			if err != nil {
				t.Fatal(err)
			}
			if c.Height() != tc.height {
				t.Fatalf("height=%d, want %d", c.Height(), tc.height)
			}
			if tc.width != 0 && c.Width() != tc.width {
				t.Fatalf("width=%d, want %d", c.Width(), tc.width)
			}
			if tc.width == 0 && c.Width() < 20 {
				t.Fatalf("linear symbol unexpectedly narrow: %d", c.Width())
			}
			data, err := c.PNGBytes()
			if err != nil {
				t.Fatal(err)
			}
			img, err := png.Decode(bytes.NewReader(data))
			if err != nil {
				t.Fatal(err)
			}
			if img.Bounds().Empty() {
				t.Fatal("PNG has empty bounds")
			}
			svg, err := c.SVGBytes()
			if err != nil {
				t.Fatal(err)
			}
			if !strings.HasPrefix(string(svg), "<svg ") {
				t.Fatal("SVG root element missing")
			}
		})
	}
}
func TestInvalidInputAndContrast(t *testing.T) {
	if _, err := EncodeQR("", Options{}); err == nil {
		t.Fatal("empty QR input accepted")
	}
	if _, err := EncodeCode39("lowercase?", Options{}); err == nil {
		t.Fatal("invalid Code 39 input accepted")
	}
	if _, err := EncodeCode128("café", Options{}); err == nil {
		t.Fatal("non-ASCII Code 128 input accepted")
	}
	if _, err := EncodeQR("ok", Options{Foreground: color.White, Background: color.White}); err == nil {
		t.Fatal("low-contrast colors accepted")
	}
}
func TestLogoAndShapeExport(t *testing.T) {
	logo := image.NewRGBA(image.Rect(0, 0, 10, 6))
	for y := 0; y < 6; y++ {
		for x := 0; x < 10; x++ {
			logo.Set(x, y, color.RGBA{R: 220, A: 255})
		}
	}
	c, err := EncodeQR("logo", Options{Shape: ShapeMixed, Logo: logo})
	if err != nil {
		t.Fatal(err)
	}
	pngBytes, err := c.PNGBytes()
	if err != nil {
		t.Fatal(err)
	}
	if _, err = png.Decode(bytes.NewReader(pngBytes)); err != nil {
		t.Fatal(err)
	}
	svg, err := c.SVGBytes()
	if err != nil {
		t.Fatal(err)
	}
	var document any
	if err := xml.Unmarshal(svg, &document); err != nil {
		t.Fatalf("invalid SVG XML: %v", err)
	}
	if !bytes.Contains(svg, []byte("data:image/png;base64,")) {
		t.Fatal("SVG logo not embedded")
	}
}

func TestConcurrentRendering(t *testing.T) {
	code, err := EncodeQR("parallel render", Options{})
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 8)
	for i := 0; i < cap(done); i++ {
		go func() {
			_, e := code.PNGBytes()
			if e == nil {
				_, e = code.SVGBytes()
			}
			done <- e
		}()
	}
	for i := 0; i < cap(done); i++ {
		if err := <-done; err != nil {
			t.Fatal(err)
		}
	}
}

func TestQRCapacityAndCorrectionLevels(t *testing.T) {
	for _, level := range []ErrorCorrection{ErrorCorrectionL, ErrorCorrectionM, ErrorCorrectionQ, ErrorCorrectionH} {
		code, err := EncodeQR("correction", Options{ErrorCorrection: level})
		if err != nil {
			t.Fatalf("level %s: %v", level, err)
		}
		if code.Width() < 21 {
			t.Fatalf("level %s: invalid width %d", level, code.Width())
		}
	}
	// Lowercase forces Byte Mode, so this checks the version 40-L byte limit
	// instead of the more compact Alphanumeric Mode capacity of uppercase text.
	code, err := EncodeQR(strings.Repeat("a", 2953), Options{})
	if err != nil {
		t.Fatal(err)
	}
	if code.Width() != 177 {
		t.Fatalf("maximum-capacity symbol width=%d, want 177", code.Width())
	}
	if _, err := EncodeQR(strings.Repeat("a", 2954), Options{}); err == nil {
		t.Fatal("over-capacity input accepted")
	}
}

func TestAllQRShapesProduceValidSVG(t *testing.T) {
	for _, shape := range []Shape{ShapeSquare, ShapeCircle, ShapeRounded, ShapeConnected, ShapeBars, ShapeMixed} {
		code, err := EncodeQR("shapes", Options{Shape: shape})
		if err != nil {
			t.Fatal(err)
		}
		svg, err := code.SVGBytes()
		if err != nil {
			t.Fatal(err)
		}
		var document any
		if err := xml.Unmarshal(svg, &document); err != nil {
			t.Fatalf("shape %s: invalid SVG: %v", shape, err)
		}
	}
}

func TestLinearRestrictions(t *testing.T) {
	if _, err := EncodeCode39("A*B", Options{}); err == nil {
		t.Fatal("reserved Code 39 asterisk accepted")
	}
	if _, err := EncodeCode128("ABC", Options{Shape: ShapeCircle}); err == nil {
		t.Fatal("QR shape accepted for Code 128")
	}
	if _, err := EncodeQR("contrast", Options{Foreground: color.White, Background: color.Black}); err == nil {
		t.Fatal("inverted colors accepted")
	}
}

func BenchmarkEncodeQR(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if _, err := EncodeQR("https://example.com/invoice/123456", Options{ErrorCorrection: ErrorCorrectionM}); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEncodeQRPayloadSizes(b *testing.B) {
	for _, tc := range []struct {
		name string
		text string
	}{
		{"short", "hello 123"},
		{"medium", strings.Repeat("invoice-1234567890-", 32)},
		{"maximum", strings.Repeat("a", 2953)},
	} {
		b.Run(tc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if _, err := EncodeQR(tc.text, Options{}); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
