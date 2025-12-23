package musashi

// opcodes.go - Opcode dispatch table and decoder

// decodeAndExecute decodes and executes a single instruction
func (cpu *CPU) decodeAndExecute(opcode uint16) {
	// Decode based on top 4 bits and dispatch
	switch opcode >> 12 {
	case 0x0:
		cpu.decode0(opcode)
	case 0x1, 0x2, 0x3:
		cpu.decodeMOVE(opcode)
	case 0x4:
		cpu.decode4(opcode)
	case 0x5:
		cpu.decode5(opcode)
	case 0x6:
		cpu.decode6(opcode)
	case 0x7:
		cpu.decodeMOVEQ(opcode)
	case 0x8:
		cpu.decode8(opcode)
	case 0x9, 0xD:
		cpu.decode9D(opcode)
	case 0xB:
		cpu.decodeB(opcode)
	case 0xC:
		cpu.decodeC(opcode)
	case 0xE:
		cpu.decodeE(opcode)
	default:
		cpu.opIllegal(opcode)
	}
}

// decode0 handles opcodes starting with 0x0
func (cpu *CPU) decode0(opcode uint16) {
	if opcode&0x0100 == 0 {
		// Bit 8 = 0
		switch (opcode >> 6) & 0x03 {
		case 0: // ORI, ANDI, SUBI, ADDI, EORI, CMPI
			switch (opcode >> 9) & 0x07 {
			case 0: // ORI
				if opcode&0x003F == 0x003C { // to SR
					cpu.opORItoCCR(opcode)
				} else {
					cpu.opORI(opcode)
				}
			case 1: // ANDI
				if opcode&0x003F == 0x003C { // to SR
					cpu.opANDItoCCR(opcode)
				} else {
					cpu.opANDI(opcode)
				}
			case 2: // SUBI
				cpu.opSUBI(opcode)
			case 3: // ADDI
				cpu.opADDI(opcode)
			case 5: // EORI
				if opcode&0x003F == 0x003C { // to SR
					cpu.opEORItoCCR(opcode)
				} else {
					cpu.opEORI(opcode)
				}
			case 6: // CMPI
				cpu.opCMPI(opcode)
			default:
				cpu.opIllegal(opcode)
			}
		case 1: // BTST, BCHG, BCLR, BSET (dynamic)
			cpu.opBitDynamic(opcode)
		case 2: // BTST, BCHG, BCLR, BSET (static), MOVEP
			if opcode&0x0138 == 0x0108 {
				cpu.opMOVEP(opcode)
			} else {
				cpu.opBitStatic(opcode)
			}
		case 3: // BTST, BCHG, BCLR, BSET (static), MOVEP
			if opcode&0x0138 == 0x0108 {
				cpu.opMOVEP(opcode)
			} else {
				cpu.opBitStatic(opcode)
			}
		}
	} else {
		// Bit 8 = 1
		cpu.opIllegal(opcode)
	}
}

// decodeMOVE handles MOVE instructions
func (cpu *CPU) decodeMOVE(opcode uint16) {
	// Check if it's MOVEA
	destMode := (opcode >> 6) & 7
	if destMode == 1 {
		cpu.opMOVEA(opcode)
	} else {
		cpu.opMOVE(opcode)
	}
}

