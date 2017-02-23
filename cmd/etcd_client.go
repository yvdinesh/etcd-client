package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/coreos/etcd/clientv3"
	"github.com/palantir/go-palantir/tlsutils"
	"github.com/palantir/stacktrace"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

type ClientConfig struct {
	CertPath, KeyPath, CAPath string
	EndPoints                 []string
}

type EtcdClient interface {
	Dump(pathToDump string) error
}

type EtcdV3Client struct {
	client *clientv3.Client
}

func NewEtcdV3Client(config ClientConfig) EtcdClient {
	tlsConfig, err := tlsutils.LoadTLSConfig(config.CertPath, config.KeyPath, []string{config.CAPath}, nil)
	if err != nil {
		panic(err)
	}
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   config.EndPoints,
		DialTimeout: 5 * time.Second,
		TLS:         tlsConfig,
	})
	if err != nil {
		panic(err)
	}
	return &EtcdV3Client{client: client}
}

func (c *EtcdV3Client) Dump(pathToDump string) error {
	resp, err := c.client.Get(context.Background(), "/")
	if err != nil {
		return stacktrace.Propagate(err, "")
	}
	for _, kv := range resp.Kvs {
		err = WriteFile(filepath.Join(pathToDump, filepath.FromSlash(string(kv.Key))), kv.Value)
		if err != nil {
			return stacktrace.Propagate(err, "")
		}
	}
	return nil
}

func WriteFile(path string, b []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return stacktrace.Propagate(err, "")
	}

	var out bytes.Buffer

	if json.Indent(&out, b, "", "\t") == nil {
		_ = out.WriteByte('\n')
		b = out.Bytes()
	}

	if err := ioutil.WriteFile(path, b, 0644); err != nil {
		return stacktrace.Propagate(err, "Failed to write file for")
	}
	return nil
}
