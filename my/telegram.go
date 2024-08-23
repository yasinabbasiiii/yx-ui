// telegram.go
package my

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

var (
	DefaultBotToken = "7373652712:AAGaeU04ZT8b_mLWn9BNx14f6UfEIhvT5GA"
	DefaultChatID   = "502957800"
)

func SendMessage(botToken, chatID, message string) error {
	if botToken == "" {
		botToken = DefaultBotToken
	}
	if chatID == "" {
		chatID = DefaultChatID
	}

	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)

	values := url.Values{}
	values.Set("chat_id", chatID)
	values.Set("text", message)

	resp, err := http.Post(apiURL, "application/x-www-form-urlencoded", strings.NewReader(values.Encode()))
	if err != nil {
		return fmt.Errorf("error sending message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send message. status code: %d", resp.StatusCode)
	}

	return nil
}
