package musashi

// instructions.go - Core M68000 instruction implementations

// NOP - No Operation (0x4E71)
func (cpu *CPU) opNOP() {
	cpu.useCycles(4)
}

// MOVE - Move data
func (cpu *CPU) opMOVE(opcode uint16) {
	// MOVE format: 00ss DDDd ddMM Mmmm
	// ss = size, DDD = dest reg, ddd = dest mode, MMM = src mode, mmm = src reg
	size := getSize(opcode, 12)
	if size == 32 {
		size = 32 // 00 = byte, 01 = word, 10 = long
	} else if (opcode>>12)&3 == 1 {
		size = 8
	} else if (opcode>>12)&3 == 3 {
		size = 16
	} else {
		size = 32
	}

	srcMode := int((opcode >> 3) & 7)
	srcReg := int(opcode & 7)
	destMode := int((opcode >> 6) & 7)
	destReg := int((opcode >> 9) & 7)

	// Read source
	value := cpu.readEA(srcMode, srcReg, size)

	// Write destination
	cpu.writeEA(destMode, destReg, size, value)

	// Set flags
	cpu.setFlagsLogical(value, size)

	cpu.useCycles(4)
}

// MOVEA - Move to address register
func (cpu *CPU) opMOVEA(opcode uint16) {
	// MOVEA format: 00ss AAA0 01MM Mmmm
	size := 32
	if (opcode>>12)&3 == 3 {
		size = 16
	}

	srcMode := int((opcode >> 3) & 7)
	srcReg := int(opcode & 7)
	destReg := int((opcode >> 9) & 7)

	value := cpu.readEA(srcMode, srcReg, size)

	// Sign extend if word
	if size == 16 {
		value = signExtend16(value)
	}

	cpu.a[destReg] = value

	cpu.useCycles(4)
}

// ADD - Add
func (cpu *CPU) opADD(opcode uint16) {
	// ADD format: 1101 RRRd ssEE Emmm
	// RRR = data register, d = direction, ss = size, EE = EA mode, mmm = EA reg
	dataReg := int((opcode >> 9) & 7)
	direction := (opcode >> 8) & 1
	size := getSize(opcode, 6)
	eaMode := int((opcode >> 3) & 7)
	eaReg := int(opcode & 7)

	if direction == 0 {
		// EA + Dn -> Dn
		src := cpu.readEA(eaMode, eaReg, size)
		dest := maskValue(cpu.d[dataReg], size)
		result := dest + src
		cpu.setFlagsAdd(dest, src, result, size)
		cpu.writeEA(0, dataReg, size, result)
	} else {
		// Dn + EA -> EA
		src := maskValue(cpu.d[dataReg], size)
		dest := cpu.readEA(eaMode, eaReg, size)
		result := dest + src
		cpu.setFlagsAdd(dest, src, result, size)
		cpu.writeEA(eaMode, eaReg, size, result)
	}

	cpu.useCycles(4)
}

// ADDA - Add to address register
func (cpu *CPU) opADDA(opcode uint16) {
	// ADDA format: 1101 RRRs 11EE Emmm
	addrReg := int((opcode >> 9) & 7)
	size := 16
	if opcode&0x0100 != 0 {
		size = 32
	}
	eaMode := int((opcode >> 3) & 7)
	eaReg := int(opcode & 7)

	src := cpu.readEA(eaMode, eaReg, size)
	if size == 16 {
		src = signExtend16(src)
	}

	cpu.a[addrReg] += src

	cpu.useCycles(8)
}

// ADDI - Add immediate
func (cpu *CPU) opADDI(opcode uint16) {
	// ADDI format: 0000 0110 ssEE Emmm
	size := getSize(opcode, 6)
	eaMode := int((opcode >> 3) & 7)
	eaReg := int(opcode & 7)

	src := cpu.readEA(7, 4, size) // Mode 7, reg 4 = immediate
	dest := cpu.readEA(eaMode, eaReg, size)
	result := dest + src

	cpu.setFlagsAdd(dest, src, result, size)
	cpu.writeEA(eaMode, eaReg, size, result)

	cpu.useCycles(8)
}

