package model

// EntryPrice
// EntryCostLoss
// EntryQuantity
// StopLossRecover
// NextProfitSeLLPrice
// NextInvestBuYPrice

type TradingSystemData struct {
	ID                       uint
	Symbol                   string    `json:"symbol"`
	ClosingPrices            []float64   `json:"closing_prices"`
	Timestamps               []int64     `json:"timestamps"`
	Signals                  []string    `json:"signals"`
	NextInvestBuYPrice       []float64 `json:"next_invest_buy_price"`
	NextProfitSeLLPrice      []float64 `json:"next_profit_sell_price"`
	CommissionPercentage     float64   `json:"commission_percentage"`
	InitialCapital           float64   `json:"initial_capital"`
	PositionSize             float64   `json:"position_size"`
	EntryPrice               []float64 `json:"entry_price"`
	InTrade                  bool      `json:"in_trade"`
	QuoteBalance             float64   `json:"quote_balance"`
	BaseBalance              float64   `json:"base_balance"`
	RiskCost                 float64   `json:"risk_cost"`
	DataPoint                int       `json:"data_point"`
	CurrentPrice             float64   `json:"current_price"`
	EntryQuantity            []float64 `json:"entry_quantity"`
	EntryCostLoss            []float64 `json:"entry_cost_loss"`
	TradeCount               int       `json:"trade_count"`
	TradingLevel             int       `json:"trading_level"`
	ClosedWinTrades          int       `json:"closed_win_trades"`
	EnableStoploss           bool      `json:"enable_stoploss"`
	StopLossTrigered         bool      `json:"stop_loss_triggered"`
	StopLossRecover          []float64 `json:"stop_loss_recover"`
	RiskFactor               float64   `json:"risk_factor"`
	MaxDataSize              int       `json:"max_data_size"`
	RiskProfitLossPercentage float64   `json:"risk_profit_loss_percentage"`
	BaseCurrency             string    `json:"base_currency"`
	QuoteCurrency            string    `json:"quote_currency"`
	MiniQty                  float64   `json:"mini_qty"`
	MaxQty                   float64   `json:"max_qty"`
	MinNotional              float64   `json:"min_notional"`
	StepSize                 float64   `json:"step_size"`
}

// type AppID uint64

type AppData struct {
	ID                     uint    `gorm:"primaryKey" json:"id"`
	DataPoint              int     `json:"data_point"`
	Strategy               string  `json:"strategy"`
	ShortPeriod            int     `json:"short_period"`
	LongPeriod             int     `json:"long_period"`
	ShortEMA               float64 `json:"short_ema"`
	LongEMA                float64 `json:"long_ema"`
	TargetProfit           float64 `json:"target_profit"`
	TargetStopLoss         float64 `json:"target_stop_loss"`
	RiskPositionPercentage float64 `json:"risk_position_percentage"`
	TotalProfitLoss        float64 `json:"total_profit_loss"`
}
type BacTServices interface {
	WriteToInfluxDB(backTData *AppData) error
}
