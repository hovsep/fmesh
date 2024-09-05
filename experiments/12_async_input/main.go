package main

import (
	"fmt"
	"net/http"
	"time"
)

// This example demonstrates how fmesh can be fed input asynchronously
func main() {
	//exampleFeedBatches()
	exampleFeedSequentially()
}

// This example processes 1 url every 3 seconds
// NOTE: urls are not crawled concurrently, because fm has only 1 worker (crawler component)
func exampleFeedSequentially() {
	fm := getMesh()

	urls := []string{
		"http://fffff.com",
		"https://google.com",
		"http://habr.com",
		"http://localhost:80",
		"https://postman-echo.com/delay/1",
		"https://postman-echo.com/delay/3",
		"https://postman-echo.com/delay/5",
		"https://postman-echo.com/delay/10",
	}

	ticker := time.NewTicker(3 * time.Second)
	resultsChan := make(chan []any)
	doneChan := make(chan struct{}) // Signals when all urls are processed

	//Feeder routine
	go func() {
		for {
			select {
			case <-ticker.C:
				if len(urls) == 0 {
					close(resultsChan)
					return
				}
				//Pop an url
				url := urls[0]
				urls = urls[1:]

				fmt.Println("feed this url:", url)

				fm.Components.byName("web crawler").inputs.byName("url").putSignal(newSignal(url))
				_, err := fm.run()
				if err != nil {
					fmt.Println("fmesh returned error ", err)
				}

				if fm.Components.byName("web crawler").outputs.byName("headers").hasSignal() {
					results := fm.Components.byName("web crawler").outputs.byName("headers").getSignal().AllValues()
					fm.Components.byName("web crawler").outputs.byName("headers").clearSignal() //@TODO maybe we can add fm.Reset() for cases when FMesh is reused (instead of cleaning ports explicitly)
					resultsChan <- results
				}
			}
		}
	}()

	//Result reader routine
	go func() {
		for {
			select {
			case r, ok := <-resultsChan:
				if !ok {
					fmt.Println("results chan is closed. shutting down the reader")
					doneChan <- struct{}{}
					return
				}
				fmt.Println(fmt.Sprintf("got results from channel: %v", r))
			}
		}
	}()

	<-doneChan
}

// This example leverages signal aggregation, so urls are pushed into fmesh all at once
// so we wait for all urls to be processed and only them we can read results
func exampleFeedBatches() {
	batch := []string{
		"http://fffff.com",
		"https://google.com",
		"http://habr.com",
		"http://localhost:80",
	}

	fm := getMesh()

	for _, url := range batch {
		fm.Components.byName("web crawler").inputs.byName("url").putSignal(newSignal(url))
	}

	_, err := fm.run()
	if err != nil {
		fmt.Println("fmesh returned error ", err)
	}

	if fm.Components.byName("web crawler").outputs.byName("headers").hasSignal() {
		results := fm.Components.byName("web crawler").outputs.byName("headers").getSignal().AllValues()
		fmt.Printf("results: %v", results)
	}
}

func getMesh() *FMesh {
	//Setup dependencies
	client := &http.Client{}

	//Define components
	crawler := &Component{
		name:        "web crawler",
		description: "gets http headers from given url",
		inputs: Ports{
			"url": &Port{},
		},
		outputs: Ports{
			"errors":  &Port{},
			"headers": &Port{},
		},
		handler: func(inputs Ports, outputs Ports) error {
			if !inputs.byName("url").hasSignal() {
				return errWaitingForInputResetInputs
			}

			for _, sigVal := range inputs.byName("url").getSignal().AllValues() {

				url := sigVal.(string)
				//All urls incoming as aggregatet signal will be crawled sequentially
				// in order to call them concurrently we need run each request in separate goroutine and handle synchronization (e.g. waitgroup)
				response, err := client.Get(url)
				if err != nil {
					outputs.byName("errors").putSignal(newSignal(fmt.Errorf("got error: %w from url: %s", err, url)))
					continue
				}

				if len(response.Header) == 0 {
					outputs.byName("errors").putSignal(newSignal(fmt.Errorf("no headers for url %s", url)))
					continue
				}

				outputs.byName("headers").putSignal(newSignal(map[string]http.Header{
					url: response.Header,
				}))
			}

			return nil
		},
	}

	logger := &Component{
		name:        "error logger",
		description: "logs http errors",
		inputs: Ports{
			"error": &Port{},
		},
		outputs: nil,
		handler: func(inputs Ports, outputs Ports) error {
			if !inputs.byName("error").hasSignal() {
				return errWaitingForInputResetInputs
			}

			for _, sigVal := range inputs.byName("error").getSignal().AllValues() {
				err := sigVal.(error)
				if err != nil {
					fmt.Println("Error logger says:", err)
				}
			}

			return nil
		},
	}

	//Define pipes
	crawler.outputs.byName("errors").CreatePipesTo(logger.inputs.byName("error"))

	//Build mesh
	return &FMesh{
		Components:            Components{crawler, logger},
		ErrorHandlingStrategy: StopOnFirstError,
	}

}
