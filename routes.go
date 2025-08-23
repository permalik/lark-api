package main

import (
	"net/http"

	"github.com/rs/cors"
)

func (app *application) Router() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/prompt", app.PostPrompt)
	handler := cors.Default().Handler(mux)
	return handler
}
