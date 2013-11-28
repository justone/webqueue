package main

import (
	"github.com/codegangsta/martini"

	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/justone/webqueue/queue"
)

func main() {
	m := martini.Classic()

	queue.Init()

	m.Get("/queue/:name", func(res http.ResponseWriter, params martini.Params) {
		queue := queue.Get(params["name"])

		// read from the queue with a 30 second timeout
		select {
		case val := <-queue:
			fmt.Fprintln(res, val)
		case <-time.After(time.Second * 30):
			http.Error(res, "Timeout", http.StatusRequestTimeout)
		}

	})
	m.Post("/queue/:name", func(res http.ResponseWriter, req *http.Request, params martini.Params) {
		queue := queue.Get(params["name"])

		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			http.Error(res, "Unable to read request body: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// send to the queue unless it's full
		select {
		case queue <- string(body):
		default:
			http.Error(res, "Queue Full", http.StatusNotAcceptable)
		}
	})
	m.Run()
}
