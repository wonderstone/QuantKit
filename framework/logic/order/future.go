package order

// import (
// 	"time"

// 	"github.com/wonderstone/QuantKit/framework/entity/account"
// 	"github.com/wonderstone/QuantKit/framework/entity/contract"
// 	"github.com/wonderstone/QuantKit/config"
// 	"github.com/wonderstone/QuantKit/tools/recorder"
// )


// type FutureOrder struct {
// 	contract       contract.Contract
// 	isEligible     bool
// 	isExecuted     bool
// 	orderTime      string
// 	orderPrice     float64
// 	orderNum       float64
// 	orderDirection string
// 	orderType      string
// }

// func (f *FutureOrder) DoRTInsert() error {
// 	//TODO implement me
// 	panic("implement me")
// }

// func (f *FutureOrder) DoRTCancel() error {
// 	//TODO implement me
// 	panic("implement me")
// }

// func (f *FutureOrder) DoResume(record OrderRecord) {
// 	// TODO implement me
// 	panic("implement me")
// }

// func (f *FutureOrder) CommissionFrozen() float64 {
// 	// TODO implement me
// 	panic("implement me")
// }

// func (f *FutureOrder) DoReject(rejectTime time.Time, err error) {
// 	// TODO implement me
// 	panic("implement me")
// }

// func (f *FutureOrder) Contract() contract.Contract {
// 	// TODO implement me
// 	panic("implement me")
// }

// func (f *FutureOrder) PositionDirection() config.PositionDirection {
// 	// TODO implement me
// 	panic("implement me")
// }

// func (f *FutureOrder) TransactionType() config.TransactionType {
// 	// TODO implement me
// 	panic("implement me")
// }

// func (f *FutureOrder) DoTradeUpdate(
// 	tradePrice float64, tradeQty float64, tradeTime time.Time, recorder recorder.Handler,
// ) {
// 	// TODO implement me
// 	panic("implement me")
// }

// func (f *FutureOrder) DoCancelUpdate(cancelTime time.Time) {
// 	// TODO implement me
// 	panic("implement me")
// }

// func (f *FutureOrder) ID() int64 {
// 	// TODO implement me
// 	panic("implement me")
// }

// func (f *FutureOrder) Account() account.Account {
// 	// TODO implement me
// 	panic("implement me")
// }

// func (f *FutureOrder) IsEligible() bool {
// 	// TODO implement me
// 	panic("implement me")
// }

// func (f *FutureOrder) IsExecuted() bool {
// 	// TODO implement me
// 	panic("implement me")
// }

// func (f *FutureOrder) OrderTime() time.Time {
// 	// TODO implement me
// 	panic("implement me")
// }

// func (f *FutureOrder) OrderPrice() float64 {
// 	// TODO implement me
// 	panic("implement me")
// }

// func (f *FutureOrder) OrderQty() float64 {
// 	// TODO implement me
// 	panic("implement me")
// }

// func (f *FutureOrder) OrderAmt() float64 {
// 	// TODO implement me
// 	panic("implement me")
// }

// func (f *FutureOrder) OrderDirection() config.OrderDirection {
// 	// TODO implement me
// 	panic("implement me")
// }

// func (f *FutureOrder) OrderType() config.OrderType {
// 	// TODO implement me
// 	panic("implement me")
// }

// func (f *FutureOrder) OrderStatus() config.OrderStatus {
// 	// TODO implement me
// 	panic("implement me")
// }

// func (f *FutureOrder) Margin() float64 {
// 	// TODO implement me
// 	panic("implement me")
// }

// func (f *FutureOrder) MarketValue() float64 {
// 	// TODO implement me
// 	panic("implement me")
// }

// func (f *FutureOrder) TradePrice() float64 {
// 	// TODO implement me
// 	panic("implement me")
// }

// func (f *FutureOrder) TradeQty() float64 {
// 	// TODO implement me
// 	panic("implement me")
// }

// func (f *FutureOrder) TradeAmt() float64 {
// 	// TODO implement me
// 	panic("implement me")
// }

// func (f *FutureOrder) TradeTime() time.Time {
// 	// TODO implement me
// 	panic("implement me")
// }

// func (f *FutureOrder) InstID() string {
// 	// TODO implement me
// 	panic("implement me")
// }

// func (f *FutureOrder) Commission() float64 {
// 	// TODO implement me
// 	panic("implement me")
// }

// func (f *FutureOrder) DoSettle(tm time.Time, recorder recorder.Handler) {
// 	// TODO implement me
// 	panic("implement me")
// }
