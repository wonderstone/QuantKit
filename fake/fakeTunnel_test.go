package fake

// test the fake tunnel

import (
	"testing"

	"github.com/wonderstone/QuantKit/config"
	"github.com/wonderstone/QuantKit/framework/entity/tunnel"
	"github.com/wonderstone/QuantKit/framework/setting"
)

func TestFakeTunnel(t *testing.T) {
	fake := setting.MustNewTunnelHandler(
		config.HandlerTypeDefault,
		tunnel.QuoteOnly(),
		tunnel.WithConfig(&config.Runtime{}),
	
	)

	// fake should not be nil
	if fake == nil {
		t.Error("fake tunnel handler is nil")
	}

	// run fake quote part
	fake.RegTick([]string{"not important for this test"})

	// run fake trade part
	fake.PlaceOrder(nil)
	
	// run fake cancel part
	fake.CancelOrder("fake order id")



}
