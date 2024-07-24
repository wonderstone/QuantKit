package account

// Asset 资产
type Asset struct {
	Initial     float64 // 初始资金
	Total       float64 // 总资产
	Available   float64 // 可用资金
	Frozen      float64 // 冻结资金
	Margin      float64 // 保证金
	MarketValue float64 // 市值
	Commission  float64 // 手续费
}

// PnL 盈亏
type PnL struct {
	Profit float64
}

func (p PnL) Add(r PnL) PnL {
	return PnL{
		Profit: p.Profit + r.Profit,
	}
}