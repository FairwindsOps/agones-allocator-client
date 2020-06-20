package allocator

import (
	"fmt"
	"log"
	"net"
	"time"

	"k8s.io/klog"
)

// TestUDP tests a series of connections to the simple-udp server gameserver example
func (a *Allocation) TestUDP() error {
	endpoint := fmt.Sprintf("%s:%d", a.Address, a.Port)

	conn, err := net.ListenPacket("udp", ":0")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	dst, err := net.ResolveUDPAddr("udp", endpoint)
	if err != nil {
		log.Fatal(err)
	}

	klog.Infof("connected to gameserver and sending hello")

	_, err = conn.WriteTo([]byte("Hello Gameserver!"), dst)
	if err != nil {
		return err
	}
	klog.V(2).Infof("sleeping 10 seconds to view logs")
	time.Sleep(10 * time.Second)
	klog.Info("sending EXIT command")

	_, err = conn.WriteTo([]byte("EXIT"), dst)
	if err != nil {
		return err
	}
	return nil
}
