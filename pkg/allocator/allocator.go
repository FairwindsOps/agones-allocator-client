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
}

// NewClient builds a new client object
func NewClient(keyFile string, certFile string, cacertFile string, externalIP string, namespace string, multiCluster bool) (*Client, error) {
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
func (c *Client) AllocateGameserver() error {
	request := &pb.AllocationRequest{
		Namespace: c.Namespace,
		MultiClusterSetting: &pb.MultiClusterSetting{
			Enabled: c.Multicluster,
		},
	}

	err := c.makeRequest(request)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) makeRequest(request *pb.AllocationRequest) error {
	conn, err := grpc.Dial(c.Endpoint, c.DialOpts)
	if err != nil {
		return err
	}
	defer conn.Close()

	grpcClient := pb.NewAllocationServiceClient(conn)
	response, err := grpcClient.Allocate(context.Background(), request)
	if err != nil {
		return err
	}
	klog.Infof("response: %s", response.String())
	return nil
}
