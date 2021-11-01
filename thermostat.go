//go:generate stringer -type=ThermostatMode,Unit,FanSpeed -linecomment
// Copyright 2019 Andrew Bates
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

package insteon

import (
	"errors"
	"fmt"
	"strings"
)

type ZoneInfo [4]int

func (zi ZoneInfo) Temperature() int {
	return zi[0] / 2
}

func (zi ZoneInfo) Deadband() int {
	return zi[1] / 2
}

func (zi ZoneInfo) Setpoint() int {
	return zi[2] / 2
}

func (zi ZoneInfo) Humidity() int {
	return zi[3]
}

type ThermostatMode byte

const (
	ThermostatOff ThermostatMode = 0x00 // Thermostat Off
	Heat          ThermostatMode = 0x01 // Heat Mode
	Cool          ThermostatMode = 0x02 // Cool Mode
	Auto          ThermostatMode = 0x03 // Auto Mode
	FanOn         ThermostatMode = 0x04 // Fan On
	ProgramAuto   ThermostatMode = 0x05 // Program Auto
	ProgramHeat   ThermostatMode = 0x06 // Program Heat
	ProgramCool   ThermostatMode = 0x07 // Program Cool
	FanOff        ThermostatMode = 0x08 // Fan Off
)

type Unit int

const (
	Fahrenheit Unit = 0x00 // °F
	Celsius    Unit = 0x01 // °C
)

type EquipmentState byte

func (es EquipmentState) CoolActive() bool                  { return es&0x01 == 0x01 }
func (es EquipmentState) HeatActive() bool                  { return es&0x02 == 0x02 }
func (es EquipmentState) ProgrammableOutputAvailable() bool { return es&0x04 == 0x04 }
func (es EquipmentState) ProgrammableOutputState() bool     { return es&0x08 == 0x08 }

func (es EquipmentState) String() string {
	values := []string{}
	if es.CoolActive() {
		values = append(values, "cooling")
	}
	if es.HeatActive() {
		values = append(values, "heating")
	}
	if es.ProgrammableOutputAvailable() {
		values = append(values, "programmable output available")
	}
	if es.ProgrammableOutputState() {
		values = append(values, "programmable output state")
	}
	return strings.Join(values, ",")
}

type FanSpeed int

const (
	SingleSpeed FanSpeed = 0x00 // Single Speed
	LowSpeed    FanSpeed = 0x01 // Low Speed
	MedSpeed    FanSpeed = 0x02 // Medium Speed
	HighSpeed   FanSpeed = 0x03 // High Speed
)

type ThermostatStatus struct {
	Temperature int
	Humidity    int
	Setpoint    int
	Deadband    int
	Unit        Unit
	Mode        ThermostatMode
	State       EquipmentState
}

type ThermostatFlags byte

func (tf ThermostatFlags) LinkingLock() bool { return tf&0x01 == 0x01 }
func (tf ThermostatFlags) ButtonBeep() bool  { return tf&0x02 == 0x02 }
func (tf ThermostatFlags) ButtonLock() bool  { return tf&0x04 == 0x04 }
func (tf ThermostatFlags) TempFormat() Unit  { return Unit((tf & 0x08) >> 3) }
func (tf ThermostatFlags) TimeFormat() int   { return int((tf & 0x10) >> 5) }

type ThermostatInfo struct {
	// Data Set 1
	Temp              float32
	Humidity          int
	TempOffset        int
	HumidityOffset    int
	Mode              ThermostatMode
	FanMode           int
	BacklightSeconds  int
	HysteresisMinutes int
	Flags             ThermostatFlags

	// Data Set 2
	HumidityLow         int
	HumidityHigh        int
	Rev                 int
	CoolSetPoint        int
	HeatSetPoint        int
	RFOffset            int
	EnergySetbackPoint  int
	ExternalTempOffset  int
	StatusReportEnabled bool
	ExternalPower       bool
	ExternalTemp        bool
}

func (ti *ThermostatInfo) UnmarshalBinary(data []byte) error {
	if len(data) < 14 {
		return newBufError(ErrBufferTooShort, 14, len(data))
	}
	if data[2] == 0x00 {
		// temp high byte is data[13], temp low byte is data[3]
		temp := int(data[13])<<8 | int(data[3])
		ti.Humidity = int(data[4])
		ti.TempOffset = int(data[5])
		ti.HumidityOffset = int(data[6])
		// These next two fields don't seem right to me
		ti.Mode = ThermostatMode(data[7])
		ti.FanMode = int(data[8])
		ti.BacklightSeconds = int(data[9])
		ti.HysteresisMinutes = int(data[10])
		// Data[11] unused
		ti.Flags = ThermostatFlags(data[12])
		ti.Rev = int(data[13])

		// Local temp is stored as tenths of a degree celsius
		ti.Temp = float32(temp) / 10
		// Convert temp if necessary
		if ti.Flags.TempFormat() == Fahrenheit {
			ti.Temp = (ti.Temp * 9 / 5) + 32
		}
	} else if data[2] == 0x01 {
		ti.HumidityLow = int(data[4])
		ti.HumidityHigh = int(data[3])
		ti.Rev = int(data[5])
		ti.CoolSetPoint = int(data[6])
		ti.HeatSetPoint = int(data[7])
		ti.RFOffset = int(data[8])
		ti.EnergySetbackPoint = int(data[9])
		ti.ExternalTempOffset = int(data[10])
		ti.StatusReportEnabled = (data[11] == 0x01)
		ti.ExternalPower = (data[12] == 0x01)
		ti.ExternalTemp = (data[13] == 0x02)
	} else {
		// TODO need a better error here
		return errors.New("Invalid return indicator")
	}

	return nil
}

