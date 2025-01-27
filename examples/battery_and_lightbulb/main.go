package main

import (
	"fmt"
	"github.com/hovsep/fmesh"
	"github.com/hovsep/fmesh/component"
	"github.com/hovsep/fmesh/port"
	"github.com/hovsep/fmesh/signal"
	"log"
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
	//Init battery level (state stored outside the component)
	batteryLevel := 1000

	battery := component.New("battery").
		WithDescription("electric battery with initial charge level").
		WithInputs("power_demand").
		WithOutputs("power_supply").
		WithActivationFunc(func(inputs *port.Collection, outputs *port.Collection, log *log.Logger) error {
			log.Println("level: ", batteryLevel)
			//Power demand/supply cycle
			if inputs.ByName("power_demand").HasSignals() {
				demandedCurrent := inputs.ByName("power_demand").FirstSignalPayloadOrDefault(0).(int)

				//Emit current represented as a number
				suppliedCurrent := min(batteryLevel, demandedCurrent)
				if suppliedCurrent > 0 {
					outputs.ByName("power_supply").PutSignals(signal.New(suppliedCurrent))
					log.Println("supplying power:", suppliedCurrent)

					if suppliedCurrent < demandedCurrent {
						log.Println("LOW BATTERY")
					}

					//Discharge
					batteryLevel = max(0, batteryLevel-suppliedCurrent)
					log.Println("discharged to: ", batteryLevel)
				} else {
					log.Println("BATTERY DIED")
				}
			}

			return nil
		})

	//Init lightbulb state
	lightbulbTemperature := 26.0 //room temperature

	lightbulb := component.New("lightbulb").
		WithDescription("electric lightbulb").
		WithInputs("power_supply", "start_power_demand").
		WithOutputs("light_supply", "power_demand").
		WithActivationFunc(func(inputs *port.Collection, outputs *port.Collection, log *log.Logger) error {
			//Skip power consumption on start (as power is not demanded yet)
			if !inputs.ByName("start_power_demand").HasSignals() {
				//Power consumption cycle (at constant rate)
				inputPower := inputs.ByName("power_supply").FirstSignalPayloadOrDefault(0).(int)
				log.Println("got power: ", inputPower)

				if inputPower >= lightBulbPowerConsumption {
					//Emit light (here we simulate how the lighbulb performance degrades with warming)
					lightEmission := lightBulbLuminousFlux / lightbulbTemperature * 100
					if lightbulbTemperature >= lightbulbOverheatingThreshold {
						log.Println("OVERHEATING. LIGHT EMISSION WILL SIGNIFICANTLY DEGRADE")
						lightEmission *= lightbulbOverheatDegradation
					}
					outputs.ByName("light_supply").PutSignals(signal.New(lightEmission))
					log.Println("emited light: ", lightEmission)
					log.Println("temperature:", lightbulbTemperature)

					//Get warmer
					lightbulbTemperature += lightBulbWarmingPerCycle

					if lightbulbTemperature > lightbulbMaxWorkingTemperature {
						log.Println("BURNOUT")
						return nil
					}
				} else {
					log.Println("POWER STARVATION")
				}
			}

			//Always continue demanding power
			outputs.ByName("power_demand").PutSignals(signal.New(lightBulbPowerConsumption))
			return nil
		})

	battery.OutputByName("power_supply").PipeTo(lightbulb.InputByName("power_supply"))
	lightbulb.OutputByName("power_demand").PipeTo(battery.InputByName("power_demand"))

	fm := fmesh.New("battery_and_lightbulb").
		WithDescription("simple electric simulation").
		WithComponents(battery, lightbulb).
		WithConfig(fmesh.Config{
			ErrorHandlingStrategy: fmesh.StopOnFirstErrorOrPanic,
		})

	// Turn on the lightbulb (yes you can init an output port)
	lightbulb.InputByName("start_power_demand").PutSignals(signal.New("start"))

	cycles, err := fm.Run()

	if err != nil {
		fmt.Println("Simulation failed with error: ", err)
		return
	}

	fmt.Println(fmt.Sprintf("Simulation finished after %d cycles", len(cycles)))
}
