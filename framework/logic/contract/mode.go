package contract

import (
	"fmt"
	"regexp"
	"sync"

	"github.com/wonderstone/QuantKit/config"
	"github.com/wonderstone/QuantKit/framework/entity/contract"
	"github.com/wonderstone/QuantKit/framework/entity/handler"
	"github.com/wonderstone/QuantKit/framework/setting"
)

type Handler struct {
	handler.Resource
	// * 合约信息外部yaml配置，所以设在config包下
	stock  map[string]*config.StockContract
	future map[string]*config.FutureContract
	mutex  sync.RWMutex
}

func (c *Handler) Init(option ...contract.WithOption) (handler.Contract, error) {
	op := contract.NewOp(option...)

	c.stock = make(map[string]*config.StockContract)
	c.future = make(map[string]*config.FutureContract)
	c.mutex = sync.RWMutex{}

	for i := range op.Property.TargetProp.Stock {
		c.stock[op.Property.TargetProp.Stock[i].Name] = op.Property.TargetProp.Stock[i]
	}

	for i := range op.Property.TargetProp.Future {
		c.future[op.Property.TargetProp.Future[i].Name] = op.Property.TargetProp.Future[i]
	}

	return c, nil
}

func (c *Handler) GetContract(instId string) contract.Contract {
	// 判断市场 期货还是股票
	instIdSlice := regexp.MustCompile("[.]").Split(instId, -1)
	if len(instIdSlice) != 3 {
		panic("合约代码格式错误，应该为: 合约代码.交易所Mic码.合约类型，例如：600000.XSHG.CS")
	}

	// 判断合约类型
	if instIdSlice[2] == "CS" {
		prop, ok := c.getStockContract(&instId, &instIdSlice[0], &instIdSlice[1])
		if !ok {
			panic(fmt.Sprintf("未支持的股票合约: %s", instId))
		}

		return &StockContract{
			instId:        instId,
			StockContract: prop,
		}
	} else if instIdSlice[2] == "CF" {
		prop, ok := c.getFutureContract(&instId, &instIdSlice[0], &instIdSlice[1])
		if !ok {
			panic(fmt.Sprintf("未支持的期货合约: %s", instId))
		}

		return &FutureContract{
			instId:         instId,
			FutureContract: prop,
		}
	}

	panic(fmt.Sprintf("未知的合约类型: %s", instIdSlice[2]))
}

// 科创板/创业板股票默认合约
func makeStarDefaultContract(instId *string) (*config.StockContract, bool) {
	return &config.StockContract{
		Name: *instId,
		Basic: config.ContractBasic{
			MinOrderVol:  200,
			ContractSize: 1,
			TickSize:     0.01,
			TPlus:        1,
		},
		StockFee: config.StockFee{
			TransferFeeRate: 0.00001,
			TaxRate:         0.001,
			CommBrokerRate:  0.0003,
			MinFees:         5,
		},
	}, true
}

// 主板股票默认合约
func makeMainBoardDefaultContract(instId *string) (*config.StockContract, bool) {
	return &config.StockContract{
		Name: *instId,
		Basic: config.ContractBasic{
			MinOrderVol:  100,
			ContractSize: 100,
			TickSize:     0.01,
			TPlus:        1,
		},
		StockFee: config.StockFee{
			TransferFeeRate: 0.00001,
			TaxRate:         0.001,
			CommBrokerRate:  0.0003,
			MinFees:         5,
		},
	}, true
}

// ETF默认合约
func makeETFDefaultContract(instId *string) (*config.StockContract, bool) {
	return &config.StockContract{
		Name: *instId,
		Basic: config.ContractBasic{
			MinOrderVol:  100,
			ContractSize: 100,
			TickSize:     0.001,
			TPlus:        1,
		},
		StockFee: config.StockFee{
			TransferFeeRate: 0.0,
			TaxRate:         0.0,
			CommBrokerRate:  0.00005,
			MinFees:         5,
		},
	}, true
}

// 合约代码格式: 合约代码.交易所.合约类型
// code 为合约代码
// market 为交易所mic代码
func (c *Handler) getStockContract(instId, code, market *string) (*config.StockContract, bool) {
	var prop *config.StockContract = nil
	// 如果配置指定了合约具体的属性，则使用指定的属性, 加锁防止并发
	c.mutex.RLock()
	v, ok := c.stock[*instId]

	if ok {
		c.mutex.RUnlock()
		return v, ok
	}

	// TODO 否则使用合约特殊类型的属性
	switch (*code)[:3] {
	case "688":
		prop, ok = c.stock["star"]
		if !ok {
			prop, ok = makeStarDefaultContract(instId)
		}
	case "300":
		prop, ok = c.stock["chi-next"]
		if !ok {
			prop, ok = makeStarDefaultContract(instId)
		}
	case "159", "510":
		prop, ok = c.stock["etf"]
		if !ok {
			prop, ok = makeETFDefaultContract(instId)
		}
	default:
		switch (*code)[:2] {
		case "60", "00":
			prop, ok = c.stock["main-board"]
			if !ok {
				prop, ok = makeMainBoardDefaultContract(instId)
			}
		default:
			prop, ok = makeMainBoardDefaultContract(instId)
		}

	}

	c.mutex.RUnlock()

	c.mutex.Lock()
	// 找到一次之后就直接建立缓存，下次就不用再找了
	c.stock[*instId] = prop
	c.mutex.Unlock()
	return prop, ok
}

func (c *Handler) getFutureContract(instId, code, market *string) (*config.FutureContract, bool) {
	if v, ok := c.future[*instId]; ok {
		return v, ok
	}

	// TODO 期货未完全实现
	panic(fmt.Sprintf("未知的期货合约: %s", *instId))
}

func init() {
	setting.RegisterContractHandler(
		&Handler{},
		config.HandlerTypeDefault,
		config.HandlerTypeConfig,
	)
}