// decode4 handles opcodes starting with 0x4
func (cpu *CPU) decode4(opcode uint16) {
	switch opcode {
	case 0x4E70:
		cpu.opRESET()
	case 0x4E71:
		cpu.opNOP()
	case 0x4E72:
		cpu.opSTOP()
	case 0x4E73:
		cpu.opRTE()
	case 0x4E75:
		cpu.opRTS()
	case 0x4E76:
		cpu.opTRAPV()
	case 0x4E77:
		cpu.opRTR()
	default:
		switch (opcode >> 6) & 0x07 {
		case 0: // NEGX, CLR, NEG, NOT
			switch (opcode >> 9) & 0x07 {
			case 0:
				cpu.opNEGX(opcode)
			case 1:
				cpu.opCLR(opcode)
			case 2:
				cpu.opNEG(opcode)
			case 3:
				cpu.opNOT(opcode)
			default:
				cpu.opIllegal(opcode)
			}
		case 1: // NBCD, SWAP, PEA
			switch (opcode >> 3) & 0x3F {
			case 0x00, 0x01: // NBCD
				cpu.opNBCD(opcode)
			case 0x08, 0x09: // SWAP, EXT
				if opcode&0x0008 != 0 {
					cpu.opSWAP(opcode)
				} else {
					cpu.opEXT(opcode)
				}
			default:
				if opcode&0x01C0 == 0x01C0 {
					cpu.opPEA(opcode)
				} else {
					cpu.opIllegal(opcode)
				}
			}
		case 2: // MOVEM, EXT
			if opcode&0x0800 != 0 {
				cpu.opMOVEMtoMem(opcode)
			} else if opcode&0x0B80 == 0x0880 {
				cpu.opEXT(opcode)
			} else {
				cpu.opMOVEMtoReg(opcode)
			}
		case 3: // TST, TAS, ILLEGAL, TRAP, LINK, UNLK, MOVE USP
			if opcode&0x0FC0 == 0x0AC0 {
				cpu.opTAS(opcode)
			} else if opcode&0x0B80 == 0x0880 {
				cpu.opEXT(opcode)
			} else if opcode&0x0F00 == 0x0E00 {
				// TRAP, LINK, UNLK, MOVE USP
				if opcode&0x0080 == 0 {
					cpu.opLINK(opcode)
				} else if opcode&0x0008 == 0 {
					cpu.opUNLK(opcode)
				} else {
					cpu.opMOVEUSP(opcode)
				}
			} else {
				cpu.opTST(opcode)
			}
		case 4, 6: // CHK, LEA
			if opcode&0x01C0 == 0x01C0 {
				cpu.opLEA(opcode)
			} else {
				cpu.opCHK(opcode)
			}
		case 5, 7: // ADDQ, SUBQ, Scc, DBcc
			if opcode&0x00C0 == 0x00C0 {
				if opcode&0x0038 == 0x0008 {
					cpu.opDBcc(opcode)
				} else {
					cpu.opScc(opcode)
				}
			} else {
				if opcode&0x0100 == 0 {
					cpu.opADDQ(opcode)
				} else {
					cpu.opSUBQ(opcode)
				}
			}
		default:
			if opcode&0x0FC0 == 0x04C0 {
				cpu.opMOVEMtoReg(opcode)
			} else if opcode&0x0B80 == 0x0080 {
				cpu.opMOVEMtoMem(opcode)
			} else if opcode&0x01C0 == 0x01C0 {
				switch (opcode >> 9) & 0x07 {
				case 4:
					cpu.opJSR(opcode)
				case 6:
					cpu.opJMP(opcode)
				default:
					cpu.opIllegal(opcode)
				}
			} else {
				cpu.opIllegal(opcode)
			}
		}
	}
}

// decode5 handles ADDQ, SUBQ, Scc, DBcc
func (cpu *CPU) decode5(opcode uint16) {
	if opcode&0x00C0 == 0x00C0 {
		// Scc or DBcc
		if opcode&0x0038 == 0x0008 {
			cpu.opDBcc(opcode)
		} else {
			cpu.opScc(opcode)
		}
	} else {
		// ADDQ or SUBQ
		if opcode&0x0100 == 0 {
			cpu.opADDQ(opcode)
		} else {
			cpu.opSUBQ(opcode)
		}
	}
}

// decode6 handles Bcc, BSR, BRA
func (cpu *CPU) decode6(opcode uint16) {
	cond := (opcode >> 8) & 0x0F
	switch cond {
	case 0: // BRA
		cpu.opBRA(opcode)
	case 1: // BSR
		cpu.opBSR(opcode)
	default: // Bcc
		cpu.opBcc(opcode)
	}
}

// decodeMOVEQ handles MOVEQ
func (cpu *CPU) decodeMOVEQ(opcode uint16) {
	if opcode&0x0100 == 0 {
		cpu.opMOVEQ(opcode)
	} else {
		cpu.opIllegal(opcode)
	}
}

// decode8 handles OR, DIVU, SBCD
func (cpu *CPU) decode8(opcode uint16) {
	if opcode&0x01C0 == 0x0100 {
		cpu.opSBCD(opcode)
	} else if opcode&0x01F0 == 0x0100 {
		cpu.opSBCD(opcode)
	} else if opcode&0x01C0 == 0x01C0 {
		cpu.opDIVU(opcode)
	} else {
		cpu.opOR(opcode)
	}
}

