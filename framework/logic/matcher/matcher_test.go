package matcher

import (
	"testing"

	"github.com/wonderstone/QuantKit/config"
	"github.com/wonderstone/QuantKit/framework/entity/handler"
	"github.com/wonderstone/QuantKit/framework/setting"
)

func TestNewMatcher(t *testing.T) {

	setting.RegisterMatcher(new(NextQuote), config.HandlerTypeNextMatch)
	mtch ,_ := setting.NewMatcher(config.HandlerTypeNextMatch,handler.WithStockSlippage(0.1),handler.WithFutureSlippage(0.2))
	// mtch 不为空，说明生成成功
	if mtch == nil {
		t.Error("matcher is nil")
	}

}