# Musashi-Go

A pure Go port of the Musashi Motorola 68000 emulator library.

## Overview

Musashi-Go is a complete port of Karl Stenerud's Musashi M68000 emulator from C to Go. It provides accurate emulation of the Motorola 68000 family of processors, including:

- **Motorola 68000** - The original 16/32-bit processor
- **Motorola 68010** - Enhanced with virtual memory support
- **Motorola 68020** - Full 32-bit processor
- **Motorola 68030** - Added MMU and data cache
- **Motorola 68040** - Added FPU and improved performance

## Features

- ✅ **Pure Go Implementation** - No CGo, no C dependencies
- ✅ **Idiomatic Go API** - Follows Go conventions and best practices
- ✅ **Thread-Safe Design** - No global state, multiple CPU instances supported
- ✅ **Flexible Memory Interface** - Implement your own memory handlers via interfaces
- ✅ **Complete Instruction Set** - All M68000 family instructions implemented
- ✅ **Accurate Emulation** - Cycle-accurate timing and behavior
- ✅ **Disassembler** - Built-in disassembler for all CPU types

## Installation

```bash
go get github.com/hansbonini/musashi-go
```

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/hansbonini/musashi-go"
)

// Implement the memory interface
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
    // Create a new CPU instance
    cpu := musashi.NewCPU(musashi.CPU68000)
    
    // Set up memory
    memory := &SimpleMemory{}
    cpu.SetMemoryHandler(memory)
    
    // Load a simple program into memory
    // Example: MOVE.W #$1234, D0 ; RTS
    program := []byte{
        0x30, 0x3C, 0x12, 0x34, // MOVE.W #$1234, D0
        0x4E, 0x75,             // RTS
    }
    for i, b := range program {
        memory.Write8(uint32(i), b)
    }
    
    // Reset the CPU
    cpu.Reset()
    
    // Execute some cycles
    cyclesExecuted := cpu.Execute(100)
    fmt.Printf("Executed %d cycles\n", cyclesExecuted)
    
    // Read register values
    d0 := cpu.GetRegister(musashi.RegD0)
    fmt.Printf("D0 = 0x%08X\n", d0)
}
```

## API Documentation

### CPU Creation and Management

```go
// Create a new CPU instance
cpu := musashi.NewCPU(cpuType musashi.CPUType) *musashi.CPU

// CPU types
musashi.CPU68000
musashi.CPU68010
musashi.CPU68020
musashi.CPU68030
musashi.CPU68040
```

### CPU Control

```go
// Reset the CPU
cpu.Reset()

// Execute instructions for a number of cycles
cyclesUsed := cpu.Execute(cycles int) int

// Set interrupt request level (0-7)
cpu.SetIRQ(level int)

// Pulse the HALT pin
cpu.PulseHalt()

// Trigger a bus error
cpu.PulseBusError()
```

### Register Access

```go
// Get register value
value := cpu.GetRegister(reg musashi.Register) uint32

// Set register value
cpu.SetRegister(reg musashi.Register, value uint32)

// Convenience methods
pc := cpu.GetPC()
cpu.SetPC(address uint32)
sp := cpu.GetSP()
cpu.SetSP(address uint32)
sr := cpu.GetSR()
cpu.SetSR(value uint16)
```

### Available Registers

- **Data Registers**: `RegD0` through `RegD7`
- **Address Registers**: `RegA0` through `RegA7`
- **Special Registers**: `RegPC`, `RegSR`, `RegSP`, `RegUSP`, `RegISP`, `RegMSP`
- **68010+ Registers**: `RegVBR`, `RegSFC`, `RegDFC`
- **68020+ Registers**: `RegCACR`, `RegCAAR`

### Memory Interface

Implement the `MemoryHandler` interface to provide memory access:

```go
type MemoryHandler interface {
    Read8(address uint32) uint8
    Read16(address uint32) uint16
    Read32(address uint32) uint32
    Write8(address uint32, value uint8)
    Write16(address uint32, value uint16)
    Write32(address uint32, value uint32)
}

cpu.SetMemoryHandler(handler MemoryHandler)
```

### Context Management (Multiple CPUs)

```go
// Get CPU context for saving
context := cpu.GetContext() *musashi.Context

// Restore CPU context
cpu.SetContext(context *musashi.Context)

// Get context size
size := cpu.ContextSize() int
```

### Disassembler

```go
// Disassemble instruction at address
instruction, size := cpu.Disassemble(address uint32) (string, int)
```

## Comparison with Original C Library

| C API | Go API | Notes |
|-------|--------|-------|
| `m68k_init()` | `musashi.NewCPU(type)` | Constructor pattern |
| `m68k_execute(cycles)` | `cpu.Execute(cycles)` | Method on CPU struct |
| `m68k_set_cpu_type(type)` | `cpu.SetCPUType(type)` | Method on CPU struct |
| `m68k_pulse_reset()` | `cpu.Reset()` | Cleaner naming |
| `m68k_get_reg(ctx, reg)` | `cpu.GetRegister(reg)` | Simplified API |
| `m68k_set_reg(reg, val)` | `cpu.SetRegister(reg, val)` | Simplified API |
| Global functions | Methods on `*CPU` | No global state |
| Callbacks via function pointers | Interface implementation | Go idiomatic |

## Performance Notes

Musashi-Go aims to maintain the performance characteristics of the original C implementation:

- Instruction dispatch uses computed goto (via switch in Go)
- Critical hot paths are optimized
- Memory access is abstracted but efficient
- No reflection used in hot paths

While Go's garbage collector adds some overhead compared to C, the overall performance is suitable for most emulation needs, including real-time emulation of 68000-based systems.

## Examples

See the `examples/` directory for complete working examples:

- **simple** - Basic CPU initialization and execution
- **memory** - Custom memory implementation with ROM/RAM
- **interrupts** - Interrupt handling example
- **disassembler** - Using the built-in disassembler

## Testing

Run the test suite:

```bash
go test ./...
```

Run with coverage:

```bash
go test -cover ./...
```

## Original Musashi

This is a port of the original Musashi emulator by Karl Stenerud:
- Repository: https://github.com/kstenerud/Musashi
- Version: 4.10

## License

MIT License - See LICENSE file for details.

This project maintains the same MIT license as the original Musashi library.

Copyright (c) 2025 Hans Bonini (Go port)  
Copyright (c) 1998-2002 Karl Stenerud (Original Musashi)

## Contributing

Contributions are welcome! Please ensure:

1. Code follows Go conventions (`go fmt`, `go vet`)
2. All tests pass
3. New features include tests
4. API remains idiomatic Go

## Acknowledgments

- **Karl Stenerud** - Original Musashi emulator author
- The **MAME project** - Where Musashi has been battle-tested for years
- All contributors to the original Musashi project
