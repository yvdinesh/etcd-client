// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
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
// limitations under the License.

package cmd

import (
	"github.com/spf13/cobra"
)

var (
	certPath   string
	keyPath    string
	caPath     string
	endpoints  []string
	pathToDump string
)

var dumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "Dumps the keyspace with key as path and value as contents of the file",
	Run: func(cmd *cobra.Command, args []string) {
		clientV3 := NewEtcdV3Client(ClientConfig{
			CertPath:  certPath,
			KeyPath:   keyPath,
			EndPoints: endpoints,
			CAPath:    caPath,
		})
		err := clientV3.Dump(pathToDump)
		if err != nil {
			panic(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(dumpCmd)
	dumpCmd.Flags().StringVar(&certPath, "cert-path", "", "Path to the etcd certificate")
	dumpCmd.Flags().StringVar(&keyPath, "key-path", "", "Path to the etcd key")
	dumpCmd.Flags().StringVar(&caPath, "ca-path", "", "Path to the etcd ca")
	dumpCmd.Flags().StringArrayVar(&endpoints, "endpoints", []string{""}, "Endpoints on which etcd is listening")
	dumpCmd.Flags().StringVar(&pathToDump, "destination", "", "Destination directory to which the data will be dumped")
}