// ADDQ - Add quick
func (cpu *CPU) opADDQ(opcode uint16) {
	// ADDQ format: 0101 qqq0 ssEE Emmm
	data := uint32((opcode >> 9) & 7)
	if data == 0 {
		data = 8
	}
	size := getSize(opcode, 6)
	eaMode := int((opcode >> 3) & 7)
	eaReg := int(opcode & 7)

	if eaMode == 1 { // Address register - no flags
		cpu.a[eaReg] += data
	} else {
		dest := cpu.readEA(eaMode, eaReg, size)
		result := dest + data
		cpu.setFlagsAdd(dest, data, result, size)
		cpu.writeEA(eaMode, eaReg, size, result)
	}

	cpu.useCycles(4)
}

// SUB - Subtract
func (cpu *CPU) opSUB(opcode uint16) {
	dataReg := int((opcode >> 9) & 7)
	direction := (opcode >> 8) & 1
	size := getSize(opcode, 6)
	eaMode := int((opcode >> 3) & 7)
	eaReg := int(opcode & 7)

	if direction == 0 {
		// Dn - EA -> Dn
		dest := maskValue(cpu.d[dataReg], size)
		src := cpu.readEA(eaMode, eaReg, size)
		result := dest - src
		cpu.setFlagsSub(dest, src, result, size)
		cpu.writeEA(0, dataReg, size, result)
	} else {
		// EA - Dn -> EA
		dest := cpu.readEA(eaMode, eaReg, size)
		src := maskValue(cpu.d[dataReg], size)
		result := dest - src
		cpu.setFlagsSub(dest, src, result, size)
		cpu.writeEA(eaMode, eaReg, size, result)
	}

	cpu.useCycles(4)
}

// SUBA - Subtract from address register
func (cpu *CPU) opSUBA(opcode uint16) {
	addrReg := int((opcode >> 9) & 7)
	size := 16
	if opcode&0x0100 != 0 {
		size = 32
	}
	eaMode := int((opcode >> 3) & 7)
	eaReg := int(opcode & 7)

	src := cpu.readEA(eaMode, eaReg, size)
	if size == 16 {
		src = signExtend16(src)
	}

	cpu.a[addrReg] -= src

	cpu.useCycles(8)
}

// SUBI - Subtract immediate
func (cpu *CPU) opSUBI(opcode uint16) {
	size := getSize(opcode, 6)
	eaMode := int((opcode >> 3) & 7)
	eaReg := int(opcode & 7)

	src := cpu.readEA(7, 4, size)
	dest := cpu.readEA(eaMode, eaReg, size)
	result := dest - src

	cpu.setFlagsSub(dest, src, result, size)
	cpu.writeEA(eaMode, eaReg, size, result)

	cpu.useCycles(8)
}

// SUBQ - Subtract quick
func (cpu *CPU) opSUBQ(opcode uint16) {
	data := uint32((opcode >> 9) & 7)
	if data == 0 {
		data = 8
	}
	size := getSize(opcode, 6)
	eaMode := int((opcode >> 3) & 7)
	eaReg := int(opcode & 7)

	if eaMode == 1 { // Address register - no flags
		cpu.a[eaReg] -= data
	} else {
		dest := cpu.readEA(eaMode, eaReg, size)
		result := dest - data
		cpu.setFlagsSub(dest, data, result, size)
		cpu.writeEA(eaMode, eaReg, size, result)
	}

	cpu.useCycles(4)
}

// AND - Logical AND
func (cpu *CPU) opAND(opcode uint16) {
	dataReg := int((opcode >> 9) & 7)
	direction := (opcode >> 8) & 1
	size := getSize(opcode, 6)
	eaMode := int((opcode >> 3) & 7)
	eaReg := int(opcode & 7)

	if direction == 0 {
		// EA & Dn -> Dn
		src := cpu.readEA(eaMode, eaReg, size)
		dest := maskValue(cpu.d[dataReg], size)
		result := dest & src
		cpu.setFlagsLogical(result, size)
		cpu.writeEA(0, dataReg, size, result)
	} else {
		// Dn & EA -> EA
		src := maskValue(cpu.d[dataReg], size)
		dest := cpu.readEA(eaMode, eaReg, size)
		result := dest & src
		cpu.setFlagsLogical(result, size)
		cpu.writeEA(eaMode, eaReg, size, result)
	}

	cpu.useCycles(4)
}

// ANDI - AND immediate
func (cpu *CPU) opANDI(opcode uint16) {
	size := getSize(opcode, 6)
	eaMode := int((opcode >> 3) & 7)
	eaReg := int(opcode & 7)

	src := cpu.readEA(7, 4, size)
	dest := cpu.readEA(eaMode, eaReg, size)
	result := dest & src

	cpu.setFlagsLogical(result, size)
	cpu.writeEA(eaMode, eaReg, size, result)

	cpu.useCycles(8)
}

