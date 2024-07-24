package setting

import (
	"fmt"
	"reflect"

	"github.com/wonderstone/QuantKit/config"
	"github.com/wonderstone/QuantKit/framework/entity/tunnel"
)

var tunnelCreator = make(map[config.HandlerType]reflect.Type)

func RegisterTunnelHandler(elem interface{}, name ...config.HandlerType) {
	for _, n := range name {
		tunnelCreator[n] = reflect.TypeOf(elem).Elem()
	}
}

func NewTunnelHandler(name config.HandlerType, options ...tunnel.WithOption) (tunnel.Tunnel, error) {
	if t, ok := tunnelCreator[name]; ok {
		tunnelHandler := reflect.New(t).Interface().(tunnel.Tunnel)
		err := tunnelHandler.Init(options...)
		if err != nil {
			return nil, err
		}

		return tunnelHandler, nil
	}

	support := make([]string, 0, len(tunnelCreator))
	for k := range tunnelCreator {
		support = append(support, string(k))
	}
	return nil, fmt.Errorf("没有找到对应的合约处理器，类型: %s, 目前支持: %v", name, support)
}

func MustNewTunnelHandler(name config.HandlerType, options ...tunnel.WithOption) tunnel.Tunnel {
	f, err := NewTunnelHandler(name, options...)
	if err != nil {
		config.ErrorF(err.Error())
	}
	return f
}

