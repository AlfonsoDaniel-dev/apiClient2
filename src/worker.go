package apiClient2

import (
	"io"
	"net/http"
	"time"
)

// worker.go contains the worker implementation used by the pool to
// process attempter tasks. Each worker owns an http.Client and listens
// for work or cancellation signals.

type worker struct {
	Id         int
	Pool       *pool
	Client     *http.Client
	CancelChan chan struct{}
}

// worker represents a single worker that pulls tasks from the pool's
// work channel and executes them using an http.Client.

func newWorker(threadPool *pool, id int) *worker {
	// Create an HTTP client with a sensible timeout for API calls.
	client := &http.Client{}

	client.Timeout = time.Second * 10

	return &worker{
		Id:     id,
		Pool:   threadPool,
		Client: client,
	}
}

func (w *worker) work() {
	for {
		select {
		case attempt := <-w.Pool.WorkChan:

			// Build and execute the request, then deliver the response
			// to the attempt. If callApi returns an error the attempt's
			// error is set; otherwise the raw bytes and status code are
			// passed to ReadResponse which handles unmarshalling.
			req := attempt.CreateRequest()
			data, statusCode, err := w.callApi(req)
			if err != nil {
				attempt.setError(err)
			} else {
				attempt.ReadResponse(data, statusCode)
			}

			attempt.sendDone()

		case <-w.CancelChan:
			close(w.CancelChan)
			return
		}
	}
}

func (w *worker) callApi(req *http.Request) ([]byte, int, error) {

	var body []byte

	resp, err := w.Client.Do(req)
	if err != nil {
		return nil, 0, err
	}

	defer resp.Body.Close()

	body, err = io.ReadAll(resp.Body)

	return body, resp.StatusCode, err
}

// callApi executes the provided request using the worker's http.Client
// and returns the response bytes and HTTP status code. Any I/O error
// encountered while reading the response body is returned to the
// caller.
