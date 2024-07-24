package perf

type IndicateType string // 性能指标类型

const (
	TotalReturn      IndicateType = "total-return"      // 总收益率
	AnnualizedReturn IndicateType = "annualized-return" // 年化收益率
	MaxDrawdown      IndicateType = "max-drawdown"      // 最大回撤
	SharpeRatio      IndicateType = "sharpe-ratio"      // 夏普比率
	SortinoRatio     IndicateType = "sortino-ratio"     // 索提诺比率
	CalmarRatio      IndicateType = "calmar-ratio"      // 卡玛比率
	WinRate          IndicateType = "win-rate"          // 胜率
	ProfitFactor     IndicateType = "profit-factor"     // 盈亏比
	Alpha            IndicateType = "alpha"             // alpha
	Beta             IndicateType = "beta"              // beta
	Volatility       IndicateType = "volatility"        // 波动率
	InformationRatio IndicateType = "information-ratio" // 信息比率
	TrackingError    IndicateType = "tracking-error"    // 跟踪误差
	TreynorRatio     IndicateType = "treynor-ratio"     // 特雷诺比率
	SterlingRatio    IndicateType = "sterling-ratio"    // 斯特林比率
	DownsideRisk     IndicateType = "downside-risk"     // 下行风险
	UpsidePotential  IndicateType = "upside-potential"  // 上行潜力
	R2               IndicateType = "r2"                // R2
	AlphaBetaRatio   IndicateType = "alpha-beta-ratio"  // alpha-beta比率
	GainLossRatio    IndicateType = "gain-loss-ratio"   // 盈亏比
)
