package musashi

import (
	"testing"
)

// SimpleMemory is a basic memory implementation for testing
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

func TestNewCPU(t *testing.T) {
	cpu := NewCPU(CPU68000)
	if cpu == nil {
		t.Fatal("NewCPU returned nil")
	}
	if cpu.GetCPUType() != CPU68000 {
		t.Errorf("Expected CPU type 68000, got %v", cpu.GetCPUType())
	}
}

func TestCPUTypeString(t *testing.T) {
	tests := []struct {
		cpuType CPUType
		want    string
	}{
		{CPU68000, "68000"},
		{CPU68010, "68010"},
		{CPU68020, "68020"},
		{CPU68030, "68030"},
		{CPU68040, "68040"},
		{CPUInvalid, "Invalid"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.cpuType.String(); got != tt.want {
				t.Errorf("CPUType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCPUReset(t *testing.T) {
	cpu := NewCPU(CPU68000)
	memory := &SimpleMemory{}

	// Set up initial vectors
	memory.Write32(0, 0x00001000) // Initial SSP
	memory.Write32(4, 0x00000400) // Initial PC

	cpu.SetMemoryHandler(memory)
	cpu.Reset()

	if cpu.GetSP() != 0x00001000 {
		t.Errorf("Expected SP = 0x00001000, got 0x%08X", cpu.GetSP())
	}
	if cpu.GetPC() != 0x00000400 {
		t.Errorf("Expected PC = 0x00000400, got 0x%08X", cpu.GetPC())
	}

	// Check supervisor mode is set
	sr := cpu.GetSR()
	if sr&0x2000 == 0 {
		t.Error("Expected supervisor mode to be set after reset")
	}
}

func TestRegisterAccess(t *testing.T) {
	cpu := NewCPU(CPU68000)

	// Test data registers
	for i := RegD0; i <= RegD7; i++ {
		testValue := uint32(0x12345678 + i)
		cpu.SetRegister(i, testValue)
		if got := cpu.GetRegister(i); got != testValue {
			t.Errorf("D%d: expected 0x%08X, got 0x%08X", i-RegD0, testValue, got)
		}
	}

	// Test address registers
	for i := RegA0; i <= RegA7; i++ {
		testValue := uint32(0x87654321 + i)
		cpu.SetRegister(i, testValue)
		if got := cpu.GetRegister(i); got != testValue {
			t.Errorf("A%d: expected 0x%08X, got 0x%08X", i-RegA0, testValue, got)
		}
	}

	// Test PC
	cpu.SetPC(0xABCD1234)
	if got := cpu.GetPC(); got != 0xABCD1234 {
		t.Errorf("PC: expected 0xABCD1234, got 0x%08X", got)
	}

	// Test SR
	cpu.SetSR(0x2715)
	if got := cpu.GetSR(); got != 0x2715 {
		t.Errorf("SR: expected 0x2715, got 0x%04X", got)
	}
}

func TestMemoryHandler(t *testing.T) {
	cpu := NewCPU(CPU68000)
	memory := &SimpleMemory{}
	cpu.SetMemoryHandler(memory)

	// Test 8-bit access
	memory.Write8(0x100, 0x42)
	if got := memory.Read8(0x100); got != 0x42 {
		t.Errorf("Read8: expected 0x42, got 0x%02X", got)
	}

	// Test 16-bit access
	memory.Write16(0x200, 0x1234)
	if got := memory.Read16(0x200); got != 0x1234 {
		t.Errorf("Read16: expected 0x1234, got 0x%04X", got)
	}

	// Test 32-bit access
	memory.Write32(0x300, 0x12345678)
	if got := memory.Read32(0x300); got != 0x12345678 {
		t.Errorf("Read32: expected 0x12345678, got 0x%08X", got)
	}
}

func TestIRQHandling(t *testing.T) {
	cpu := NewCPU(CPU68000)
	memory := &SimpleMemory{}

	// Set up interrupt vectors
	memory.Write32(0, 0x00001000)    // Initial SSP
	memory.Write32(4, 0x00000400)    // Initial PC
	memory.Write32(0x64, 0x00000500) // Level 1 autovector

	cpu.SetMemoryHandler(memory)
	cpu.Reset()

	// Set IRQ level
	cpu.SetIRQ(1)

	// IRQ level should be set
	if cpu.irqLevel != 1 {
		t.Errorf("Expected IRQ level 1, got %d", cpu.irqLevel)
	}
}

func TestVirtualIRQ(t *testing.T) {
	cpu := NewCPU(CPU68000)

	// Set multiple virtual IRQ lines
	cpu.SetVIRQ(3, true)
	cpu.SetVIRQ(5, true)
	cpu.SetVIRQ(2, true)

	// Should select highest (5)
	if cpu.irqLevel != 5 {
		t.Errorf("Expected IRQ level 5, got %d", cpu.irqLevel)
	}

	// Clear highest
	cpu.SetVIRQ(5, false)

	// Should now be 3
	if cpu.irqLevel != 3 {
		t.Errorf("Expected IRQ level 3, got %d", cpu.irqLevel)
	}
}

func TestContextSaveRestore(t *testing.T) {
	cpu := NewCPU(CPU68000)

	// Set up some state
	cpu.SetRegister(RegD0, 0x11111111)
	cpu.SetRegister(RegD7, 0x77777777)
	cpu.SetRegister(RegA0, 0xAAAAAAAA)
	cpu.SetRegister(RegA7, 0xBBBBBBBB)
	cpu.SetPC(0x12345678)
	cpu.SetSR(0x2700)

	// Save context
	ctx := cpu.GetContext()

	// Modify CPU state
	cpu.SetRegister(RegD0, 0xFFFFFFFF)
	cpu.SetPC(0x00000000)
	cpu.SetSR(0x0000)

	// Restore context
	cpu.SetContext(ctx)

	// Verify state was restored
	if got := cpu.GetRegister(RegD0); got != 0x11111111 {
		t.Errorf("D0: expected 0x11111111, got 0x%08X", got)
	}
	if got := cpu.GetRegister(RegD7); got != 0x77777777 {
		t.Errorf("D7: expected 0x77777777, got 0x%08X", got)
	}
	if got := cpu.GetPC(); got != 0x12345678 {
		t.Errorf("PC: expected 0x12345678, got 0x%08X", got)
	}
	if got := cpu.GetSR(); got != 0x2700 {
		t.Errorf("SR: expected 0x2700, got 0x%04X", got)
	}
}

func TestCycleAccounting(t *testing.T) {
	cpu := NewCPU(CPU68000)
	memory := &SimpleMemory{}

	// Set up vectors
	memory.Write32(0, 0x00001000) // Initial SSP
	memory.Write32(4, 0x00000400) // Initial PC

	// Put NOP instructions in memory
	for i := uint32(0x400); i < 0x500; i += 2 {
		memory.Write16(i, 0x4E71) // NOP
	}

	cpu.SetMemoryHandler(memory)
	cpu.Reset()

	// Execute some cycles
	cycles := cpu.Execute(100)

	if cycles <= 0 {
		t.Error("Expected positive cycle count")
	}
	if cycles > 100 {
		t.Errorf("Expected cycles <= 100, got %d", cycles)
	}
}

func TestCPUTypeChange(t *testing.T) {
	cpu := NewCPU(CPU68000)
	if cpu.GetCPUType() != CPU68000 {
		t.Errorf("Expected CPU68000, got %v", cpu.GetCPUType())
	}

	cpu.SetCPUType(CPU68020)
	if cpu.GetCPUType() != CPU68020 {
		t.Errorf("Expected CPU68020, got %v", cpu.GetCPUType())
	}
}

func TestCallbacks(t *testing.T) {
	cpu := NewCPU(CPU68000)

	// Test int ack callback
	cpu.SetIntAckCallback(func(level int) uint32 {
		return IntAckAutovector
	})

	// Test reset callback
	cpu.SetResetCallback(func() {
	})

	// Test PC changed callback
	pcChangedCalled := false
	cpu.SetPCChangedCallback(func(newPC uint32) {
		pcChangedCalled = true
	})

	// Trigger PC changed
	cpu.SetPC(0x1000)
	if !pcChangedCalled {
		t.Error("PC changed callback was not called")
	}

	// Test instruction hook
	instrHookCalled := false
	cpu.SetInstrHookCallback(func(pc uint32) {
		instrHookCalled = true
	})

	// Set up minimal memory for execution
	memory := &SimpleMemory{}
	memory.Write32(0, 0x00001000) // Initial SSP
	memory.Write32(4, 0x00000400) // Initial PC
	memory.Write16(0x400, 0x4E71) // NOP

	cpu.SetMemoryHandler(memory)
	cpu.Reset()
	cpu.Execute(10)

	if !instrHookCalled {
		t.Error("Instruction hook callback was not called")
	}
}

// Example test showing basic usage
func ExampleCPU() {
	// Create a new 68000 CPU
	cpu := NewCPU(CPU68000)

	// Create simple memory
	memory := &SimpleMemory{}

	// Set up initial vectors
	memory.Write32(0, 0x00001000) // Stack pointer
	memory.Write32(4, 0x00000400) // Program counter

	// Set memory handler
	cpu.SetMemoryHandler(memory)

	// Reset CPU
	cpu.Reset()

	// Execute instructions
	cpu.Execute(1000)
}
