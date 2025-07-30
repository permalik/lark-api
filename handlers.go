package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (app *application) PostPrompt(w http.ResponseWriter, r *http.Request) {
	type Prompt struct {
		Id      int    `json:"id"`
		Content string `json:"content"`
	}
	var prompt Prompt
	err := json.NewDecoder(r.Body).Decode(&prompt)
	if err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	fmt.Printf("Received prompt: %+v\n", prompt)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(prompt)
}
