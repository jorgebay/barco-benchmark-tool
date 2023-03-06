module github.com/jorgebay/polar-benchmark-tool

go 1.19

require (
	github.com/HdrHistogram/hdrhistogram-go v1.1.2
	github.com/google/uuid v1.3.0
	github.com/polarstreams/go-client v0.4.1-0.20230220120450-051b496e6a6c
	golang.org/x/net v0.7.0
)

require (
	github.com/klauspost/compress v1.15.15 // indirect
	golang.org/x/text v0.7.0 // indirect
)

replace github.com/polarstreams/go-client => ../go-client
