package hipchat

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

type Room struct {
	// The ID of the room.
	RoomId int `json:"room_id"`

	// The name of the room.
	Name string

	// The current room topic.
	Topic string

	// Time of last activity (sent message) in the room in UNIX time (UTC).
	// May be 0 in rare cases when the time is unknown.
	LastActive int `json:"last_active"`

	// Time the room was created in UNIX time (UTC).
	Created int

	// Whether or not this room is archived.
	Archived bool `json:"is_archived"`

	// Whether or not this room is private.
	Private bool `json:"is_private"`

	// User ID of the room owner.
	OwnerUserId int `json:"owner_user_id"`

	// XMPP/Jabber ID of the room.
	XMPPJabberId string `json:"xmpp_jid"`
}

func (c *Client) RoomList() ([]Room, error) {
	uri := fmt.Sprintf("%s/rooms/list?auth_token=%s", baseURL, url.QueryEscape(c.AuthToken))

	resp, err := http.Get(uri)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		var errResp ErrorResponse
		if err := json.Unmarshal(body, &errResp); err != nil {
			return nil, err
		}
		return nil, errors.New(errResp.Error.Message)
	}
	roomsResp := &struct{ Rooms []Room }{}
	if err := json.Unmarshal(body, roomsResp); err != nil {
		return nil, err
	}

	return roomsResp.Rooms, nil
}
