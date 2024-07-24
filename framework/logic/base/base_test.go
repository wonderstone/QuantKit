package base

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/wonderstone/QuantKit/config"
	"github.com/wonderstone/QuantKit/framework/entity/base"
	"github.com/wonderstone/QuantKit/framework/setting"
)

// 测试base
func TestBase(t *testing.T) {

	pth := config.NewPath(
		config.CalcMode,
		config.WithBaseDir("./"))

	setting.RegisterBase(&FullLoader{}, config.HandlerTypeFullLoad)

	resource := setting.NewResource(
		setting.WithBaseHandler(nil),
	)

	baseHDL := setting.MustBaseNewHandler(
		config.HandlerTypeFullLoad,
		resource,
		base.WithPath(pth))

	baseHDL.Init(resource, base.WithPath(pth))
	// 如果baseHDL不为空，说明生成成功
	require.NotNil(t, baseHDL)


	// get 2020.06.02T08:30:00.000,2020.06.01T16:00:00.000 

	forma := "2006-01-02T15:04:05.000"
	tm,_ := time.Parse(forma,"2020.06.02T08:30:00.000")
	tm1, _ := time.Parse(forma,"2020.06.01T16:00:00.000")
	tmp ,_:= baseHDL.GetXrxd("603908.XSHG.CS", tm)
	tmp1 ,_:= baseHDL.GetXrxd("603908.XSHG.CS", tm1)
	fmt.Println(tmp)
	fmt.Println(tmp1)
	// 如果tmp不为空，说明生成成功
	require.NotNil(t, tmp)
	require.NotNil(t, tmp1)

}
