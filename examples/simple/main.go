package main

import (
	"fmt"

	"github.com/hansbonini/musashi-go"
)

// SimpleMemory implements a basic memory system
type SimpleMemory struct {
	ram [1024 * 1024]byte // 1MB of RAM
}

func (m *SimpleMemory) Read8(address uint32) uint8 {
	return m.ram[address&0xFFFFF]
}

func (m *SimpleMemory) Read16(address uint32) uint16 {
	addr := address & 0xFFFFF
	return uint16(m.ram[addr])<<8 | uint16(m.ram[addr+1])
}

func (m *SimpleMemory) Read32(address uint32) uint32 {
	addr := address & 0xFFFFF
	return uint32(m.ram[addr])<<24 | uint32(m.ram[addr+1])<<16 |
		uint32(m.ram[addr+2])<<8 | uint32(m.ram[addr+3])
}

func (m *SimpleMemory) Write8(address uint32, value uint8) {
	m.ram[address&0xFFFFF] = value
}

func (m *SimpleMemory) Write16(address uint32, value uint16) {
	addr := address & 0xFFFFF
	m.ram[addr] = uint8(value >> 8)
	m.ram[addr+1] = uint8(value)
}

func (m *SimpleMemory) Write32(address uint32, value uint32) {
	addr := address & 0xFFFFF
	m.ram[addr] = uint8(value >> 24)
	m.ram[addr+1] = uint8(value >> 16)
	m.ram[addr+2] = uint8(value >> 8)
	m.ram[addr+3] = uint8(value)
}

func main() {
	fmt.Println("Musashi M68000 Emulator - Simple Example")
	fmt.Println("=========================================")
	fmt.Println()

	// Create a new 68000 CPU
	cpu := musashi.NewCPU(musashi.CPU68000)
	fmt.Printf("Created CPU: %s\n", cpu.GetCPUType())

	// Create and set up memory
	memory := &SimpleMemory{}
	cpu.SetMemoryHandler(memory)

	// Set up initial vectors
	// At address 0: Initial SSP (supervisor stack pointer)
	// At address 4: Initial PC (program counter)
	memory.Write32(0, 0x00001000) // Stack starts at 0x1000
	memory.Write32(4, 0x00000400) // Program starts at 0x400

	// Write a simple program to memory
	// This program will:
	// 1. Load some values into registers
	// 2. Perform some operations
	// 3. Loop forever (for now, since we don't have full instruction set)

	programStart := uint32(0x400)

	// For now, just fill with NOPs as placeholder
	// NOP instruction = 0x4E71
	for i := uint32(0); i < 0x100; i += 2 {
		memory.Write16(programStart+i, 0x4E71) // NOP
	}

	fmt.Println("\nInitializing CPU...")

	// Reset the CPU
	cpu.Reset()

	fmt.Printf("After reset:\n")
	fmt.Printf("  PC = 0x%08X\n", cpu.GetPC())
	fmt.Printf("  SP = 0x%08X\n", cpu.GetSP())
	fmt.Printf("  SR = 0x%04X\n", cpu.GetSR())

	// Set up some initial register values
	cpu.SetRegister(musashi.RegD0, 0x00001234)
	cpu.SetRegister(musashi.RegD1, 0x00005678)
	cpu.SetRegister(musashi.RegA0, 0x00002000)

	fmt.Println("\nRegister values before execution:")
	for i := musashi.RegD0; i <= musashi.RegD7; i++ {
		fmt.Printf("  D%d = 0x%08X\n", i-musashi.RegD0, cpu.GetRegister(i))
	}
	for i := musashi.RegA0; i <= musashi.RegA7; i++ {
		fmt.Printf("  A%d = 0x%08X\n", i-musashi.RegA0, cpu.GetRegister(i))
	}

	// Execute some cycles
	fmt.Println("\nExecuting 1000 cycles...")
	cyclesExecuted := cpu.Execute(1000)
	fmt.Printf("Executed %d cycles\n", cyclesExecuted)

	fmt.Printf("\nAfter execution:\n")
	fmt.Printf("  PC = 0x%08X\n", cpu.GetPC())
	fmt.Printf("  Cycles run = %d\n", cpu.CyclesRun())

	fmt.Println("\nRegister values after execution:")
	for i := musashi.RegD0; i <= musashi.RegD7; i++ {
		fmt.Printf("  D%d = 0x%08X\n", i-musashi.RegD0, cpu.GetRegister(i))
	}

	fmt.Println("\n========================================")
	fmt.Println("Note: Full instruction set implementation")
	fmt.Println("is in progress. This example demonstrates")
	fmt.Println("the basic CPU structure and API.")
}
