package musashi

import (
	"fmt"
)

// Disassemble disassembles a single instruction at the specified address.
// Returns the disassembled string and the size of the instruction in bytes.
func (cpu *CPU) Disassemble(address uint32) (string, int) {
	if cpu.memory == nil {
		return "???", 2
	}

	// Read opcode
	opcode := cpu.memory.Read16(address)
	pc := address + 2

	// Decode based on opcode
	switch opcode >> 12 {
	case 0x0:
		return cpu.disasm0(opcode, address, pc)
	case 0x1, 0x2, 0x3:
		return cpu.disasmMOVE(opcode, address, pc)
	case 0x4:
		return cpu.disasm4(opcode, address, pc)
	case 0x5:
		return cpu.disasm5(opcode, address, pc)
	case 0x6:
		return cpu.disasm6(opcode, address, pc)
	case 0x7:
		if opcode&0x0100 == 0 {
			data := int8(opcode & 0xFF)
			return fmt.Sprintf("MOVEQ\t#$%02X,D%d", data, (opcode>>9)&7), 2
		}
	case 0x8:
		return cpu.disasm8(opcode, address, pc)
	case 0x9, 0xD:
		return cpu.disasm9D(opcode, address, pc)
	case 0xB:
		return cpu.disasmB(opcode, address, pc)
	case 0xC:
		return cpu.disasmC(opcode, address, pc)
	case 0xE:
		return cpu.disasmE(opcode, address, pc)
	}

	return fmt.Sprintf("DC.W\t$%04X", opcode), 2
}

func (cpu *CPU) disasm0(opcode uint16, address, pc uint32) (string, int) {
	if opcode&0x0100 == 0 {
		switch (opcode >> 9) & 0x07 {
		case 0:
			if opcode&0x003F == 0x003C {
				imm := cpu.memory.Read16(pc)
				return fmt.Sprintf("ORI\t#$%02X,CCR", imm&0xFF), 4
			}
			return fmt.Sprintf("ORI\t<ea>"), 2
		case 1:
			if opcode&0x003F == 0x003C {
				imm := cpu.memory.Read16(pc)
				return fmt.Sprintf("ANDI\t#$%02X,CCR", imm&0xFF), 4
			}
			return fmt.Sprintf("ANDI\t<ea>"), 2
		case 2:
			return fmt.Sprintf("SUBI\t<ea>"), 2
		case 3:
			return fmt.Sprintf("ADDI\t<ea>"), 2
		case 5:
			if opcode&0x003F == 0x003C {
				imm := cpu.memory.Read16(pc)
				return fmt.Sprintf("EORI\t#$%02X,CCR", imm&0xFF), 4
			}
			return fmt.Sprintf("EORI\t<ea>"), 2
		case 6:
			return fmt.Sprintf("CMPI\t<ea>"), 2
		}
	}
	return fmt.Sprintf("DC.W\t$%04X", opcode), 2
}

func (cpu *CPU) disasmMOVE(opcode uint16, address, pc uint32) (string, int) {
	destMode := (opcode >> 6) & 7
	if destMode == 1 {
		return fmt.Sprintf("MOVEA\t<ea>"), 2
	}
	return fmt.Sprintf("MOVE\t<ea>"), 2
}

func (cpu *CPU) disasm4(opcode uint16, address, pc uint32) (string, int) {
	switch opcode {
	case 0x4E70:
		return "RESET", 2
	case 0x4E71:
		return "NOP", 2
	case 0x4E72:
		imm := cpu.memory.Read16(pc)
		return fmt.Sprintf("STOP\t#$%04X", imm), 4
	case 0x4E73:
		return "RTE", 2
	case 0x4E75:
		return "RTS", 2
	case 0x4E76:
		return "TRAPV", 2
	case 0x4E77:
		return "RTR", 2
	}

	switch (opcode >> 6) & 0x07 {
	case 0:
		switch (opcode >> 9) & 0x07 {
		case 0:
			return fmt.Sprintf("NEGX\t<ea>"), 2
		case 1:
			return fmt.Sprintf("CLR\t<ea>"), 2
		case 2:
			return fmt.Sprintf("NEG\t<ea>"), 2
		case 3:
			return fmt.Sprintf("NOT\t<ea>"), 2
		}
	case 1:
		if opcode&0x0008 != 0 {
			return fmt.Sprintf("SWAP\tD%d", opcode&7), 2
		}
		if opcode&0x0040 == 0 {
			return fmt.Sprintf("EXT.W\tD%d", opcode&7), 2
		}
		return fmt.Sprintf("EXT.L\tD%d", opcode&7), 2
	case 3:
		return fmt.Sprintf("TST\t<ea>"), 2
	case 4, 6:
		if opcode&0x01C0 == 0x01C0 {
			return fmt.Sprintf("LEA\t<ea>"), 2
		}
		return fmt.Sprintf("CHK\t<ea>"), 2
	}

	if opcode&0x01C0 == 0x01C0 {
		switch (opcode >> 9) & 0x07 {
		case 4:
			return fmt.Sprintf("JSR\t<ea>"), 2
		case 6:
			return fmt.Sprintf("JMP\t<ea>"), 2
		}
	}

	return fmt.Sprintf("DC.W\t$%04X", opcode), 2
}

