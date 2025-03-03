package server

import (
	"fmt"
	"time"

	"github.com/lib/pq"
)

func (server Server) ListenForChange() error {
	minReconn := 10 * time.Second
	maxReconn := 1 * time.Minute

	listener := pq.NewListener(server.dbConnStr, minReconn, maxReconn, server.reportChange)
	err := listener.Listen("change")
	if err != nil {
		return err
	}

	for {
		notification := <-listener.Notify
		fmt.Println(notification.Extra)
	}
}

func (server Server) reportChange(ev pq.ListenerEventType, err error) {
	if err != nil {
		fmt.Println(err.Error())
	}
}
