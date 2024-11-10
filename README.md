
<div align="center">
  <img src="./assets/img/logo.png" width="200" height="200" alt="f-mesh"/>
  <h1>f-mesh</h1>
  <p>Flow Based Programming inspired framework in Go</p>
  <p><a href="https://jpaulm.github.io/fbp/">Learn more about FBP</a> (originally discovered by @jpaulm)</p>
</div>

<h1>What is it?</h1>
<p>F-Mesh is a simplistic FBP-inspired framework in Go. 
It allows you to express your program as a mesh of interconnected components.
You can think of it as a simple functions orchestrator.
</p>
<h3>Main concepts:</h3>
<ul>
<li>F-Mesh consists of multiple <b>Components</b> - the main building blocks</li>
<li>Components have unlimited number of input and output <b>Ports</b></li>
<li>The main job of each component is to read inputs and provide outputs</li>
<li>Any output port can be connected to any input port via <b>Pipes</b></li>
<li>The component behaviour is defined by its <b>Activation function</b></li>
<li>The framework checks when components are ready to be activated and calls their activation functions concurrently</li>
<li>One such iteration is called <b>Activation cycle</b></li>
<li>On each activation cycle the framework does same things: activates all the components ready for activation, flushes the data through pipes and disposes input <b>Signals (the data chunks flowing between components)</b></li>
<li>Ports and pipes are type agnostic, any data can be transferred or aggregated on any port</li>
<li>The framework works in discrete time, not it wall time. The quant of time is 1 activation cycle, which gives you "logical parallelism" out of the box</li>
<li>F-Mesh is suitable for logical wireframing, simulation, functional-style computations and implementing simple concurrency patterns without using the concurrency primitives like channels or any sort of locks</li>
</ul>

<h1>What it is not?</h1>
<p>F-mesh is not a classical FBP implementation, and it is not fully async. It does not support long-running components or wall-time events (like timers and tickers)</p>
<p>The framework is not suitable for implementing complex concurrent systems</p>

<h2>Example:</h2>

```go
	// Create f-mesh
	fm := fmesh.New("hello world").
		WithComponents(
			component.New("concat").
				WithInputs("i1", "i2").
				WithOutputs("res").
				WithActivationFunc(func(inputs port.Collection, outputs port.Collection) error {
					word1 := inputs.ByName("i1").Signals().FirstPayload().(string)
					word2 := inputs.ByName("i2").Signals().FirstPayload().(string)

					outputs.ByName("res").PutSignals(signal.New(word1 + word2))
					return nil
				}),
			component.New("case").
				WithInputs("i1").
				WithOutputs("res").
				WithActivationFunc(func(inputs port.Collection, outputs port.Collection) error {
					inputString := inputs.ByName("i1").Signals().FirstPayload().(string)

					outputs.ByName("res").PutSignals(signal.New(strings.ToTitle(inputString)))
					return nil
				})).
                .WithConfig(fmesh.Config{
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
	results := fm.Components().ByName("case").Outputs().ByName("res").Signals().FirstPayload()
	fmt.Printf("Result is :%v", results)
```

<h2>Version 0.1.0 coming soon</h2>
