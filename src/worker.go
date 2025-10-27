package apiClient2

import (
	"io"
	"net/http"
	"time"
)

type worker struct {
	Id         int
	Pool       *pool
	Client     *http.Client
	CancelChan chan struct{}
}

func newWorker(threadPool *pool, id int) *worker {

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
