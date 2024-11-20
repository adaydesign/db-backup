package handlers

import (
	"db-backup/models"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

func PostMessage(results []models.ResultMessage) error {
	url := os.Getenv("DISCORD_WEBHOOK")
	botName := os.Getenv("DISCORD_BOT_NAME")
	botAvatar := os.Getenv("DISCORD_BOT_AVATAR")
	content := "backup result of date : " + time.Now().Format("2006-01-02 15:04:05")
	method := "POST"

	//make result message
	embeds := ""
	for i, v := range results {
		color := "15599624" // error
		if v.Success {
			color = "456521" // success
		}
		a := fmt.Sprintf(`{
		"title": "%s",
		"description": "%s",
		"color":%s
		}`, v.ServerName, v.Message, color)

		embeds = embeds + a
		if i < len(results)-1 {
			embeds = embeds + ","
		}
	}

	payloadStr := fmt.Sprintf(`{
	"username": "%s",
	"avatar_url": "%s",
	"content": "%s",
	"embeds": [%s]
	}`, botName, botAvatar, content, embeds)
	payload := strings.NewReader(payloadStr)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	_, err = io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	return nil
}
