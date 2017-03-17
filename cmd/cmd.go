package cmd

import (
	"github.com/spf13/cobra"
	"path/filepath"
	"sync"
)

var (
	certPath   string
	keyPath    string
	caPath     string
	endpoints  []string
	pathToDump string
	enableV3   bool
	root       string
	numGets    int
	etcdKey    string
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
		if enableV3 {
			client = NewEtcdV3Client(config)
		} else {
			client = NewEtcdv2Client(config)
		}
		var wg sync.WaitGroup
		wg.Add(numGets)
		for i := 0; i < numGets; i++ {
			go func() {
				defer wg.Done()
				_, err := client.Get(etcdKey)
				if err != nil {
					panic(err)
				}
			}()
		}
		wg.Wait()
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
}
