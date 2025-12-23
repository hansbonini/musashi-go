package musashi

// Addressing mode calculation functions

// Effective address calculation
type effectiveAddress struct {
	address uint32
	value   uint32
	mode    int
}

// Addressing mode types
const (
	ModeDataDirect   = 0  // Dn
	ModeAddrDirect   = 1  // An
	ModeAddrIndirect = 2  // (An)
	ModeAddrPostInc  = 3  // (An)+
	ModeAddrPreDec   = 4  // -(An)
	ModeAddrDisplace = 5  // (d16,An)
	ModeAddrIndex    = 6  // (d8,An,Xn)
	ModeAbsShort     = 7  // (xxx).W
	ModeAbsLong      = 8  // (xxx).L
	ModePCDisplace   = 9  // (d16,PC)
	ModePCIndex      = 10 // (d8,PC,Xn)
	ModeImmediate    = 11 // #<data>
)

// getEAMode extracts addressing mode from opcode
func getEAMode(opcode uint16) int {
	return int((opcode >> 3) & 0x07)
}

// getEAReg extracts register number from opcode
func getEAReg(opcode uint16) int {
	return int(opcode & 0x07)
}

// readEA reads a value using the specified effective address
func (cpu *CPU) readEA(mode, reg, size int) uint32 {
	switch mode {
	case 0: // Dn - Data register direct
		return maskValue(cpu.d[reg], size)

	case 1: // An - Address register direct
		return cpu.a[reg]

	case 2: // (An) - Address register indirect
		return cpu.readMem(cpu.a[reg], size)

	case 3: // (An)+ - Address register indirect with postincrement
		addr := cpu.a[reg]
		val := cpu.readMem(addr, size)
		inc := uint32(size / 8)
		if size == 8 && reg == 7 {
			inc = 2 // SP always increments by 2
		}
		cpu.a[reg] += inc
		return val

	case 4: // -(An) - Address register indirect with predecrement
		dec := uint32(size / 8)
		if size == 8 && reg == 7 {
			dec = 2 // SP always decrements by 2
		}
		cpu.a[reg] -= dec
		return cpu.readMem(cpu.a[reg], size)

	case 5: // (d16,An) - Address register indirect with displacement
		disp := signExtend16(uint32(cpu.readImmediate16()))
		addr := cpu.a[reg] + disp
		return cpu.readMem(addr, size)

	case 6: // (d8,An,Xn) - Address register indirect with index
		ext := uint32(cpu.readImmediate16())
		disp := signExtend8(ext & 0xFF)
		xn := int((ext >> 12) & 0x0F)
		var index uint32
		if ext&0x8000 != 0 { // Address register
			index = cpu.a[xn&7]
		} else { // Data register
			index = cpu.d[xn&7]
		}
		if ext&0x800 == 0 { // Word index
			index = signExtend16(index)
		}
		addr := cpu.a[reg] + disp + index
		return cpu.readMem(addr, size)

	case 7: // Special modes based on register
		switch reg {
		case 0: // (xxx).W - Absolute short
			addr := signExtend16(uint32(cpu.readImmediate16()))
			return cpu.readMem(addr, size)

		case 1: // (xxx).L - Absolute long
			addr := cpu.readImmediate32()
			return cpu.readMem(addr, size)

		case 2: // (d16,PC) - PC with displacement
			oldPC := cpu.pc
			disp := signExtend16(uint32(cpu.readImmediate16()))
			addr := oldPC + disp
			return cpu.readMem(addr, size)

		case 3: // (d8,PC,Xn) - PC with index
			oldPC := cpu.pc
			ext := uint32(cpu.readImmediate16())
			disp := signExtend8(ext & 0xFF)
			xn := int((ext >> 12) & 0x0F)
			var index uint32
			if ext&0x8000 != 0 { // Address register
				index = cpu.a[xn&7]
			} else { // Data register
				index = cpu.d[xn&7]
			}
			if ext&0x800 == 0 { // Word index
				index = signExtend16(index)
			}
			addr := oldPC + disp + index
			return cpu.readMem(addr, size)

		case 4: // #<data> - Immediate
			switch size {
			case 8:
				return uint32(cpu.readImmediate16() & 0xFF)
			case 16:
				return uint32(cpu.readImmediate16())
			case 32:
				return cpu.readImmediate32()
			}
		}
	}

	return 0
}

