package fake

import (
	"fmt"
	"net/url"
	"time"

	"github.com/wonderstone/QuantKit/config"
	"github.com/wonderstone/QuantKit/framework/entity/tunnel"
	q "github.com/wonderstone/QuantKit/framework/logic/quote"

	"github.com/wonderstone/QuantKit/framework/setting"
)

// & Order related Request struct
type OrderRequest struct {
	RequestID   string  `json:"request_id"`
	InstID      string  `json:"inst_id"`
	Price       float64 `json:"price"`
	Volume      int     `json:"volume"`
	Direction   string  `json:"direction"`
	CashAccount string  `json:"cash_account"`
	OrderType   string  `json:"order_type"`
}

type CancelOrderRequest struct {
	InstID       string `json:"inst_id"`
	CashAccount  string `json:"cash_account"`
	LocalOrderID string `json:"local_order_id"`
}

// & FakeTunnel is a fake implementation of the Tunnel interface
type FakeTunnel struct {
	quoteChan   chan tunnel.Tick
	tradeSource string // ip:port
	op          config.Runtime
}

func (k *FakeTunnel) Init(options ...tunnel.WithOption) error {
	op := tunnel.NewOp(options...)

	k.op = *op.Runtime

	if !op.TradeOnly {
		k.quoteChan = op.TickChannel
		go func() {
			ticker := time.NewTicker(time.Second * 10)
			defer ticker.Stop()
			// maybe add somefunc to ping the tunnel make sure it's alive

			for {
				select {
				case <-ticker.C:
					// maybe add somefunc to ping the tunnel make sure it's alive
				}
			}
		}()
	}

	if !op.QuoteOnly {
		tunnelUrl, err := url.Parse(op.TunnelUrl)
		if err != nil {
			config.ErrorF("解析交易通道地址失败: %v", err)
			return err
		}
		k.tradeSource = tunnelUrl.Host
	}

	return nil
}

func (k *FakeTunnel) RegTick(inst []string) error {
	// let's fake a market data source
	// build a market
	var market q.TmpMarket
	// initialize the market with data already
	market.Init("./testdata", "30min", "510300.XSHG.CS")

	// subscribe to the market
	k.quoteChan = market.Subscribe()
	go market.Publish()
	for d := range k.quoteChan {
		fmt.Println(d.Time(), d.Close())
	}

	// & fake market, no fail
	var err error
	if err != nil {
		config.ErrorF("订阅失败: %v", err)
		return err
	}

	// wait for the market to shutdown
	market.Sim.WaitForShutdown()

	return nil
}


func (k *FakeTunnel) PlaceOrder(order tunnel.Order) (orderID string, err error) {
	// TODO implement me
	// logic for placing order
	fmt.Println("Placing order with fake tunnel..." )
	// get orderID from trade source
	// some fancy logic related the OrderRequest struct
	return orderID, err
}

func (k *FakeTunnel) CancelOrder(orderID string) error {
	// TODO implement me
	// some fancy logic related the CancelOrderRequest struct

	return nil
}







func init() {
	setting.RegisterTunnelHandler(&FakeTunnel{}, config.HandlerTypeDefault)
}
