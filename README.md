# price-tracker-widget

A simple, frameless, always-on-top desktop widget that displays real-time prices from Binance.

## Features

* **Frameless and Always-on-Top:** Designed to be unobtrusive and stay visible while at work.
* **Customizable Symbols:** Add symbols to track to `~/.pricetrack.json`.
* **Real-Time Updates:** Prices are updated every 2 minutes.
* **Draggable:** Easily move the widget around your screen by dragging it.
* **Color Toggle:** Right-click on the widget to toggle between magenta and green text color.

## Installation

1. **Prerequisites:**
   * Go (1.18 or higher)
   * Qt (5.15 or higher)
   * Therecipe/qt bindings (ensure `qmake` is in your `PATH`)

1. **Clone the repository:**
   ```bash
   git clone https://github.com/zewebdev1337/price-tracker-widget.git
   ```

2. **Build and run:**
   ```bash
   cd pricetrack
   go build
   ./price-tracker-widget
   ```

## Configuration

The widget reads the list of symbols to track from `~/.pricetrack.json`. 

**Example `~/.pricetrack.json`:**

```json
["TRX", "PEPE", "SOL", "BNB"]
```

If the configuration file does not exist, the widget will create one with the default symbols: `BTC`, `ETH`, `SOL`.
