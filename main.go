package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/widgets"
)

var defaultSymbols = []string{"BTC", "ETH", "SOL"}

type BinanceWidget struct {
	widgets.QWidget

	apiUrl string

	// priceLabels stores pointers to QLabel widgets.
	// Each QLabel widget displays the price of a specific pair.
	// The key of the map is the symbol string, and the value is a pointer to the corresponding QLabel widget.
	priceLabels map[string]*widgets.QLabel

	// symbols is a slice that stores the symbols to be tracked.
	symbols []string

	// oldPos is a pointer to a QPoint object that stores the previous position of the mouse cursor.
	// This is used to calculate the distance moved by the mouse cursor and update the position of the widget accordingly.
	oldPos *core.QPoint

	// layout is a pointer to a QVBoxLayout object that manages the layout of the widgets in the BinanceWidget.
	layout *widgets.QVBoxLayout
}

// NewBinanceWidget creates a new instance of BinanceWidget with the given parent, window type, and symbols.
// It initializes the widget's API URL, price labels, symbols, and old position.
// It then calls the initUI method to set up the widget's UI.
func NewBinanceWidget(parent widgets.QWidget_ITF, fo core.Qt__WindowType, symbols []string) *BinanceWidget {
	widget := &BinanceWidget{
		apiUrl:      "https://api.binance.com/api/v3/ticker/price?symbol=",
		QWidget:     *widgets.NewQWidget(parent, fo),
		oldPos:      core.NewQPoint(),
		priceLabels: make(map[string]*widgets.QLabel),
		symbols:     symbols,
	}
	widget.initUI()
	return widget
}

// initUI sets up the UI for the BinanceWidget.
// It sets the window flags and attributes, creates a new vertical box layout, and adds a QLabel widget for each symbol.
// It then sets the layout for the widget, updates the price, and starts a timer to update the price every 120 seconds.
// It also connects the mouse press, move, and context menu events to their respective methods.
func (w *BinanceWidget) initUI() {
	w.SetWindowFlags(core.Qt__FramelessWindowHint | core.Qt__WindowStaysOnTopHint)
	w.SetAttribute(core.Qt__WA_X11NetWmWindowTypeDock, true)
	w.SetAttribute(core.Qt__WA_TranslucentBackground, true)
	w.layout = widgets.NewQVBoxLayout()

	for _, symbol := range w.symbols {
		label := widgets.NewQLabel2("Loading...", nil, 0)
		label.SetStyleSheet("color: rgb(0, 255, 0)")
		w.layout.AddWidget(label, 0, 0)
		w.priceLabels[symbol] = label
	}

	w.SetLayout(w.layout)
	w.updatePrice()

	timer := core.NewQTimer(nil)
	timer.ConnectTimeout(w.updatePrice)
	timer.Start(120000)

	w.ConnectMousePressEvent(w.mousePressEvent)
	w.ConnectMouseMoveEvent(w.mouseMoveEvent)
	w.ConnectContextMenuEvent(w.contextMenuEvent)
}

// updatePrice sends HTTP requests to the Binance API to update the price for each symbol in the widget.
// It uses a WaitGroup to ensure that all requests have completed before returning.
// Each request is made in a separate goroutine.
func (w *BinanceWidget) updatePrice() {
	var wg sync.WaitGroup
	for _, symbol := range w.symbols {
		wg.Add(1)
		go func(symbol string) {
			defer wg.Done()
			w.updateLabel(symbol)
		}(symbol)
	}
	wg.Wait()
}

// updateLabel sends an HTTP request to the Binance API to update the price for a single symbol.
// It updates the corresponding QLabel widget with the new price.
// If there is an error, it updates the QLabel widget with the error message.
func (w *BinanceWidget) updateLabel(symbol string) {
	url := fmt.Sprintf("%s%sUSDT", w.apiUrl, symbol)
	resp, err := http.Get(url)
	if err != nil {
		w.priceLabels[symbol].SetText(fmt.Sprintf("Error: %v", err))
		return
	}
	defer resp.Body.Close()
	var data map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		w.priceLabels[symbol].SetText(fmt.Sprintf("Error: %v", err))
		return
	}
	price := data["price"]
	// Remove trailing zeroes
	price = strings.TrimRight(price, "0")
	// If the price ends with a decimal point, remove it
	if strings.HasSuffix(price, ".") {
		price = price[:len(price)-1]
	}
	w.priceLabels[symbol].SetText(fmt.Sprintf("%s/USDT: %s", symbol, price))
	w.priceLabels[symbol].Font().SetPointSize(12)
}

