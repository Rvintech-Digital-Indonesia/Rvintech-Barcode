// Package barcode provides QR Code, Code 39, and Code 128 generation with
// deterministic PNG and SVG renderers. It has no third-party dependencies.
package barcode

import (
	"image"
	"io"

	"github.com/pt-rvintech/rvintech-barcode/internal/code128"
	"github.com/pt-rvintech/rvintech-barcode/internal/code39"
	"github.com/pt-rvintech/rvintech-barcode/internal/qr"
	"github.com/pt-rvintech/rvintech-barcode/internal/symbol"
)

type Shape = symbol.Shape

const (
	ShapeSquare    = symbol.ShapeSquare
	ShapeCircle    = symbol.ShapeCircle
	ShapeRounded   = symbol.ShapeRounded
	ShapeConnected = symbol.ShapeConnected
	ShapeBars      = symbol.ShapeBars
	ShapeMixed     = symbol.ShapeMixed
)

type ErrorCorrection = symbol.ErrorCorrection

const (
	ErrorCorrectionL = symbol.ErrorCorrectionL
	ErrorCorrectionM = symbol.ErrorCorrectionM
	ErrorCorrectionQ = symbol.ErrorCorrectionQ
	ErrorCorrectionH = symbol.ErrorCorrectionH
)

type Options = symbol.Options
type Code = symbol.Code

var ErrEmptyInput = symbol.ErrEmptyInput
var ErrUnsupported = symbol.ErrUnsupported

// EncodeQR encodes text as QR Code Model 2 with configurable error correction.
func EncodeQR(text string, options Options) (*Code, error) { return qr.Encode(text, options) }

// EncodeCode39 encodes the Code 39 alphabet with an optional modulo-43 checksum.
func EncodeCode39(text string, options Options) (*Code, error) { return code39.Encode(text, options) }

// EncodeCode128 encodes ASCII input using automatic Code Set A/B/C selection.
func EncodeCode128(text string, options Options) (*Code, error) { return code128.Encode(text, options) }

// LoadLogo decodes PNG, JPEG, or GIF image data using the Go standard library.
func LoadLogo(r io.Reader) (image.Image, error) { return symbol.LoadLogo(r) }
