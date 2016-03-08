package main

import (
	"fmt"
	"github.com/dgruber/loadsensor"
	"os"
)

func main() {
	// create an array of sensors (here with one entry which reports the hostname
	// for the string complex "name")
	sensors := []loadsensor.Sensor{
		{
			HostNameFunction:     loadsensor.LocalHostname,
			ResourceNameFunction: func() (string, error) { return "name", nil },
			MeasurementFunction:  func() (string, error) { return os.Hostname() },
		},
	}

	// create the load sensor by providing all sensors
	ctx, err := loadsensor.Create(sensors)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// start the load sensor
	ctx.Run()
}
