package cmd

import (
	"github.com/spf13/cobra"
	"path/filepath"
)

var (
	certPath   string
	keyPath    string
	caPath     string
	endpoints  []string
	pathToDump string
	enableV3   bool
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
		err := client.Dump(pathToDump)
		if err != nil {
			panic(err)
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
}