func (ti *ThermostatInfo) String() string {
	builder := &strings.Builder{}
	builder.WriteString(fmt.Sprintf("                Temp: %.01f %s\n", ti.Temp, ti.Flags.TempFormat()))
	builder.WriteString(fmt.Sprintf("            Humidity: %d %%\n", ti.Humidity))
	builder.WriteString(fmt.Sprintf("          Temp Offset: %d\n", ti.TempOffset))
	builder.WriteString(fmt.Sprintf("      Humidity Offset: %d\n", ti.HumidityOffset))
	builder.WriteString(fmt.Sprintf("                 Mode: %s\n", ti.Mode))
	builder.WriteString(fmt.Sprintf("             Fan Mode: %d\n", ti.FanMode))
	builder.WriteString(fmt.Sprintf("    Backlight Seconds: %d\n", ti.BacklightSeconds))
	builder.WriteString(fmt.Sprintf("   Hysteresis Minutes: %d\n", ti.HysteresisMinutes))
	builder.WriteString(fmt.Sprintf(" Energy Set Backpoint: %d\n", ti.EnergySetbackPoint))
	builder.WriteString(fmt.Sprintf("   Temperature Format: %s\n", ti.Flags.TempFormat()))
	builder.WriteString(fmt.Sprintf("         Humidity Low: %d\n", ti.HumidityLow))
	builder.WriteString(fmt.Sprintf("        Humidity High: %d\n", ti.HumidityHigh))
	builder.WriteString(fmt.Sprintf("                  Rev: %d\n", ti.Rev))
	builder.WriteString(fmt.Sprintf("       Cool Set Point: %d\n", ti.CoolSetPoint))
	builder.WriteString(fmt.Sprintf("       Heat Set Point: %d\n", ti.HeatSetPoint))
	builder.WriteString(fmt.Sprintf("            RF Offset: %d\n", ti.RFOffset))
	builder.WriteString(fmt.Sprintf(" Energy Set Backpoint: %d\n", ti.EnergySetbackPoint))
	builder.WriteString(fmt.Sprintf(" External Temp Offset: %d\n", ti.ExternalTempOffset))
	builder.WriteString(fmt.Sprintf("Status Report Enabled: %v\n", ti.StatusReportEnabled))
	builder.WriteString(fmt.Sprintf("       External Power: %v\n", ti.ExternalPower))
	builder.WriteString(fmt.Sprintf("        External Temp: %v\n", ti.ExternalTemp))
	return builder.String()
}

type Thermostat struct {
	*device
	info DeviceInfo
}

// NewThermostat will return a configured Thermostat object
func NewThermostat(d *device, info DeviceInfo) *Thermostat {
	therm := &Thermostat{device: d, info: info}

	return therm
}

func (therm *Thermostat) Address() Address { return therm.info.Address }

func (therm *Thermostat) IncreaseTemp(delta int) error {
	return therm.SendCommand(CmdIncreaseTemp.SubCommand(delta*2), nil)
}

func (therm *Thermostat) DecreaseTemp(delta int) error {
	return therm.SendCommand(CmdDecreaseTemp.SubCommand(delta*2), nil)
}

func (therm *Thermostat) GetZoneInfo(zone int) (zi ZoneInfo, err error) {
	commands := []Command{
		CmdGetZoneInfo.SubCommand(zone & 0x0f),
		CmdGetZoneInfo.SubCommand(0x10 | zone&0x0f),
		CmdGetZoneInfo.SubCommand(0x20 | zone&0x0f),
		CmdGetZoneInfo.SubCommand(0x30 | zone&0x0f),
	}

	var ack *Message
	for i := 0; i < len(commands) && err == nil; i++ {
		ack, err = therm.Write(&Message{Command: commands[i]})
		if err == nil {
			zi[i] = ack.Command.Command2()
		}
	}
	return
}

func (therm *Thermostat) GetMode() (mode ThermostatMode, err error) {
	ack, err := therm.Send(CmdGetThermostatMode, nil)
	if err == nil {
		mode = ThermostatMode(ack.Command2())
	}
	return
}

func (therm *Thermostat) GetAmbientTemp() (temp int, err error) {
	ack, err := therm.Send(CmdGetAmbientTemp, nil)
	if err == nil {
		temp = int(ack.Command2())
	}
	return
}

