package main

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"encoding/json"
)

const (
	SHOWROOM_TIMETABLE_API_URL = "https://www.showroom-live.com/api/time_table/time_tables"
)

type TodayPickResponse struct {
	TimeTable []Room `json:"time_tables"`
}
type Room struct {
	RoomID     uint   `json:"room_id"`
	RoomURLKey string `json:"room_url_key"`
	MainName   string `json:"main_name"`
	StartedAt  uint64 `json:"started_at"`
	IsOnlive   bool   `json:"is_onlive"`
}

func GetTodayPick() ([]Room, error) {
	u, _ := url.Parse(SHOWROOM_TIMETABLE_API_URL)
	q := u.Query()
	q.Add("order", "asc")
	q.Add("started_at", strconv.FormatInt(time.Now().Unix(), 10))
	u.RawQuery = q.Encode()

	var todayPickResponse TodayPickResponse

	resp, err := http.Get(u.String())
	if err != nil {
		logger.WithError(err).Error("showroom api error")
		return nil, err
	}

	if err := json.NewDecoder(resp.Body).Decode(&todayPickResponse); err != nil {
		logger.WithError(err).Error("cant decode showroom response")
		return nil, err
	}

	var rooms []Room
	for _, room := range todayPickResponse.TimeTable {
		if strings.Contains(room.MainName, "乃木坂46") {
			rooms = append(rooms, room)
		}
	}

	logger.WithField("rooms", todayPickResponse.TimeTable).Info("today's pick")
	return rooms, nil
}

func extraTitle(title string) (string, string) {
	t := strings.ReplaceAll(title, " ", "")
	t = strings.ReplaceAll(t, "）", "")
	ts := strings.Split(t, "（")
	return ts[0], ts[1]
}

func (r Room) ParseToEvent() Event {
	member, group := extraTitle(r.MainName)
	return Event{
		URL:       "https://www.showroom-live.com/" + r.RoomURLKey,
		Member:    member,
		Group:     group,
		RoomID:    r.RoomID,
		StartAt:   uint64(r.StartedAt),
		CreatedAt: uint64(time.Now().Unix()),
	}
}
