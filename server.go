package main

import (
	"github.com/codegangsta/martini"

	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

type ChanRequest struct {
	name     string
	response chan chan int
}

func main() {
	m := martini.Classic()

	var requestChan = make(chan ChanRequest)

	go func(req chan ChanRequest) {
		var channels = make(map[string]chan int, 10)

		for {
			r := <-req

			queue, ok := channels[r.name]
			if ok {
				r.response <- queue
			} else {
				fmt.Println("adding new channel")
				newchan := make(chan int, 10)
				channels[r.name] = newchan
				r.response <- newchan
			}
		}
	}(requestChan)

	m.Get("/queue/:name", func(res http.ResponseWriter, params martini.Params) {
		response := make(chan chan int)
		request := ChanRequest{name: params["name"], response: response}
		requestChan <- request
		queue := <-response

		// read from the queue with a 30 second timeout
		select {
		case val := <-queue:
			fmt.Fprintln(res, val)
		case <-time.After(time.Second * 30):
			res.WriteHeader(408)
			fmt.Fprintln(res, "TIMEOUT")
		}

	})
	m.Put("/queue/:name", func(res http.ResponseWriter, req *http.Request, params martini.Params) {
		response := make(chan chan int)
		request := ChanRequest{name: params["name"], response: response}
		requestChan <- request
		queue := <-response

		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			fmt.Fprintln(res, err)
		}

		intval, err := strconv.Atoi(string(body))
		if err != nil {
			fmt.Fprintln(res, err)
		}

		// send to the queue unless it's full
		select {
		case queue <- intval:
		default:
			res.WriteHeader(406)
			fmt.Fprintln(res, "QUEUE FULL")
		}
	})
	m.Run()
}
