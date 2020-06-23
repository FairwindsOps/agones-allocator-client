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
	"io/ioutil"

	pb "agones.dev/agones/pkg/allocation/go"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"k8s.io/klog"
)

// Client is all of the things necessary to build an allocator request
type Client struct {
	CA           []byte
	ClientCert   []byte
	ClientKey    []byte
	Endpoint     string
	Namespace    string
	Multicluster bool
	DialOpts     grpc.DialOption
	MatchLabels  map[string]string
}

// Allocation is a game server allocation
type Allocation struct {
	Address string
	Port    int32
}

// NewClient builds a new client object
func NewClient(keyFile, certFile, cacertFile, externalIP, namespace string, multiCluster bool, labelSelector map[string]string) (*Client, error) {
	endpoint := externalIP + ":443"
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
		Endpoint:     endpoint,
		Multicluster: multiCluster,
		Namespace:    namespace,
		MatchLabels:  labelSelector,
	}
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
