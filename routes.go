package main

import (
	"net/http"
)

func (app *application) Router() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/prompt", app.PostPrompt)
	return mux
}
