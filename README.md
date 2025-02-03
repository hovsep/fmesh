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
<li>F-Mesh consists of <b>Components</b> - the main building blocks</li>
<li>Components have unlimited number of input and output <b>Ports</b></li>
<li>Ports can be connected via <b>Pipes</b></li>
<li>Ports and pipes are type agnostic, any data can be transferred to any port</li>
<li>The framework works in discrete time, not it wall time. The quant of time is 1 activation cycle, which gives you "logical parallelism" out of the box (activation function is running in "frozen time")</li>
<li>
	
Learn more in [documentation](https://github.com/hovsep/fmesh/wiki)
</li>
</ul>

<h1>What it is not?</h1>
<p>F-mesh is not a classical FBP implementation, it does not support long-running components or wall-time events (like timers and tickers)</p>


<h2>Example:</h2>

```go
	t.Run("readme test", func(t *testing.T) {
		fm := fmesh.New("hello world").
			WithComponents(
				component.New("concat").
					WithInputs("i1", "i2").
					WithOutputs("res").
					WithActivationFunc(func(this *component.Component) error {
						word1 := this.InputByName("i1").FirstSignalPayloadOrDefault("").(string)
						word2 := this.InputByName("i2").FirstSignalPayloadOrDefault("").(string)
						this.OutputByName("res").PutSignals(signal.New(word1 + word2))
						return nil
					}),
				component.New("case").
					WithInputs("i1").
					WithOutputs("res").
					WithActivationFunc(func(this *component.Component) error {
						inputString := this.InputByName("i1").FirstSignalPayloadOrDefault("").(string)
						this.OutputByName("res").PutSignals(signal.New(strings.ToTitle(inputString)))
						return nil
					})).
			WithConfig(fmesh.Config{
				ErrorHandlingStrategy: fmesh.StopOnFirstErrorOrPanic,
				CyclesLimit:           10,
			})

		fm.Components().ByName("concat").Outputs().ByName("res").PipeTo(
			fm.Components().ByName("case").Inputs().ByName("i1"),
		)

		// Init inputs
		fm.Components().ByName("concat").Inputs().ByName("i1").PutSignals(signal.New("hello "))
		fm.Components().ByName("concat").Inputs().ByName("i2").PutSignals(signal.New("world !"))

		// Run the mesh
		_, err := fm.Run()

		// Check for errors
		if err != nil {
			fmt.Println("F-Mesh returned an error")
			os.Exit(1)
		}

		//Extract results
		results := fm.Components().ByName("case").Outputs().ByName("res").FirstSignalPayloadOrNil()
		fmt.Printf("Result is :%v", results)
```

See more in [examples](https://github.com/hovsep/fmesh/tree/main/examples) directory.
<h2>Version <a href="https://github.com/hovsep/fmesh/releases/tag/v0.0.1-alpha">0.1.0-Sugunia</a> is already released!</h2>
