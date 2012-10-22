Hipchat
=====
This project implements a [Go](http://golang.org) client library for the [Hipchat API](https://www.hipchat.com/docs/api/).

Pull requests are welcome as <b>the only supported call right now is posting a message to a room.</b>

Installing
----------
Run
```bash
go get github.com/andybons/hipchat
```

Example usage:
```go
package main

import (
  "github.com/andybons/hipchat"
	"log"
)

func main() {
	c := hipchat.Client{AuthToken: "<PUT YOUR AUTH TOKEN HERE>"}
	req := hipchat.MessageRequest{
		RoomId:        "Rat Man's Den",
		From:          "GLaDOS",
		Message:       "Bad news: Combustible lemons failed.",
		Color:         hipchat.ColorPurple,
		MessageFormat: hipchat.FormatText,
		Notify:        true,
	}

	if err := c.PostMessage(req); err != nil {
		log.Printf("Expected no error, but got %q", err)
	}
}
```
