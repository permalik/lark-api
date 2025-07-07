package main

import "net/http"

func (app *application) Router() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("127.0.0.1/turn", app.PostPrompt)
	return mux
}
