# Rvintech Barcode

Rvintech Barcode is a dependency-free Go library for generating QR Code Model 2, Code 39, and Code 128 symbols. The
module is maintained by PT Rvintech Digital Indonesia.

The project implements barcode encoding and rendering itself. It uses Go's standard library for PNG output and PNG,
JPEG, and GIF logo decoding. It requires no network access at runtime and has no third-party module dependencies.

## Supported formats

| Format   | Current support                                                      |
|----------|----------------------------------------------------------------------|
| QR Code  | Byte mode, versions 1–40, error-correction levels L/M/Q/H            |
| Code 39  | Standard alphabet, optional modulo-43 checksum                       |
| Code 128 | ASCII, automatic Code Set A/B/C switching, mandatory symbol checksum |
| PNG      | All formats                                                          |
| SVG      | All formats, with embedded PNG logo for QR Codes                     |

The QR encoder supports all 40 Model 2 versions and four error-correction levels. It uses byte mode; numeric and
alphanumeric mode compaction is not implemented. Logo requests select level H and a version whose center avoids QR
function patterns. See [DOC.md](DOC.md) for API details and practical integration examples.

## Install

```sh
go get github.com/pt-rvintech/rvintech-barcode
```

## Quick start

```go
package main

import (
	"os"

	barcode "github.com/pt-rvintech/rvintech-barcode"
)

func main() {
	code, err := barcode.EncodeQR("https://example.com", barcode.Options{})
	if err != nil {
		panic(err)
	}
	file, err := os.Create("qr.png")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	if err := code.PNG(file); err != nil {
		panic(err)
	}
}
```

For Code 39, use `barcode.EncodeCode39("RVT-39", barcode.Options{Checksum: true})`. For Code 128, use
`barcode.EncodeCode128("INV-2026-001234", barcode.Options{})`.

Run the standalone example with `go run ./examples/demo -format qr -text "Rvintech" -out ./barcode-output`. It writes
both PNG and SVG files.

## Package layout

- `barcode.go`: stable public API and aliases.
- `internal/qr`: QR encoding, Reed–Solomon error correction, masking, and matrix construction.
- `internal/code39`, `internal/code128`: linear symbology encoders.
- `internal/symbol`: shared options, symbol matrix, PNG and SVG rendering.
- `examples/demo/main.go`: standalone application that generates PNG and SVG files.

Documentation: [English](DOC.md) · [Bahasa Indonesia](DOC.id.md) · [简体中文](DOC.zh.md).
README: [English](README.md) · [Bahasa Indonesia](README.id.md) · [简体中文](README.zh.md). The custom source-available
license is in [LICENSE](LICENSE); it is not an OSI-approved open-source license and needs legal review before release.
