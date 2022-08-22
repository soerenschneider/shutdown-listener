package httphandler

import (
	"errors"
	"github.com/soerenschneider/shutdown-listener/internal"
	"io"
	"net/http"
	"sync"
)

const Name = "http"

type HttpHandler struct {
	msgQueue chan string
	addr     string
	path     string
}

func NewHttpHandler(addr, path string) (*HttpHandler, error) {
	if len(addr) == 0 {
		return nil, errors.New("no addr given")
	}

	if len(path) == 0 {
		return nil, errors.New("no path given")
	}

	return &HttpHandler{addr: addr, path: path}, nil
}

func (handler *HttpHandler) Name() string {
	return Name
}

func (handler *HttpHandler) Shutdown() error {
	return nil
}

func (handler *HttpHandler) Start(queue chan string) error {
	handler.msgQueue = queue

	srv := &http.Server{
		Addr:    handler.addr,
		Handler: handler,
	}

	var err error
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		err = srv.ListenAndServe()
		wg.Done()
	}()

	wg.Wait()
	return err
}

func (handler *HttpHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		internal.MetricHttpRequestErrors.Inc()
	}

	handler.msgQueue <- string(body)
}
