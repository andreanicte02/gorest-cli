//go:build js && wasm

// Command wasm exposes the gorest client to JavaScript so the same Go code
// that powers the CLI can run in the browser demo.
//
// Every function takes a single JSON string and returns a Promise that
// resolves to a JSON string or rejects with an error message.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"syscall/js"
	"time"

	"github.com/andreanicte02/gorest-cli/gorest"
)

type request struct {
	Token   string      `json:"token"`
	ID      int         `json:"id"`
	Page    int         `json:"page"`
	PerPage int         `json:"perPage"`
	User    gorest.User `json:"user"`
}

func main() {
	api := map[string]any{
		"list": handler(func(ctx context.Context, c *gorest.Client, r request) (any, error) {
			return c.ListUsers(ctx, r.Page, r.PerPage)
		}),
		"get": handler(func(ctx context.Context, c *gorest.Client, r request) (any, error) {
			return c.GetUser(ctx, r.ID)
		}),
		"create": handler(func(ctx context.Context, c *gorest.Client, r request) (any, error) {
			return c.CreateUser(ctx, r.User)
		}),
		"update": handler(func(ctx context.Context, c *gorest.Client, r request) (any, error) {
			return c.UpdateUser(ctx, r.ID, r.User)
		}),
		"delete": handler(func(ctx context.Context, c *gorest.Client, r request) (any, error) {
			return map[string]bool{"deleted": true}, c.DeleteUser(ctx, r.ID)
		}),
	}
	js.Global().Set("gorest", js.ValueOf(api))
	js.Global().Call("dispatchEvent", js.Global().Get("Event").New("gorest-ready"))
	select {} // keep the Go runtime alive
}

type opFunc func(context.Context, *gorest.Client, request) (any, error)

func handler(op opFunc) js.Func {
	return js.FuncOf(func(_ js.Value, args []js.Value) any {
		var input string
		if len(args) > 0 {
			input = args[0].String()
		}
		return newPromise(func() (string, error) {
			var r request
			if input != "" {
				if err := json.Unmarshal([]byte(input), &r); err != nil {
					return "", errors.New("invalid request: " + err.Error())
				}
			}
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()
			res, err := op(ctx, gorest.New(r.Token), r)
			if err != nil {
				return "", err
			}
			out, err := json.MarshalIndent(res, "", "  ")
			return string(out), err
		})
	})
}

// newPromise runs fn in a goroutine (HTTP calls block, which is not allowed on
// the JS event loop) and settles a JS Promise with its result.
func newPromise(fn func() (string, error)) js.Value {
	var executor js.Func
	executor = js.FuncOf(func(_ js.Value, args []js.Value) any {
		resolve, reject := args[0], args[1]
		go func() {
			defer executor.Release()
			out, err := fn()
			if err != nil {
				reject.Invoke(js.Global().Get("Error").New(err.Error()))
				return
			}
			resolve.Invoke(out)
		}()
		return nil
	})
	return js.Global().Get("Promise").New(executor)
}
