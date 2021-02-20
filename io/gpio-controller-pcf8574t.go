package io

import (
	"fmt"
	"log"
	"sync"
)

const pinCount = 8

// GPIOControllerPCF8574T type
type GPIOControllerPCF8574T struct {
	bus          *I2C
	mutex        sync.Mutex
	address      byte
	valuePinMask byte
	inputPinMask byte
}

// NewGPIOControllerPCF8574T factory
func NewGPIOControllerPCF8574T(bus *I2C, address byte) *GPIOControllerPCF8574T {
	return &GPIOControllerPCF8574T{
		bus:          bus,
		address:      address,
		inputPinMask: 0,
		valuePinMask: 0,
	}
}

func (g *GPIOControllerPCF8574T) isInput(index int) bool {
	return (byte(1<<index) & g.inputPinMask) > 0
}

func (g *GPIOControllerPCF8574T) isOn(index int) bool {
	return (byte(1<<index) & g.valuePinMask) > 0
}

// Sync needs to be called to write the values to the controllser
func (g *GPIOControllerPCF8574T) sync() {
	err := g.bus.WriteByte(g.address, g.valuePinMask)
	if err != nil {
		fmt.Printf("Error syncing gpio controller: %v\n", err)
	}
}

// Count returns number of ports
func (g *GPIOControllerPCF8574T) Count() int {
	return pinCount
}

// Get reads current value
func (g *GPIOControllerPCF8574T) Get(index int) bool {

	if index < 0 || index > pinCount {
		fmt.Printf("Input out of range for gpio: %d", index)
		return false
	}

	g.mutex.Lock()
	defer g.mutex.Unlock()

	if !g.isInput(index) {
		return g.isOn(index)
	}

	log.Fatal("Input not implemented")

	return false
}

// Set sets a value
func (g *GPIOControllerPCF8574T) Set(index int, on bool) {
	if index < 0 || index > pinCount {
		fmt.Printf("Input out of range for gpio: %d", index)
		return
	}

	g.mutex.Lock()
	defer g.mutex.Unlock()
	if on {
		g.valuePinMask |= (1 << index)
	} else {
		g.valuePinMask &= ((1 << index) ^ 0xff)
	}
	g.sync()
}

// IsInput returns true if pin is input
func (g *GPIOControllerPCF8574T) IsInput(index int) bool {
	if index < 0 || index > pinCount {
		fmt.Printf("Input out of range for gpio: %d", index)
		return false
	}

	return false
}

// SetInput sets if a pin is input
func (g *GPIOControllerPCF8574T) SetInput(index int, on bool) {
	if index < 0 || index > pinCount {
		fmt.Printf("Input out of range for gpio: %d", index)
		return
	}

	if !on {
		return
	}

	log.Fatal("Input not implemented")
}
