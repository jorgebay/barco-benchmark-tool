package internal

import (
	"fmt"
	"strings"
	"sync/atomic"
)

type urlIterator struct {
	values []string
	index  int64
}

func newUrlIterator(hosts string, port int, path string) *urlIterator {
	urls := make([]string, 0)
	for _, h := range strings.Split(hosts, ",") {
		urls = append(urls, fmt.Sprintf("http://%s:%d%s", h, port, path))
	}

	return &urlIterator{
		values: urls,
		index:  0,
	}
}

func (u *urlIterator) next() string {
	i := int(atomic.AddInt64(&u.index, 1)) % len(u.values)
	return u.values[i]
}
