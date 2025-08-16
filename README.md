# CT-Bot

A crypto trading bot focused on short-term, trend-following strategies, combining rule-based trading techniques with machine learning to intelligently scale trade sizes based on confidence.

## Project Overview

**CT-Bot** is a hybrid crypto trading system designed to automate day trades using a mix of traditional technical analysis and machine learning. It leverages both deterministic trading signals (e.g., support/resistance, ADX, trendline breaks) and probabilistic model outputs to make position-weighted decisions.

This project currently supports the **Solana (SOL)**/USDT pair for, with plans to extend to more markets.

## ✅ What’s Been Built So Far

- Historical candlestick data fetcher for SOL/USDT using Binance API
  - Efficient, rate-limited multi-worker downloader in Go
  - Parses and sorts OHLCV candlestick data
- Live candlestick data obtained through a WebSocket connection (Binance), currently being converted to direct Solana block-chain connection
- Technical indicator module
  - EMA, ADX
- Connection to MySQL Database
  - Contains tables to store both historical candle data, live candle data, ML training data and table to log bot orders
- Sliding Window
  - Automated sliding window, initially propogated with most recent historical candles
- Trendlines module complete
  - Creation of Support and Resistance trendlines over the sliding window
  - Detection of breakouts from trendline
- Dataset preparation for training and rule-based logic integration

## 📐 Planned Strategy Pipeline

### Rule-Based Trading Engine

- Trendline detection via sliding windows
- Support/resistance level recognition
- Drawdown and volatility-based trade logic
- Break and retest pattern detection

### Machine Learning Integration

- ML model trained on enriched historical dataset as well as backtested results for original trading logic (OHLCV + indicators + historical trade signals)
- Outputs float confidence value in range [0.0, 1.0]
- Model does **not** directly decide trades — it **modulates** trade weight

**Example:**

If the rule-based logic flags a trade, and the ML model returns `confidence = 0.65`, the position size might be reduced from `$5` to `$3.25`, scaling risk and conviction together.

### Other Key Features (Planned)

- Paper trading mode for safe backtesting
- Modular design: Go for backend logic, Python for ML
- Future live trading using connected crypto wallets (starting with Solana)
- Long-term goal: Market Sentiment Analysis as well as expansion to support BTC, ETH, and other pairs

## Getting Started

### Prerequisites

- Go 1.21+
- Python 3.9+
- No paid APIs required for initial development
- Binance public endpoint used for historical data
- Sol RPC public endpoint used for raw trade data

_(Installation instructions will follow as live features are added.)_

## Roadmap

- [x] Historical data fetcher with proper sorting and rate-limiting
- [x] Base `CandleStick` struct and dataset pipeline
- [x] Compute and append technical indicators to candle struct
- [x] Train ML model
- [ ] Confidence-weighted trading logic (hybrid ML + rules)
- [ ] Drawdown and trendline-based rule logic
- [ ] Paper trading simulator
- [x] Real-time price feed integration
- [ ] Wallet connection and execution engine
- [ ] Analysis of Market Sentiment (Twitter, Reddit etc)
- [ ] Extend to other assets (BTC, ETH, etc.)

## 📄 License

MIT License

