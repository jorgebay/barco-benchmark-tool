# Barco Benchmarking Tool

Uses HTTP/2 clients to benchmark Barco's producer API.

```shell
go build .
./barco-benchmark-tool --help
```

The load tool works in a similar way as [h2load] but it has the ability to generate a pseudo random payload to
try to mimic production values.

[h2load]: https://nghttp2.org/documentation/h2load-howto.html
