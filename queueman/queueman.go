package queueman

type Queue chan string

type ChanRequest struct {
	name     string
	response chan Queue
}

var requestChan chan ChanRequest

func Init() {
	requestChan = make(chan ChanRequest)

	go func(req chan ChanRequest) {
		var channels = make(map[string]Queue, 10)

		for {
			r := <-req

			// create if there's no existing queue
			if _, ok := channels[r.name]; !ok {
				channels[r.name] = make(Queue, 10)
			}

			r.response <- channels[r.name]
		}
	}(requestChan)

	return
}

func Get(name string) Queue {
	request := ChanRequest{name: name, response: make(chan Queue)}
	requestChan <- request
	queue := <-request.response
	close(request.response)

	return queue
}
