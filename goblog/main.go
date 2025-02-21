package main

import (
	"fmt"
	"frame"
	"net/http"
)

func main() {
	engine := frame.New()
	g := engine.Group("user")
	g.Add("/hello", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "hello mszlu.com")
	})
	engine.Run()
}
