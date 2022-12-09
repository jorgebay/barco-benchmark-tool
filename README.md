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
  -m int
    	Max concurrent requests to issue per client (default 32)
  -mr int
    	Number of messages per request in the workload (when supported) (default 16)
  -n int
    	Number of  requests across all  clients (default 100)
  -u string
    	The uri(s) of the endpoint(s)
  -w string
    	The name of the workload (default "default")
```

### Example

Starting a pseudo random workload with 1 million of post requests with 64 messages per request targeting 3 brokers.

```shell
./polar-benchmark-tool -c 32 -n 1000000 -m 16 -mr 64 -ch 16 \
    -u http://10.0.0.100:9251/v1/topic/a-topic/messages,http://10.0.0.101:9251/v1/topic/a-topic/messages,http://10.0.0.102:9251/v1/topic/a-topic/messages

Finished. Total responses 1000000 in 44212ms.
Latency in ms p50: 10.3; p999: 34.9 max: 37.5.
Throughput 723744 messages/s (22617 req/s)
```

[h2load]: https://nghttp2.org/documentation/h2load-howto.html
