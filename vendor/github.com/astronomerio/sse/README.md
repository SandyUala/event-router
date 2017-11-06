# sse

An sse client for Go

## Example usage

```golang
package main

import (
    "log"

    "github.com/astronomerio/sse"
)

func main() {
    go func() {
        clickstreamEvents := sse.Subscribe("clickstream")
        heartbeat := sse.Subscribe("heartbeat")
        // runtime status events
        events := sse.Events()
        for {
            select {
            case e := <-clickstreamEvents:
                log.Println(e.Event)
                log.Println(string(e.Data))
            case e := <-heartbeat:
                log.Println(e.Event)
            case e := <-events:
                log.Println(e.Code)
                log.Println(e.Error.Error())
            }
        }
    }()

    sse.WithURI("https://houston.astronomer.io/broadcast")
    sse.WithHTTPHeaders(map[string]string{
        "authorization": "API_KEY",
    })

    log.Fatal(sse.Listen())
}

```