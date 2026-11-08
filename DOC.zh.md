# Rvintech Barcode 开发文档

## 概述

Go 根包名为 `barcode`。编码函数返回 `*barcode.Code`，可重复渲染并支持并发使用。输入与选项相同时，编码和渲染结果具有确定性。渲染期间请勿修改传入 `Options.Logo` 的图片。

## 安装

```sh
go get github.com/pt-rvintech/rvintech-barcode
```

模块面向 Go 1.23 或更高版本，不依赖第三方 Go 模块。

## API

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

`Width` 和 `Height` 返回不含静区的模块数。输入为空或无效、选项不受支持、颜色对比不足或 QR 数据超出容量时，编码函数会返回错误。

## 选项

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

零值默认使用每模块 8 像素、QR 静区 4 个模块（线性条码为 10 个模块）、黑色前景、白色背景、方形模块和 L 级 QR 纠错。QR 静区不得少于 4 个模块；Code 39 与 Code 128 静区不得少于 10 个模块。每模块最大缩放为 128 像素，静区最大为 64 个模块，PNG 最大为 3200 万像素。线性条码输入最多 512 字节。

前景和背景颜色必须完全不透明。前景亮度必须比背景低至少 0.55（线性亮度）。该校验有助于保持可读性，但无法代替实际扫描测试。

`Checksum` 为 Code 39 添加可选的 modulo-43 校验字符。Code 128 始终添加必需的 modulo-103 校验字符。`Shape` 支持 `ShapeSquare`、`ShapeCircle`、`ShapeRounded`、`ShapeConnected`、`ShapeBars` 和 `ShapeMixed`。装饰形状只适用于 QR 数据模块，QR 功能图案仍保持方形。线性条码只支持方形条纹。

## 生成 QR Code

`EncodeQR` 生成 QR Code Model 2 字节模式符号，支持版本 1–40 和纠错等级 L、M、Q、H，默认 L。编码器会选择能够容纳输入的最小版本。输入字符串按 UTF-8 转为字节。L 级字节模式最大容量为 2,953 字节；当前不使用数字或字母数字专用模式压缩数据。

```go
code, err := barcode.EncodeQR("Hello, QR", barcode.Options{
    ErrorCorrection: barcode.ErrorCorrectionM,
    Shape: barcode.ShapeRounded,
})
```

编码器会评估全部八种掩码，并在本项目代码中生成格式信息和 Reed–Solomon 纠错码字。

### 添加标志图片

将 `image.Image` 赋给 `Options.Logo`。可使用 `LoadLogo` 读取 PNG、JPEG 或 GIF。图片会保持宽高比例，缩放至图像短边的 6% 以内并置于 QR 中央。使用标志时，编码器自动启用 H 级纠错，并选择中心区域不覆盖定位、时序、校准、格式或版本图案的 QR 版本。

```go
file, err := os.Open("logo.png")
if err != nil { return err }
defer file.Close()
logo, err := barcode.LoadLogo(file)
if err != nil { return err }
code, err := barcode.EncodeQR("RVINTECH", barcode.Options{Logo: logo})
if err != nil { return err }
```

## 生成线性条码

### Code 39

`EncodeCode39` 接受标准 Code 39 字符集，并将小写字母转换为大写。`*` 是保留的起始/结束符，不能作为输入数据。设置 `Options{Checksum: true}` 可添加可选的 modulo-43 校验字符。

### Code 128

`EncodeCode128` 当前接受 ASCII 字节。数字连续段会优先使用 Code Set C，其余字符会根据需要在 A/B 之间切换。函数始终添加 modulo-103 校验字符。FNC 字符和扩展模式尚未通过公开 API 提供。

## PNG、SVG 与集成

`PNG` 使用 Go 标准库 `image/png` 写入 `io.Writer`。`SVG` 输出独立文档；QR 标志会以 PNG data URI 嵌入。两种格式都会包含静区。`PNGBytes` 和 `SVGBytes` 便于集成 HTTP handler、网站资源和文档生成流程。

```go
svgBytes, err := code.SVGBytes()
if err != nil { return err }
if err := os.WriteFile("code.svg", svgBytes, 0644); err != nil { return err }
```

## 可运行示例

```sh
go run ./examples/demo -format qr -text "RVINTECH-123" -out ./barcode-output
```

示例会同时生成 PNG 和 SVG。将 `-format` 改为 `code39` 或 `code128` 可尝试其他格式。其他选项包括 `-shape`、`-ecc`、`-fg`、`-bg`、`-logo`、`-scale`、`-quiet` 和 `-checksum`。运行 `go run ./examples/demo -help` 查看说明。
