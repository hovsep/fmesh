package main

import (
	"fmt"
	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/signal"
)

const (
	// lightBulbPowerConsumption specifies the amount of energy (in watts or equivalent unit)
	// that the lightbulb consumes during a single activation cycle.
	lightBulbPowerConsumption = 22

	// lightBulbLuminousFlux indicates the total amount of visible light (in lumens)
	// emitted by the lightbulb during one activation cycle.
	lightBulbLuminousFlux = 6000

	// lightBulbWarmingPerCycle defines the temperature increase (in degrees Celsius)
	// experienced by the lightbulb after completing one activation cycle.
	lightBulbWarmingPerCycle = 0.13

	// lightbulbOverheatingThreshold represents the temperature (in degrees Celsius)
	// at which the lightbulb's performance begins to degrade significantly, such as reduced brightness or efficiency.
	lightbulbOverheatingThreshold = 30

	// lightbulbMaxWorkingTemperature is the critical temperature (in degrees Celsius)
	// beyond which the lightbulb will fail or burn out permanently due to thermal stress.
	lightbulbMaxWorkingTemperature = 50

	// lightbulbOverheatDegradation represents the performance degradation factor (as a multiplier)
	// when the lightbulb operates above the overheating threshold.
	// For example, a value of 0.7 means the lightbulb will perform at 70% of its normal capacity.
	lightbulbOverheatDegradation = 0.7
)

// This example simulates a simple electric circuit consisting of a battery and a lightbulb.
// The battery serves as the power source, and the lightbulb acts as the load, consuming power to emit light.
// Key aspects of the simulation include:
//   - The battery maintains a charge level and supplies power based on the lightbulb's demand.
//   - The lightbulb consumes power, emits light, and experiences a temperature increase during operation.
//   - The simulation models realistic behaviors, such as power starvation, overheating, and burnout.
//   - Power demand and supply are managed through fmesh components and ports, simulating current flow in the circuit.
//   - Logs provide detailed insights into each component's state, such as battery level, power consumption,
//     light emission, and temperature changes.
//
// The simulation runs until a termination condition is met, such as the battery depleting or the lightbulb burning out.
func main() {
	battery := component.New("battery").
		WithDescription("electric battery with initial charge level").
		WithInputs("power_demand").
		WithOutputs("power_supply").
		WithInitialState(func(state component.State) {
			state.Set("level", 1000)
		}).
		WithActivationFunc(func(this *component.Component) error {
			// Read state
			level := this.State().Get("level").(int)

			defer func() {
				// Write state
				this.State().Set("level", level)
			}()

			this.Logger().Println("level: ", level)
			// Power demand/supply cycle
			if this.InputByName("power_demand").HasSignals() {
				demandedCurrent := this.InputByName("power_demand").FirstSignalPayloadOrDefault(0).(int)

				// Emit current represented as a number
				suppliedCurrent := min(level, demandedCurrent)
				if suppliedCurrent > 0 {
					this.OutputByName("power_supply").PutSignals(signal.New(suppliedCurrent))
					this.Logger().Println("supplying power:", suppliedCurrent)

					if suppliedCurrent < demandedCurrent {
						this.Logger().Println("LOW BATTERY")
					}

					// Discharge
					level = max(0, level-suppliedCurrent)
					this.Logger().Println("discharged to: ", level)
				} else {
					this.Logger().Println("BATTERY DIED")
				}
			}

			return nil
		})

	lightbulb := component.New("lightbulb").
		WithDescription("electric lightbulb").
		WithInputs("power_supply", "start_power_demand").
		WithOutputs("light_supply", "power_demand").
		WithInitialState(func(state component.State) {
			state.Set("temperature", 26.0)
		}).
		WithActivationFunc(func(this *component.Component) error {

			// Read state
			temperature := this.State().Get("temperature").(float64)

			defer func() {
				// Write state
				this.State().Set("temperature", temperature)
			}()

			// Skip power consumption on start (as power is not demanded yet)
			if !this.InputByName("start_power_demand").HasSignals() {
				// Power consumption cycle (at constant rate)
				inputPower := this.InputByName("power_supply").FirstSignalPayloadOrDefault(0).(int)
				this.Logger().Println("got power: ", inputPower)

				if inputPower >= lightBulbPowerConsumption {
					// Emit light (here we simulate how the lightbulb performance degrades with warming)
					lightEmission := lightBulbLuminousFlux / temperature * 100
					if temperature >= lightbulbOverheatingThreshold {
						this.Logger().Println("OVERHEATING. LIGHT EMISSION WILL SIGNIFICANTLY DEGRADE")
						lightEmission *= lightbulbOverheatDegradation
					}
					this.OutputByName("light_supply").PutSignals(signal.New(lightEmission))
					this.Logger().Println("emitted light: ", lightEmission)
					this.Logger().Println("temperature:", temperature)

					// Get warmer
					temperature += lightBulbWarmingPerCycle

					if temperature > lightbulbMaxWorkingTemperature {
						this.Logger().Println("BURNOUT")
						return nil
					}
				} else {
					this.Logger().Println("POWER STARVATION")
				}
			}

			// Always continue demanding power
			this.OutputByName("power_demand").PutSignals(signal.New(lightBulbPowerConsumption))
			return nil
		})

	battery.OutputByName("power_supply").PipeTo(lightbulb.InputByName("power_supply"))
	lightbulb.OutputByName("power_demand").PipeTo(battery.InputByName("power_demand"))

	fm := fmesh.NewWithConfig("battery_and_lightbulb", &fmesh.Config{
		ErrorHandlingStrategy: fmesh.StopOnFirstErrorOrPanic,
		Debug:                 false,
	}).
		WithDescription("simple electric simulation").
		WithComponents(battery, lightbulb)

	// Turn on the lightbulb (yes you can init an output port)
	lightbulb.InputByName("start_power_demand").PutSignals(signal.New("start"))

	runResult, err := fm.Run()

	if err != nil {
		fmt.Println("Simulation failed with error: ", err)
		return
	}

	fmt.Printf("Simulation finished after %d cycles \n", runResult.Cycles.Len())
}
