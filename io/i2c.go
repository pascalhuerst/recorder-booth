package io

import (
	"fmt"
	"os"
	"sync"
	"syscall"
)

// I2C abstracts an i2c-dev device
type I2C struct {
	busNumber byte
	file      *os.File
	addr      byte
	mu        sync.Mutex

	initialized bool
}

// NewI2C factory
func NewI2C(busNumber byte) *I2C {
	return &I2C{busNumber: busNumber}
}

const (
	delay    = 20
	slaveCmd = 0x0703 // Cmd to set slave address
	rdrwCmd  = 0x0707 // Cmd to read/write data together
	rd       = 0x0001
)

type i2cMsg struct {
	addr  uint16
	flags uint16
	len   uint16
	buf   uintptr
}

type i2cRdwrIoctlData struct {
	msgs uintptr
	nmsg uint32
}

func (b *I2C) init() error {
	if b.initialized {
		return nil
	}

	var err error
	if b.file, err = os.OpenFile(fmt.Sprintf("/dev/i2c-%v", b.busNumber), os.O_RDWR, os.ModeExclusive); err != nil {
		return err
	}

	b.initialized = true
	return nil
}

func (b *I2C) setAddress(addr byte) error {
	if addr != b.addr {
		if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, b.file.Fd(), slaveCmd, uintptr(addr)); errno != 0 {
			return syscall.Errno(errno)
		}

		b.addr = addr
	}

	return nil
}

// ReadByte reads a byte from i2c
func (b *I2C) ReadByte(addr byte) (byte, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if err := b.init(); err != nil {
		return 0, err
	}

	if err := b.setAddress(addr); err != nil {
		return 0, err
	}

	bytes := make([]byte, 1)
	n, _ := b.file.Read(bytes)

	if n != 1 {
		return 0, fmt.Errorf("i2c: Unexpected number (%v) of bytes read", n)
	}

	return bytes[0], nil
}

func (b *I2C) WriteByte(addr, value byte) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if err := b.init(); err != nil {
		return err
	}

	if err := b.setAddress(addr); err != nil {
		return err
	}

	n, err := b.file.Write([]byte{value})

	if n != 1 {
		err = fmt.Errorf("i2c: Unexpected number (%v) of bytes written in WriteByte", n)
	}

	return err
}

func (b *I2C) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if !b.initialized {
		return nil
	}

	return b.file.Close()
}
