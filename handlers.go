package main

import (
	"fmt"
	"net/http"
)

func (app *application) PostPrompt(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.Body)
	// jsonData, err := CreatePromptService(app.ctx)
	// if err != nil {
	// 	http.Error(w, "Error creating prompt", http.StatusInternalServerError)
	// 	return
	// }
	// w.Header().Set("Content-Type", "application/json")
	// w.Write(jsonData)
}
