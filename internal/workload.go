package internal

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	messageLength = 1024
	totalPayloads = 1024
)

var alphabet []rune = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz")
var loremIpsum = strings.Split("Lorem ipsum dolor sit amet, consectetur adipiscing elit. Donec convallis, metus nec ullamcorper ultrices, eros urna ullamcorper ligula, dignissim ultricies est augue at est. Curabitur non porta urna. Phasellus rutrum, sapien eu pretium pharetra, velit tellus consectetur elit, eget egestas augue sapien ut justo. Vivamus id eros sapien. Nulla non elit tincidunt, laoreet arcu eu, commodo urna. Vivamus tincidunt ligula orci, a tempor velit elementum et. Suspendisse eget ex eu nisi porta molestie. Sed eu dui in mi sollicitudin vestibulum. Pellentesque in ex cursus, lacinia dolor at, accumsan odio. Nunc imperdiet magna fringilla libero ultrices elementum. In a justo sed lectus dapibus lobortis id sed dolor. Maecenas et leo dictum, consequat lacus a, vestibulum tellus. Curabitur gravida, ligula vel eleifend posuere, urna nibh aliquet nulla, nec malesuada nisi mi quis quam. Quisque varius dapibus nisi, quis tempor ligula consectetur ac. Nam ligula augue, finibus maximus pretium gravida, vestibulum a purus. Suspendisse facilisis orci ac lectus iaculis, sed aliquam orci posuere. Quisque a venenatis tellus. Nam lacus massa, auctor vitae dolor sed, pretium pharetra ex. Sed interdum tellus sit amet laoreet pellentesque. Nunc et sem vel leo efficitur sodales. Donec blandit sollicitudin tellus, ut facilisis felis gravida at. Pellentesque cursus nisl sit amet elit pharetra, vitae laoreet mauris posuere. Nam tincidunt nec diam vitae vulputate.", " ")

type Workload interface {
	Url() string
	Method() string
	ContentType() string
	Init()
	Body(index int) io.Reader
	MessagesPerPayload() int
}

func BuildWorkload(name string, url string, messagesPerRequest int) Workload {
	if name == "default" || name == "random" {
		return newRandomWorkload(url, messagesPerRequest)
	}
	if name == "get" {
		return newGetWorkload(url)
	}

	panic(fmt.Sprintf("Workload '%s' not found", name))
}

// Describes a workload with large portion of random data
type randomWorkload struct {
	url                string
	messagesPerRequest int
	payloads           [][]byte
}

func newRandomWorkload(url string, messagesPerRequest int) Workload {
	return &randomWorkload{
		url:                url,
		messagesPerRequest: messagesPerRequest,
		payloads:           make([][]byte, totalPayloads),
	}
}

func (w *randomWorkload) Url() string {
	return w.url
}

func (w *randomWorkload) Method() string {
	return http.MethodPost
}

func (w *randomWorkload) ContentType() string {
	return "application/x-ndjson"
}

func (w *randomWorkload) MessagesPerPayload() int {
	return w.messagesPerRequest
}

func (w *randomWorkload) Init() {
	// Create the values in advance
	const format = `{"id": %d, "sub_id": "%s", "date": "%s", "category": "%s", "wd": "%s", "arr": [-1, %d, %d], "ref": "%s", "sample_bool": %v, "about": "%s", "rnd_text": "`

	// Use some values that follow a pattern (like words from a dictionary), alongside pure random values
	buf := new(bytes.Buffer)
	for i := 0; i < totalPayloads; i++ {
		buf.Reset()
		for j := 0; j < w.messagesPerRequest; j++ {
			if j > 0 {
				buf.WriteRune('\n')
			}
			tid, _ := uuid.NewUUID()
			value := fmt.Sprintf(
				format,
				i,
				tid,
				time.Now().Add(time.Duration(i)*time.Second).Format(time.RFC3339),
				fmt.Sprintf("category-%d", i),
				time.Weekday(i%7),
				1000+i,
				(i+5)%32,
				uuid.New(),
				i%4,
				tokenString(500),
			)

			rem := messageLength - 2 - len(value)
			value += randomString(rem)
			value += `"}`
			buf.WriteString(value)
		}
		w.payloads[i] = buf.Bytes()
	}
}

func (w *randomWorkload) Body(index int) io.Reader {
	return bytes.NewReader(w.payloads[index%totalPayloads])
}

func randomString(n int) string {
	length := len(alphabet)
	var builder strings.Builder
	for i := 0; i < n; i++ {
		ch := alphabet[rand.Intn(length)]
		builder.WriteRune(ch)
	}
	return builder.String()
}

func tokenString(maxLength int) string {
	length := len(loremIpsum)
	var builder strings.Builder
	for {
		ch := loremIpsum[rand.Intn(length)]
		if builder.Len()+len(ch)+1 > maxLength {
			break
		}
		builder.WriteRune(' ')
		builder.WriteString(ch)
	}
	return builder.String()
}

// Sample GET workload
type getWorkload struct {
	url string
}

func newGetWorkload(url string) Workload {
	return &getWorkload{url}
}

func (w *getWorkload) Url() string {
	return w.url
}

func (w *getWorkload) Method() string {
	return http.MethodGet
}

func (w *getWorkload) ContentType() string {
	return ""
}

func (w *getWorkload) MessagesPerPayload() int {
	return 1
}

func (w *getWorkload) Init() {
}

func (w *getWorkload) Body(index int) io.Reader {
	return nil
}
