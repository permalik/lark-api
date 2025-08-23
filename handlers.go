package main

import (
	"encoding/json"
	"net/http"
)

func (app *application) PostPrompt(w http.ResponseWriter, r *http.Request) {

	type Prompt struct {
		MsgId   int    `json:"msgId"`
		Content string `json:"content"`
	}
	var prompt Prompt
	err := json.NewDecoder(r.Body).Decode(&prompt)
	if err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	app.logger.Debugw("received prompt:",
		"prompt", prompt,
	)

	cleanedPrompt := Prompt{
		MsgId:   prompt.MsgId,
		Content: prompt.Content,
	}
	promptBytes, err := json.Marshal(cleanedPrompt)
	if err != nil {
		app.logger.Errorw("error marshaling prompt:",
			"err", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	promptString := string(promptBytes)
	app.ProducePromptRaw(promptString)

	app.logger.Infow("produced:",
		"prompt", promptString)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(prompt)
}
