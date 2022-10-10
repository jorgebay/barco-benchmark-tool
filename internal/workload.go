package internal

import (
	"io"
	"net/http"
)

type Workload interface {
	Url() string
	Method() string
	Init()
	Body(index int) io.Reader
}

// Describes a workload with large portion of random data
type randomWorkload struct {
	url string
}

func NewRandomWorkload(url string) Workload {
	return &randomWorkload{url}
}

func (w *randomWorkload) Url() string {
	return w.url
}

func (w *randomWorkload) Method() string {
	return http.MethodGet
}

func (w *randomWorkload) Init() {

}

func (w *randomWorkload) Body(index int) io.Reader {
	return nil
}
