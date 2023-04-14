package internal

import (
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	polar "github.com/polarstreams/go-client"
	"github.com/polarstreams/go-client/types"
	"golang.org/x/net/http2"
)

const producerPort = 9251
const producerBinaryPort = 9254

type WorkloadClient interface {
	DoRequest(index int) error
}

type httpClient struct {
	client      *http.Client
	url         *urlIterator
	method      string
	contentType string
	workload    Workload
}

func (c *httpClient) DoRequest(index int) error {
	req, err := http.NewRequest(c.method, c.Url(), c.workload.Body(index))
	if c.contentType != "" {
		req.Header.Add("Content-Type", c.contentType)
	}
	if err != nil {
		panic(err)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}
		return fmt.Errorf(string(body))
	}

	io.Copy(io.Discard, resp.Body)
	return nil
}

func (c *httpClient) Url() string {
	return c.url.next()
}

func NewHttpClient(
	workload Workload,
	maxConnectionsPerHost int,
	protocolInfo string,
	hosts string,
	path string,
	method string,
	contentType string,
) WorkloadClient {
	return &httpClient{
		client:      createHttpClient(maxConnectionsPerHost, protocolInfo),
		url:         newUrlIterator(hosts, producerPort, path),
		method:      method,
		contentType: contentType,
		workload:    workload,
	}
}

func createHttpClient(maxConnectionsPerHost int, protocolInfo string) *http.Client {
	var transport http.RoundTripper
	if protocolInfo != HTTP2 {
		transport = &http.Transport{
			MaxConnsPerHost:     maxConnectionsPerHost,
			MaxIdleConnsPerHost: maxConnectionsPerHost,
		}
	} else {
		transport = &http2.Transport{
			StrictMaxConcurrentStreams: true,
			AllowHTTP:                  true,
			DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
				// Pretend we are dialing a TLS endpoint.
				return net.Dial(network, addr)
			},
			ReadIdleTimeout: 1 * time.Second,
		}
	}

	return &http.Client{
		Transport: transport,
		Timeout:   10 * time.Second,
	}
}

type binaryClient struct {
	client             polar.Producer
	workload           Workload
	partitionKeyGetter func(int) string
}

func NewBinaryClient(w Workload, host string, maxConnectionsPerHost int, ordered bool) WorkloadClient {
	producer := newBinaryProducer(host, maxConnectionsPerHost)
	partitionKeyGetter := func(int) string {
		return ""
	}

	if ordered {
		partitionKeyGetter = func(index int) string {
			return fmt.Sprintf("p%d", index)
		}
	}

	return &binaryClient{
		client:             producer,
		workload:           w,
		partitionKeyGetter: partitionKeyGetter,
	}
}

func (c *binaryClient) DoRequest(index int) error {
	return c.client.Send("test-topic", c.workload.Body(index), c.partitionKeyGetter(index))
}

func newBinaryProducer(host string, maxConnectionsPerHost int) polar.Producer {
	serviceUrl := fmt.Sprintf("polar://%s", host)
	options := types.ProducerOptions{
		FlushThresholdBytes:  0, // Leave at default
		ConnectionsPerBroker: maxConnectionsPerHost,
		Logger:               types.StdLogger,
	}
	producer, err := polar.NewProducerWithOpts(serviceUrl, options)
	if err != nil {
		panic(err)
	}
	return producer
}
