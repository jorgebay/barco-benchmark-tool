# PolarStreams Benchmarking Tool

Uses HTTP/2 clients to benchmark PolarStreams's producer API.

The load tool works in a similar way as [h2load] but it has the ability to generate a pseudo random payload to
try to mimic production values.

## Building

Use go 1.19+ to build the tool.

```shell
go build .
```

## Usage

```shell
./polar-benchmark-tool -h
Usage of ./polar-benchmark-tool:
  -c int
    	Number of clients (default 1)
  -ch int
    	Connections per host. (default 16)
  -h2
    	For workloads that allow it, use HTTP/2
  -hosts string
    	The host addresses of one of brokers in the cluster, the client will discover the rest of the brokers
  -m int
    	Max concurrent requests to issue per client. (default 32)
  -mr int
    	Number of messages per request in the workload (when supported) (default 16)
  -n int
    	Number of  requests across all  clients (default 100)
  -w string
    	The name of the workload (binary, get, http) (default "binary")
```

### Example

Starting a pseudo random workload with 2 million of messages with 6 clients targeting all hosts in the cluster.

```shell
./polar-benchmark-tool -w binary -hosts 10.0.0.100 -c 6 -n 2000000 -m 1024 -ch 1

Finished. Total responses 2000000 in 1224ms.
Latency in ms p50: 1.8; p999: 6.2 max: 6.2.
Throughput 1633986 messages/s (1633986 req/s)
```

[h2load]: https://nghttp2.org/documentation/h2load-howto.html