// OR - Logical OR
func (cpu *CPU) opOR(opcode uint16) {
	dataReg := int((opcode >> 9) & 7)
	direction := (opcode >> 8) & 1
	size := getSize(opcode, 6)
	eaMode := int((opcode >> 3) & 7)
	eaReg := int(opcode & 7)

	if direction == 0 {
		// EA | Dn -> Dn
		src := cpu.readEA(eaMode, eaReg, size)
		dest := maskValue(cpu.d[dataReg], size)
		result := dest | src
		cpu.setFlagsLogical(result, size)
		cpu.writeEA(0, dataReg, size, result)
	} else {
		// Dn | EA -> EA
		src := maskValue(cpu.d[dataReg], size)
		dest := cpu.readEA(eaMode, eaReg, size)
		result := dest | src
		cpu.setFlagsLogical(result, size)
		cpu.writeEA(eaMode, eaReg, size, result)
	}

	cpu.useCycles(4)
}

// ORI - OR immediate
func (cpu *CPU) opORI(opcode uint16) {
	size := getSize(opcode, 6)
	eaMode := int((opcode >> 3) & 7)
	eaReg := int(opcode & 7)

	src := cpu.readEA(7, 4, size)
	dest := cpu.readEA(eaMode, eaReg, size)
	result := dest | src

	cpu.setFlagsLogical(result, size)
	cpu.writeEA(eaMode, eaReg, size, result)

	cpu.useCycles(8)
}

// EOR - Logical exclusive OR
func (cpu *CPU) opEOR(opcode uint16) {
	dataReg := int((opcode >> 9) & 7)
	size := getSize(opcode, 6)
	eaMode := int((opcode >> 3) & 7)
	eaReg := int(opcode & 7)

	src := maskValue(cpu.d[dataReg], size)
	dest := cpu.readEA(eaMode, eaReg, size)
	result := dest ^ src

	cpu.setFlagsLogical(result, size)
	cpu.writeEA(eaMode, eaReg, size, result)

	cpu.useCycles(4)
}

// EORI - EOR immediate
func (cpu *CPU) opEORI(opcode uint16) {
	size := getSize(opcode, 6)
	eaMode := int((opcode >> 3) & 7)
	eaReg := int(opcode & 7)

	src := cpu.readEA(7, 4, size)
	dest := cpu.readEA(eaMode, eaReg, size)
	result := dest ^ src

	cpu.setFlagsLogical(result, size)
	cpu.writeEA(eaMode, eaReg, size, result)

	cpu.useCycles(8)
}

// NOT - Logical complement
func (cpu *CPU) opNOT(opcode uint16) {
	size := getSize(opcode, 6)
	eaMode := int((opcode >> 3) & 7)
	eaReg := int(opcode & 7)

	dest := cpu.readEA(eaMode, eaReg, size)
	result := ^dest

	cpu.setFlagsLogical(result, size)
	cpu.writeEA(eaMode, eaReg, size, result)

	cpu.useCycles(4)
}

// CLR - Clear operand
func (cpu *CPU) opCLR(opcode uint16) {
	size := getSize(opcode, 6)
	eaMode := int((opcode >> 3) & 7)
	eaReg := int(opcode & 7)

	cpu.writeEA(eaMode, eaReg, size, 0)

	// Set flags: N=0, Z=1, V=0, C=0
	cpu.sr &^= (FlagN | FlagV | FlagC)
	cpu.sr |= FlagZ

	cpu.useCycles(4)
}

// NEG - Negate
func (cpu *CPU) opNEG(opcode uint16) {
	size := getSize(opcode, 6)
	eaMode := int((opcode >> 3) & 7)
	eaReg := int(opcode & 7)

	dest := cpu.readEA(eaMode, eaReg, size)
	result := uint32(0) - dest

	cpu.setFlagsSub(0, dest, result, size)
	cpu.writeEA(eaMode, eaReg, size, result)

	cpu.useCycles(4)
}

// CMP - Compare
func (cpu *CPU) opCMP(opcode uint16) {
	dataReg := int((opcode >> 9) & 7)
	size := getSize(opcode, 6)
	eaMode := int((opcode >> 3) & 7)
	eaReg := int(opcode & 7)

	dest := maskValue(cpu.d[dataReg], size)
	src := cpu.readEA(eaMode, eaReg, size)
	result := dest - src

	cpu.setFlagsSub(dest, src, result, size)

	cpu.useCycles(4)
}

