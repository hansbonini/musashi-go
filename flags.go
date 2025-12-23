package musashi

// Condition code flags (in SR lower byte)
const (
	FlagC = 0x01 // Carry
	FlagV = 0x02 // Overflow
	FlagZ = 0x04 // Zero
	FlagN = 0x08 // Negative
	FlagX = 0x10 // Extend
)

// Condition codes for conditional instructions
const (
	CondT  = 0  // True
	CondF  = 1  // False
	CondHI = 2  // High
	CondLS = 3  // Low or Same
	CondCC = 4  // Carry Clear
	CondCS = 5  // Carry Set
	CondNE = 6  // Not Equal
	CondEQ = 7  // Equal
	CondVC = 8  // Overflow Clear
	CondVS = 9  // Overflow Set
	CondPL = 10 // Plus
	CondMI = 11 // Minus
	CondGE = 12 // Greater or Equal
	CondLT = 13 // Less Than
	CondGT = 14 // Greater Than
	CondLE = 15 // Less or Equal
)

// setFlagsLogical sets condition codes for logical operations
func (cpu *CPU) setFlagsLogical(result uint32, size int) {
	// Clear V and C
	cpu.sr &^= (FlagV | FlagC)

	// Set N and Z based on result
	switch size {
	case 8:
		if result&0x80 != 0 {
			cpu.sr |= FlagN
		} else {
			cpu.sr &^= FlagN
		}
		if result&0xFF == 0 {
			cpu.sr |= FlagZ
		} else {
			cpu.sr &^= FlagZ
		}
	case 16:
		if result&0x8000 != 0 {
			cpu.sr |= FlagN
		} else {
			cpu.sr &^= FlagN
		}
		if result&0xFFFF == 0 {
			cpu.sr |= FlagZ
		} else {
			cpu.sr &^= FlagZ
		}
	case 32:
		if result&0x80000000 != 0 {
			cpu.sr |= FlagN
		} else {
			cpu.sr &^= FlagN
		}
		if result == 0 {
			cpu.sr |= FlagZ
		} else {
			cpu.sr &^= FlagZ
		}
	}
}

// setFlagsAdd sets condition codes for addition
func (cpu *CPU) setFlagsAdd(dest, src, result uint32, size int) {
	var sm, dm, rm bool

	switch size {
	case 8:
		sm = (src & 0x80) != 0
		dm = (dest & 0x80) != 0
		rm = (result & 0x80) != 0
		// Carry
		if result&0x100 != 0 {
			cpu.sr |= (FlagC | FlagX)
		} else {
			cpu.sr &^= (FlagC | FlagX)
		}
	case 16:
		sm = (src & 0x8000) != 0
		dm = (dest & 0x8000) != 0
		rm = (result & 0x8000) != 0
		// Carry
		if result&0x10000 != 0 {
			cpu.sr |= (FlagC | FlagX)
		} else {
			cpu.sr &^= (FlagC | FlagX)
		}
	case 32:
		sm = (src & 0x80000000) != 0
		dm = (dest & 0x80000000) != 0
		rm = (result & 0x80000000) != 0
		// Carry (check for overflow in 64-bit space)
		src64 := uint64(src)
		dest64 := uint64(dest)
		result64 := src64 + dest64
		if result64 > 0xFFFFFFFF {
			cpu.sr |= (FlagC | FlagX)
		} else {
			cpu.sr &^= (FlagC | FlagX)
		}
	}

	// Overflow: (Sm & Dm & !Rm) | (!Sm & !Dm & Rm)
	if (sm && dm && !rm) || (!sm && !dm && rm) {
		cpu.sr |= FlagV
	} else {
		cpu.sr &^= FlagV
	}

	// Set N and Z
	cpu.setFlagsLogical(result, size)
}

// setFlagsSub sets condition codes for subtraction
func (cpu *CPU) setFlagsSub(dest, src, result uint32, size int) {
	var sm, dm, rm bool

	switch size {
	case 8:
		sm = (src & 0x80) != 0
		dm = (dest & 0x80) != 0
		rm = (result & 0x80) != 0
		// Carry (borrow)
		if result&0x100 != 0 {
			cpu.sr |= (FlagC | FlagX)
		} else {
			cpu.sr &^= (FlagC | FlagX)
		}
	case 16:
		sm = (src & 0x8000) != 0
		dm = (dest & 0x8000) != 0
		rm = (result & 0x8000) != 0
		// Carry (borrow)
		if result&0x10000 != 0 {
			cpu.sr |= (FlagC | FlagX)
		} else {
			cpu.sr &^= (FlagC | FlagX)
		}
	case 32:
		sm = (src & 0x80000000) != 0
		dm = (dest & 0x80000000) != 0
		rm = (result & 0x80000000) != 0
		// Carry (borrow)
		if src > dest {
			cpu.sr |= (FlagC | FlagX)
		} else {
			cpu.sr &^= (FlagC | FlagX)
		}
	}

	// Overflow: (!Sm & Dm & !Rm) | (Sm & !Dm & Rm)
	if (!sm && dm && !rm) || (sm && !dm && rm) {
		cpu.sr |= FlagV
	} else {
		cpu.sr &^= FlagV
	}

	// Set N and Z
	cpu.setFlagsLogical(result, size)
}

// testCondition tests a condition code
func (cpu *CPU) testCondition(cond int) bool {
	c := (cpu.sr & FlagC) != 0
	v := (cpu.sr & FlagV) != 0
	z := (cpu.sr & FlagZ) != 0
	n := (cpu.sr & FlagN) != 0

	switch cond {
	case CondT:
		return true
	case CondF:
		return false
	case CondHI:
		return !c && !z
	case CondLS:
		return c || z
	case CondCC:
		return !c
	case CondCS:
		return c
	case CondNE:
		return !z
	case CondEQ:
		return z
	case CondVC:
		return !v
	case CondVS:
		return v
	case CondPL:
		return !n
	case CondMI:
		return n
	case CondGE:
		return (n && v) || (!n && !v)
	case CondLT:
		return (n && !v) || (!n && v)
	case CondGT:
		return (n && v && !z) || (!n && !v && !z)
	case CondLE:
		return z || (n && !v) || (!n && v)
	default:
		return false
	}
}

// signExtend8 sign extends an 8-bit value to 32-bit
func signExtend8(value uint32) uint32 {
	if value&0x80 != 0 {
		return value | 0xFFFFFF00
	}
	return value & 0xFF
}

// signExtend16 sign extends a 16-bit value to 32-bit
func signExtend16(value uint32) uint32 {
	if value&0x8000 != 0 {
		return value | 0xFFFF0000
	}
	return value & 0xFFFF
}

// maskValue masks a value to the specified size
func maskValue(value uint32, size int) uint32 {
	switch size {
	case 8:
		return value & 0xFF
	case 16:
		return value & 0xFFFF
	case 32:
		return value
	default:
		return value
	}
}