// decode9D handles SUB, SUBA, SUBX, ADD, ADDA, ADDX
func (cpu *CPU) decode9D(opcode uint16) {
	isAdd := (opcode & 0xF000) == 0xD000

	if opcode&0x00C0 == 0x00C0 {
		// ADDA or SUBA
		if isAdd {
			cpu.opADDA(opcode)
		} else {
			cpu.opSUBA(opcode)
		}
	} else if opcode&0x0130 == 0x0100 {
		// ADDX or SUBX
		if isAdd {
			cpu.opADDX(opcode)
		} else {
			cpu.opSUBX(opcode)
		}
	} else {
		// ADD or SUB
		if isAdd {
			cpu.opADD(opcode)
		} else {
			cpu.opSUB(opcode)
		}
	}
}

// decodeB handles CMP, CMPA, CMPM, EOR
func (cpu *CPU) decodeB(opcode uint16) {
	if opcode&0x00C0 == 0x00C0 {
		// CMPA
		cpu.opCMPA(opcode)
	} else if opcode&0x0138 == 0x0108 {
		// CMPM
		cpu.opCMPM(opcode)
	} else if opcode&0x0100 == 0x0100 {
		// EOR
		cpu.opEOR(opcode)
	} else {
		// CMP
		cpu.opCMP(opcode)
	}
}

// decodeC handles AND, MULU, ABCD, EXG
func (cpu *CPU) decodeC(opcode uint16) {
	if opcode&0x01C0 == 0x0100 {
		cpu.opABCD(opcode)
	} else if opcode&0x01F0 == 0x0100 {
		cpu.opABCD(opcode)
	} else if opcode&0x01C0 == 0x01C0 {
		cpu.opMULU(opcode)
	} else if opcode&0x0130 == 0x0100 {
		cpu.opEXG(opcode)
	} else {
		cpu.opAND(opcode)
	}
}

// decodeE handles shift/rotate instructions
func (cpu *CPU) decodeE(opcode uint16) {
	if opcode&0x00C0 == 0x00C0 {
		// Memory shifts
		cpu.opShiftMem(opcode)
	} else {
		// Register shifts
		cpu.opShiftReg(opcode)
	}
}

// Stub implementations for missing instructions
func (cpu *CPU) opIllegal(opcode uint16) {
	// TODO: Generate illegal instruction exception
	cpu.useCycles(4)
}

func (cpu *CPU) opMOVEQ(opcode uint16) {
	reg := int((opcode >> 9) & 7)
	data := signExtend8(uint32(opcode & 0xFF))
	cpu.d[reg] = data
	cpu.setFlagsLogical(data, 32)
	cpu.useCycles(4)
}

func (cpu *CPU) opRESET() {
	if cpu.resetCallback != nil {
		cpu.resetCallback()
	}
	cpu.useCycles(132)
}

func (cpu *CPU) opSTOP() {
	// Read immediate data (new SR)
	newSR := cpu.readImmediate16()
	cpu.sr = newSR
	cpu.stopped = true
	cpu.useCycles(4)
}

func (cpu *CPU) opRTE() {
	// Return from exception
	cpu.sr = cpu.popWord()
	cpu.pc = cpu.popLong()
	cpu.useCycles(20)
}

func (cpu *CPU) opTRAPV() {
	if cpu.sr&FlagV != 0 {
		// TODO: Generate TRAPV exception
	}
	cpu.useCycles(4)
}

func (cpu *CPU) opRTR() {
	// Return and restore condition codes
	ccr := cpu.popWord()
	cpu.sr = (cpu.sr & 0xFF00) | (ccr & 0x00FF)
	cpu.pc = cpu.popLong()
	cpu.useCycles(20)
}

func (cpu *CPU) opNEGX(opcode uint16) {
	// TODO: Implement NEGX
	cpu.useCycles(4)
}

func (cpu *CPU) opNBCD(opcode uint16) {
	// TODO: Implement NBCD
	cpu.useCycles(6)
}

func (cpu *CPU) opMOVEMtoReg(opcode uint16) {
	// TODO: Implement MOVEM to registers
	cpu.useCycles(12)
}

