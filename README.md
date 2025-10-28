<div align="center">
  <img src="./assets/img/logo.png" width="200" height="200" alt="f-mesh"/>
  <h1>f-mesh</h1>
  <p>Flow Based Programming inspired framework in Go</p>
  
[Learn more about FBP](https://jpaulm.github.io/fbp/) (originally discovered by @jpaulm) or read the [documentation](https://github.com/hovsep/fmesh/wiki)
</div>

<img src="https://github.com/user-attachments/assets/045bb7ac-0852-4a0d-9158-6af2d6e66dbb" width="500px">


<h1>What is it?</h1>
<p>F-Mesh is a functions orchestrator inspired by FBP. 
It allows you to express your program as a mesh of interconnected components (or more formally as a computational graph).
</p>
<h3>Main concepts:</h3>
<ul>
<li>F-Mesh consists of <a href="https://github.com/hovsep/fmesh/wiki/5.-Component">Components</a> - the main building blocks</li>
<li>Components have unlimited number of input and output <a href="https://github.com/hovsep/fmesh/wiki/3.-Ports">Ports</a></li>
<li>Ports can be connected via <a href="https://github.com/hovsep/fmesh/wiki/4.-Pipes">Pipes</a></li>
<li>Ports and pipes are type agnostic, any data can be transferred to any port in form of <a href="https://github.com/hovsep/fmesh/wiki/2.-Signals">Signals</a></li>
<li>The framework works in discrete time, not in wall time. The quant of time is 1 <a href="https://github.com/hovsep/fmesh/wiki/6.-Scheduling-rules#phases-of-an-activation-cycle">activation cycle</a>, which gives you "logical parallelism" out of the box (activation function is running in "frozen time")</li>
<li>
	
Learn more in [documentation](https://github.com/hovsep/fmesh/wiki)
</li>
</ul>

<h1>Limitations</h1>
<p>F-mesh is not a classical FBP implementation, it is not suited for long-running components or wall-time events (like timers and tickers)</p>


<h2>Example:</h2>

```go
fm := fmesh.New("hello world").
	WithComponents(
		component.New("concat").
			WithInputs("i1", "i2").
			WithOutputs("res").
			WithActivationFunc(func(this *component.Component) error {
				word1 := this.InputByName("i1").Signals().FirstPayloadOrDefault("").(string)
				word2 := this.InputByName("i2").Signals().FirstPayloadOrDefault("").(string)
				this.OutputByName("res").PutSignals(signal.New(word1 + word2))
				return nil
			}),
		component.New("case").
			WithInputs("i1").
			WithOutputs("res").
			WithActivationFunc(func(this *component.Component) error {
				inputString := this.InputByName("i1").Signals().FirstPayloadOrDefault("").(string)
				this.OutputByName("res").PutSignals(signal.New(strings.ToTitle(inputString)))
				return nil
			}))

fm.ComponentByName("concat").OutputByName("res").PipeTo(fm.ComponentByName("case").InputByName("i1"))

// Init inputs
fm.ComponentByName("concat").InputByName("i1").PutSignals(signal.New("hello "))
fm.ComponentByName("concat").InputByName("i2").PutSignals(signal.New("world !"))

// Run the mesh
_, err := fm.Run()

// Check for errors
if err != nil {
	fmt.Println("F-Mesh returned an error")
	os.Exit(1)
}

//Extract results
results := fm.ComponentByName("case").OutputByName("res").Signals().FirstPayloadOrNil()
fmt.Printf("Result is : %v", results) // Result is : HELLO WORLD !
```

See more in [examples repo](https://github.com/hovsep/fmesh-examples).
