// Copyright 2020 Fairwinds
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License

package cmd

import (
	"flag"
	"fmt"
	"os"

	"github.com/fairwindsops/agones-allocator-client/pkg/allocator"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/klog"
)

var (
	version       string
	versionCommit string
	keyFile       string
	certFile      string
	caCertFile    string
	host          string
	namespace     string
	multicluster  bool
	demoCount     int
	demoDelay     int
	demoDuration  int
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&keyFile, "key", "", "", "The path to the client key file in PEM format")
	rootCmd.PersistentFlags().StringVarP(&certFile, "cert", "", "", "The path the client cert file in PEM format")
	rootCmd.PersistentFlags().StringVar(&caCertFile, "ca-cert", "", "The path the CA cert file in PEM format")
	rootCmd.PersistentFlags().StringVarP(&host, "host", "", "", "The hostname or IP address of the allocator server")
	rootCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", "The namespace of gameservers to request from")
	rootCmd.PersistentFlags().BoolVarP(&multicluster, "multicluster", "m", false, "If true, multicluster allocation will be requested")

	rootCmd.AddCommand(allocateCmd)
	rootCmd.AddCommand(loadTestCmd)

	loadTestCmd.PersistentFlags().IntVarP(&demoCount, "count", "c", 10, "The number of connections to make during the demo.")
	loadTestCmd.PersistentFlags().IntVar(&demoDelay, "delay", 2, "The number of seconds to wait between connections")
	loadTestCmd.PersistentFlags().IntVarP(&demoDuration, "duration", "d", 10, "The number of seconds to leave each connection open.")

	klog.InitFlags(nil)
	flag.Parse()
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	envMap := map[string]string{
		"AGONES_CLIENT_CERT":  "cert",
		"AGONES_CLIENT_KEY":   "key",
		"AGONES_CA_CERT":      "ca-cert",
		"AGONES_HOST":         "host",
		"AGONES_GS_NAMESPACE": "namespace",
	}

	for env, flagName := range envMap {
		flag := rootCmd.PersistentFlags().Lookup(flagName)
		if flag == nil {
			klog.Errorf("Could not find flag %s", flagName)
			continue
		}
		flag.Usage = fmt.Sprintf("%v [%v]", flag.Usage, env)
		if value := os.Getenv(env); value != "" {
			err := flag.Value.Set(value)
			if err != nil {
				klog.Errorf("Error setting flag %v to %s from environment variable %s", flag, value, env)
			}
		}
	}

}

var rootCmd = &cobra.Command{
	Use:   "agones-allocator-client",
	Short: "agones-allocator-client",
	Long:  `A tool to test the agones allocator service`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("You must specify a sub-command.")
		err := cmd.Help()
		if err != nil {
			klog.Error(err)
		}
		os.Exit(1)
	},
}

var allocateCmd = &cobra.Command{
	Use:     "allocate",
	Short:   "allocate",
	Long:    `Request an allocated server`,
	PreRunE: argsValidator,
	Run: func(cmd *cobra.Command, args []string) {
		allocatorClient, err := allocator.NewClient(keyFile, certFile, caCertFile, host, namespace, multicluster)
		if err != nil {
			klog.Error(err)
		}
		allocation, err := allocatorClient.AllocateGameserver()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Printf("Got allocation %s %d\n", allocation.Address, allocation.Port)
	},
}

var loadTestCmd = &cobra.Command{
	Use:     "load-test",
	Short:   "load-test",
	Long:    `Allocates a set of servers, communicates with them, and then closes the connection.`,
	PreRunE: argsValidator,
	Run: func(cmd *cobra.Command, args []string) {
		allocatorClient, err := allocator.NewClient(keyFile, certFile, caCertFile, host, namespace, multicluster)
		if err != nil {
			klog.Error(err)
		}
		err = allocatorClient.RunUDPLoad(demoCount, demoDelay, demoDuration)
	},
}

// Execute the stuff
func Execute(VERSION string, COMMIT string) {
	version = VERSION
	versionCommit = COMMIT
	if err := rootCmd.Execute(); err != nil {
		klog.Error(err)
		os.Exit(1)
	}
}

func fileExists(path string) (bool, error) {
	if _, err := os.Stat(path); err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	} else {
		return false, err
	}
}

func argsValidator(cmd *cobra.Command, args []string) error {
	if namespace == "" {
		return fmt.Errorf("you must specify a namespace")
	}

	if host == "" {
		return fmt.Errorf("host must not be blank")
	}

	exists, err := fileExists(keyFile)
	if !exists {
		return fmt.Errorf("key file %s does not exist", keyFile)
	}
	if err != nil {
		return err
	}

	exists, err = fileExists(caCertFile)
	if !exists {
		return fmt.Errorf("ca cert %s does not exist", caCertFile)
	}
	if err != nil {
		return err
	}

	exists, err = fileExists(certFile)
	if !exists {
		return fmt.Errorf("client cert %s does not exist", certFile)
	}
	if err != nil {
		return err
	}

	return nil
}
