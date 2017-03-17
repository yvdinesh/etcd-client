package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	clientv2 "github.com/coreos/etcd/client"
	"github.com/coreos/etcd/clientv3"
	"github.com/palantir/go-palantir/tlsutils"
	"github.com/palantir/stacktrace"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type ClientConfig struct {
	CertPath, KeyPath, CAPath string
	EndPoints                 []string
}

type EtcdClient interface {
	Dump(root, pathToDump string) error
	Get(key string) (string, error)
}

type EtcdV3Client struct {
	client *clientv3.Client
}

type EtcdV2Client struct {
	client clientv2.KeysAPI
}

func NewEtcdv2Client(config ClientConfig) EtcdClient {
	tlsConfig, err := tlsutils.LoadTLSConfig(config.CertPath, config.KeyPath, []string{config.CAPath}, nil)
	if err != nil {
		panic(err)
	}
	client, err := clientv2.New(clientv2.Config{
		Endpoints: config.EndPoints,
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).Dial,
			TLSHandshakeTimeout: 10 * time.Second,
			TLSClientConfig:     tlsConfig,
			MaxIdleConnsPerHost: 64,
		},
	})
	if err != nil {
		panic(err)
	}
	return &EtcdV2Client{client: clientv2.NewKeysAPI(client)}
}

func (c *EtcdV2Client) Dump(root, pathToDump string) error {
	resp, err := c.client.Get(context.Background(), root, &clientv2.GetOptions{Recursive: true})
	if err != nil {
		return stacktrace.Propagate(err, "")
	}
	if resp == nil {
		return stacktrace.Propagate(err, "empty response from etcd")
	}
	return dump(resp.Node, pathToDump)
}

func (c *EtcdV2Client) Get(key string) (string, error) {
	resp, err := c.client.Get(context.Background(), key, nil)
	if err != nil {
		return "", stacktrace.Propagate(err, "")
	}
	return resp.Node.Value, nil
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

func (c *EtcdV3Client) Dump(root, pathToDump string) error {
	resp, err := c.client.Get(context.Background(), root, clientv3.WithPrefix())
	if err != nil {
		return stacktrace.Propagate(err, "")
	}
	fmt.Printf("number of keys to dump:%v\n", len(resp.Kvs))
	for _, kv := range resp.Kvs {
		err = writeFile(filepath.Join(pathToDump, filepath.FromSlash(string(kv.Key))), kv.Value)
		if err != nil {
			return stacktrace.Propagate(err, "")
		}
	}
	return nil
}

func (c *EtcdV3Client) Get(key string) (string, error) {
	resp, err := c.client.Get(context.Background(), key, nil)
	if err != nil {
		return "", stacktrace.Propagate(err, "")
	}
	if len(resp.Kvs) == 0 {
		return "", stacktrace.NewError("No key found")
	}
	return string(resp.Kvs[0].Value), nil
}

func dump(node *clientv2.Node, pathToDump string) error {
	// Reached the end of the tree
	if node == nil {
		return nil
	}

	// Reached a file
	if !node.Dir {
		return writeFile(filepath.Join(pathToDump, filepath.FromSlash(node.Key)), []byte(node.Value))
	}

	for _, child := range node.Nodes {
		if err := dump(child, pathToDump); err != nil {
			return stacktrace.Propagate(err, "Failed to dump %q", child.Key)
		}
	}
	return nil
}
func writeFile(path string, b []byte) error {
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
