# Go Price Monitor

Minimalist TUI for tracking real-time stock (Finnhub) and crypto (Binance) prices.

![Go Price Monitor Screenshot](https://media.discordapp.net/attachments/1318707911731576834/1439101293804392600/image.png?ex=69194b05&is=6917f985&hm=27212071bd5c3fe5e0803df7096ab6f517dea440228b47641becd37a5afc0692&=&format=webp&quality=lossless)

## Quick Start

1.  **Get API Key:** Sign up at [Finnhub.io](https://finnhub.io/).
2.  **Setup:** Paste your key into a file named `finnhub_key.txt`.
3.  **Install:** `go mod tidy`
4.  **Run:**
    * Default: `go run main.go`
    * Custom list: `go run main.go -p my_stocks.txt`