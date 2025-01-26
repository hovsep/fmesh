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
	// lightBulbPowerConsumption represents how much energy the lightbulb consumes per cycle
	lightBulbPowerConsumption = 37

	// lightBulbLuminousFlux represents how much light is generated during one activation cycle
	lightBulbLuminousFlux = 1
)

func main() {
	//Init battery level (state stored outside the component)
	batteryLevel := 100

	battery := component.New("battery").
		WithDescription("electric battery with initial charge level").
		WithInputs("power_demand").
		WithOutputs("power_supply").
		WithActivationFunc(func(inputs *port.Collection, outputs *port.Collection, log *log.Logger) error {
			log.Println("activated, level= ", batteryLevel)
			//Power demand/supply cycle
			if inputs.ByName("power_demand").HasSignals() {
				demandedCurrent := inputs.ByName("power_demand").FirstSignalPayloadOrDefault(0).(int)
				log.Println("consumption = ", demandedCurrent)

				//Emit current represented as a number
				suppliedCurrent := min(batteryLevel, demandedCurrent)
				if suppliedCurrent > 0 {
					outputs.ByName("power_supply").PutSignals(signal.New(suppliedCurrent))
					log.Println("emiting power ", suppliedCurrent)

					//Discharge
					batteryLevel = max(0, batteryLevel-suppliedCurrent)
					log.Println("discharged to", batteryLevel)
				} else {
					log.Println("LOW POWER")
				}
			}

			return nil
		})

	lightbulb := component.New("lightbulb").
		WithDescription("electric lightbulb").
		WithInputs("power_supply", "start_power_demand").
		WithOutputs("light_supply", "power_demand").
		WithActivationFunc(func(inputs *port.Collection, outputs *port.Collection, log *log.Logger) error {
			log.Println("activated")
			//Power consumption cycle (at constant rate)
			inputPower := inputs.ByName("power_supply").FirstSignalPayloadOrDefault(0).(int)
			log.Println("got power: ", inputPower)

			if inputPower >= lightBulbPowerConsumption {
				//Emit light
				outputs.ByName("light_supply").PutSignals(signal.New(lightBulbLuminousFlux))
				log.Println("emited light: ", lightBulbLuminousFlux)
			} else {
				log.Println("LOW POWER")
			}

			//Always continue demanding power
			outputs.ByName("power_demand").PutSignals(signal.New(lightBulbPowerConsumption))
			log.Println("demanded power: ", lightBulbPowerConsumption)
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

	fmt.Println(fmt.Sprintf("Simulation finished with %d cycles", len(cycles)))
}
