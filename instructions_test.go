package musashi

import (
	"testing"
)

// TestMOVEQInstruction tests the MOVEQ instruction
func TestMOVEQInstruction(t *testing.T) {
	cpu := NewCPU(CPU68000)
	memory := &SimpleMemory{}
	cpu.SetMemoryHandler(memory)

	// Set up vectors
	memory.Write32(0, 0x00001000) // Initial SSP
	memory.Write32(4, 0x00000400) // Initial PC

	// MOVEQ #$42, D0 = 0x7042
	memory.Write16(0x400, 0x7042)

	cpu.Reset()
	cpu.Execute(10)

	// Check D0 = 0x42
	if cpu.d[0] != 0x00000042 {
		t.Errorf("Expected D0 = 0x42, got 0x%08X", cpu.d[0])
	}

	// Check Z flag clear, N flag clear
	if cpu.sr&FlagZ != 0 {
		t.Error("Z flag should be clear")
	}
	if cpu.sr&FlagN != 0 {
		t.Error("N flag should be clear")
	}
}

// TestADDQInstruction tests the ADDQ instruction
func TestADDQInstruction(t *testing.T) {
	cpu := NewCPU(CPU68000)
	memory := &SimpleMemory{}
	cpu.SetMemoryHandler(memory)

	// Set up vectors
	memory.Write32(0, 0x00001000)
	memory.Write32(4, 0x00000400)

	cpu.Reset()

	// Set D0 = 10
	cpu.d[0] = 10

	// ADDQ.L #5, D0 = 0x5A80 (5 << 9 | size=2 << 6 | mode=0 << 3 | reg=0)
	memory.Write16(0x400, 0x5A80)

	cpu.Execute(10)

	// Check D0 = 15
	if cpu.d[0] != 15 {
		t.Errorf("Expected D0 = 15, got %d", cpu.d[0])
	}
}

// TestSUBQInstruction tests the SUBQ instruction
func TestSUBQInstruction(t *testing.T) {
	cpu := NewCPU(CPU68000)
	memory := &SimpleMemory{}
	cpu.SetMemoryHandler(memory)

	memory.Write32(0, 0x00001000)
	memory.Write32(4, 0x00000400)

	cpu.Reset()

	// Set D0 = 10
	cpu.d[0] = 10

	// SUBQ.L #3, D0 = 0x5780 (3 << 9 | dir=1 << 8 | size=2 << 6 | mode=0 << 3 | reg=0)
	memory.Write16(0x400, 0x5780)

	cpu.Execute(10)

	// Check D0 = 7
	if cpu.d[0] != 7 {
		t.Errorf("Expected D0 = 7, got %d", cpu.d[0])
	}
}

// TestANDInstruction tests the AND instruction
func TestANDInstruction(t *testing.T) {
	cpu := NewCPU(CPU68000)
	memory := &SimpleMemory{}
	cpu.SetMemoryHandler(memory)

	memory.Write32(0, 0x00001000)
	memory.Write32(4, 0x00000400)

	cpu.Reset()

	// Set D0 = 0xFF, D1 = 0x0F
	cpu.d[0] = 0xFF
	cpu.d[1] = 0x0F

	// AND.B D1, D0 = 0xC001 (reg=0 << 9 | dir=0 << 8 | size=0 << 6 | mode=0 << 3 | reg=1)
	memory.Write16(0x400, 0xC001)

	cpu.Execute(10)

	// Check D0 lower byte = 0x0F
	if (cpu.d[0] & 0xFF) != 0x0F {
		t.Errorf("Expected D0 = 0x0F, got 0x%02X", cpu.d[0]&0xFF)
	}
}

// TestORInstruction tests the OR instruction
func TestORInstruction(t *testing.T) {
	cpu := NewCPU(CPU68000)
	memory := &SimpleMemory{}
	cpu.SetMemoryHandler(memory)

	memory.Write32(0, 0x00001000)
	memory.Write32(4, 0x00000400)

	cpu.Reset()

	// Set D0 = 0xF0, D1 = 0x0F
	cpu.d[0] = 0xF0
	cpu.d[1] = 0x0F

	// OR.B D1, D0 = 0x8001
	memory.Write16(0x400, 0x8001)

	cpu.Execute(10)

	// Check D0 lower byte = 0xFF
	if (cpu.d[0] & 0xFF) != 0xFF {
		t.Errorf("Expected D0 = 0xFF, got 0x%02X", cpu.d[0]&0xFF)
	}
}

// TestEORInstruction tests the EOR instruction
func TestEORInstruction(t *testing.T) {
	cpu := NewCPU(CPU68000)
	memory := &SimpleMemory{}
	cpu.SetMemoryHandler(memory)

	memory.Write32(0, 0x00001000)
	memory.Write32(4, 0x00000400)

	cpu.Reset()

	// Set D0 = 0xFF, D1 = 0x0F
	cpu.d[0] = 0xFF
	cpu.d[1] = 0x0F

	// EOR.B D1, D0 = 0xB100 (reg=1 << 9 | size=0 << 6 | mode=0 << 3 | reg=0)
	memory.Write16(0x400, 0xB100)

	cpu.Execute(10)

	// Check D0 lower byte = 0xF0
	if (cpu.d[0] & 0xFF) != 0xF0 {
		t.Errorf("Expected D0 = 0xF0, got 0x%02X", cpu.d[0]&0xFF)
	}
}

