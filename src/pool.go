package apiClient2

import (
	"errors"
	"log"
	"net/http"

	"github.com/google/uuid"
)

type attempter interface {
	sendDone()
	setError(error)
	validatePayload() error
	CreateRequest() *http.Request
	ReadResponse([]byte, int)
}

type pool struct {
	Id         uuid.UUID
	MaxWorkers int
	Workers    map[int]*worker
	WorkChan   chan attempter
}

func (p *pool) addWorkers() {

	for i := 0; i <= p.MaxWorkers; i++ {
		w := newWorker(p, i)

		p.Workers[w.Id] = w

		go w.work()

	}

	return
}

func NewRequest[requestBody any, responseBody any](Pool *pool, method, url string, body requestBody, respBody *responseBody) (*responseBody, error) {

	attempt, err := newFetchAttempt[requestBody, responseBody](url, method, body, respBody)
	log.Println("attempt created")

	if err != nil {
		attempt.setError(err)

		return nil, err
	}

	Pool.WorkChan <- attempt
	log.Println("attempt sent to the work chan")

	<-attempt.Resolve.done
	log.Println("attempt resolved")

	return respBody, err
}

func NewPool(maxWorkers int) (*pool, error) {
	if maxWorkers <= 0 {
		return nil, errors.New("max workers must be greater than 0")
	}
	threadPool := &pool{
		Id:         uuid.New(),
		MaxWorkers: maxWorkers,
		Workers:    make(map[int]*worker),
		WorkChan:   make(chan attempter, maxWorkers),
	}

	threadPool.addWorkers()

	return threadPool, nil

}

func (p *pool) closeWorkers() {

	for _, w := range p.Workers {
		w.CancelChan <- struct{}{}
	}

	close(p.WorkChan)

	log.Fatalf("pool %s closed", p.Id)
}
