# Dokumentasi Pengembang Rvintech Barcode

## Ringkasan

Paket Go di root bernama `barcode`. Setiap encoder mengembalikan `*barcode.Code` yang dapat dirender berulang kali dan aman digunakan secara bersamaan, selama gambar logo yang diberikan tidak diubah saat rendering. Input serta opsi yang sama menghasilkan simbol yang deterministik.

## Instalasi

```sh
go get github.com/pt-rvintech/rvintech-barcode
```

Modul menargetkan Go 1.23 atau yang lebih baru dan tidak memiliki dependency modul eksternal.

## API publik

```go
func EncodeQR(text string, options Options) (*Code, error)
func EncodeCode39(text string, options Options) (*Code, error)
func EncodeCode128(text string, options Options) (*Code, error)
func LoadLogo(r io.Reader) (image.Image, error)
func (c *Code) Width() int
func (c *Code) Height() int
func (c *Code) PNG(w io.Writer) error
func (c *Code) PNGBytes() ([]byte, error)
func (c *Code) SVG(w io.Writer) error
func (c *Code) SVGBytes() ([]byte, error)
```

`Width` dan `Height` menyatakan jumlah modul tanpa quiet zone. Encoder mengembalikan error jika input kosong atau tidak valid, opsi tidak didukung, kontras rendah, atau data QR melampaui kapasitas.

## Opsi

```go
type Options struct {
    Scale int
    QuietZone int
    Foreground color.Color
    Background color.Color
    Shape Shape
    ErrorCorrection ErrorCorrection
    Checksum bool
    Logo image.Image
}
```

Nilai nol memilih skala 8 piksel per modul, quiet zone 4 modul untuk QR atau 10 modul untuk barcode linear, modul hitam, latar putih, bentuk kotak, dan koreksi QR level L. Quiet zone minimal adalah 4 modul untuk QR serta 10 modul untuk Code 39 dan Code 128. Skala maksimal 128 piksel per modul, quiet zone maksimal 64 modul, dan PNG maksimal 32 juta piksel. Panjang input barcode linear maksimal 512 byte.

Warna harus sepenuhnya opak. Foreground harus lebih gelap daripada background dengan selisih luminans linear minimal 0,55. Pemeriksaan ini membantu menjaga keterbacaan, tetapi kondisi pemindaian tetap bergantung pada ukuran dan media keluaran.

`Checksum` menambahkan karakter pemeriksa modulo-43 pada Code 39. Code 128 selalu menyertakan checksum modulo-103. `Shape` dapat berupa `ShapeSquare`, `ShapeCircle`, `ShapeRounded`, `ShapeConnected`, `ShapeBars`, atau `ShapeMixed`. Bentuk dekoratif hanya berlaku untuk QR; pola fungsi QR tetap kotak. Barcode linear hanya menerima `ShapeSquare`.

## Membuat QR Code

`EncodeQR` membuat simbol QR Code Model 2 dalam byte mode, versi 1–40, dengan level koreksi L, M, Q, atau H. Versi terkecil yang memuat data dipilih otomatis. Teks diubah menjadi byte UTF-8. Kapasitas byte mode terbesar adalah 2.953 byte pada level L. Pemadatan mode numerik dan alfanumerik belum tersedia.

```go
code, err := barcode.EncodeQR("Halo, QR", barcode.Options{
    ErrorCorrection: barcode.ErrorCorrectionM,
    Shape: barcode.ShapeRounded,
})
```

Encoder mengevaluasi delapan mask serta membangun informasi format dan kode koreksi Reed–Solomon di dalam proyek ini.

### Menambahkan logo

Isi `Options.Logo` dengan `image.Image`. Gunakan `LoadLogo` untuk membaca PNG, JPEG, atau GIF. Logo diskalakan sambil mempertahankan rasio aspek dan ditempatkan di tengah dengan ukuran maksimum 6% sisi pendek gambar. Permintaan logo otomatis memakai koreksi H dan memilih versi yang tidak menempatkan logo di atas finder, timing, alignment, format, maupun version pattern.

```go
file, err := os.Open("logo.png")
if err != nil { return err }
defer file.Close()
logo, err := barcode.LoadLogo(file)
if err != nil { return err }
code, err := barcode.EncodeQR("RVINTECH", barcode.Options{Logo: logo})
if err != nil { return err }
```

## Membuat barcode linear

### Code 39

`EncodeCode39` menerima alfabet Code 39 standar dan mengubah huruf kecil menjadi huruf besar. Tanda `*` disediakan untuk start/stop dan tidak boleh menjadi data. Aktifkan checksum opsional dengan `Options{Checksum: true}`.

### Code 128

`EncodeCode128` menerima byte ASCII. Code Set C dipilih untuk rangkaian angka, sedangkan A/B dipakai sesuai karakter. Checksum simbol modulo-103 selalu ditambahkan. FNC dan mode Code 128 lanjutan belum tersedia melalui API ini.

## PNG, SVG, dan integrasi

`PNG` menulis ke `io.Writer` memakai `image/png` dari standard library Go. `SVG` menghasilkan dokumen mandiri; logo QR disematkan sebagai data URI PNG. Keduanya menyertakan quiet zone. Gunakan `PNGBytes` atau `SVGBytes` untuk handler HTTP dan penyimpanan dokumen.

```go
svgBytes, err := code.SVGBytes()
if err != nil { return err }
if err := os.WriteFile("code.svg", svgBytes, 0644); err != nil { return err }
```

## Contoh yang dapat dijalankan

```sh
go run ./examples/demo -format qr -text "RVINTECH-123" -out ./barcode-output
```

Contoh tersebut membuat PNG dan SVG sekaligus. Gunakan `-format code39` atau `-format code128` untuk format lain. Opsi lain: `-shape`, `-ecc`, `-fg`, `-bg`, `-logo`, `-scale`, `-quiet`, dan `-checksum`. Lihat `go run ./examples/demo -help` untuk penjelasan tiap opsi.
