package main

import (
	"fmt"
	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/port"
	"github.com/hovsep/fmesh/signal"
	"log"
	"net/http"
	"time"
)

// This example processes 1 url every 3 seconds
// NOTE: urls are not crawled concurrently, because fm has only 1 worker (crawler component)
func main() {
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

	//Producer goroutine
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

				fmt.Println("produce:", url)

				fm.Components().ByName("web crawler").InputByName("url").PutSignals(signal.New(url))
				_, err := fm.Run()
				if err != nil {
					fmt.Println("fmesh returned error ", err)
				}

				if fm.Components().ByName("web crawler").OutputByName("headers").HasSignals() {
					results, err := fm.Components().ByName("web crawler").OutputByName("headers").AllSignalsPayloads()
					if err != nil {
						fmt.Println("Failed to get results ", err)
					}
					fm.Components().ByName("web crawler").OutputByName("headers").Clear() //@TODO maybe we can add fm.Reset() for cases when FMesh is reused (instead of cleaning ports explicitly)
					resultsChan <- results
				}
			}
		}
	}()

	//Consumer goroutine
	go func() {
		for {
			select {
			case r, ok := <-resultsChan:
				if !ok {
					fmt.Println("results chan is closed. shutting down the reader")
					doneChan <- struct{}{}
					return
				}
				fmt.Println(fmt.Sprintf("consume: %v", r))
			}
		}
	}()

	<-doneChan
}

func getMesh() *fmesh.FMesh {
	//Setup dependencies
	client := &http.Client{}

	//Define components
	crawler := component.New("web crawler").
		WithDescription("gets http headers from given url").
		WithInputs("url").
		WithOutputs("errors", "headers").
		WithActivationFunc(func(inputs *port.Collection, outputs *port.Collection, log *log.Logger) error {
			if !inputs.ByName("url").HasSignals() {
				return component.NewErrWaitForInputs(false)
			}

			allUrls, err := inputs.ByName("url").AllSignalsPayloads()
			if err != nil {
				return err
			}

			for _, urlVal := range allUrls {

				url := urlVal.(string)
				//All urls will be crawled sequentially
				// in order to call them concurrently we need run each request in separate goroutine and handle synchronization (e.g. waitgroup)
				response, err := client.Get(url)
				if err != nil {
					outputs.ByName("errors").PutSignals(signal.New(fmt.Errorf("got error: %w from url: %s", err, url)))
					continue
				}

				if len(response.Header) == 0 {
					outputs.ByName("errors").PutSignals(signal.New(fmt.Errorf("no headers for url %s", url)))
					continue
				}

				outputs.ByName("headers").PutSignals(signal.New(map[string]http.Header{
					url: response.Header,
				}))
			}

			return nil
		})

	logger := component.New("error logger").
		WithDescription("logs http errors").
		WithInputs("error").
		WithActivationFunc(func(inputs *port.Collection, outputs *port.Collection, log *log.Logger) error {
			if !inputs.ByName("error").HasSignals() {
				return component.NewErrWaitForInputs(false)
			}

			allErrors, err := inputs.ByName("error").AllSignalsPayloads()
			if err != nil {
				return err
			}

			for _, errVal := range allErrors {
				e := errVal.(error)
				if e != nil {
					fmt.Println("Error logger says:", e)
				}
			}

			return nil
		})

	//Define pipes
	crawler.OutputByName("errors").PipeTo(logger.InputByName("error"))

	return fmesh.New("web scraper").WithConfig(fmesh.Config{
		ErrorHandlingStrategy: fmesh.StopOnFirstErrorOrPanic,
	}).WithComponents(crawler, logger)

}
