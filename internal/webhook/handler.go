package webhook

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"fritzcable_exporter/internal/client"
)

var resyncLock = &sync.Mutex{}
var resyncTimeout *time.Timer
var interval = 5 * time.Minute

func Handler(c *client.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		d := json.NewDecoder(r.Body)
		var body AlertManagerMessage
		if err := d.Decode(&body); err != nil {
			fmt.Println("failed to decode webhook message:", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if body.Version != "4" {
			fmt.Println("unsupported version", body.Version)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		for _, alert := range body.Alerts {
			if alert.Status == "firing" {
				if resyncLock.TryLock() {
					resyncTimeout = time.NewTimer(interval)
					go func() {
						<-resyncTimeout.C
						resyncLock.Unlock()
					}()
					if err := c.Resync(); err != nil {
						fmt.Println("failed to resync: ", err)
						w.WriteHeader(http.StatusInternalServerError)
						return
					} else {
						fmt.Println("resync triggered")
						w.WriteHeader(http.StatusOK)
						return
					}
				} else {
					fmt.Println("already waiting for resync")
					w.WriteHeader(http.StatusOK)
					return
				}
			}
		}
		// Nothing to do
		w.WriteHeader(http.StatusOK)
	}
}