// TestNOTInstruction tests the NOT instruction
func TestNOTInstruction(t *testing.T) {
	cpu := NewCPU(CPU68000)
	memory := &SimpleMemory{}
	cpu.SetMemoryHandler(memory)

	memory.Write32(0, 0x00001000)
	memory.Write32(4, 0x00000400)

	cpu.Reset()

	// Set D0 = 0xAA
	cpu.d[0] = 0xAA

	// NOT.B D0 = 0x4600
	memory.Write16(0x400, 0x4600)

	cpu.Execute(10)

	// Check D0 lower byte = 0x55
	if (cpu.d[0] & 0xFF) != 0x55 {
		t.Errorf("Expected D0 = 0x55, got 0x%02X", cpu.d[0]&0xFF)
	}
}

// TestCLRInstruction tests the CLR instruction
func TestCLRInstruction(t *testing.T) {
	cpu := NewCPU(CPU68000)
	memory := &SimpleMemory{}
	cpu.SetMemoryHandler(memory)

	memory.Write32(0, 0x00001000)
	memory.Write32(4, 0x00000400)

	cpu.Reset()

	// Set D0 = 0xFFFFFFFF
	cpu.d[0] = 0xFFFFFFFF

	// CLR.L D0 = 0x4280
	memory.Write16(0x400, 0x4280)

	cpu.Execute(10)

	// Check D0 = 0
	if cpu.d[0] != 0 {
		t.Errorf("Expected D0 = 0, got 0x%08X", cpu.d[0])
	}

	// Check Z flag set
	if cpu.sr&FlagZ == 0 {
		t.Error("Z flag should be set")
	}
}

// TestNEGInstruction tests the NEG instruction
func TestNEGInstruction(t *testing.T) {
	cpu := NewCPU(CPU68000)
	memory := &SimpleMemory{}
	cpu.SetMemoryHandler(memory)

	memory.Write32(0, 0x00001000)
	memory.Write32(4, 0x00000400)

	cpu.Reset()

	// Set D0 = 5
	cpu.d[0] = 5

	// NEG.B D0 = 0x4400
	memory.Write16(0x400, 0x4400)

	cpu.Execute(10)

	// Check D0 lower byte = 0xFB (-5 in two's complement)
	if (cpu.d[0] & 0xFF) != 0xFB {
		t.Errorf("Expected D0 = 0xFB, got 0x%02X", cpu.d[0]&0xFF)
	}

	// Check N flag set
	if cpu.sr&FlagN == 0 {
		t.Error("N flag should be set")
	}
}

// TestCMPInstruction tests the CMP instruction
func TestCMPInstruction(t *testing.T) {
	cpu := NewCPU(CPU68000)
	memory := &SimpleMemory{}
	cpu.SetMemoryHandler(memory)

	memory.Write32(0, 0x00001000)
	memory.Write32(4, 0x00000400)

	cpu.Reset()

	// Set D0 = 10, D1 = 10
	cpu.d[0] = 10
	cpu.d[1] = 10

	// CMP.B D1, D0 = 0xB001
	memory.Write16(0x400, 0xB001)

	cpu.Execute(10)

	// D0 should still be 10 (CMP doesn't modify)
	if cpu.d[0] != 10 {
		t.Errorf("Expected D0 = 10, got %d", cpu.d[0])
	}

	// Check Z flag set (equal)
	if cpu.sr&FlagZ == 0 {
		t.Error("Z flag should be set (values are equal)")
	}
}

// TestTSTInstruction tests the TST instruction
func TestTSTInstruction(t *testing.T) {
	cpu := NewCPU(CPU68000)
	memory := &SimpleMemory{}
	cpu.SetMemoryHandler(memory)

	memory.Write32(0, 0x00001000)
	memory.Write32(4, 0x00000400)

	cpu.Reset()

	// Set D0 = 0
	cpu.d[0] = 0

	// TST.L D0 = 0x4A80
	memory.Write16(0x400, 0x4A80)

	cpu.Execute(10)

	// Check Z flag set
	if cpu.sr&FlagZ == 0 {
		t.Error("Z flag should be set")
	}
}

// TestBRAInstruction tests the BRA (branch always) instruction
func TestBRAInstruction(t *testing.T) {
	cpu := NewCPU(CPU68000)
	memory := &SimpleMemory{}
	cpu.SetMemoryHandler(memory)

	memory.Write32(0, 0x00001000)
	memory.Write32(4, 0x00000400)

	cpu.Reset()

	// BRA with displacement of +4 = 0x6004
	memory.Write16(0x400, 0x6004)
	// NOP at 0x402
	memory.Write16(0x402, 0x4E71)
	// Target at 0x406
	memory.Write16(0x406, 0x4E71)

	cpu.Execute(20)

	// PC should be at 0x408 (skipped the NOP at 0x402, executed NOP at 0x406)
	if cpu.pc != 0x408 {
		t.Errorf("Expected PC = 0x408, got 0x%08X", cpu.pc)
	}
}