// writeEA writes a value using the specified effective address
func (cpu *CPU) writeEA(mode, reg, size int, value uint32) {
	value = maskValue(value, size)

	switch mode {
	case 0: // Dn - Data register direct
		switch size {
		case 8:
			cpu.d[reg] = (cpu.d[reg] & 0xFFFFFF00) | (value & 0xFF)
		case 16:
			cpu.d[reg] = (cpu.d[reg] & 0xFFFF0000) | (value & 0xFFFF)
		case 32:
			cpu.d[reg] = value
		}

	case 1: // An - Address register direct
		cpu.a[reg] = value

	case 2: // (An) - Address register indirect
		cpu.writeMem(cpu.a[reg], value, size)

	case 3: // (An)+ - Address register indirect with postincrement
		addr := cpu.a[reg]
		cpu.writeMem(addr, value, size)
		inc := uint32(size / 8)
		if size == 8 && reg == 7 {
			inc = 2
		}
		cpu.a[reg] += inc

	case 4: // -(An) - Address register indirect with predecrement
		dec := uint32(size / 8)
		if size == 8 && reg == 7 {
			dec = 2
		}
		cpu.a[reg] -= dec
		cpu.writeMem(cpu.a[reg], value, size)

	case 5: // (d16,An) - Address register indirect with displacement
		disp := signExtend16(uint32(cpu.readImmediate16()))
		addr := cpu.a[reg] + disp
		cpu.writeMem(addr, value, size)

	case 6: // (d8,An,Xn) - Address register indirect with index
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
		addr := cpu.a[reg] + disp + index
		cpu.writeMem(addr, value, size)

	case 7: // Special modes
		switch reg {
		case 0: // (xxx).W - Absolute short
			addr := signExtend16(uint32(cpu.readImmediate16()))
			cpu.writeMem(addr, value, size)

		case 1: // (xxx).L - Absolute long
			addr := cpu.readImmediate32()
			cpu.writeMem(addr, value, size)
		}
	}
}

// readMem reads from memory with the specified size
func (cpu *CPU) readMem(address uint32, size int) uint32 {
	if cpu.memory == nil {
		return 0
	}

	switch size {
	case 8:
		return uint32(cpu.memory.Read8(address))
	case 16:
		return uint32(cpu.memory.Read16(address))
	case 32:
		return cpu.memory.Read32(address)
	default:
		return 0
	}
}

// writeMem writes to memory with the specified size
func (cpu *CPU) writeMem(address, value uint32, size int) {
	if cpu.memory == nil {
		return
	}

	switch size {
	case 8:
		cpu.memory.Write8(address, uint8(value))
	case 16:
		cpu.memory.Write16(address, uint16(value))
	case 32:
		cpu.memory.Write32(address, value)
	}
}

// readImmediate16 reads a 16-bit immediate value from the instruction stream
func (cpu *CPU) readImmediate16() uint16 {
	if cpu.memory == nil {
		return 0
	}
	value := cpu.memory.Read16(cpu.pc)
	cpu.pc += 2
	return value
}

// readImmediate32 reads a 32-bit immediate value from the instruction stream
func (cpu *CPU) readImmediate32() uint32 {
	if cpu.memory == nil {
		return 0
	}
	value := cpu.memory.Read32(cpu.pc)
	cpu.pc += 4
	return value
}

// getSize extracts size from opcode (bits 6-7)
// Returns 8, 16, or 32
func getSize(opcode uint16, shift int) int {
	size := (opcode >> shift) & 0x03
	switch size {
	case 0:
		return 8
	case 1:
		return 16
	case 2:
		return 32
	default:
		return 8
	}
}

// getSizeBits extracts size bits from opcode
func getSizeBits(opcode uint16, shift int) int {
	return int((opcode >> shift) & 0x03)
}
