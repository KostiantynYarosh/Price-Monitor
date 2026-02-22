# Golang Price Monitor

Minimalist TUI for tracking real-time stock (Finnhub) and crypto (Binance) prices.

<img width="505" height="587" alt="Image" src="https://github.com/user-attachments/assets/213873be-226c-43da-9f59-5c93855b0514" />

## Quick Start

1.  **Get API Key:** Sign up at [Finnhub.io](https://finnhub.io/).
2.  **Setup:** Paste your key into a file named `finnhub_key.txt`.
3.  **Install:** `go mod tidy`
4.  **Run:**
    * Default: `go run main.go`
    * Custom list: `go run main.go -p my_stocks.txt`
