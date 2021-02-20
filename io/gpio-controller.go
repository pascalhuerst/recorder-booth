package io

// GPIOController interface
type GPIOController interface {
	Count() int
	Get(index int) bool
	Set(index int, on bool)
	IsInput(index int) bool
	SetInput(index int, on bool)
}
