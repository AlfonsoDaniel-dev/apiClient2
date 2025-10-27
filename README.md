# apiClient2

Small concurrent HTTP client helper that provides a simple worker
pool and typed request/response handling using Go generics.

Features
- Create a pool of HTTP workers with `NewPool(maxWorkers)`.
- Submit typed requests with `NewRequest[Req,Resp](pool, method, url, &req, &resp)`.
- Responses are unmarshalled into the provided response pointer and
  the call blocks until the request completes.

Quick usage

```go
package main

import (
    "fmt"
    "log"

    "github.com/your/module/apiClient2"
)

type MyReq struct {
    Name string `json:"name"`
}

type MyResp struct {
    ID   string `json:"id"`
    Name string `json:"name"`
}

func main() {
    p, err := apiClient2.NewPool(4)
    if err != nil {
        log.Fatal(err)
    }

    req := MyReq{Name: "alice"}
    var resp MyResp

    // Note: resp must be passed as a pointer to a zero-valued value.
    result, err := apiClient2.NewRequest[MyReq, MyResp](p, "POST", "https://example.com/users", req, &resp)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Response: %+v (status code: %d)\n", result, result)
}
```

Important notes
- When calling `NewRequest`, pass a pointer to the response value
  (for example `&MyResp{}`) so the package can unmarshal JSON into it.
- This package uses an internal `FetchResolve` object and a done
  channel; `NewRequest` blocks until the request finishes.
- The pool's work channel is buffered to `maxWorkers`.

Build

Run `go build ./...` in the repository root to compile the package.
