package musashi

import (
	"strings"
	"testing"
)

func TestDisassembler(t *testing.T) {
	cpu := NewCPU(CPU68000)
	memory := &SimpleMemory{}
	cpu.SetMemoryHandler(memory)

	tests := []struct {
		name    string
		address uint32
		opcode  uint16
		want    string
	}{
		{"NOP", 0x1000, 0x4E71, "NOP"},
		{"RESET", 0x1000, 0x4E70, "RESET"},
		{"RTS", 0x1000, 0x4E75, "RTS"},
		{"RTE", 0x1000, 0x4E73, "RTE"},
		{"RTR", 0x1000, 0x4E77, "RTR"},
		{"TRAPV", 0x1000, 0x4E76, "TRAPV"},
		{"MOVEQ", 0x1000, 0x7042, "MOVEQ"},
		{"SWAP", 0x1000, 0x4840, "SWAP"},
		{"EXT.W", 0x1000, 0x4880, "EXT.W"},
		{"EXT.L", 0x1000, 0x48C0, "EXT.L"},
		{"EXG", 0x1000, 0xC141, "EXG"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			memory.Write16(tt.address, tt.opcode)
			result, size := cpu.Disassemble(tt.address)

			if !strings.Contains(result, tt.want) {
				t.Errorf("Disassemble() = %v, want %v", result, tt.want)
			}

			if size < 2 {
				t.Errorf("Disassemble() size = %v, want >= 2", size)
			}
		})
	}
}

func TestDisassembleBranches(t *testing.T) {
	cpu := NewCPU(CPU68000)
	memory := &SimpleMemory{}
	cpu.SetMemoryHandler(memory)

	// BRA with 8-bit displacement
	memory.Write16(0x1000, 0x6004) // BRA +4
	result, size := cpu.Disassemble(0x1000)

	if !strings.Contains(result, "BRA") {
		t.Errorf("Expected BRA instruction, got %s", result)
	}
	if size != 2 {
		t.Errorf("Expected size 2, got %d", size)
	}

	// BSR with 16-bit displacement
	memory.Write16(0x2000, 0x6100) // BSR
	memory.Write16(0x2002, 0x0100) // displacement = +256
	result, size = cpu.Disassemble(0x2000)

	if !strings.Contains(result, "BSR") {
		t.Errorf("Expected BSR instruction, got %s", result)
	}
	if size != 4 {
		t.Errorf("Expected size 4, got %d", size)
	}
}

func TestDisassembleConditions(t *testing.T) {
	cpu := NewCPU(CPU68000)
	memory := &SimpleMemory{}
	cpu.SetMemoryHandler(memory)

	// BEQ (branch if equal)
	memory.Write16(0x1000, 0x6704)
	result, _ := cpu.Disassemble(0x1000)

	if !strings.Contains(result, "BEQ") {
		t.Errorf("Expected BEQ instruction, got %s", result)
	}

	// BNE (branch if not equal)
	memory.Write16(0x1000, 0x6604)
	result, _ = cpu.Disassemble(0x1000)

	if !strings.Contains(result, "BNE") {
		t.Errorf("Expected BNE instruction, got %s", result)
	}

	// DBcc
	memory.Write16(0x1000, 0x51C8) // DBRA D0
	memory.Write16(0x1002, 0xFFFE) // displacement
	result, size := cpu.Disassemble(0x1000)

	if !strings.Contains(result, "DBF") || !strings.Contains(result, "D0") {
		t.Errorf("Expected DBRA D0 instruction, got %s", result)
	}
	if size != 4 {
		t.Errorf("Expected size 4, got %d", size)
	}
}
