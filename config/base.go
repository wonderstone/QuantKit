package config

import (
	"math"
	"time"
)

// Xrxd 除权除息数据 Ex-Right Dividend
type Xrxd struct {
	InstID         string  `csv:"inst_id"`          // 合约代码
	ExDate         *Time   `csv:"ex_date"`          // 除权除息日
	RegDate        *Time   `csv:"regi_date"`        // 股权登记日
	RegiClosePrice float64 `csv:"regi_close_price"` // 股权登记日收盘价
	DivCash        float64 `csv:"div_cash"`         // 每股派现
	DivShare       float64 `csv:"div_share"`        // 每股送股
	TrabShare      float64 `csv:"trab_share"`       // 每股转增股
	PlaceRate      float64 `csv:"place_rate"`       // 配股比例
	PlacePrice     float64 `csv:"place_price"`      // 配股价
	ExchRate       float64 `csv:"exch_rate"`        // 换股比例
	ExchInstID     string  `csv:"exch_inst"`        // 换股代码
	ExFactor       float64 `csv:"ex_factor"`        // 复权因子
	CumlExFactor   float64 `csv:"cuml_ex_factor"`   // 累计复权因子
	XrxdPrice      float64 `csv:"ex_div_price"`     // 除息除权价格
}

func adjustedPrice(regiPrice, DivCash, DivShare, PlaceRate, TrabShare, PlacePrice float64) float64 {
	return (regiPrice - DivCash + (PlacePrice * PlaceRate)) / (1 + DivShare + PlaceRate + TrabShare)
}

func (x *Xrxd) CalcPos(qty, price float64, closePrice ...float64) (
	settleQty, settlePrice, settleLastPrice, dividend, tax float64, instID string,
) {
	settleQty = qty
	if x.XrxdPrice > 0 {
		settleLastPrice = x.XrxdPrice
	} else if (len(closePrice) > 0) && (closePrice[0] > 0) {
		settleLastPrice = adjustedPrice(closePrice[0], x.DivCash, x.DivShare, x.PlaceRate, x.TrabShare, x.PlacePrice)
	}

	if x.ExchRate > 0 {
		settleQty = math.Round(x.ExchRate * qty)
		instID = x.ExchInstID
	}

	if x.DivCash > 0 {
		dividend += x.DivCash * qty
	}

	settleQty += (x.DivShare + x.TrabShare + x.PlaceRate) * qty

	if x.PlaceRate > 0 {
		dividend -= x.PlaceRate * qty * x.PlacePrice
	}

	settlePrice = (price*qty - dividend) / settleQty

	if dividend > 0 {
		// 资本利得税
		tax = dividend * 0.2
	}

	return
}

type Time struct {
	time.Time
}

func (t *Time) UnmarshalCSV(text string) error {
	if t == nil {
		t.Time = time.Time{}
	}

	tm, err := time.ParseInLocation(TimeFormatDefault, text, time.Local)
	if err != nil {
		ErrorF("解析时间失败, 格式: %s, text: %v, err: %v", TimeFormatDefault, text, err)
	}

	t.Time = tm

	return nil
}