// mousePressEvent is called when the mouse button is pressed on the widget.
// It updates the oldPos field with the current position of the mouse cursor.
func (w *BinanceWidget) mousePressEvent(event *gui.QMouseEvent) {
	w.oldPos = event.GlobalPos()
}

// mouseMoveEvent is called when the mouse cursor is moved over the widget.
// It calculates the distance moved by the mouse cursor and updates the position of the widget accordingly.
// It also updates the oldPos field with the current position of the mouse cursor.
func (w *BinanceWidget) mouseMoveEvent(event *gui.QMouseEvent) {
	deltaX := event.GlobalPos().X() - w.oldPos.X()
	deltaY := event.GlobalPos().Y() - w.oldPos.Y()
	w.Move2(w.X()+deltaX, w.Y()+deltaY)
	w.oldPos = event.GlobalPos()
}

// contextMenuEvent is called when a context menu event is triggered on the widget.
// It toggles the color of the text in all price labels.
// The new color is the inverse of the current color.
func (w *BinanceWidget) contextMenuEvent(event *gui.QContextMenuEvent) {
	currentColor := w.priceLabels[w.symbols[0]].Palette().Color(gui.QPalette__Active, gui.QPalette__WindowText)
	newColor := gui.NewQColor3(255-currentColor.Red(), 255-currentColor.Green(), 255-currentColor.Blue(), 255)

	// Create a new style sheet string with the new color
	styleSheet := fmt.Sprintf("color: rgb(%d, %d, %d)", newColor.Red(), newColor.Green(), newColor.Blue())
	// Set the style sheet of all price labels to the new style sheet
	for _, label := range w.priceLabels {
		label.SetStyleSheet(styleSheet)
		label.Font().SetPointSize(12)
	}
}

// loadSymbolsFromConfig loads the symbols to track from a config file.
// If the config file does not exist, it creates a new one with default symbols.
// It returns the symbols to track.
func loadSymbolsFromConfig() ([]string, error) {
	homePath, _ := os.UserHomeDir()
	configPath := fmt.Sprintf("%s/.pricetrack.json", homePath)
	configFile, err := os.Open(configPath)

	// If the config file does not exist, create a new one with default symbols
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("Config file not found, creating a new one with default symbols: %v\n", defaultSymbols)
			createDefaultConfig(configPath)
			return defaultSymbols, nil
		} else {
			// If there is an error opening the config file, print the error and return default symbols
			fmt.Println("Error opening config file:", err)
			return defaultSymbols, nil
		}
	}

	defer configFile.Close()
	stringPointers, err := parseSymbolsFromConfig(configFile)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	strings := []string{}
	for _, strPtr := range stringPointers {
		strings = append(strings, *strPtr)
	}
	return strings, nil
}

// createDefaultConfig creates a new config file with default symbols.
func createDefaultConfig(path string) error {
	defaultConfig, err := json.Marshal(defaultSymbols)
	if err != nil {
		return fmt.Errorf("can't marshal default symbols: %w", err)
	}
	if err := os.WriteFile(path, defaultConfig, 0644); err != nil {
		return fmt.Errorf("create config file: %w", err)
	}
	return nil
}

// parseSymbolsFromConfig parses the symbols to track from a config file.
func parseSymbolsFromConfig(file *os.File) ([]*string, error) {
	byteValue, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("can't marshal default symbols: %w", err)
	}
	var symbols []*string
	json.Unmarshal(byteValue, &symbols)
	return symbols, nil
}

func main() {
	widgets.NewQApplication(len([]string{}), []string{})
	symbols, err := loadSymbolsFromConfig()
	if err != nil {
		log.Fatal(err)
	}
	widget := NewBinanceWidget(nil, 0, symbols)
	widget.Show()
	widgets.QApplication_Exec()
}
