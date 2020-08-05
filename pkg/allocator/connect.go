/*
Copyright 2020 Fairwinds

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License
*/

package allocator

import (
	"bufio"
	"fmt"
	"net"
	"sync"
	"time"

	"k8s.io/klog"
)

// RunLoad runs many concurrent game connections on a simple UDP or TCP server
// This is designed to test the allocator service and autoscaling of the game servers.
func (c *Client) RunLoad(count int, delay int, duration int, proto string) error {
	var wg sync.WaitGroup
	for i := 0; i < count; i++ {
		wg.Add(1)
		go c.testConnection(i, &wg, duration, proto)
		time.Sleep(time.Duration(delay) * time.Second)
	}
	wg.Wait()
	return nil
}

func (c *Client) testConnection(id int, wg *sync.WaitGroup, duration int, proto string) {
	defer wg.Done()

	a, err := c.AllocateGameserverWithRetry()
	if err != nil {
		klog.Error(err.Error())
		return
	}

	klog.V(3).Infof("%d - got allocation %s %d. Proceeding to connection...\n", id, a.Address, a.Port)
	err = a.testConnection(id, duration, proto)
	if err != nil {
		klog.Error(err)
	}
}

// testConnection tests a series of connections to the simple-udp server gameserver example
func (a *Allocation) testConnection(id int, duration int, proto string) error {
	endpoint := fmt.Sprintf("%s:%d", a.Address, a.Port)

	switch proto {
	case "tcp":
		conn, err := net.Dial("tcp", endpoint)
		if err != nil {
			return err
		}
		klog.V(2).Infof("%d - connected to gameserver and sending hello", id)
		fmt.Fprintf(conn, "HELLO\n")
		status, _ := bufio.NewReader(conn).ReadString('\n')
		klog.V(3).Infof("%d - response: %s", id, status)

		klog.V(3).Infof("%d - sleeping %d seconds to view logs", id, duration)
		time.Sleep(time.Duration(duration) * time.Second)

		klog.V(3).Infof("%d - closing connection", id)
		fmt.Fprintf(conn, "EXIT\n")
		return nil

	case "udp":
		conn, err := net.ListenPacket(proto, ":0")
		if err != nil {
			return err
		}
		defer conn.Close()

		dst, err := net.ResolveUDPAddr(proto, endpoint)
		if err != nil {
			return err
		}

		klog.V(2).Infof("%d - connected to gameserver and sending hello", id)

		// Hello
		msg := fmt.Sprintf("Hello from process %d!", id)
		_, err = conn.WriteTo([]byte(msg), dst)
		if err != nil {
			return err
		}

		// Wait
		klog.V(3).Infof("%d - sleeping %d seconds to view logs", id, duration)
		time.Sleep(time.Duration(duration) * time.Second)

		// Goodbye
		msg = fmt.Sprintf("Goodbye from process %d.", id)
		_, err = conn.WriteTo([]byte(msg), dst)
		if err != nil {
			return err
		}

		klog.V(3).Infof("%d - closing connection", id)
		_, err = conn.WriteTo([]byte("EXIT"), dst)
		if err != nil {
			return err
		}
		return nil
	default:
		return fmt.Errorf("proto must be one of (udp|tcp)")
	}
}
