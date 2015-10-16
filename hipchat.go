// Package hipchat provides a client library for the Hipchat REST API.
package hipchat

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

const (
	defaultBaseURL = "https://api.hipchat.com/v1"

	ColorYellow = "yellow"
	ColorRed    = "red"
	ColorGreen  = "green"
	ColorPurple = "purple"
	ColorGray   = "gray"
	ColorRandom = "random"

	FormatText = "text"
	FormatHTML = "html"

	ResponseStatusSent = "sent"
)

type MessageRequest struct {
	// Required. ID or name of the room.
	RoomId string

	// Required. Name the message will appear to be sent from. Must be less
	// than 15 characters long. May contain letters, numbers, -, _, and spaces.
	From string

	// Required. The message body. 10,000 characters max.
	Message string

	// Determines how the message is treated by our server and rendered
	// inside HipChat applications.
	// html - Message is rendered as HTML and receives no special treatment.
	// Must be valid HTML and entities must be escaped (e.g.: &amp; instead of &).
	// May contain basic tags: a, b, i, strong, em, br, img, pre, code.
	// Special HipChat features such as @mentions, emoticons, and image previews
	// are NOT supported when using this format.
	// text - Message is treated just like a message sent by a user. Can include
	// @mentions, emoticons, pastes, and auto-detected URLs (Twitter, YouTube, images, etc).
	// (default: html)
	MessageFormat string

	// Whether or not this message should trigger a notification for people
	// in the room (change the tab color, play a sound, etc). Each recipient's
	// notification preferences are taken into account. 0 = false, 1 = true.
	// (default: 0)
	Notify bool

	// Background color for message. One of "yellow", "red", "green",
	// "purple", "gray", or "random".
	// (default: yellow)
	Color string

	// Whether to test authentication. Note: the normal actions will NOT be performed.
	// (default: false)
	AuthTest bool
}

type AuthResponse struct {
	Success, Error *HipchatError
}

type HipchatError struct {
	Code    int
	Type    string
	Message string
}

func (e HipchatError) Error() string {
	return e.Message
}

type ErrorResponse struct {
	Error HipchatError
}

type Client struct {
	AuthToken string
	BaseURL   string
}

// NewClient allocates and returns a Client with the given authToken.
// By default, the client will use the publicly available HipChat servers.
// For internal or custom servers, set the BaseURL field of the Client.
func NewClient(authToken string) Client {
	return Client{AuthToken: authToken, BaseURL: defaultBaseURL}
}

func urlValuesFromMessageRequest(req MessageRequest) (url.Values, error) {
	if len(req.RoomId) == 0 || len(req.From) == 0 || len(req.Message) == 0 {
		return nil, errors.New("The RoomId, From, and Message fields are all required.")
	}
	payload := url.Values{
		"room_id": {req.RoomId},
		"from":    {req.From},
		"message": {req.Message},
	}
	if req.Notify == true {
		payload.Add("notify", "1")
	}
	if len(req.Color) > 0 {
		payload.Add("color", req.Color)
	}
	if len(req.MessageFormat) > 0 {
		payload.Add("message_format", req.MessageFormat)
	}
	return payload, nil
}

func (c *Client) PostMessage(req MessageRequest) error {
	if len(c.BaseURL) == 0 {
		c.BaseURL = defaultBaseURL
	}
	uri := fmt.Sprintf("%s/rooms/message?auth_token=%s", c.BaseURL, url.QueryEscape(c.AuthToken))
	if req.AuthTest {
		uri += "&auth_test=true"
	}

	payload, err := urlValuesFromMessageRequest(req)
	if err != nil {
		return err
	}

	resp, err := http.PostForm(uri, payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if req.AuthTest {
		msgResp := &AuthResponse{}
		if err := json.Unmarshal(body, msgResp); err != nil {
			return err
		}
		if msgResp.Error != nil {
			return msgResp.Error
		}
	} else {
		msgResp := &struct{ Status string }{}
		if err := json.Unmarshal(body, msgResp); err != nil {
			return err
		}
		if msgResp.Status != ResponseStatusSent {
			return getError(body)
		}
	}

	return nil
}

func (c *Client) RoomHistory(id, date, tz string) ([]Message, error) {
	if len(c.BaseURL) == 0 {
		c.BaseURL = defaultBaseURL
	}
	uri := fmt.Sprintf("%s/rooms/history?auth_token=%s&room_id=%s&date=%s&timezone=%s",
		c.BaseURL, url.QueryEscape(c.AuthToken), url.QueryEscape(id), url.QueryEscape(date), url.QueryEscape(tz))

	resp, err := http.Get(uri)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, getError(body)
	}
	msgResp := &struct{ Messages []Message }{}
	if err := json.Unmarshal(body, msgResp); err != nil {
		return nil, err
	}

	return msgResp.Messages, nil
}

func (c *Client) RoomList() ([]Room, error) {
	if len(c.BaseURL) == 0 {
		c.BaseURL = defaultBaseURL
	}
	uri := fmt.Sprintf("%s/rooms/list?auth_token=%s", c.BaseURL, url.QueryEscape(c.AuthToken))

	resp, err := http.Get(uri)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, getError(body)
	}
	roomsResp := &struct{ Rooms []Room }{}
	if err := json.Unmarshal(body, roomsResp); err != nil {
		return nil, err
	}

	return roomsResp.Rooms, nil
}

// getError unmarshals a HipChat error response from the request body and
// returns its error field.
func getError(body []byte) error {
	var errResp ErrorResponse
	if err := json.Unmarshal(body, &errResp); err != nil {
		return err
	}
	return errResp.Error
}
