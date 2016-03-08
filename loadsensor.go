/*
   Copyright 2016 Daniel Gruber, Univa

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

// Package loadsensor implements helper functions for writing an
// Univa Grid Engine loadsensor in Go.
package loadsensor

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Arch executes the Univa Grid Engine architecture detection
// script once and returns the correct UGE architecture string.
// This is required to create the correct path to the UGE binaries.
// The result is cached since the archtecture string does not change
// during the runtime of the load sensor.
func Arch() (string, error) {
	path := fmt.Sprintf("%s/util/arch", os.Getenv("SGE_ROOT"))
	arch, err := exec.Command(path).Output()
	return strings.TrimSpace(string(arch)), err
}

// LocalHostname returns the local hostname determined by
// the Univa Grid Engine gethostname binary. Using this hostname
// prevents issues when the host is known by multiple hostnames.
// You should not rely on the OS hostname call.
func LocalHostname() (string, error) {
	arch, err := Arch()
	if err != nil {
		return "", err
	}
	path := fmt.Sprintf("%s/utilbin/%s/gethostname", os.Getenv("SGE_ROOT"), arch)
	hostname, errExec := exec.Command(path, "-name").Output()
	return strings.TrimSpace(string(hostname)), errExec
}

// Sensor is a data structure which contains all functions required for
// performing one load measurement.
type Sensor struct {
	HostNameFunction     func() (string, error)
	ResourceNameFunction func() (string, error)
	MeasurementFunction  func() (string, error)
}

// Context of the whole load sensor. Contains all sensors which make the
// individual measurements.
type Context struct {
	sensors []Sensor
}

// Create initializes a new load sensor context with the given sensors.
func Create(s []Sensor) (*Context, error) {
	for i := range s {
		if s[i].HostNameFunction == nil {
			return nil, errors.New("HostNameFunction is not set")
		}
		if s[i].ResourceNameFunction == nil {
			return nil, errors.New("ResourceNameFunction is not set")
		}
		if s[i].MeasurementFunction == nil {
			return nil, errors.New("MeasurementFunction is not set")
		}
	}
	c := Context{
		sensors: s,
	}
	return &c, nil
}

// Run implements the Univa Grid Engine load sensor protocol and
// executes in each load report interval the measrements given by
// the list of structs implementing the Sesorer interface.
func (ctx Context) Run() {
	stdin := bufio.NewReader(os.Stdin)
	//  the UGE load sensor protocol
	for {
		line, _, err := stdin.ReadLine()
		if err != nil {
			os.Exit(1)
		}
		if string(line) == "quit" {
			os.Exit(0)
		}
		fmt.Println("begin")
		for _, sensor := range ctx.sensors {
			host, errHost := sensor.HostNameFunction()
			if errHost != nil {
				fmt.Fprintf(os.Stderr, "error during hostname function call: %s\n", errHost)
				continue
			}
			resource, errResource := sensor.ResourceNameFunction()
			if errResource != nil {
				fmt.Fprintf(os.Stderr, "error during resource name function call: %s\n", errResource)
				continue
			}
			measurement, errMeasurement := sensor.MeasurementFunction()
			if errMeasurement != nil {
				fmt.Fprintf(os.Stderr, "error during measurement function call: %s\n", errMeasurement)
				continue
			}
			// write load value for resource for the given host
			fmt.Printf("%s:%s:%s\n", host, resource, measurement)
		}
		fmt.Println("end")
	}
}
