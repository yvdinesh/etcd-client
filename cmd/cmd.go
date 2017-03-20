package cmd

import (
	"github.com/spf13/cobra"
	"math/rand"
	"path/filepath"
	"sync"
	"time"
)

var (
	certPath        string
	keyPath         string
	caPath          string
	endpoints       []string
	pathToDump      string
	enableV3        bool
	root            string
	numGets         int
	etcdKey         string
	refreshInterval int
	maxWait         int
)

func mustMakeAbs(path string) string {
	ret, err := filepath.Abs(path)
	if err != nil {
		panic(err)
	}
	return ret
}

var dumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "Dumps the keyspace with key as path and value as contents of the file",
	Run: func(cmd *cobra.Command, args []string) {
		var client EtcdClient
		config := ClientConfig{
			CertPath:  mustMakeAbs(certPath),
			KeyPath:   mustMakeAbs(keyPath),
			EndPoints: endpoints,
			CAPath:    mustMakeAbs(caPath),
		}
		if enableV3 {
			client = NewEtcdV3Client(config)
		} else {
			client = NewEtcdv2Client(config)
		}
		err := client.Dump(root, pathToDump)
		if err != nil {
			panic(err)
		}
	},
}

var getOverloadCmd = &cobra.Command{
	Use:   "get-overload",
	Short: "does how many number of gets you want",
	Run: func(cmd *cobra.Command, args []string) {
		var client EtcdClient
		config := ClientConfig{
			CertPath:  mustMakeAbs(certPath),
			KeyPath:   mustMakeAbs(keyPath),
			EndPoints: endpoints,
			CAPath:    mustMakeAbs(caPath),
		}
		var wg sync.WaitGroup
		wg.Add(numGets)
		if refreshInterval == 0 {
			refreshInterval = numGets
		}
		for g := 0; g < numGets; {
			if enableV3 {
				client = NewEtcdV3Client(config)
			} else {
				client = NewEtcdv2Client(config)
			}
			for i := 0; i < refreshInterval; i++ {
				go func() {
					defer wg.Done()
					time.Sleep(time.Duration(rand.Intn(maxWait)+1) * time.Second)
					_, err := client.Get(etcdKey)
					if err != nil {
						panic(err)
					}
				}()
			}
			wg.Wait()
			err := client.Close()
			if err != nil {
				panic(err)
			}
			g += refreshInterval
		}
	},
}

func init() {
	RootCmd.AddCommand(dumpCmd)
	dumpCmd.Flags().StringVar(&certPath, "cert-path", "", "Path to the etcd certificate")
	dumpCmd.Flags().BoolVar(&enableV3, "enable-v3", false, "Enable v3 client")
	dumpCmd.Flags().StringVar(&keyPath, "key-path", "", "Path to the etcd key")
	dumpCmd.Flags().StringVar(&caPath, "ca-path", "", "Path to the etcd ca")
	dumpCmd.Flags().StringArrayVar(&endpoints, "endpoints", []string{""}, "Endpoints on which etcd is listening")
	dumpCmd.Flags().StringVar(&pathToDump, "destination", "", "Destination directory to which the data will be dumped")
	dumpCmd.Flags().StringVar(&root, "root", "/", "Root from which dump will be taken")

	RootCmd.AddCommand(getOverloadCmd)
	getOverloadCmd.Flags().StringVar(&certPath, "cert-path", "", "Path to the etcd certificate")
	getOverloadCmd.Flags().BoolVar(&enableV3, "enable-v3", false, "Enable v3 client")
	getOverloadCmd.Flags().StringVar(&keyPath, "key-path", "", "Path to the etcd key")
	getOverloadCmd.Flags().StringVar(&caPath, "ca-path", "", "Path to the etcd ca")
	getOverloadCmd.Flags().StringArrayVar(&endpoints, "endpoints", []string{""}, "Endpoints on which etcd is listening")
	getOverloadCmd.Flags().IntVar(&numGets, "numgets", 1, "Number of gets")
	getOverloadCmd.Flags().StringVar(&etcdKey, "etcd-key", "", "Key in etcd to get")
	getOverloadCmd.Flags().IntVar(&refreshInterval, "refresh-interval", 0, "Number of gets after which client will be refreshed")
	getOverloadCmd.Flags().IntVar(&maxWait, "max-wait", 10, "Upper bound on seconds to wait before calling the next get.")
}