func (cpu *CPU) opMOVEMtoMem(opcode uint16) {
	// TODO: Implement MOVEM to memory
	cpu.useCycles(8)
}

func (cpu *CPU) opTAS(opcode uint16) {
	// TODO: Implement TAS
	cpu.useCycles(4)
}

func (cpu *CPU) opLINK(opcode uint16) {
	reg := int(opcode & 7)
	disp := int32(int16(cpu.readImmediate16()))

	// Push An
	cpu.pushLong(cpu.a[reg])

	// An = SP
	cpu.a[reg] = cpu.a[7]

	// SP = SP + disp
	cpu.a[7] = uint32(int32(cpu.a[7]) + disp)

	cpu.useCycles(16)
}

func (cpu *CPU) opUNLK(opcode uint16) {
	reg := int(opcode & 7)

	// SP = An
	cpu.a[7] = cpu.a[reg]

	// Pop An
	cpu.a[reg] = cpu.popLong()

	cpu.useCycles(12)
}

func (cpu *CPU) opMOVEUSP(opcode uint16) {
	reg := int(opcode & 7)
	if opcode&0x0008 != 0 {
		// USP to An
		cpu.a[reg] = cpu.usp
	} else {
		// An to USP
		cpu.usp = cpu.a[reg]
	}
	cpu.useCycles(4)
}

func (cpu *CPU) opCHK(opcode uint16) {
	// TODO: Implement CHK
	cpu.useCycles(10)
}

func (cpu *CPU) opBSR(opcode uint16) {
	disp := int32(int8(opcode & 0xFF))
	if disp == 0 {
		disp = int32(int16(cpu.readImmediate16()))
	}

	// Push return address
	cpu.pushLong(cpu.pc)

	// Branch
	cpu.pc = uint32(int32(cpu.pc) + disp)
	cpu.useCycles(18)
}

func (cpu *CPU) opDIVU(opcode uint16) {
	// TODO: Implement DIVU
	cpu.useCycles(140)
}

func (cpu *CPU) opSBCD(opcode uint16) {
	// TODO: Implement SBCD
	cpu.useCycles(6)
}

func (cpu *CPU) opADDX(opcode uint16) {
	// TODO: Implement ADDX
	cpu.useCycles(4)
}

func (cpu *CPU) opSUBX(opcode uint16) {
	// TODO: Implement SUBX
	cpu.useCycles(4)
}

func (cpu *CPU) opCMPM(opcode uint16) {
	// TODO: Implement CMPM
	cpu.useCycles(4)
}

func (cpu *CPU) opABCD(opcode uint16) {
	// TODO: Implement ABCD
	cpu.useCycles(6)
}

func (cpu *CPU) opMULU(opcode uint16) {
	// TODO: Implement MULU
	cpu.useCycles(70)
}

func (cpu *CPU) opShiftMem(opcode uint16) {
	// TODO: Implement memory shifts
	cpu.useCycles(8)
}

func (cpu *CPU) opShiftReg(opcode uint16) {
	// TODO: Implement register shifts
	cpu.useCycles(6)
}

func (cpu *CPU) opBitDynamic(opcode uint16) {
	// TODO: Implement dynamic bit operations
	cpu.useCycles(4)
}

func (cpu *CPU) opBitStatic(opcode uint16) {
	// TODO: Implement static bit operations
	cpu.useCycles(8)
}

func (cpu *CPU) opMOVEP(opcode uint16) {
	// TODO: Implement MOVEP
	cpu.useCycles(16)
}

func (cpu *CPU) opORItoCCR(opcode uint16) {
	data := cpu.readImmediate16() & 0xFF
	cpu.sr = (cpu.sr & 0xFF00) | ((cpu.sr | uint16(data)) & 0x00FF)
	cpu.useCycles(20)
}

func (cpu *CPU) opANDItoCCR(opcode uint16) {
	data := cpu.readImmediate16() & 0xFF
	cpu.sr = (cpu.sr & 0xFF00) | ((cpu.sr & uint16(data)) & 0x00FF)
	cpu.useCycles(20)
}

func (cpu *CPU) opEORItoCCR(opcode uint16) {
	data := cpu.readImmediate16() & 0xFF
	cpu.sr = (cpu.sr & 0xFF00) | ((cpu.sr ^ uint16(data)) & 0x00FF)
	cpu.useCycles(20)
}
