package account

import (
	"fmt"
	"sync"
	"time"

	"github.com/wonderstone/QuantKit/config"
	"github.com/wonderstone/QuantKit/framework/entity/account"
	"github.com/wonderstone/QuantKit/framework/entity/handler"

	"github.com/wonderstone/QuantKit/framework/setting"
	"github.com/wonderstone/QuantKit/tools/container/orderedmap"
	"github.com/wonderstone/QuantKit/tools/dataframe"
	"github.com/wonderstone/QuantKit/tools/idgen"
	"github.com/wonderstone/QuantKit/tools/math"
	"github.com/wonderstone/QuantKit/tools/recorder"
)



type DefaultHandler struct {
	handler.Resource

	ai *idgen.AutoInc

	framework setting.ReplayFramework
	realtime  bool

	mode string

	accounts        map[string]handler.Account2
	market2accounts map[config.MarketType][]account.Account

	recorder map[config.RecordType]recorder.Handler // 记录器

	wg sync.WaitGroup
}

func (d *DefaultHandler) GenOrderId() int64 {
	return d.ai.Id()
}

func (d *DefaultHandler) GetAccount(market config.MarketType) []account.Account {
	return d.market2accounts[market]
}

func (d *DefaultHandler) getLastAssetRecords() (record []recorder.AssetRecord, resumeDate string) {
	date := time.Now().AddDate(0, 0, -1).Format(config.TimeFormatDate2)

	recs := d.recorder[config.RecordTypeAsset].QueryRecord(
		recorder.WithSQL(
			fmt.Sprintf(
				"SELECT * FROM asset_records WHERE date = (SELECT MAX(date) FROM asset_records WHERE date <= '%s')",
				date,
			),
		),
	)

	for _, r := range recs {
		record = append(record, r.(recorder.AssetRecord))
	}

	if len(record) == 0 {
		config.ErrorF("未找到资金记录")
	}

	return record, record[0].Date
}

func (d *DefaultHandler) getPositionRecords(date string) []recorder.PositionRecord {
	var records []recorder.PositionRecord
	recs := d.recorder[config.RecordTypePosition].QueryRecord(
		recorder.WithSQL(
			fmt.Sprintf(
				"SELECT * FROM position_records WHERE date = '%s'",
				date,
			),
		),
	)
	for _, r := range recs {
		records = append(records, r.(recorder.PositionRecord))
	}

	return records

}

func (d *DefaultHandler) getOrderMaxID() int64 {
	var maxID int64
	recs := d.recorder[config.RecordTypeOrder].QueryRecord(
		recorder.WithSQL(
			fmt.Sprintf(
				"SELECT * FROM order_records WHERE order_date = '%s' ORDER BY order_id DESC LIMIT 1",
				time.Now().Format(config.TimeFormatDate2),
			),
		),
	)

	for _, r := range recs {
		maxID = r.(recorder.OrderRecord).OrderId
	}

	return maxID
}

func (d *DefaultHandler) getLastTillNowOrderRecords(date string) []recorder.OrderRecord {
	var records []recorder.OrderRecord
	recs := d.recorder[config.RecordTypeOrder].QueryRecord(
		recorder.WithSQL(
			fmt.Sprintf(
				"SELECT * FROM order_records WHERE order_date > '%s'", date,
			),
		),
	)

	for _, r := range recs {
		records = append(records, r.(recorder.OrderRecord))
	}

	return records
}

func (d *DefaultHandler) DoResume() {
	// 恢复资金记录, 根据前一天的资金和持仓数据叠加今天的订单数据还原出今天交易的数据
	accRec, date := d.getLastAssetRecords()
	posRec := d.getPositionRecords(date)
	orderRec := d.getLastTillNowOrderRecords(date)
	maxID := d.getOrderMaxID()

	for _, acc := range d.accounts {
		acc.DoResume(accRec, posRec, orderRec)
	}

	d.ai = idgen.New(maxID+1, 1)
}

func (d *DefaultHandler) GetHistoryAssetRecord(records *[]recorder.AssetRecord) error {
	if r, ok := d.recorder[config.RecordTypeAsset]; ok {
		err := r.Read(records)
		if err != nil {
			return err
		}

		return nil
	}

	config.ErrorF("未找到资金记录")
	return nil
}

func (d *DefaultHandler) SetResource(resource handler.Resource) {
	d.Resource = resource
}

func (d *DefaultHandler) Base() handler.Basic {
	return d.framework.Base()
}

func (d *DefaultHandler) CalcPositionPnL(
	tm time.Time, indicators orderedmap.OrderedMap[string, dataframe.StreamingRecord],
) {
	for _, acc := range d.accounts {
		acc.CalcPositionPnL(tm, indicators)
	}
}

func (d *DefaultHandler) DoMatch(tm time.Time, indicators orderedmap.OrderedMap[string, dataframe.StreamingRecord]) {
	for _, acc := range d.accounts {
		acc.DoMatch(tm, indicators, d.framework.Matcher())
	}
}

func (d *DefaultHandler) GetCurrTime() *time.Time {
	return d.framework.CurrTime()
}

func (d *DefaultHandler) Contract() handler.Contract {
	return d.framework.Contract()
}

func NewDefaultHandler() *DefaultHandler {
	return &DefaultHandler{}
}

