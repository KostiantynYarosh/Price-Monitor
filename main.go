package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type stockDataMsg struct {
	symbol string
	price  float64
	change float64
}

type tickMsg time.Time
type uiTickMsg time.Time

type viewMode int

const (
	viewCustom viewMode = iota
	viewTopCrypto
	viewTopStocks
	viewTechStocks
	viewAll
)

type profile struct {
	name    string
	symbols []string
}

var profiles = map[viewMode]profile{
	viewTopCrypto: {
		name: "Top 10 Crypto",
		symbols: []string{
			"BTCUSDT", "ETHUSDT", "BNBUSDT", "XRPUSDT", "ADAUSDT",
			"DOGEUSDT", "SOLUSDT", "DOTUSDT", "MATICUSDT", "AVAXUSDT",
		},
	},
	viewTopStocks: {
		name: "Top 10 Stocks",
		symbols: []string{
			"AAPL", "MSFT", "GOOGL", "AMZN", "NVDA",
			"META", "TSLA", "V", "JPM", "WMT",
		},
	},
	viewTechStocks: {
		name: "Tech Stocks",
		symbols: []string{
			"AAPL", "MSFT", "GOOGL", "META", "NVDA",
			"AMD", "INTC", "NFLX", "ADBE", "CRM",
		},
	},
	viewAll: {
		name: "All Markets",
		symbols: []string{
			"BTCUSDT", "ETHUSDT", "AAPL", "GOOGL", "MSFT",
			"TSLA", "NVDA", "BNBUSDT", "XRPUSDT", "META",
		},
	},
}

type model struct {
	choices       []string
	cursor        int
	selected      map[string]bool
	textInput     textinput.Model
	inputMode     bool
	profileMode   bool
	stockData     map[string]stockInfo
	lastUpdate    time.Time
	apiKey        string
	currentView   viewMode
	profileCursor int
}

type stockInfo struct {
	price  float64
	change float64
}

var (
	// 🚀 ПУНКТ 1: Створюємо один клієнт для перевикористання
	apiClient = &http.Client{Timeout: 10 * time.Second}

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00FFFF")).
			Background(lipgloss.Color("#1a1a2e")).
			Padding(0, 1)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#874BFD")).
			Padding(1, 2).
			Width(70)

	selectedItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFD700")).
				Bold(true)

	greenStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00")).
			Bold(true)

	redStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true)

	grayStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888"))

	cursorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF00FF")).
			Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			Italic(true)

	cryptoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFA500"))

	stockStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#4169E1"))

	profileStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00")).
			Bold(true)
)

type FinnhubResponse struct {
	C  float64 `json:"c"`
	D  float64 `json:"d"`
	DP float64 `json:"dp"`
	H  float64 `json:"h"`
	L  float64 `json:"l"`
	O  float64 `json:"o"`
	PC float64 `json:"pc"`
}

type BinanceResponse struct {
	Symbol             string `json:"symbol"`
	LastPrice          string `json:"lastPrice"`
	PriceChangePercent string `json:"priceChangePercent"`
}

func loadAPIKey(filename string) (string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func loadSymbolsFromFile(filename string) ([]string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading profile file: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	var symbols []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			symbols = append(symbols, strings.ToUpper(trimmed))
		}
	}

	if len(symbols) == 0 {
		return nil, fmt.Errorf("profile file is empty or contains no valid symbols")
	}

	return symbols, nil
}

func initialModel(initialChoices []string) model {
	ti := textinput.New()
	ti.Placeholder = "Enter symbol (AAPL for stock, BTCUSDT for crypto)..."
	ti.CharLimit = 50

	apiKey, err := loadAPIKey("finnhub_key.txt")
	if err != nil {
		apiKey = ""
	}

	selectedMap := make(map[string]bool)
	for _, symbol := range initialChoices {
		selectedMap[symbol] = true
	}

	return model{
		choices:       initialChoices,
		selected:      selectedMap,
		textInput:     ti,
		inputMode:     false,
		profileMode:   false,
		cursor:        0,
		profileCursor: 0,
		stockData:     make(map[string]stockInfo),
		apiKey:        apiKey,
		currentView:   viewCustom,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		tickCmd(),
		uiTickCmd(),
		// 📉 ПУНКТ 2: Ми передаємо m.selected для початкового завантаження
		fetchAllSymbolsCmd(m.choices, m.selected, m.apiKey),
	)
}

