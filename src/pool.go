package apiClient2

import (
	"errors"
	"log"
	"net/http"

	"github.com/google/uuid"
)

// pool.go provides a simple worker pool that accepts generic fetch
// attempts and processes them concurrently. Users create a pool with
// NewPool and submit requests using NewRequest. The pool dispatches
// attempters to workers via a buffered channel.

type attempter interface {
	sendDone()
	setError(error)
	validatePayload() error
	CreateRequest() *http.Request
	ReadResponse([]byte, int)
}

// attempter defines what a work item must implement to be processed by
// the pool. It is kept unexported because callers should construct
// attempts via the provided helper functions.

type pool struct {
	Id         uuid.UUID
	MaxWorkers int
	Workers    map[int]*worker
	WorkChan   chan attempter
}

// pool represents the thread pool and manages a set of workers and a
// work channel used to hand off attempter tasks.

func (p *pool) addWorkers() {
	// Start workers and register them in the pool. Worker IDs are
	// simple integers and the pool keeps a map for bookkeeping.
	for i := 0; i <= p.MaxWorkers; i++ {
		w := newWorker(p, i)

		p.Workers[w.Id] = w

		go w.work()

	}

	return
}

// NewRequest constructs a typed fetch attempt and submits it to the
// provided pool. The function is generic over the request and response
// body types. The respBody argument should be a pointer to the zero
// value of the expected response (for example: &MyResponse{}). The
// function blocks until the attempt is processed and then returns the
// populated respBody and any error encountered.
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

// NewPool creates a pool with the requested number of workers and a
// buffered work channel sized to the same value. The function validates
// the argument and returns the constructed pool.
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

// closeWorkers signals cancellation to all workers and closes the work
// channel. This is a forceful shutdown; any attempts still in flight
// may be interrupted.
func (p *pool) closeWorkers() {

	for _, w := range p.Workers {
		w.CancelChan <- struct{}{}
	}

	close(p.WorkChan)

	log.Fatalf("pool %s closed", p.Id)
}
