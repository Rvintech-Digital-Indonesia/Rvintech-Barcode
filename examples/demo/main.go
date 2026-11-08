// Command demo generates a PNG and an SVG with the Rvintech Barcode library.
package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"strings"

	barcode "github.com/Rvintech-Digital-Indonesia/Rvintech-Barcode"
)

func main() {
	format := flag.String("format", "qr", "Barcode format: qr, code39, or code128")
	value := flag.String("text", "RVINTECH-123", "Text to encode")
	output := flag.String("out", "barcode-output", "Output directory")
	shape := flag.String("shape", "square", "QR shape: square, circle, rounded, connected, bars, or mixed")
	ecc := flag.String("ecc", "L", "QR error correction: L, M, Q, or H")
	foreground := flag.String("fg", "#000000", "Foreground color as #RRGGBB")
	background := flag.String("bg", "#FFFFFF", "Background color as #RRGGBB")
	logoPath := flag.String("logo", "", "Optional PNG, JPEG, or GIF logo path for QR")
	scale := flag.Int("scale", 8, "Pixels per module in PNG")
	quiet := flag.Int("quiet", 0, "Quiet zone in modules; zero selects the format default")
	checksum := flag.Bool("checksum", false, "Add the optional Code 39 checksum")
	flag.Parse()
	fg, err := parseColor(*foreground)
	if err != nil {
		exit(err)
	}
	bg, err := parseColor(*background)
	if err != nil {
		exit(err)
	}
	options := barcode.Options{Scale: *scale, QuietZone: *quiet, Foreground: fg, Background: bg, Shape: barcode.Shape(*shape), ErrorCorrection: barcode.ErrorCorrection(strings.ToUpper(*ecc)), Checksum: *checksum}
	if *logoPath != "" {
		file, err := os.Open(*logoPath)
		if err != nil {
			exit(err)
		}
		var logo image.Image
		logo, err = barcode.LoadLogo(file)
		file.Close()
		if err != nil {
			exit(err)
		}
		options.Logo = logo
	}
	var code *barcode.Code
	switch strings.ToLower(*format) {
	case "qr":
		code, err = barcode.EncodeQR(*value, options)
	case "code39":
		code, err = barcode.EncodeCode39(*value, options)
	case "code128":
		code, err = barcode.EncodeCode128(*value, options)
	default:
		exit(fmt.Errorf("unknown format %q", *format))
	}
	if err != nil {
		exit(err)
	}
	if err := os.MkdirAll(*output, 0755); err != nil {
		exit(err)
	}
	base := strings.ToLower(*format)
	pngPath := filepath.Join(*output, base+".png")
	svgPath := filepath.Join(*output, base+".svg")
	file, err := os.Create(pngPath)
	if err != nil {
		exit(err)
	}
	renderErr := code.PNG(file)
	closeErr := file.Close()
	if renderErr != nil {
		exit(renderErr)
	}
	if closeErr != nil {
		exit(closeErr)
	}
	svg, err := code.SVGBytes()
	if err != nil {
		exit(err)
	}
	if err := os.WriteFile(svgPath, svg, 0644); err != nil {
		exit(err)
	}
	fmt.Printf("Created %s and %s (%d x %d modules)\n", pngPath, svgPath, code.Width(), code.Height())
}
func parseColor(s string) (color.RGBA, error) {
	var r, g, b uint8
	if len(s) != 7 || s[0] != '#' {
		return color.RGBA{}, fmt.Errorf("invalid color %q: expected #RRGGBB", s)
	}
	if _, err := fmt.Sscanf(s, "#%02x%02x%02x", &r, &g, &b); err != nil {
		return color.RGBA{}, fmt.Errorf("invalid color %q: %w", s, err)
	}
	return color.RGBA{R: r, G: g, B: b, A: 255}, nil
}
func exit(err error) { fmt.Fprintln(os.Stderr, "demo:", err); os.Exit(1) }