func tickCmd() tea.Cmd {
	return tea.Tick(10*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func uiTickCmd() tea.Cmd {
	return tea.Tick(1*time.Second, func(t time.Time) tea.Msg {
		return uiTickMsg(t)
	})
}

// 📉 ПУНКТ 2: Функція тепер приймає 'selected', щоб знати, що оновлювати
func fetchAllSymbolsCmd(symbols []string, selected map[string]bool, apiKey string) tea.Cmd {
	return func() tea.Msg {
		var cmds []tea.Cmd
		for _, symbol := range symbols {
			// Перевіряємо, чи символ обраний, перш ніж його завантажувати
			if selected[symbol] {
				cmds = append(cmds, fetchSymbolData(symbol, apiKey))
			}
		}
		return tea.Batch(cmds...)()
	}
}

func fetchSymbolData(symbol string, apiKey string) tea.Cmd {
	if isCrypto(symbol) {
		return fetchCryptoData(symbol)
	}
	return fetchStockDataFinnhub(symbol, apiKey)
}

func isCrypto(symbol string) bool {
	cryptoSuffixes := []string{"USDT", "BTC", "ETH", "BUSD", "USDC"}
	for _, suffix := range cryptoSuffixes {
		if strings.HasSuffix(symbol, suffix) {
			return true
		}
	}
	return false
}

func fetchStockDataFinnhub(symbol string, apiKey string) tea.Cmd {
	return func() tea.Msg {
		if apiKey == "" {
			return stockDataMsg{symbol: symbol, price: 0, change: 0}
		}

		url := fmt.Sprintf(
			"https://finnhub.io/api/v1/quote?symbol=%s&token=%s",
			symbol, apiKey,
		)

		// 🚀 ПУНКТ 1: Використовуємо глобальний apiClient
		resp, err := apiClient.Get(url)
		if err != nil {
			return stockDataMsg{symbol: symbol, price: 0, change: 0}
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return stockDataMsg{symbol: symbol, price: 0, change: 0}
		}

		var data FinnhubResponse
		if err := json.Unmarshal(body, &data); err != nil {
			return stockDataMsg{symbol: symbol, price: 0, change: 0}
		}

		if data.C == 0 {
			return stockDataMsg{symbol: symbol, price: 0, change: 0}
		}

		return stockDataMsg{
			symbol: symbol,
			price:  data.C,
			change: data.DP,
		}
	}
}

func fetchCryptoData(symbol string) tea.Cmd {
	return func() tea.Msg {
		url := fmt.Sprintf("https://api.binance.com/api/v3/ticker/24hr?symbol=%s", symbol)

		// 🚀 ПУНКТ 1: Використовуємо глобальний apiClient
		resp, err := apiClient.Get(url)
		if err != nil {
			return stockDataMsg{symbol: symbol, price: 0, change: 0}
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return stockDataMsg{symbol: symbol, price: 0, change: 0}
		}

		var data BinanceResponse
		if err := json.Unmarshal(body, &data); err != nil {
			return stockDataMsg{symbol: symbol, price: 0, change: 0}
		}

		var price, change float64
		fmt.Sscanf(data.LastPrice, "%f", &price)
		fmt.Sscanf(data.PriceChangePercent, "%f", &change)

		return stockDataMsg{
			symbol: symbol,
			price:  price,
			change: change,
		}
	}
}

func (m *model) loadProfile(view viewMode) tea.Cmd {
	m.currentView = view
	if view == viewCustom {
		return nil
	}

	profile := profiles[view]
	m.choices = profile.symbols
	m.selected = make(map[string]bool)
	for _, symbol := range m.choices {
		m.selected[symbol] = true
	}
	m.cursor = 0

	// 📉 ПУНКТ 2: Передаємо m.selected при завантаженні профілю
	return fetchAllSymbolsCmd(m.choices, m.selected, m.apiKey)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tickMsg:
		return m, tea.Batch(
			tickCmd(),
			// 📉 ПУНКТ 2: Оновлюємо лише обрані символи
			fetchAllSymbolsCmd(m.choices, m.selected, m.apiKey),
		)

	case uiTickMsg:
		return m, uiTickCmd()

	case stockDataMsg:
		m.stockData[msg.symbol] = stockInfo{
			price:  msg.price,
			change: msg.change,
		}
		m.lastUpdate = time.Now()
		return m, nil

	case tea.KeyMsg:
		if m.profileMode {
			switch msg.String() {
			case "up", "k":
				if m.profileCursor > 0 {
					m.profileCursor--
				}
			case "down", "j":
				if m.profileCursor < 4 {
					m.profileCursor++
				}
			case "enter":
				m.profileMode = false
				return m, m.loadProfile(viewMode(m.profileCursor))
			case "esc", "p":
				m.profileMode = false
			case "q", "ctrl+c":
				return m, tea.Quit
			}
		} else if m.inputMode {
			switch msg.String() {
			case "enter":
				if m.textInput.Value() != "" {
					newSymbol := strings.ToUpper(m.textInput.Value())
					m.choices = append(m.choices, newSymbol)

					m.selected[newSymbol] = true

					m.textInput.Reset()
					m.inputMode = false
					m.currentView = viewCustom
					return m, fetchSymbolData(newSymbol, m.apiKey)
				}
				m.inputMode = false
			case "esc":
				m.inputMode = false
				m.textInput.Reset()
			default:
				m.textInput, cmd = m.textInput.Update(msg)
				return m, cmd
			}
		} else {
			switch msg.String() {
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < len(m.choices)-1 {
					m.cursor++
				}
			case "enter", " ":
				symbol := m.choices[m.cursor]
				m.selected[symbol] = !m.selected[symbol]
				m.currentView = viewCustom

				// 📉 ПУНКТ 2 (Бонус): Якщо ми щойно обрали символ,
				// ми можемо негайно запросити для нього дані,
				// не чекаючи наступного 10-секундного тіка.
				if m.selected[symbol] {
					return m, fetchSymbolData(symbol, m.apiKey)
				}
			case "a":
				m.inputMode = true
				m.textInput.Focus()
				return m, textinput.Blink
			case "d":
				if len(m.choices) > 0 {
					deletedSymbol := m.choices[m.cursor]

					m.choices = append(m.choices[:m.cursor], m.choices[m.cursor+1:]...)

					delete(m.selected, deletedSymbol)
					delete(m.stockData, deletedSymbol)

					if m.cursor >= len(m.choices) && m.cursor > 0 {
						m.cursor--
					}
					m.currentView = viewCustom
				}
			case "p":
				m.profileMode = true
			case "q", "ctrl+c":
				return m, tea.Quit
			}
		}
	}

	return m, nil
}

func (m model) View() string {
	if m.profileMode {
		return m.renderProfileMenu()
	}

	if m.inputMode {
		return boxStyle.Render(
			titleStyle.Render("Add New Symbol") + "\n\n" +
				m.textInput.View() + "\n\n" +
				helpStyle.Render("Stock: AAPL | Crypto: BTCUSDT\nenter: save | esc: cancel"),
		)
	}

	var content strings.Builder

	profileName := "Custom"
	if m.currentView != viewCustom {
		profileName = profiles[m.currentView].name
	}

	content.WriteString(titleStyle.Render("PRICE MONITOR") + " " +
		profileStyle.Render(fmt.Sprintf("[%s]", profileName)) + "\n\n")

	hasSelected := false
	for _, symbol := range m.choices {
		if m.selected[symbol] {
			hasSelected = true
			info, ok := m.stockData[symbol]

			symbolLabel := symbol
			if isCrypto(symbol) {
				symbolLabel = cryptoStyle.Render(symbol)
			} else {
				symbolLabel = stockStyle.Render(symbol)
			}

			if ok && info.price > 0 {
				priceStr := fmt.Sprintf("$%.2f", info.price)
				var changeStr string

				if info.change > 0 {
					changeStr = greenStyle.Render(fmt.Sprintf("+%.2f%%", info.change))
				} else if info.change < 0 {
					changeStr = redStyle.Render(fmt.Sprintf("%.2f%%", info.change))
				} else {
					changeStr = grayStyle.Render("0.00%")
				}

				content.WriteString(fmt.Sprintf("%-10s %13s  %s\n",
					symbolLabel, priceStr, changeStr))
			} else {
				content.WriteString(fmt.Sprintf("%-10s %s\n",
					symbolLabel, grayStyle.Render("Loading...")))
			}
		}
	}

	if !hasSelected {
		content.WriteString(grayStyle.Render("No symbols selected (press space to select)") + "\n")
	}

	content.WriteString("\n")

	if !m.lastUpdate.IsZero() {
		elapsed := int(time.Since(m.lastUpdate).Seconds())
		content.WriteString(grayStyle.Render(fmt.Sprintf("Updated: %ds ago", elapsed)) + "\n")
	} else {
		content.WriteString(grayStyle.Render("Updated: never") + "\n")
	}

	content.WriteString("\n" + strings.Repeat("─", 60) + "\n\n")

	content.WriteString("Select monitors:\n\n")

	for i, symbol := range m.choices {
		cursor := "  "
		if m.cursor == i {
			cursor = cursorStyle.Render("> ")
		}

		checkbox := "[ ]"
		symbolText := symbol
		if m.selected[symbol] {
			checkbox = selectedItemStyle.Render("[*]")
			symbolText = selectedItemStyle.Render(symbol)
		}

		typeLabel := "[STOCK]"
		if isCrypto(symbol) {
			typeLabel = cryptoStyle.Render("[CRYPTO]")
		} else {
			typeLabel = stockStyle.Render("[STOCK]")
		}

		content.WriteString(fmt.Sprintf("%s%s %-10s %s\n", cursor, checkbox, symbolText, typeLabel))
	}

	content.WriteString("\n" + helpStyle.Render("space/enter: toggle | a: add | d: delete | p: profiles | q: quit"))

	return boxStyle.Render(content.String())
}

func (m model) renderProfileMenu() string {
	var content strings.Builder

	content.WriteString(titleStyle.Render("SELECT PROFILE") + "\n\n")

	profileList := []viewMode{viewCustom, viewTopCrypto, viewTopStocks, viewTechStocks, viewAll}

	for i, view := range profileList {
		cursor := "  "
		if m.profileCursor == i {
			cursor = cursorStyle.Render("> ")
		}

		name := "Custom"
		if view != viewCustom {
			name = profiles[view].name
		}

		active := ""
		if m.currentView == view {
			active = selectedItemStyle.Render(" ✓")
		}

		content.WriteString(fmt.Sprintf("%s%s%s\n", cursor, name, active))
	}

	content.WriteString("\n" + helpStyle.Render("↑/↓: navigate | enter: select | esc: cancel | q: quit"))

	return boxStyle.Render(content.String())
}

func main() {
	profileFile := flag.String("p", "", "Path to a file containing a list of stock/crypto symbols (one per line).")
	flag.Parse()

	initialChoices := []string{"AAPL", "GOOGL", "MSFT", "BTCUSDT", "ETHUSDT"}

	if *profileFile != "" {
		symbols, err := loadSymbolsFromFile(*profileFile)
		if err != nil {
			fmt.Printf("Error loading profile from file: %v\n", err)
			os.Exit(1)
		}
		initialChoices = symbols
	}

	p := tea.NewProgram(initialModel(initialChoices))
	if _, err := p.Run(); err != nil {
		fmt.Printf("An error occurred: %v\n", err)
		os.Exit(1)
	}
}
