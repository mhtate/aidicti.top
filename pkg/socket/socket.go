package socket

import (
	"math/rand"
	"net"
	"time"

	"aidicti.top/pkg/logging"
	"aidicti.top/pkg/utils"
)

func NewRandomPort() uint64 {
	// random port from free range [9702-9746]
	return uint64(9702) + (rand.New(rand.NewSource(time.Now().UnixNano())).Uint64() % (9746 - 9702 - 1))
}

func NewListener(addrPort string) net.Listener {
	listener, err := net.Listen("tcp", addrPort)
	if err == nil {
		logging.Info("listen started", "addr", addrPort)

		return listener
	}
	utils.Assert(err == nil, "listen failed")

	return listener
}
