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

package cmd

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/fairwindsops/agones-allocator-client/pkg/allocator"
	"github.com/fairwindsops/agones-allocator-client/pkg/ping"

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
	hosts         []string
	pingServers   []string
	namespace     string
	multicluster  bool
	demoCount     int
	demoDelay     int
	demoDuration  int
	labelSelector map[string]string
	pingTargets   []string
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&keyFile, "key", "", "", "The path to the client key file in PEM format")
	rootCmd.PersistentFlags().StringVarP(&certFile, "cert", "", "", "The path the client cert file in PEM format")
	rootCmd.PersistentFlags().StringVar(&caCertFile, "ca-cert", "", "The path the CA cert file in PEM format")
	rootCmd.PersistentFlags().StringSliceVar(&hosts, "hosts", nil, "A list of possible allocation servers.")
	rootCmd.PersistentFlags().StringSliceVar(&pingServers, "ping-servers", nil, "A list of ping servers that correspond to the allocation servers. If nil, then a random server is chosen.")
	rootCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", "The namespace of gameservers to request from")
	rootCmd.PersistentFlags().BoolVarP(&multicluster, "multicluster", "m", false, "If true, multicluster allocation will be requested")
	rootCmd.PersistentFlags().StringToStringVar(&labelSelector, "labels", nil, "A map of labels to match on the allocation.")

	rootCmd.AddCommand(allocateCmd)
	rootCmd.AddCommand(loadTestCmd)

	loadTestCmd.PersistentFlags().IntVarP(&demoCount, "count", "c", 10, "The number of connections to make during the demo.")
	loadTestCmd.PersistentFlags().IntVar(&demoDelay, "delay", 2, "The number of seconds to wait between connections")
	loadTestCmd.PersistentFlags().IntVarP(&demoDuration, "duration", "d", 10, "The number of seconds to leave each connection open.")

	rootCmd.AddCommand(pingTestCmd)
	pingTestCmd.PersistentFlags().StringSliceVarP(&pingTargets, "targets", "t", nil, "The list of targets to ping.")

	klog.InitFlags(nil)
	flag.Parse()
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	envMap := map[string]string{
		"AGONES_CLIENT_CERT":  "cert",
		"AGONES_CLIENT_KEY":   "key",
		"AGONES_CA_CERT":      "ca-cert",
		"AGONES_HOSTS":        "hosts",
		"AGONES_PING_SERVERS": "ping-servers",
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
		allocatorClient, err := allocator.NewClient(keyFile, certFile, caCertFile, namespace, multicluster, labelSelector, hosts, pingServers)
		if err != nil {
			klog.Fatal(err)
		}
		allocation, err := allocatorClient.AllocateGameserver()
		if err != nil {
			klog.Fatal(err)
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
		allocatorClient, err := allocator.NewClient(keyFile, certFile, caCertFile, namespace, multicluster, labelSelector, hosts, pingServers)
		if err != nil {
			klog.Fatal(err)
		}
		err = allocatorClient.RunUDPLoad(demoCount, demoDelay, demoDuration)
	},
}

var pingTestCmd = &cobra.Command{
	Use:   "ping-test",
	Short: "ping-test",
	Long:  `Pings a list of ping servers and prints out their response and response time.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if pingTargets == nil {
			return fmt.Errorf("You must pass a list of target hostanmes or IP addresses")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		results := []ping.Trace{}
		for _, target := range pingTargets {
			trace := ping.Trace{
				Host: target,
			}
			err := trace.Run()
			if err != nil {
				klog.Fatal(err)
			}
			results = append(results, trace)
		}
		output, err := json.MarshalIndent(results, "", "  ")
		if err != nil {
			klog.Fatal(err)
		}
		fmt.Println(string(output))
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

	if hosts == nil {
		return fmt.Errorf("hosts must not be empty")
	}

	if len(pingServers) > 0 {
		if len(pingServers) != len(hosts) {
			return fmt.Errorf("if passing ping-servers, the length of hosts and ping-servers must be equal")
		}
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
