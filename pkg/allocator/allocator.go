package allocator

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io/ioutil"

	pb "agones.dev/agones/pkg/allocation/go"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Run does the thing
func Run(keyFile string, certFile string, cacertFile string, externalIP string, namespace string, multicluster bool) {

	flag.Parse()

	endpoint := externalIP + ":443"
	cert, err := ioutil.ReadFile(certFile)
	if err != nil {
		panic(err)
	}
	key, err := ioutil.ReadFile(keyFile)
	if err != nil {
		panic(err)
	}
	cacert, err := ioutil.ReadFile(cacertFile)
	if err != nil {
		panic(err)
	}

	request := &pb.AllocationRequest{
		Namespace: namespace,
		MultiClusterSetting: &pb.MultiClusterSetting{
			Enabled: multicluster,
		},
	}

	dialOpts, err := createRemoteClusterDialOption(cert, key, cacert)
	if err != nil {
		panic(err)
	}
	conn, err := grpc.Dial(endpoint, dialOpts)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	grpcClient := pb.NewAllocationServiceClient(conn)
	response, err := grpcClient.Allocate(context.Background(), request)
	if err != nil {
		panic(err)
	}
	fmt.Printf("response: %s\n", response.String())
}

// createRemoteClusterDialOption creates a grpc client dial option with TLS configuration.
func createRemoteClusterDialOption(clientCert, clientKey, caCert []byte) (grpc.DialOption, error) {
	// Load client cert
	cert, err := tls.X509KeyPair(clientCert, clientKey)
	if err != nil {
		return nil, err
	}

	tlsConfig := &tls.Config{Certificates: []tls.Certificate{cert}}
	if len(caCert) != 0 {
		// Load CA cert, if provided and trust the server certificate.
		// This is required for self-signed certs.
		tlsConfig.RootCAs = x509.NewCertPool()
		if !tlsConfig.RootCAs.AppendCertsFromPEM(caCert) {
			return nil, errors.New("only PEM format is accepted for server CA")
		}
	}

	return grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)), nil
}
