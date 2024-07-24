package main

import (
	"github.com/wonderstone/QuantKit/framework"
	"github.com/wonderstone/QuantKit/framework/entity/handler"
	"github.com/wonderstone/QuantKit/strategy"
	// "fmt"
	// "log"
	// "time"
	// "github.com/wonderstone/QuantKit/ctp"
	// "github.com/wonderstone/QuantKit/ctp/thost"
)

// type CTPInfo struct {
// 	BrokerID    string
// 	InvestorID  string
// 	Password    string
// 	ClientAppID string
// 	AppID       string
// 	AuthCode    string
// }

// var ctpInfo = CTPInfo{
// 	BrokerID:    "1080",
// 	InvestorID:  "902100188",
// 	Password:    "tieqi811001",
// 	ClientAppID: "",
// 	AppID:       "client_wtq_1.0",
// 	AuthCode:    "4URBICAH7VTABP3X",
// }

// var addressMap = map[string]string{
// 	"LT01": "tcp://140.206.101.109:42213",
// 	"LT02": "tcp://140.206.101.110:42213",
// 	"LT03": "tcp://140.207.168.9:42213",
// 	"LT04": "tcp://140.207.168.10:42213",
// 	"LT05": "tcp://111.205.217.41:42213",
// 	"LT06": "tcp://111.205.217.40:42213",
// 	"DX01": "tcp://180.169.112.52:42213",
// 	"DX02": "tcp://180.169.112.53:42213",
// 	"DX03": "tcp://180.169.112.54:42213",
// 	"DX04": "tcp://180.169.112.55:42213",
// 	"DX05": "tcp://106.37.231.6:42213",
// 	"DX06": "tcp://106.37.231.7:42213",
// }

// type baseSpi struct {
// 	ctp.BaseMdSpi
// 	// ctp.BaseMdSpi
// 	mdapi thost.MdApi
// }

// func CreateBaseSpi() *baseSpi {
// 	s := &baseSpi{}

// 	s.OnFrontConnectedCallback = func() {
// 		log.Printf("OnFrontConnected\n")

// 		loginR := &thost.CThostFtdcReqUserLoginField{}
// 		copy(loginR.BrokerID[:], ctpInfo.BrokerID)
// 		// copy(loginR.BrokerID[:], "9999")
// 		// use ctpInfo InvestorId as UserID
// 		copy(loginR.UserID[:], ctpInfo.InvestorID)
// 		// copy(loginR.UserID[:], "2011")
// 		copy(loginR.Password[:], ctpInfo.Password)

// 		ret := s.mdapi.ReqUserLogin(loginR, 1)

// 		log.Printf("user log: %v\n", ret)
// 	}
// 	s.OnFrontDisconnectedCallback = func(nReason int) {
// 		log.Printf("OnFrontDisconnected: %v\n", nReason)
// 	}
// 	s.OnRspUserLoginCallback = func(pRspUserLogin *thost.CThostFtdcRspUserLoginField, pRspInfo *thost.CThostFtdcRspInfoField, nRequestID int, bIsLast bool) {
// 		log.Printf("RspUserLogin: %+v\nRspInfo: %+v\n", pRspUserLogin, nil)
// 		s.mdapi.SubscribeMarketData("rb2410")
// 	}
// 	s.OnRtnDepthMarketDataCallback = func(pDepthMarketData *thost.CThostFtdcDepthMarketDataField) {
// 		// log.Printf all data
// 		// log.Printf("OnRtnDepthMarketData: %+v\n", pDepthMarketData)
// 		fmt.Println("LastPrice:", float64(pDepthMarketData.LastPrice))
// 		fmt.Println("TradingDay:", pDepthMarketData.TradingDay.String())
// 		fmt.Println("UpdateTime:", pDepthMarketData.UpdateTime.String())
// 		fmt.Println("UpdateMillisec:", int32(pDepthMarketData.UpdateMillisec))
// 		fmt.Println("ActionDay:", pDepthMarketData.ActionDay.String())
// 		fmt.Println("InstrumentID:", string(pDepthMarketData.InstrumentID.String()))
// 	}
// 	return s

// }

// func sample() {
// 	mdapi := ctp.CreateMdApi(ctp.MdFlowPath("./data/"), ctp.MdUsingUDP(false), ctp.MdMultiCast(false))
// 	baseSpi := CreateBaseSpi()
// 	baseSpi.mdapi = mdapi
// 	mdapi.RegisterSpi(baseSpi)
// 	// use addressMap["LT01"] for test
// 	mdapi.RegisterFront(addressMap["LT01"])
// 	// mdapi.RegisterFront("tcp://140.206.244.33:11616")
// 	mdapi.Init()

// 	println(mdapi.GetApiVersion())
// 	println(mdapi.GetTradingDay())

// 	// print out the response data sended from the server

// 	mdapi.Join()

// 	// mdapi.Join()
// 	for {
// 		time.Sleep(10 * time.Second)
// 	}

// }













func main() {

	framework.Run(
		framework.WithStrategyCreator(
			func() handler.Strategy {
				// return &strategy.SortBuy{}
				// return &strategy.DMT{}
				// return &strategy.DMTGS{}
				// return &strategy.CQ{}
				// return &strategy.VS{}
				// return &strategy.VS02{}
				// return &strategy.VS03{}
				// return &strategy.VS04{}
				return &strategy.VSO{}
			},
		),
	)
}

