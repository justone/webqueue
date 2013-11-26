package main

import (
	"github.com/codegangsta/martini"

	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type Queue chan string

type ChanRequest struct {
	name     string
	response chan Queue
}

func main() {
	m := martini.Classic()

	var requestChan = make(chan ChanRequest)

	go func(req chan ChanRequest) {
		var channels = make(map[string]Queue, 10)

		for {
			r := <-req

			queue, ok := channels[r.name]
			if ok {
				r.response <- queue
			} else {
				newchan := make(Queue, 10)
				channels[r.name] = newchan
				r.response <- newchan
			}
		}
	}(requestChan)

	m.Get("/queue/:name", func(res http.ResponseWriter, params martini.Params) {
		request := ChanRequest{name: params["name"], response: make(chan Queue)}
		requestChan <- request
		queue := <-request.response
		close(request.response)

		// read from the queue with a 30 second timeout
		select {
		case val := <-queue:
			fmt.Fprintln(res, val)
		case <-time.After(time.Second * 30):
			http.Error(res, "Timeout", http.StatusRequestTimeout)
		}

	})
	m.Post("/queue/:name", func(res http.ResponseWriter, req *http.Request, params martini.Params) {
		request := ChanRequest{name: params["name"], response: make(chan Queue)}
		requestChan <- request
		queue := <-request.response
		close(request.response)

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