func (d *DefaultHandler) Init(framework any, recorder map[config.RecordType]recorder.Handler) error {
	d.framework = framework.(setting.ReplayFramework)
	d.Resource = framework.(handler.Resource)

	d.realtime = d.framework.Config().Framework.Realtime
	d.wg = sync.WaitGroup{}

	if d.realtime {
		d.mode = "Real"
	} else {
		d.mode = "Sim"
	}

	d.accounts = make(map[string]handler.Account2)
	d.market2accounts = make(map[config.MarketType][]account.Account)

	d.recorder = recorder
	// 增加账户记录
	if assetRecorder := recorder[config.RecordTypeAsset]; assetRecorder != nil {
		d.wg.Add(1)
		go func() {
			defer d.wg.Done()
			err := assetRecorder.RecordChan()
			if err != nil {
				config.InfoF("账户记录写入结束")
			}
		}()
	}

	// 增加订单记录
	if orderRecorder := recorder[config.RecordTypeOrder]; orderRecorder != nil {
		d.wg.Add(1)
		go func() {
			defer d.wg.Done()
			err := orderRecorder.RecordChan()
			if err != nil {
				config.InfoF("订单记录写入结束")
			}
		}()
	}

	// 增加持仓记录
	if positionRecorder := recorder[config.RecordTypePosition]; positionRecorder != nil {
		d.wg.Add(1)
		go func() {
			defer d.wg.Done()
			err := positionRecorder.RecordChan()
			if err != nil {
				config.InfoF("持仓记录写入结束")
			}
		}()
	}

	d.market2accounts[config.MarketTypeStock] = make([]account.Account, 0)
	if stockAccConfig := d.framework.Config().Framework.Stock; stockAccConfig.Cash > 0 {
		acc := setting.NewAccount(config.AccountTypeStockSimple)
		err := acc.Init(d, stockAccConfig)
		if err != nil {
			config.ErrorF("股票账户初始化失败: %v", err)
			return err
		}

		d.accounts[config.AccountTypeStockSimple] = acc
		d.market2accounts[config.MarketTypeStock] = append(d.market2accounts[config.MarketTypeStock], acc)
	}

	d.market2accounts[config.MarketTypeFuture] = make([]account.Account, 0)
	if futureAccConfig := d.framework.Config().Framework.Future; futureAccConfig.Cash > 0 {
		acc := setting.NewAccount(config.AccountTypeFuture)
		err := acc.Init(d, futureAccConfig)
		if err != nil {
			config.ErrorF("期货账户初始化失败: %v", err)
			return err
		}

		d.accounts[config.AccountTypeFuture] = acc
		d.market2accounts[config.MarketTypeFuture] = append(d.market2accounts[config.MarketTypeFuture], acc)
	}

	if d.Config().Mode == config.RunMode {
		d.DoResume()
	} else {
		d.ai = idgen.New(100000, 1)
	}

	return nil
}

func MakeRecord(
	tm time.Time, accountID string, asset account.Asset, pnl account.PnL,
) recorder.AssetRecord {
	return recorder.AssetRecord{
		Date: tm.Format(config.TimeFormatDate2),
		Time: tm.Format(config.TimeFormatTime2),

		MarketValue: math.Round(asset.MarketValue, 2),
		Margin:      math.Round(asset.Margin, 2),
		FundAvail:   math.Round(asset.Available, 2),
		TotalAsset:  math.Round(asset.Total, 2),
		Profit:      math.Round(pnl.Profit, 2),
		Commission:  math.Round(asset.Commission, 2),
		Account:     accountID,
		Mode:        "VMT",
	}
}

func (d *DefaultHandler) GetAccounts() map[string]account.Account {
	accs := make(map[string]account.Account)
	for _, acc := range d.accounts {
		accs[acc.AccountID()] = acc
	}

	return accs
}

func (d *DefaultHandler) GetAccountByID(accountID string) account.Account {
	return d.accounts[accountID]
}

func (d *DefaultHandler) DoSettle(tm time.Time, indicators *orderedmap.OrderedMap[string, dataframe.StreamingRecord]) {
	for _, acc := range d.accounts {
		acc.DoSettle(tm, indicators, d.recorder)
	}

}

func (d *DefaultHandler) Release() {
	defer d.wg.Wait()
	for _, rec := range d.recorder {
		rec.Release()
	}

	for _, acc := range d.accounts {
		acc.Release()
	}

	d.ai.Close()
}

func matchOrder(o handler.Order, filters ...account.WithOrderFilter) bool {
	for _, f := range filters {
		if !f(o) {
			return false
		}
	}

	return true
}

func (d *DefaultHandler) IterOrders(instId string, filter ...account.WithOrderFilter) <-chan account.Order {
	c := make(chan account.Order)

	go func() {
		for _, acc := range d.accounts {
			orders := acc.GetOrders(instId)
			if orders != nil {
				for _, o := range orders {
					if matchOrder(o.(handler.Order), filter...) {
						c <- o
					}
				}
			}
		}
		close(c)
	}()
	return c
}

func (d *DefaultHandler) GetOrdersByInst(instID string) []account.Order {
	var orderList []account.Order
	for _, acc := range d.accounts {
		orders := acc.GetOrders(instID)
		if orders != nil {
			for _, o := range orders {
				if o.InstID() == instID {
					orderList = append(orderList, o.(handler.Order))
				}
			}
		}
	}
	return orderList
}

func (d *DefaultHandler) NewOrder(
	accountID, instID string, qty float64, option ...account.WithOrderOption,
) (account.Order, error) {
	return d.accounts[accountID].NewOrder(instID, qty, option...)
}

func (d *DefaultHandler) InsertOrder(
	order account.Order, options ...account.WithOrderOption,
) error {
	return order.Account().InsertOrder(order, options...)
}

func (d *DefaultHandler) CancelOrder(tm time.Time, id int64, accountId string) {
	d.accounts[accountId].CancelOrder(tm, id)
}

func ResumeOrder() account.WithOrderOption {
	return func(opts *account.OrderOp) {
		opts.Resumed = true
	}
}


