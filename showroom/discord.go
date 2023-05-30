package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"text/template"
	"time"
)

const (
	message = `本日 Showroom 直播:
成員: {{ .member }}
時間: {{ .startAt }}
URL: {{ .url }}`
)

func renderMessage(event Event) (string, error) {
	loc, _ := time.LoadLocation("Asia/Taipei")
	t, _ := template.New("Discord").Parse(message)
	b := new(bytes.Buffer)

	sMap := make(map[string]string)
	sMap["member"] = event.Member
	sMap["url"] = event.URL
	sMap["startAt"] = time.Unix(int64(event.StartAt), 0).In(loc).Format("2006-01-02 15:04:05")

	if err := t.Execute(b, sMap); err != nil {
		logger.WithError(err).Error("template execute error")
		return "", nil
	}
	return b.String(), nil
}

type Discord struct {
	endpoint string
}

func NewDiscord(endpoint string) (*Discord, error) {
	if _, err := url.Parse(endpoint); err != nil {
		return nil, err
	}
	return &Discord{endpoint: endpoint}, nil
}

func (dc Discord) PostMessage(event Event) error {

	msg, err := renderMessage(event)
	if err != nil {
		logger.WithError(err).Error("render message error")
		return err
	}

	dcMsg := make(map[string]string)
	dcMsg["content"] = msg

	body := new(bytes.Buffer)
	json.NewEncoder(body).Encode(dcMsg)

	resp, err := http.Post(dc.endpoint, "application/json; charset=utf-8", body)
	if err != nil {
		logger.WithField("status_code", resp.StatusCode).Error("discord api return error")
		return err
	}
	return nil
}
