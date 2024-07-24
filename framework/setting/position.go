package setting

import (
	"reflect"
	"time"

	"github.com/wonderstone/QuantKit/config"
	"github.com/wonderstone/QuantKit/tools/recorder"
)


// & position采用了注册模式
// & NewPosition 创建新仓位动作在logic中完成了！

var TypeRegistry = make(map[string]reflect.Type)

func Register(elem interface{}, name ...string) {
	if len(name) > 0 {
		TypeRegistry[name[0]] = reflect.TypeOf(elem).Elem()
		return
	}

	t := reflect.TypeOf(elem).Elem()
	TypeRegistry[t.Name()] = t
}




// MakePositionRecord 生成记录
func MakePositionRecord(
	tm time.Time, instID string,
	direction config.PositionDirection,
	costPrice, lastPrice, volume, pnl float64,
) *recorder.PositionRecord {
	return &recorder.PositionRecord{
		Date:        tm.Format(config.TimeFormatDate2),
		Time:        tm.Format(config.TimeFormatTime2),
		InstID:      instID,
		Direction:   string(direction),
		CostPrice:   costPrice,
		LastPrice:   lastPrice,
		Volume:      volume,
		AvailVolume: volume,
		Amt:         volume * lastPrice,
		PnL:         pnl,
	}
}
