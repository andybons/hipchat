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
	baseURL = "https://api.hipchat.com/v1"

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
}

type MessageResponse struct {
	Status string
}

type Client struct {
	AuthToken string
}

func urlValuesFromRequest(req MessageRequest) (url.Values, error) {
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
	uri := fmt.Sprintf("%s/rooms/message?auth_token=%s", baseURL, c.AuthToken)

	payload, err := urlValuesFromRequest(req)
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

	var response MessageResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return err
	}
	if response.Status != ResponseStatusSent {
		return errors.New("PostMessage: response 'status' field was not 'sent'.")
	}

	return nil
}
