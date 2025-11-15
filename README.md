# Golang Price Monitor

Minimalist TUI for tracking real-time stock (Finnhub) and crypto (Binance) prices.

![Go Price Monitor Screenshot](https://media.discordapp.net/attachments/1318707911731576834/1439102999023849483/image.png?ex=69194c9c&is=6917fb1c&hm=949f0f24c9c84c6621bd59d123fd5b1a1b7696395504588d5ad2a8bfd6e11df5&=&format=webp&quality=lossless)

## Quick Start

1.  **Get API Key:** Sign up at [Finnhub.io](https://finnhub.io/).
2.  **Setup:** Paste your key into a file named `finnhub_key.txt`.
3.  **Install:** `go mod tidy`
4.  **Run:**
    * Default: `go run main.go`
    * Custom list: `go run main.go -p my_stocks.txt`
