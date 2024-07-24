package position

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/wonderstone/QuantKit/config"
	"github.com/wonderstone/QuantKit/framework/entity/contract"
	_ "github.com/wonderstone/QuantKit/framework/logic/contract"
	"github.com/wonderstone/QuantKit/framework/setting"
)

// 测试生成新的持仓

func TestNewPosition(t *testing.T) {
	
	
	conf, err := config.NewContractPropertyConfig("./contract.yaml")
	require.NoError(t, err)
	fmt.Println(conf)
	wp := contract.WithProperty(conf)

	handle, _ := setting.NewContractHandler(config.HandlerTypeConfig, wp)
	contr := handle.GetContract("000001.XSHE.CS")
	// 偷懒做法 没有去生成account  但尼玛确实不影响
	position := NewPosition(nil,contr,100,10.1,config.PositionLong)

	// 如果position不为空，说明生成成功
	require.NotNil(t, position)
}