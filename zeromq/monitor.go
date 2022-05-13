package zeromq

import (
	"fmt"
	zmq "github.com/pebbe/zmq4"
	log "github.com/sirupsen/logrus"
)

func monitorSocket(zctx *zmq.Context, addr, sockType string) {
	s, err := zctx.NewSocket(zmq.PAIR)
	if err != nil {
		log.Fatalf("%s failed to create monitor socket: %s", sockType, err)
		return
	}
	defer func() {
		s.SetLinger(0)
		s.Close()
	}()

	err = s.Connect(addr)
	if err != nil {
		log.Fatalf("%s failed to connect monitor socket: %s", sockType, err)
		return
	}

	for {
		a, b, _, err := s.RecvEvent(0)
		if err != nil {
			log.Errorf("%s failed to receive on monitor socket: %s", sockType, err)
			return
		}
		event := fmt.Sprint(a, " ", b)
		log.Tracef("%s monitor: %s", sockType, event)
		if a == zmq.EVENT_CLOSED {
			break
		}
	}
}