// CMPA - Compare address
func (cpu *CPU) opCMPA(opcode uint16) {
	addrReg := int((opcode >> 9) & 7)
	size := 16
	if opcode&0x0100 != 0 {
		size = 32
	}
	eaMode := int((opcode >> 3) & 7)
	eaReg := int(opcode & 7)

	dest := cpu.a[addrReg]
	src := cpu.readEA(eaMode, eaReg, size)
	if size == 16 {
		src = signExtend16(src)
	}
	result := dest - src

	cpu.setFlagsSub(dest, src, result, 32)

	cpu.useCycles(6)
}

// CMPI - Compare immediate
func (cpu *CPU) opCMPI(opcode uint16) {
	size := getSize(opcode, 6)
	eaMode := int((opcode >> 3) & 7)
	eaReg := int(opcode & 7)

	src := cpu.readEA(7, 4, size)
	dest := cpu.readEA(eaMode, eaReg, size)
	result := dest - src

	cpu.setFlagsSub(dest, src, result, size)

	cpu.useCycles(8)
}

// TST - Test operand
func (cpu *CPU) opTST(opcode uint16) {
	size := getSize(opcode, 6)
	eaMode := int((opcode >> 3) & 7)
	eaReg := int(opcode & 7)

	value := cpu.readEA(eaMode, eaReg, size)
	cpu.setFlagsLogical(value, size)

	cpu.useCycles(4)
}

// JMP - Jump
func (cpu *CPU) opJMP(opcode uint16) {
	eaMode := int((opcode >> 3) & 7)
	eaReg := int(opcode & 7)

	// Calculate EA (for control addressing modes)
	var addr uint32

	switch eaMode {
	case 2: // (An)
		addr = cpu.a[eaReg]
	case 5: // (d16,An)
		disp := signExtend16(uint32(cpu.readImmediate16()))
		addr = cpu.a[eaReg] + disp
	case 6: // (d8,An,Xn)
		ext := uint32(cpu.readImmediate16())
		disp := signExtend8(ext & 0xFF)
		xn := int((ext >> 12) & 0x0F)
		var index uint32
		if ext&0x8000 != 0 {
			index = cpu.a[xn&7]
		} else {
			index = cpu.d[xn&7]
		}
		if ext&0x800 == 0 {
			index = signExtend16(index)
		}
		addr = cpu.a[eaReg] + disp + index
	case 7:
		switch eaReg {
		case 0: // (xxx).W
			addr = signExtend16(uint32(cpu.readImmediate16()))
		case 1: // (xxx).L
			addr = cpu.readImmediate32()
		case 2: // (d16,PC)
			oldPC := cpu.pc
			disp := signExtend16(uint32(cpu.readImmediate16()))
			addr = oldPC + disp
		case 3: // (d8,PC,Xn)
			oldPC := cpu.pc
			ext := uint32(cpu.readImmediate16())
			disp := signExtend8(ext & 0xFF)
			xn := int((ext >> 12) & 0x0F)
			var index uint32
			if ext&0x8000 != 0 {
				index = cpu.a[xn&7]
			} else {
				index = cpu.d[xn&7]
			}
			if ext&0x800 == 0 {
				index = signExtend16(index)
			}
			addr = oldPC + disp + index
		}
	}

	cpu.pc = addr
	cpu.useCycles(8)
}

// JSR - Jump to subroutine
func (cpu *CPU) opJSR(opcode uint16) {
	eaMode := int((opcode >> 3) & 7)
	eaReg := int(opcode & 7)

	// Calculate EA (same as JMP)
	var addr uint32

	switch eaMode {
	case 2: // (An)
		addr = cpu.a[eaReg]
	case 5: // (d16,An)
		disp := signExtend16(uint32(cpu.readImmediate16()))
		addr = cpu.a[eaReg] + disp
	case 6: // (d8,An,Xn)
		ext := uint32(cpu.readImmediate16())
		disp := signExtend8(ext & 0xFF)
		xn := int((ext >> 12) & 0x0F)
		var index uint32
		if ext&0x8000 != 0 {
			index = cpu.a[xn&7]
		} else {
			index = cpu.d[xn&7]
		}
		if ext&0x800 == 0 {
			index = signExtend16(index)
		}
		addr = cpu.a[eaReg] + disp + index
	case 7:
		switch eaReg {
		case 0: // (xxx).W
			addr = signExtend16(uint32(cpu.readImmediate16()))
		case 1: // (xxx).L
			addr = cpu.readImmediate32()
		case 2: // (d16,PC)
			oldPC := cpu.pc
			disp := signExtend16(uint32(cpu.readImmediate16()))
			addr = oldPC + disp
		case 3: // (d8,PC,Xn)
			oldPC := cpu.pc
			ext := uint32(cpu.readImmediate16())
			disp := signExtend8(ext & 0xFF)
			xn := int((ext >> 12) & 0x0F)
			var index uint32
			if ext&0x8000 != 0 {
				index = cpu.a[xn&7]
			} else {
				index = cpu.d[xn&7]
			}
			if ext&0x800 == 0 {
				index = signExtend16(index)
			}
			addr = oldPC + disp + index
		}
	}

	// Push return address
	cpu.pushLong(cpu.pc)

	// Jump
	cpu.pc = addr
	cpu.useCycles(16)
}