// TestBccInstruction tests conditional branch
func TestBccInstruction(t *testing.T) {
	cpu := NewCPU(CPU68000)
	memory := &SimpleMemory{}
	cpu.SetMemoryHandler(memory)

	memory.Write32(0, 0x00001000)
	memory.Write32(4, 0x00000400)

	cpu.Reset()

	// Set Z flag (equal condition)
	cpu.sr |= FlagZ

	// BEQ with displacement of +4 = 0x6704
	memory.Write16(0x400, 0x6704)

	cpu.Execute(20)

	// Branch should be taken
	if cpu.pc != 0x406 {
		t.Errorf("Expected PC = 0x406, got 0x%08X", cpu.pc)
	}
}

// TestSWAPInstruction tests the SWAP instruction
func TestSWAPInstruction(t *testing.T) {
	cpu := NewCPU(CPU68000)
	memory := &SimpleMemory{}
	cpu.SetMemoryHandler(memory)

	memory.Write32(0, 0x00001000)
	memory.Write32(4, 0x00000400)

	cpu.Reset()

	// Set D0 = 0x12345678
	cpu.d[0] = 0x12345678

	// SWAP D0 = 0x4840
	memory.Write16(0x400, 0x4840)

	cpu.Execute(10)

	// Check D0 = 0x56781234
	if cpu.d[0] != 0x56781234 {
		t.Errorf("Expected D0 = 0x56781234, got 0x%08X", cpu.d[0])
	}
}

// TestEXGInstruction tests the EXG instruction
func TestEXGInstruction(t *testing.T) {
	cpu := NewCPU(CPU68000)
	memory := &SimpleMemory{}
	cpu.SetMemoryHandler(memory)

	memory.Write32(0, 0x00001000)
	memory.Write32(4, 0x00000400)

	cpu.Reset()

	// Set D0 = 0x11111111, D1 = 0x22222222
	cpu.d[0] = 0x11111111
	cpu.d[1] = 0x22222222

	// EXG D0, D1 = 0xC141 (D0 << 9 | mode=0x08 << 3 | D1)
	memory.Write16(0x400, 0xC141)

	cpu.Execute(10)

	// Check swapped
	if cpu.d[0] != 0x22222222 {
		t.Errorf("Expected D0 = 0x22222222, got 0x%08X", cpu.d[0])
	}
	if cpu.d[1] != 0x11111111 {
		t.Errorf("Expected D1 = 0x11111111, got 0x%08X", cpu.d[1])
	}
}

// TestEXTInstruction tests the EXT instruction
func TestEXTInstruction(t *testing.T) {
	cpu := NewCPU(CPU68000)
	memory := &SimpleMemory{}
	cpu.SetMemoryHandler(memory)

	memory.Write32(0, 0x00001000)
	memory.Write32(4, 0x00000400)

	cpu.Reset()

	// Set D0 = 0x000000FF (should extend to 0x0000FFFF)
	cpu.d[0] = 0x000000FF

	// EXT.W D0 = 0x4880 (byte to word)
	memory.Write16(0x400, 0x4880)

	cpu.Execute(10)

	// Check D0 = 0x0000FFFF
	if (cpu.d[0] & 0xFFFF) != 0xFFFF {
		t.Errorf("Expected D0 = 0xFFFF, got 0x%04X", cpu.d[0]&0xFFFF)
	}
}

// TestLEAInstruction tests the LEA instruction
func TestLEAInstruction(t *testing.T) {
	cpu := NewCPU(CPU68000)
	memory := &SimpleMemory{}
	cpu.SetMemoryHandler(memory)

	memory.Write32(0, 0x00001000)
	memory.Write32(4, 0x00000400)

	cpu.Reset()

	// Set A0 = 0x2000
	cpu.a[0] = 0x2000

	// LEA (8,A0), A1 = 0x43E8 0x0008 (A1 << 9 | mode=5 << 3 | A0, disp=8)
	memory.Write16(0x400, 0x43E8)
	memory.Write16(0x402, 0x0008)

	cpu.Execute(10)

	// Check A1 = 0x2008
	if cpu.a[1] != 0x2008 {
		t.Errorf("Expected A1 = 0x2008, got 0x%08X", cpu.a[1])
	}
}

// TestRTSInstruction tests the RTS instruction
func TestRTSInstruction(t *testing.T) {
	cpu := NewCPU(CPU68000)
	memory := &SimpleMemory{}
	cpu.SetMemoryHandler(memory)

	memory.Write32(0, 0x00001000)
	memory.Write32(4, 0x00000400)

	cpu.Reset()

	// Push return address 0x1000 onto stack
	cpu.pushLong(0x1000)

	// RTS = 0x4E75
	memory.Write16(0x400, 0x4E75)

	cpu.Execute(20)

	// Check PC = 0x1000
	if cpu.pc != 0x1000 {
		t.Errorf("Expected PC = 0x1000, got 0x%08X", cpu.pc)
	}
}
