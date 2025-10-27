package apiClient2

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
)

type fetchAttempt[reqBody any, respBody any] struct {
	Url     string
	Method  string
	Body    reqBody
	Status  error
	Resolve *FetchResolve[respBody]
}

func (f *fetchAttempt[reqBody, responseBody]) CreateRequest() *http.Request {

	var err error
	var req *http.Request

	req, err = newRequest[reqBody](f.Url, f.Method, f.Body)
	if err != nil {
		f.sendDone()
	}

	defer f.setError(err)

	return req
}

func (f *fetchAttempt[reqBody, responseBody]) ReadResponse(data []byte, statusCode int) {
	if data == nil {
		f.setError(errors.New("response body is nil"))
		f.sendDone()
		return
	}

	err := json.Unmarshal(data, &f.Resolve.Data)
	if err != nil {
		f.sendDone()
		return
	}

	f.Resolve.setStatusCode(statusCode)
	f.setError(err)
}

func newRequest[reqBody any](url, method string, payload reqBody) (*http.Request, error) {

	if method == "" || url == "" {
		return nil, errors.New("invalid request it has no method or url")
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		errStr := fmt.Sprintf("error creating request: %s", err.Error())
		return nil, errors.New(errStr)
	}

	request.Header.Set("Content-Type", "application/json")
	return request, nil
}

// fix
func (f *fetchAttempt[reqBody, responseBody]) validateRespBody() error {
	value := reflect.ValueOf(f.Resolve.Data)

	if !value.IsValid() || (value.Kind() == reflect.Ptr && value.IsNil()) {
		return errors.New("response body is nil")
	}

	if value.Kind() != reflect.Ptr {
		return errors.New("response body is not a pointer")
	}

	if kind := value.Elem().Kind(); kind != reflect.Struct {
		return errors.New("response body is not a struct")
	}

	return nil
}

func (f *fetchAttempt[reqBody, responseBody]) validatePayload() error {

	value := reflect.ValueOf(f.Body)

	if !value.IsValid() || (value.Kind() == reflect.Ptr && value.IsNil()) {
		return errors.New("response body is nil")
	}

	if value.Kind() != reflect.Ptr {
		return errors.New("response body is not a pointer")
	}

	if kind := value.Elem().Kind(); kind != reflect.Struct {
		return errors.New("response body is not a struct")
	}

	return nil
}

func (f *fetchAttempt[reqBody, responseBody]) setError(err error) {
	f.Resolve.Status = err
}

func (f *fetchAttempt[reqBody, responseBody]) sendDone() {
	f.Resolve.done <- struct{}{}
}

func newFetchAttempt[reqBody any, responseBody any](url string, method string, body reqBody, respBody *responseBody) (*fetchAttempt[reqBody, responseBody], error) {

	attempt := &fetchAttempt[reqBody, responseBody]{
		Url:    url,
		Method: method,
		Body:   body,
		Resolve: &FetchResolve[responseBody]{
			Data: respBody,
			done: make(chan struct{}),
		},
	}

	if method != http.MethodGet {
		err := attempt.validatePayload()
		if err != nil {
			return nil, err
		}
	}

	err := attempt.validateRespBody()
	if err != nil {
		return nil, err
	}

	return attempt, err
}

type FetchResolve[respBody any] struct {
	Data       *respBody
	StatusCode int
	Status     error
	done       chan struct{}
}

func (f *FetchResolve[respBody]) setStatusCode(code int) {
	f.StatusCode = code
}

func (f *FetchResolve[respBody]) setData(data *respBody) {
	f.Data = data
}