// RTS - Return from subroutine
func (cpu *CPU) opRTS() {
	cpu.pc = cpu.popLong()
	cpu.useCycles(16)
}

// BRA - Branch always
func (cpu *CPU) opBRA(opcode uint16) {
	disp := int32(int8(opcode & 0xFF))
	if disp == 0 {
		disp = int32(int16(cpu.readImmediate16()))
	}

	cpu.pc = uint32(int32(cpu.pc) + disp)
	cpu.useCycles(10)
}

// Bcc - Branch conditionally
func (cpu *CPU) opBcc(opcode uint16) {
	cond := int((opcode >> 8) & 0x0F)
	disp := int32(int8(opcode & 0xFF))
	if disp == 0 {
		disp = int32(int16(cpu.readImmediate16()))
	}

	if cpu.testCondition(cond) {
		cpu.pc = uint32(int32(cpu.pc) + disp)
		cpu.useCycles(10)
	} else {
		cpu.useCycles(8)
	}
}

// DBcc - Test condition, decrement and branch
func (cpu *CPU) opDBcc(opcode uint16) {
	cond := int((opcode >> 8) & 0x0F)
	reg := int(opcode & 7)
	disp := int32(int16(cpu.readImmediate16()))

	if !cpu.testCondition(cond) {
		// Decrement and test
		cpu.d[reg] = (cpu.d[reg] & 0xFFFF0000) | ((cpu.d[reg] - 1) & 0xFFFF)
		if (cpu.d[reg] & 0xFFFF) != 0xFFFF {
			cpu.pc = uint32(int32(cpu.pc) + disp - 2)
			cpu.useCycles(10)
			return
		}
	}

	cpu.useCycles(12)
}

// Scc - Set according to condition
func (cpu *CPU) opScc(opcode uint16) {
	cond := int((opcode >> 8) & 0x0F)
	eaMode := int((opcode >> 3) & 7)
	eaReg := int(opcode & 7)

	var value uint32
	if cpu.testCondition(cond) {
		value = 0xFF
	} else {
		value = 0x00
	}

	cpu.writeEA(eaMode, eaReg, 8, value)
	cpu.useCycles(4)
}

// LEA - Load effective address
func (cpu *CPU) opLEA(opcode uint16) {
	addrReg := int((opcode >> 9) & 7)
	eaMode := int((opcode >> 3) & 7)
	eaReg := int(opcode & 7)

	// Calculate EA
	var addr uint32

	switch eaMode {
	case 2: // (An)
		addr = cpu.a[eaReg]
	case 5: // (d16,An)
		disp := signExtend16(uint32(cpu.readImmediate16()))
		addr = cpu.a[eaReg] + disp
	case 6: // (d8,An,Xn)
		ext := uint32(cpu.readImmediate16())
		disp := signExtend8(ext & 0xFF)
		xn := int((ext >> 12) & 0x0F)
		var index uint32
		if ext&0x8000 != 0 {
			index = cpu.a[xn&7]
		} else {
			index = cpu.d[xn&7]
		}
		if ext&0x800 == 0 {
			index = signExtend16(index)
		}
		addr = cpu.a[eaReg] + disp + index
	case 7:
		switch eaReg {
		case 0: // (xxx).W
			addr = signExtend16(uint32(cpu.readImmediate16()))
		case 1: // (xxx).L
			addr = cpu.readImmediate32()
		case 2: // (d16,PC)
			oldPC := cpu.pc
			disp := signExtend16(uint32(cpu.readImmediate16()))
			addr = oldPC + disp
		case 3: // (d8,PC,Xn)
			oldPC := cpu.pc
			ext := uint32(cpu.readImmediate16())
			disp := signExtend8(ext & 0xFF)
			xn := int((ext >> 12) & 0x0F)
			var index uint32
			if ext&0x8000 != 0 {
				index = cpu.a[xn&7]
			} else {
				index = cpu.d[xn&7]
			}
			if ext&0x800 == 0 {
				index = signExtend16(index)
			}
			addr = oldPC + disp + index
		}
	}

	cpu.a[addrReg] = addr
	cpu.useCycles(4)
}