func (cpu *CPU) disasm5(opcode uint16, address, pc uint32) (string, int) {
	if opcode&0x00C0 == 0x00C0 {
		if opcode&0x0038 == 0x0008 {
			disp := int16(cpu.memory.Read16(pc))
			cond := int((opcode >> 8) & 0x0F)
			return fmt.Sprintf("DB%s\tD%d,$%04X", condName(cond), opcode&7, disp), 4
		}
		cond := int((opcode >> 8) & 0x0F)
		return fmt.Sprintf("S%s\t<ea>", condName(cond)), 2
	}

	data := (opcode >> 9) & 7
	if data == 0 {
		data = 8
	}
	if opcode&0x0100 == 0 {
		return fmt.Sprintf("ADDQ\t#%d,<ea>", data), 2
	}
	return fmt.Sprintf("SUBQ\t#%d,<ea>", data), 2
}

func (cpu *CPU) disasm6(opcode uint16, address, pc uint32) (string, int) {
	cond := int((opcode >> 8) & 0x0F)
	disp := int32(int8(opcode & 0xFF))
	size := 2

	if disp == 0 {
		disp = int32(int16(cpu.memory.Read16(pc)))
		size = 4
	}

	target := uint32(int32(address+2) + disp)

	switch cond {
	case 0:
		return fmt.Sprintf("BRA\t$%08X", target), size
	case 1:
		return fmt.Sprintf("BSR\t$%08X", target), size
	default:
		return fmt.Sprintf("B%s\t$%08X", condName(cond), target), size
	}
}

func (cpu *CPU) disasm8(opcode uint16, address, pc uint32) (string, int) {
	if opcode&0x01C0 == 0x01C0 {
		return fmt.Sprintf("DIVU\t<ea>"), 2
	}
	return fmt.Sprintf("OR\t<ea>"), 2
}

func (cpu *CPU) disasm9D(opcode uint16, address, pc uint32) (string, int) {
	isAdd := (opcode & 0xF000) == 0xD000

	if opcode&0x00C0 == 0x00C0 {
		if isAdd {
			return fmt.Sprintf("ADDA\t<ea>"), 2
		}
		return fmt.Sprintf("SUBA\t<ea>"), 2
	}

	if isAdd {
		return fmt.Sprintf("ADD\t<ea>"), 2
	}
	return fmt.Sprintf("SUB\t<ea>"), 2
}

func (cpu *CPU) disasmB(opcode uint16, address, pc uint32) (string, int) {
	if opcode&0x00C0 == 0x00C0 {
		return fmt.Sprintf("CMPA\t<ea>"), 2
	}
	if opcode&0x0100 == 0x0100 {
		return fmt.Sprintf("EOR\t<ea>"), 2
	}
	return fmt.Sprintf("CMP\t<ea>"), 2
}

func (cpu *CPU) disasmC(opcode uint16, address, pc uint32) (string, int) {
	if opcode&0x01C0 == 0x01C0 {
		return fmt.Sprintf("MULU\t<ea>"), 2
	}
	if opcode&0x0130 == 0x0100 {
		rx := (opcode >> 9) & 7
		ry := opcode & 7
		return fmt.Sprintf("EXG\tD%d,D%d", rx, ry), 2
	}
	return fmt.Sprintf("AND\t<ea>"), 2
}

func (cpu *CPU) disasmE(opcode uint16, address, pc uint32) (string, int) {
	return fmt.Sprintf("SHIFT\t<ea>"), 2
}

func condName(cond int) string {
	names := []string{
		"T", "F", "HI", "LS", "CC", "CS", "NE", "EQ",
		"VC", "VS", "PL", "MI", "GE", "LT", "GT", "LE",
	}
	if cond >= 0 && cond < len(names) {
		return names[cond]
	}
	return "??"
}