func (therm *Thermostat) SetMode(mode ThermostatMode) error {
	cmd := Command(0x00)
	switch mode {
	case ThermostatOff:
		cmd = CmdThermOff
	case Heat:
		cmd = CmdSetHeat
	case Cool:
		cmd = CmdSetCool
	case Auto:
		cmd = CmdSetModeAuto
	case FanOn:
		cmd = CmdSetFan
	case FanOff:
		cmd = CmdClearFan
	case ProgramAuto:
		cmd = CmdSetProgramAuto
	case ProgramHeat:
		cmd = CmdSetProgramHeat
	case ProgramCool:
		cmd = CmdSetProgramCool
	default:
		return ErrInvalidThermostatMode
	}
	return therm.SendCommand(cmd, nil)
}

func (therm *Thermostat) GetEquipmentState() (EquipmentState, error) {
	ack, err := therm.Send(CmdGetEquipmentState, nil)
	return EquipmentState(ack.Command2()), err
}

func (therm *Thermostat) GetInfo() (ti ThermostatInfo, err error) {
	//if therm, ok := therm.Device.(ExtendedGetSet); ok {
	var buf []byte
	for i := 0; i < 2 && err == nil; i++ {
		buf, err = therm.ExtendedGet([]byte{0x00, 0x00, byte(i)})
		if err == nil {
			err = ti.UnmarshalBinary(buf)
		}
	}
	//} else {
	//err = ErrNotSupported
	//}

	return ti, err
}

func (therm *Thermostat) GetTempUnit() (Unit, error) {
	ti, err := therm.GetInfo()
	return ti.Flags.TempFormat(), err
}

func (therm *Thermostat) SetTempUnit(unit Unit) error {
	if unit == Celsius {
		return therm.SendCommand(CmdExtendedGetSet, []byte{0x00, 0x04, 0x00, 0x08})
	} else if unit == Fahrenheit {
		return therm.SendCommand(CmdExtendedGetSet, []byte{0x00, 0x04, 0x00, 0x00})
	}
	return ErrInvalidUnit
}

func (therm *Thermostat) GetFanSpeed() (FanSpeed, error) {
	ack, err := therm.Send(CmdGetFanOnSpeed, nil)
	return FanSpeed(ack.Command2()), err
}

func (therm *Thermostat) SetFanSpeen(speed FanSpeed) error {
	cmd := Command(0x00)
	switch speed {
	case LowSpeed:
		cmd = CmdSetFanOnLow
	case MedSpeed:
		cmd = CmdSetFanOnMed
	case HighSpeed:
		cmd = CmdSetFanOnHigh
	default:
		return ErrInvalidFanSpeed
	}
	return therm.SendCommand(cmd, nil)
}

func (therm *Thermostat) Status() (status ThermostatStatus, err error) {
	info, err := therm.GetZoneInfo(0)
	status.Temperature = info.Temperature()
	status.Humidity = info.Humidity()
	status.Setpoint = info.Setpoint()
	status.Deadband = info.Deadband()
	if err == nil {
		status.Unit, err = therm.GetTempUnit()
	}

	if err == nil {
		status.Mode, err = therm.GetMode()
	}

	if err == nil {
		status.State, err = therm.GetEquipmentState()
	}
	return
}

func (therm *Thermostat) SetStatusMessage(enabled bool) error {
	if enabled {
		return therm.SendCommand(CmdExtendedGetSet, []byte{0x00, 0x08, 0x00})
	}
	return therm.SendCommand(CmdDisableStatusMessage, nil)
}

func (therm *Thermostat) SetCoolSetpoint(zone int, temp int) error {
	return therm.SendCommand(CmdSetCoolSetpoint.SubCommand(temp*2), nil)
}

func (therm *Thermostat) SetHeatSetpoint(zone int, temp int) error {
	return therm.SendCommand(CmdSetHeatSetpoint.SubCommand(temp*2), nil)
}

func (therm *Thermostat) IncreaseZoneTemp(zone int, delta int) error {
	return therm.SendCommand(CmdZoneTempUp.SubCommand(zone), []byte{byte(delta * 2)})
}

func (therm *Thermostat) DecreaseZoneTemp(zone int, delta int) error {
	return therm.SendCommand(CmdZoneTempDown.SubCommand(zone), []byte{byte(delta * 2)})
}

func (therm *Thermostat) SetZoneCoolSetpoint(zone int, temp, deadband int) error {
	return therm.SendCommand(CmdSetZoneCoolSetpoint.SubCommand(zone), []byte{byte(temp * 2), byte(deadband)})
}

func (therm *Thermostat) SetZoneHeatSetpoint(zone int, temp, deadband int) error {
	return therm.SendCommand(CmdSetZoneHeatSetpoint.SubCommand(zone), []byte{byte(temp * 2), byte(deadband)})
}

func (therm *Thermostat) String() string {
	return fmt.Sprintf("Thermostat (%s)", therm.info.Address)
}
