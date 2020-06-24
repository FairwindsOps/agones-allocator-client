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
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"
	"strings"

	pb "agones.dev/agones/pkg/allocation/go"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"k8s.io/klog"

	"github.com/fairwindsops/agones-allocator-client/pkg/ping"
)

// Client is all of the things necessary to build an allocator request
type Client struct {
	// CA is a list of CAs to consider valid for the respsonse
	CA []byte
	// ClientCert is the client certificate to use for gRPC communication
	ClientCert []byte
	// ClientKey is the key corresponding to ClientCert
	ClientKey []byte
	// Endpoints is a map of possible allocators and their corresponding pingServers
	// if there is no ping server for that allocator, then the value is an empty string
	Endpoints map[string]string
	// CheckPing is set to true if there is a corresponding list of ping servers. This
	// will indicate if we need to check ping times before choosing an allocation server
	CheckPing bool
	// Namespace is the namespace of the fleet or set of gameservers we wish to allocate from
	Namespace string
	// Multicluster is a boolean indicating if a multi-cluster request should be made
	Multicluster bool
	// Endpoint is the chose endpoint after checkPing is resolved
	Endpoint string
	// DialOpts is a constructed grpc DialOption that is used to make requests
	DialOpts grpc.DialOption
	// MatchLabels is a map of key/value pairs to send when asking for an allocation
	MatchLabels map[string]string
}

// Allocation is a game server allocation
type Allocation struct {
	Address string
	Port    int32
}

// NewClient builds a new client object
func NewClient(keyFile, certFile, cacertFile, namespace string, multiCluster bool, labelSelector map[string]string, hosts, pingServers []string) (*Client, error) {
	var endpoints = make(map[string]string)
	var checkPing bool

	if pingServers == nil {
		for _, server := range hosts {
			endpoints[server] = ""
		}
		checkPing = false
	} else {
		for _, server := range hosts {
			for _, pingServer := range pingServers {
				endpoints[server] = pingServer
			}
		}
		checkPing = true
	}

	cert, err := ioutil.ReadFile(certFile)
	if err != nil {
		return nil, err
	}
	key, err := ioutil.ReadFile(keyFile)
	if err != nil {
		return nil, err
	}
	cacert, err := ioutil.ReadFile(cacertFile)
	if err != nil {
		return nil, err
	}

	newClient := &Client{
		CA:           cacert,
		ClientCert:   cert,
		ClientKey:    key,
		Endpoints:    endpoints,
		CheckPing:    checkPing,
		Multicluster: multiCluster,
		Namespace:    namespace,
		MatchLabels:  labelSelector,
	}
	err = newClient.setEndpoint()
	if err != nil {
		return nil, err
	}

	if !strings.Contains(newClient.Endpoint, ":443") {
		newClient.Endpoint = fmt.Sprintf("%s:443", newClient.Endpoint)
	}

	klog.V(2).Infof("client endpoint is set to %s", newClient.Endpoint)
	err = newClient.createRemoteClusterDialOption()
	if err != nil {
		return nil, err
	}
	return newClient, nil
}

// createRemoteClusterDialOption creates a grpc client dial option with TLS configuration.
func (c *Client) createRemoteClusterDialOption() error {
	// Load client cert
	cert, err := tls.X509KeyPair(c.ClientCert, c.ClientKey)
	if err != nil {
		return err
	}

	tlsConfig := &tls.Config{Certificates: []tls.Certificate{cert}}
	if len(c.CA) != 0 {
		// Load CA cert, if provided and trust the server certificate.
		// This is required for self-signed certs.
		tlsConfig.RootCAs = x509.NewCertPool()
		if !tlsConfig.RootCAs.AppendCertsFromPEM(c.CA) {
			return errors.New("only PEM format is accepted for server CA")
		}
	}
	dialOpts := grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig))
	c.DialOpts = dialOpts

	return nil
}

// AllocateGameserver allocates a new gamserver
func (c *Client) AllocateGameserver() (*Allocation, error) {
	request := &pb.AllocationRequest{
		Namespace: c.Namespace,
		MultiClusterSetting: &pb.MultiClusterSetting{
			Enabled: c.Multicluster,
		},
		RequiredGameServerSelector: &pb.LabelSelector{
			MatchLabels: c.MatchLabels,
		},
	}

	resp, err := c.makeRequest(request)
	if err != nil {
		return nil, err
	}

	allocation := &Allocation{
		Address: resp.Address,
		Port:    resp.Ports[0].Port,
	}
	return allocation, nil
}

func (c *Client) makeRequest(request *pb.AllocationRequest) (*pb.AllocationResponse, error) {
	conn, err := grpc.Dial(c.Endpoint, c.DialOpts)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	grpcClient := pb.NewAllocationServiceClient(conn)
	response, err := grpcClient.Allocate(context.Background(), request)
	if err != nil {
		return nil, err
	}
	klog.V(2).Infof("response: %s", response.String())

	return response, nil
}

// setEndpoint picks a host from the list that has the lowest ping time
// if checkPing is false, then endpoint is set to the first host in the list
func (c *Client) setEndpoint() error {
	if c.CheckPing {
		traces := []ping.Trace{}
		for server, pingServer := range c.Endpoints {
			klog.V(2).Infof("checking ping for server: %s ping: %s", server, pingServer)
			trace := ping.Trace{
				Host: pingServer,
			}
			err := trace.Run()
			if err != nil {
				klog.V(3).Infof("trace failed on %s - %s", pingServer, err.Error())
				continue
			}
			traces = append(traces, trace)
		}
		if len(traces) < 1 {
			return fmt.Errorf("no traces succeeded, could not find a valid server")
		}
		fastest, err := ping.FastestTrace(traces)
		if err != nil {
			return err
		}
		for host, pingServer := range c.Endpoints {
			if strings.Contains(fastest.Host, pingServer) {
				klog.V(2).Infof("setting fastest endpoint to %s", host)
				c.Endpoint = host
				return nil
			}
		}
	} else {
		for server := range c.Endpoints {
			klog.V(2).Infof("checkPing is false - setting endpoint to first server in list")
			c.Endpoint = server
			return nil
		}
	}
	return fmt.Errorf("unknown error resolving hosts")
}

func isIPV4(ip string) bool {
	if net.ParseIP(ip) == nil {
		klog.V(4).Infof("not a valid ip address - %s", ip)
		return false
	}
	for i := 0; i < len(ip); i++ {
		switch ip[i] {
		case '.':
			klog.V(4).Infof("ip is v4 - %s", ip)
			return true
		case ':':
			klog.V(4).Infof("ip is v6 - %s", ip)
			return false
		}
	}
	return false
}
