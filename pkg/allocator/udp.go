package allocator

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"k8s.io/klog"
)

// RunUDPDemo runs many concurrent game connections
func (c *Client) RunUDPDemo(count int) error {
	var wg sync.WaitGroup
	for i := 0; i < count; i++ {
		wg.Add(1)
		go c.testUDP(i, &wg)
		time.Sleep(3 * time.Second)
	}
	wg.Wait()
	return nil
}

func (c *Client) testUDP(id int, wg *sync.WaitGroup) {
	defer wg.Done()
	allocation, err := c.AllocateGameserver()
	if err != nil {
		klog.Error(err)
	}
	fmt.Printf("%d - got allocation %s %d. Proceeding to connection...\n", id, allocation.Address, allocation.Port)
	err = allocation.testUDP(id)
	if err != nil {
		klog.Error(err)
	}
}

// testUDP tests a series of connections to the simple-udp server gameserver example
func (a *Allocation) testUDP(id int) error {
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

	klog.Infof("%d - connected to gameserver and sending hello", id)

	_, err = conn.WriteTo([]byte("Hello Gameserver!"), dst)
	if err != nil {
		return err
	}
	klog.V(2).Infof("%d - sleeping 10 seconds to view logs", id)
	time.Sleep(10 * time.Second)
	klog.Infof("%d - sending EXIT command", id)

	_, err = conn.WriteTo([]byte("EXIT"), dst)
	if err != nil {
		return err
	}
	return nil
}
