# Rvintech Barcode

Rvintech Barcode adalah library Go untuk membuat QR Code Model 2, Code 39, dan Code 128. Encoder dan renderer dibuat di
proyek ini tanpa dependency library barcode pihak ketiga. Standard library Go dipakai untuk PNG serta membaca logo PNG,
JPEG, dan GIF. Library dapat berjalan tanpa jaringan saat runtime.

Dikelola oleh PT Rvintech Digital Indonesia.

## Format yang didukung

| Format      | Dukungan                                                   |
|-------------|------------------------------------------------------------|
| QR Code     | Segmentasi otomatis mode numerik, alfanumerik, dan byte UTF-8; versi 1–40, L/M/Q/H           |
| Code 39     | Alfabet standar dan checksum modulo-43 opsional            |
| Code 128    | ASCII, perpindahan otomatis Code Set A/B/C, checksum wajib |
| PNG dan SVG | Semua format; logo QR disematkan dalam SVG                 |

QR otomatis memilih dan menggabungkan mode numerik, alfanumerik QR, dan byte UTF-8 agar ukuran encoding lebih efisien. Logo otomatis menggunakan
koreksi H dan versi yang menjaga pola fungsi QR di area tengah.

## Instalasi

```sh
go get github.com/Rvintech-Digital-Indonesia/Rvintech-Barcode
```

## Pemakaian singkat

```go
package main

import (
	"os"

	barcode "github.com/Rvintech-Digital-Indonesia/Rvintech-Barcode"
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

Untuk Code 39 gunakan `barcode.EncodeCode39("RVT-39", barcode.Options{Checksum: true})`. Untuk Code 128 gunakan
`barcode.EncodeCode128("INV-2026-001234", barcode.Options{})`.

## Coba secara mandiri

Dari direktori proyek, jalankan:

```sh
go run ./examples/demo -format qr -text "RVINTECH-123" -out ./barcode-output
```

Perintah tersebut membuat `qr.png` dan `qr.svg`. Ganti `-format` menjadi `code39` atau `code128` untuk mencoba barcode
linear. Jalankan `go run ./examples/demo -help` untuk melihat semua opsi, termasuk warna, bentuk QR, logo, dan level
koreksi.

## Struktur paket

- `barcode.go`: API publik.
- `internal/qr`: encoding QR, Reed–Solomon, penyusunan matriks, dan mask.
- `internal/code39` dan `internal/code128`: encoder barcode linear.
- `internal/symbol`: opsi, matriks simbol, renderer PNG dan SVG.
- `examples/demo/main.go`: aplikasi contoh mandiri.

Dokumentasi lengkap: [Bahasa Indonesia](DOC.id.md) · [English](DOC.md) · [简体中文](DOC.zh.md).
README: [English](README.md) · [简体中文](README.zh.md). Lisensi khusus source-available tetap berbahasa Inggris
di [LICENSE](LICENSE). Lisensi itu bukan lisensi open-source yang disetujui OSI dan perlu tinjauan hukum sebelum rilis.
