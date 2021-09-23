// Copyright 2018 Andrew Bates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)

type category struct {
	domain *domain
	id     int
	SKU    string
	Name   string
	Var    string
}

func (c *category) ID() string {
	return fmt.Sprintf("0x%02x", c.id)
}

func (c *category) DevCat() string {
	return fmt.Sprintf("DevCat{0x%02x, 0x%02x}", c.domain.id, c.id)
}

type domain struct {
	id         int
	Var        string
	Name       string
	Categories []category
}

func (d *domain) ID() string {
	return fmt.Sprintf("0x%02x", d.id)
}

var domains = []*domain{
	{id: 0x00, Var: "GeneralizeDomain", Name: "Generalized Controllers"},
	{id: 0x01, Var: "DimmerDomain", Name: "Dimmable Lighting Control"},
	{id: 0x02, Var: "SwitchDomain", Name: "Switched Lighting Control"},
	{id: 0x03, Var: "NetworkDomain", Name: "Network Bridges"},
	{id: 0x04, Var: "IrrigationDomain", Name: "Irrigation Control"},
	{id: 0x05, Var: "ThermostatDomain", Name: "Climate Control"},
	{id: 0x06, Var: "PoolSpaDomain", Name: "Pool and Spa Control"},
	{id: 0x07, Var: "SensorActuatorDomain", Name: "Sensors and Actuators"},
	{id: 0x08, Var: "HomeEntertainmentDomain", Name: "Home Entertainment"},
	{id: 0x09, Var: "EnergyMgmtDomain", Name: "Energy Management"},
	{id: 0x0a, Var: "ApplianceDomain", Name: "Built-In Appliance Control"},
	{id: 0x0b, Var: "PlumbingDomain", Name: "Plumbing"},
	{id: 0x0c, Var: "CommunicationDomain", Name: "Communication"},
	{id: 0x0d, Var: "ComputerDomain", Name: "Computer Control"},
	{id: 0x0e, Var: "WindowCoveringsDomain", Name: "Window Coverings"},
	{id: 0x0f, Var: "AccessDomain", Name: "Access Control"},
	{id: 0x10, Var: "SecurityDomain", Name: "Security, Health, Safety"},
	{id: 0x11, Var: "SurveillanceDomain", Name: "Surveillance"},
	{id: 0x12, Var: "AutomotiveDomain", Name: "Automotive"},
	{id: 0x13, Var: "PetCareDomain", Name: "Pet Care"},
	{id: 0x14, Var: "ToysDomain", Name: "Toys"},
	{id: 0x15, Var: "TimekeepingDomain", Name: "Timekeeping"},
	{id: 0x16, Var: "HolidayDomain", Name: "Holiday"},
	{id: 0xff, Var: "UnassignedDomain", Name: "Unassigned"},
}

func init() {
	autogenCommands["devcats"] = autogenCommand{
		inputFilename:  "internal/devcats.go.tmpl",
		outputFilename: "devcats.go",
		data:           data,
	}
}

func parseHex(str string) int {
	i, err := strconv.ParseInt(str, 0, 64)
	if err != nil {
		log.Fatalf("Failed to parse %s: %v", str, err)
	}
	return int(i)
}

func data() interface{} {
	index := make(map[int]*domain)
	for _, domain := range domains {
		index[domain.id] = domain
	}

	file, err := os.Open("internal/devcats.csv")
	if err != nil {
		log.Fatalf("Failed to open devcats.csv: %v", err)
	}

	cr := csv.NewReader(file)
	for {
		record, err := cr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		domainID := parseHex(strings.Replace(record[0], "\ufeff", "", -1))
		categoryID := parseHex(strings.Replace(record[1], "\ufeff", "", -1))
		domain, found := index[domainID]
		if !found {
			log.Fatalf("Could not find domain matching %d", domainID)
		}
		domain.Categories = append(domain.Categories, category{domain: domain, id: categoryID, SKU: strings.TrimSpace(record[2]), Name: strings.TrimSpace(record[3])})
	}

	return domains
}
