package internal

import (
	"strings"
	"sync/atomic"
)

type urlIterator struct {
	values []string
	index  int32
}

func newUrlIterator(urls string) *urlIterator {
	return &urlIterator{
		values: strings.Split(urls, ","),
		index:  0,
	}
}

func (u *urlIterator) next() string {
	i := int(atomic.AddInt32(&u.index, 1)) % len(u.values)
	return u.values[i]
}
