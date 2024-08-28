package main

import (
	"fmt"
	"math/rand"
	"strings"
)

type Request struct {
	Method  string
	Headers map[string]string
	Body    []byte
}

type Response struct {
	Status     string
	StatusCode int
	Headers    map[string]string
	Body       []byte
}

// This experiment shows how signal can carry arbitrary data (not only primitive types)
// We use a simple round-robin load balancer as an example
func main() {
	//Define components
	lb := &Component{
		name: "Load Balancer",
		inputs: Ports{
			"req:localhost:80": &Port{}, //Incoming requests

			//Responses from each backend:
			"resp:backend1:80": &Port{},
			"resp:backend2:80": &Port{},
			"resp:backend3:80": &Port{},
		},
		outputs: Ports{
			//Requests to each backend
			"req:backend1:80": &Port{},
			"req:backend2:80": &Port{},
			"req:backend3:80": &Port{},

			"resp:localhost:80": &Port{}, //Out coming responses
		},
		handler: func(inputs Ports, outputs Ports) error {
			//Process incoming requests
			reqInput := inputs.byName("req:localhost:80")
			if reqInput.hasValue() {
				reqSignals := reqInput.getValue()

				if reqSignals.IsAggregate() {
					for _, sig := range reqSignals.(*AggregateSignal).val {

						//Handle authorization
						req := sig.val.(*Request)
						authHeader, ok := req.Headers["Auth"]

						if !ok || len(authHeader) == 0 {
							//Missing auth header
							outputs.byName("resp:localhost:80").setValue(&SingleSignal{val: &Response{
								Status:     "401 Unauthorized",
								StatusCode: 401,
								Headers:    nil,
								Body:       []byte(fmt.Sprintf("response from LB: request: %s is missing auth header", string(req.Body))),
							}})
							continue
						}

						if authHeader != "Bearer 123" {
							//Auth error
							outputs.byName("resp:localhost:80").setValue(&SingleSignal{val: &Response{
								Status:     "401 Unauthorized",
								StatusCode: 401,
								Headers:    nil,
								Body:       []byte(fmt.Sprintf("response from LB: request: %s is unauthorized", string(req.Body))),
							}})
							continue
						}

						targetBackendIndex := 1 + rand.Intn(3)
						outputs.byName(fmt.Sprintf("req:backend%d:80", targetBackendIndex)).setValue(sig)

					}

				} else {
					doDispatch := true

					sig := reqSignals.(*SingleSignal)

					//Handle authorization
					req := sig.val.(*Request)
					authHeader, ok := req.Headers["Auth"]

					if !ok || len(authHeader) == 0 {
						//Missing auth header
						doDispatch = false
						outputs.byName("resp:localhost:80").setValue(&SingleSignal{val: &Response{
							Status:     "401 Unauthorized",
							StatusCode: 401,
							Headers:    nil,
							Body:       []byte(fmt.Sprintf("Request: %s is missing auth header", string(req.Body))),
						}})
					}

					if authHeader != "Bearer 123" {
						//Auth error
						doDispatch = false
						outputs.byName("resp:localhost:80").setValue(&SingleSignal{val: &Response{
							Status:     "401 Unauthorized",
							StatusCode: 401,
							Headers:    nil,
							Body:       []byte(fmt.Sprintf("Request: %s is unauthorized", string(req.Body))),
						}})
					}

					if doDispatch {
						targetBackendIndex := 1 + rand.Intn(3)
						outputs.byName(fmt.Sprintf("req:backend%d:80", targetBackendIndex)).setValue(sig)
					}
				}
			}

			//Read responses from backends and put them on main output for upstream consumer
			for pname, port := range inputs {
				if strings.Contains(pname, "resp:backend") && port.hasValue() {
					outputs.byName("resp:localhost:80").setValue(port.getValue())
				}
			}

			return nil
		},
	}

	b1 := &Component{
		name: "Backend 1",
		inputs: Ports{
			"req:localhost:80": &Port{},
		},
		outputs: Ports{
			"resp:localhost:80": &Port{},
		},
		handler: func(inputs Ports, outputs Ports) error {
			reqInput := inputs.byName("req:localhost:80")

			if !reqInput.hasValue() {
				return nil
			}

			reqSignal := reqInput.getValue()

			if reqSignal.IsSingle() {
				req := reqSignal.(*SingleSignal).GetVal().(*Request)

				resp := &SingleSignal{
					val: &Response{
						Status:     "200 OK",
						StatusCode: 200,
						Headers:    nil,
						Body:       []byte("response from b1 for req:" + string(req.Body)),
					},
				}

				outputs.byName("resp:localhost:80").setValue(resp)
			} else {
				for _, sig := range reqSignal.(*AggregateSignal).val {
					resp := &SingleSignal{
						val: &Response{
							Status:     "200 OK",
							StatusCode: 200,
							Headers:    nil,
							Body:       []byte("response from b1 for req:" + string(sig.val.(*Request).Body)),
						},
					}

					outputs.byName("resp:localhost:80").setValue(resp)
				}
			}
			return nil
		},
	}

	b2 := &Component{
		name: "Backend 2",
		inputs: Ports{
			"req:localhost:80": &Port{},
		},
		outputs: Ports{
			"resp:localhost:80": &Port{},
		},
		handler: func(inputs Ports, outputs Ports) error {
			reqInput := inputs.byName("req:localhost:80")

			if !reqInput.hasValue() {
				return nil
			}

			reqSignal := reqInput.getValue()

			if reqSignal.IsSingle() {
				req := reqSignal.(*SingleSignal).GetVal().(*Request)

				resp := &SingleSignal{
					val: &Response{
						Status:     "200 OK",
						StatusCode: 200,
						Headers:    nil,
						Body:       []byte("response from b2 for req:" + string(req.Body)),
					},
				}

				outputs.byName("resp:localhost:80").setValue(resp)
			} else {
				for _, sig := range reqSignal.(*AggregateSignal).val {
					resp := &SingleSignal{
						val: &Response{
							Status:     "200 OK",
							StatusCode: 200,
							Headers:    nil,
							Body:       []byte("response from b2 for req:" + string(sig.val.(*Request).Body)),
						},
					}

					outputs.byName("resp:localhost:80").setValue(resp)
				}
			}
			return nil
		},
	}

	b3 := &Component{
		name: "Backend 3",
		inputs: Ports{
			"req:localhost:80": &Port{},
		},
		outputs: Ports{
			"resp:localhost:80": &Port{},
		},
		handler: func(inputs Ports, outputs Ports) error {
			reqInput := inputs.byName("req:localhost:80")

			if !reqInput.hasValue() {
				return nil
			}

			reqSignal := reqInput.getValue()

			if reqSignal.IsSingle() {
				req := reqSignal.(*SingleSignal).GetVal().(*Request)

				resp := &SingleSignal{
					val: &Response{
						Status:     "200 OK",
						StatusCode: 200,
						Headers:    nil,
						Body:       []byte("response from b3 for req:" + string(req.Body)),
					},
				}

				outputs.byName("resp:localhost:80").setValue(resp)
			} else {
				for _, sig := range reqSignal.(*AggregateSignal).val {
					resp := &SingleSignal{
						val: &Response{
							Status:     "200 OK",
							StatusCode: 200,
							Headers:    nil,
							Body:       []byte("response from b3 for req:" + string(sig.val.(*Request).Body)),
						},
					}

					outputs.byName("resp:localhost:80").setValue(resp)
				}
			}
			return nil
		},
	}

	//Define pipes

	//Request path
	lb.outputs.byName("req:backend1:80").CreatePipeTo(b1.inputs.byName("req:localhost:80"))
	lb.outputs.byName("req:backend2:80").CreatePipeTo(b2.inputs.byName("req:localhost:80"))
	lb.outputs.byName("req:backend3:80").CreatePipeTo(b3.inputs.byName("req:localhost:80"))

	//Response path
	b1.outputs.byName("resp:localhost:80").CreatePipeTo(lb.inputs.byName("resp:backend1:80"))
	b2.outputs.byName("resp:localhost:80").CreatePipeTo(lb.inputs.byName("resp:backend2:80"))
	b3.outputs.byName("resp:localhost:80").CreatePipeTo(lb.inputs.byName("resp:backend3:80"))

	//Build mesh
	fm := FMesh{
		Components: Components{
			lb, b1, b2, b3,
		},
		ErrorHandlingStrategy: StopOnFirstError,
	}

	//Set inputs
	reqs := []*Request{
		{
			Method: "GET",
			Headers: map[string]string{
				"Auth": "Bearer 123",
			},
			Body: []byte("request 1"),
		},
		{
			Method: "POST",
			Headers: map[string]string{
				"Auth": "Bearer 123",
			},
			Body: []byte("request 2"),
		},
		{
			Method: "PATCH",
			Headers: map[string]string{
				"Auth": "Bearer 123",
			},
			Body: []byte("request 3"),
		},

		{
			Method: "GET",
			Headers: map[string]string{
				"Auth": "Bearer 777", //Gonna fail auth on LB:
			},
			Body: []byte("request 4"),
		},
		{
			Method: "GET",
			Headers: map[string]string{
				"Auth": "Bearer 123",
			},
			Body: []byte("request 5"),
		},
		{
			Method: "GET",
			Headers: map[string]string{
				"Auth": "", //Empty auth header
			},
			Body: []byte("request 6"),
		},
		{
			Method: "GET",
			Headers: map[string]string{
				"Content-Type": "application/json", //Missing auth
			},
			Body: []byte("request 7"),
		},
	}

	//No need to create aggregated signal literal,
	//we can perfectly set the value to same input, and it will be aggregated automatically
	for _, req := range reqs {
		lb.inputs.byName("req:localhost:80").setValue(&SingleSignal{val: req})
		//break //set just 1 request
	}

	//Run the mesh
	hops, err := fm.run()
	_ = hops

	res := lb.outputs.byName("resp:localhost:80").getValue()

	if res.IsAggregate() {
		for _, sig := range res.(*AggregateSignal).val {
			resp := sig.val.(*Response)

			fmt.Println(string(resp.Body))
		}
	}

	if err != nil {
		fmt.Println(err)
	}

}
