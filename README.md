# Golang Price Monitor

Minimalist TUI for tracking real-time stock (Finnhub) and crypto (Binance) prices.

![Image](https://github.com/user-attachments/assets/c89aa2c4-7b30-4d2c-a17d-87cf3732ff11)

## Quick Start

1.  **Get API Key:** Sign up at [Finnhub.io](https://finnhub.io/).
2.  **Setup:** Paste your key into a file named `finnhub_key.txt`.
3.  **Install:** `go mod tidy`
4.  **Run:**
    * Default: `go run main.go`
    * Custom list: `go run main.go -p my_stocks.txt`