// PEA - Push effective address
func (cpu *CPU) opPEA(opcode uint16) {
	eaMode := int((opcode >> 3) & 7)
	eaReg := int(opcode & 7)

	// Calculate EA (same as LEA)
	var addr uint32

	switch eaMode {
	case 2: // (An)
		addr = cpu.a[eaReg]
	case 5: // (d16,An)
		disp := signExtend16(uint32(cpu.readImmediate16()))
		addr = cpu.a[eaReg] + disp
	case 6: // (d8,An,Xn)
		ext := uint32(cpu.readImmediate16())
		disp := signExtend8(ext & 0xFF)
		xn := int((ext >> 12) & 0x0F)
		var index uint32
		if ext&0x8000 != 0 {
			index = cpu.a[xn&7]
		} else {
			index = cpu.d[xn&7]
		}
		if ext&0x800 == 0 {
			index = signExtend16(index)
		}
		addr = cpu.a[eaReg] + disp + index
	case 7:
		switch eaReg {
		case 0: // (xxx).W
			addr = signExtend16(uint32(cpu.readImmediate16()))
		case 1: // (xxx).L
			addr = cpu.readImmediate32()
		case 2: // (d16,PC)
			oldPC := cpu.pc
			disp := signExtend16(uint32(cpu.readImmediate16()))
			addr = oldPC + disp
		case 3: // (d8,PC,Xn)
			oldPC := cpu.pc
			ext := uint32(cpu.readImmediate16())
			disp := signExtend8(ext & 0xFF)
			xn := int((ext >> 12) & 0x0F)
			var index uint32
			if ext&0x8000 != 0 {
				index = cpu.a[xn&7]
			} else {
				index = cpu.d[xn&7]
			}
			if ext&0x800 == 0 {
				index = signExtend16(index)
			}
			addr = oldPC + disp + index
		}
	}

	cpu.pushLong(addr)
	cpu.useCycles(12)
}

// EXG - Exchange registers
func (cpu *CPU) opEXG(opcode uint16) {
	rx := int((opcode >> 9) & 7)
	ry := int(opcode & 7)
	mode := (opcode >> 3) & 0x1F

	switch mode {
	case 0x08: // Data registers
		cpu.d[rx], cpu.d[ry] = cpu.d[ry], cpu.d[rx]
	case 0x09: // Address registers
		cpu.a[rx], cpu.a[ry] = cpu.a[ry], cpu.a[rx]
	case 0x11: // Data and address
		cpu.d[rx], cpu.a[ry] = cpu.a[ry], cpu.d[rx]
	}

	cpu.useCycles(6)
}

// SWAP - Swap register halves
func (cpu *CPU) opSWAP(opcode uint16) {
	reg := int(opcode & 7)

	value := cpu.d[reg]
	cpu.d[reg] = ((value & 0xFFFF) << 16) | ((value >> 16) & 0xFFFF)

	cpu.setFlagsLogical(cpu.d[reg], 32)
	cpu.useCycles(4)
}

// EXT - Sign extend
func (cpu *CPU) opEXT(opcode uint16) {
	reg := int(opcode & 7)

	if opcode&0x0040 == 0 {
		// Byte to word
		if cpu.d[reg]&0x80 != 0 {
			cpu.d[reg] = (cpu.d[reg] & 0xFFFF0000) | 0x0000FF00 | (cpu.d[reg] & 0xFF)
		} else {
			cpu.d[reg] = (cpu.d[reg] & 0xFFFF0000) | (cpu.d[reg] & 0xFF)
		}
		cpu.setFlagsLogical(cpu.d[reg], 16)
	} else {
		// Word to long
		if cpu.d[reg]&0x8000 != 0 {
			cpu.d[reg] = 0xFFFF0000 | (cpu.d[reg] & 0xFFFF)
		} else {
			cpu.d[reg] = cpu.d[reg] & 0xFFFF
		}
		cpu.setFlagsLogical(cpu.d[reg], 32)
	}

	cpu.useCycles(4)
}
