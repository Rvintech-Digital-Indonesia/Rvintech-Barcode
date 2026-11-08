# Rvintech Barcode

Rvintech Barcode 是一个用于生成 QR Code Model 2、Code 39 和 Code 128 的 Go 库，由 PT Rvintech Digital Indonesia
维护。编码器与渲染器由本项目自行实现，不依赖第三方条码生成库。PNG 输出及 PNG、JPEG、GIF 标志图片的读取仅使用 Go 标准库；运行时不需要网络。

## 支持的格式

| 格式       | 当前支持范围                                    |
|------------|-------------------------------------------------|
| QR Code    | 自动分段数字、字母数字和 UTF-8 字节模式；版本 1–40、纠错等级 L/M/Q/H |
| Code 39    | 标准字符集，可选 modulo-43 校验字符             |
| Code 128   | ASCII，自动切换字符集 A/B/C，包含必需的校验字符 |
| PNG 与 SVG | 所有格式；QR 标志图片可嵌入 SVG                 |

QR 编码器会自动选择数字、QR 字母数字和 UTF-8 字节模式，并对混合内容进行分段，以减少编码位数。添加标志图片时会自动选择 H 级纠错，并选择中心区域不会覆盖 QR 功能图案的版本。

## 安装

```sh
go get github.com/Rvintech-Digital-Indonesia/Rvintech-Barcode
```

## 快速开始

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

Code 39 示例：`barcode.EncodeCode39("RVT-39", barcode.Options{Checksum: true})`。Code 128 示例：
`barcode.EncodeCode128("INV-2026-001234", barcode.Options{})`。

## 独立运行示例

在项目根目录执行：

```sh
go run ./examples/demo -format qr -text "RVINTECH-123" -out ./barcode-output
```

该命令会生成 `qr.png` 和 `qr.svg`。将 `-format` 改为 `code39` 或 `code128` 可体验线性条码。运行
`go run ./examples/demo -help` 可查看颜色、QR 形状、纠错等级和标志图片等参数。

## 项目结构

- `barcode.go`：公开 API。
- `internal/qr`：QR 编码、Reed–Solomon 纠错、掩码和矩阵构造。
- `internal/code39`、`internal/code128`：线性条码编码器。
- `internal/symbol`：选项、符号矩阵、PNG 与 SVG 渲染。
- `examples/demo/main.go`：独立运行的示例程序。

详细文档：[English](DOC.md) · [Bahasa Indonesia](DOC.id.md) · [简体中文](DOC.zh.md)。README：[English](README.md) · [Bahasa Indonesia](README.id.md)。自定义的 source-available
许可证见英文版 [LICENSE](LICENSE)；该许可证并非 OSI 批准的开源许可证，发布前需要法律审查。
