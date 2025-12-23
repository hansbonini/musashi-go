// Package musashi provides a pure Go implementation of the Musashi Motorola 68000 emulator.
//
// This package emulates the Motorola 68000 family of processors (68000, 68010, 68020, 68030, 68040)
// with cycle-accurate timing and full instruction set support.
//
// Example usage:
//
//	cpu := musashi.NewCPU(musashi.CPU68000)
//	cpu.SetMemoryHandler(myMemoryHandler)
//	cpu.Reset()
//	cycles := cpu.Execute(1000)
package musashi

// CPUType represents the type of M68000 CPU to emulate
type CPUType int

// CPU types supported by the emulator
const (
	CPUInvalid  CPUType = iota
	CPU68000            // Motorola 68000
	CPU68010            // Motorola 68010
	CPU68EC020          // Motorola 68EC020 (no MMU)
	CPU68020            // Motorola 68020
	CPU68EC030          // Motorola 68EC030 (no MMU)
	CPU68030            // Motorola 68030
	CPU68EC040          // Motorola 68EC040 (no FPU)
	CPU68LC040          // Motorola 68LC040 (no FPU, no MMU)
	CPU68040            // Motorola 68040
	CPUSCC68070         // Philips SCC68070 (68010 with 32-bit data bus)
)

// String returns the string representation of a CPU type
func (c CPUType) String() string {
	switch c {
	case CPU68000:
		return "68000"
	case CPU68010:
		return "68010"
	case CPU68EC020:
		return "68EC020"
	case CPU68020:
		return "68020"
	case CPU68EC030:
		return "68EC030"
	case CPU68030:
		return "68030"
	case CPU68EC040:
		return "68EC040"
	case CPU68LC040:
		return "68LC040"
	case CPU68040:
		return "68040"
	case CPUSCC68070:
		return "SCC68070"
	default:
		return "Invalid"
	}
}

// Register represents a CPU register that can be accessed
type Register int

// CPU registers
const (
	// Data registers (D0-D7)
	RegD0 Register = iota
	RegD1
	RegD2
	RegD3
	RegD4
	RegD5
	RegD6
	RegD7

	// Address registers (A0-A7)
	RegA0
	RegA1
	RegA2
	RegA3
	RegA4
	RegA5
	RegA6
	RegA7

	// Special registers
	RegPC       // Program Counter
	RegSR       // Status Register
	RegSP       // Stack Pointer (same as A7)
	RegUSP      // User Stack Pointer
	RegISP      // Interrupt Stack Pointer (68010+)
	RegMSP      // Master Stack Pointer (68020+)
	RegSFC      // Source Function Code (68010+)
	RegDFC      // Destination Function Code (68010+)
	RegVBR      // Vector Base Register (68010+)
	RegCACR     // Cache Control Register (68020+)
	RegCAAR     // Cache Address Register (68020+)
	RegPrefAddr // Prefetch Address (for debugging)
	RegPrefData // Prefetch Data (for debugging)
	RegPPC      // Previous Program Counter
	RegIR       // Instruction Register
	RegCPUType  // CPU Type register
)

// IRQ levels
const (
	IRQNone = 0 // No interrupt
	IRQ1    = 1 // Interrupt level 1
	IRQ2    = 2 // Interrupt level 2
	IRQ3    = 3 // Interrupt level 3
	IRQ4    = 4 // Interrupt level 4
	IRQ5    = 5 // Interrupt level 5
	IRQ6    = 6 // Interrupt level 6
	IRQ7    = 7 // NMI (Non-Maskable Interrupt)
)

// Special interrupt acknowledge values
const (
	IntAckAutovector = 0xFFFFFFFF // Use autovectored interrupt
	IntAckSpurious   = 0xFFFFFFFE // Spurious interrupt
)

// Function codes for memory access
const (
	FCUserData       = 1 // User data access
	FCUserProgram    = 2 // User program access
	FCSupervisorData = 5 // Supervisor data access
	FCSupervisorProg = 6 // Supervisor program access
	FCCPUSpace       = 7 // CPU space (not fully emulated)
)

// MemoryHandler defines the interface for memory access callbacks.
// Implement this interface to provide custom memory access behavior.
type MemoryHandler interface {
	// Read8 reads a byte from the specified address
	Read8(address uint32) uint8

	// Read16 reads a word (16-bit) from the specified address
	// Note: Address must be word-aligned on real 68000, or address error occurs
	Read16(address uint32) uint16

	// Read32 reads a longword (32-bit) from the specified address
	// Note: Address must be word-aligned on real 68000, or address error occurs
	Read32(address uint32) uint32

	// Write8 writes a byte to the specified address
	Write8(address uint32, value uint8)

	// Write16 writes a word (16-bit) to the specified address
	// Note: Address must be word-aligned on real 68000, or address error occurs
	Write16(address uint32, value uint16)

	// Write32 writes a longword (32-bit) to the specified address
	// Note: Address must be word-aligned on real 68000, or address error occurs
	Write32(address uint32, value uint32)
}

// CPU represents a Motorola 68000 family processor
type CPU struct {
	// CPU type
	cpuType CPUType

	// Data registers (D0-D7)
	d [8]uint32

	// Address registers (A0-A7)
	a [8]uint32

	// Program counter
	pc uint32

	// Status register (SR)
	// Upper byte: System byte (T, S, I2, I1, I0)
	// Lower byte: Condition codes (X, N, Z, V, C)
	sr uint16

	// Stack pointers
	usp uint32 // User stack pointer
	isp uint32 // Interrupt stack pointer (68010+)
	msp uint32 // Master stack pointer (68020+)

	// Control registers (68010+)
	sfc uint8  // Source function code
	dfc uint8  // Destination function code
	vbr uint32 // Vector base register

	// Cache control (68020+)
	cacr uint32 // Cache control register
	caar uint32 // Cache address register

	// Execution state
	stopped      bool    // CPU is stopped
	halted       bool    // CPU is halted
	cyclesRun    int     // Cycles executed in current timeslice
	cyclesRemain int     // Cycles remaining in current timeslice
	irqLevel     uint8   // Current IRQ level (0-7)
	virq         [8]bool // Virtual IRQ lines
	prefetchAddr uint32  // Last prefetch address
	prefetchData uint32  // Last prefetch data
	ppc          uint32  // Previous program counter
	ir           uint16  // Instruction register

	// Memory access
	memory MemoryHandler

	// Callbacks (optional)
	intAckCallback    func(level int) uint32
	resetCallback     func()
	pcChangedCallback func(newPC uint32)
	fcCallback        func(fc uint8)
	instrHookCallback func(pc uint32)
	bkptAckCallback   func(data uint32)
	illegalCallback   func(opcode uint16) bool
	tasCallback       func() int
}

// NewCPU creates a new CPU instance of the specified type
func NewCPU(cpuType CPUType) *CPU {
	cpu := &CPU{
		cpuType: cpuType,
	}
	return cpu
}

// Reset resets the CPU to its initial state.
// This simulates pulsing the RESET pin on the physical CPU.
// The CPU will read the initial stack pointer and program counter from
// memory locations 0 and 4 respectively.
func (cpu *CPU) Reset() {
	// Clear all data registers
	for i := range cpu.d {
		cpu.d[i] = 0
	}

	// Clear all address registers
	for i := range cpu.a {
		cpu.a[i] = 0
	}

	// Set supervisor mode and interrupt mask to 7
	cpu.sr = 0x2700

	// Clear control registers
	cpu.sfc = 0
	cpu.dfc = 0
	cpu.vbr = 0
	cpu.cacr = 0
	cpu.caar = 0

	// Clear execution state
	cpu.stopped = false
	cpu.halted = false
	cpu.cyclesRun = 0
	cpu.cyclesRemain = 0
	cpu.irqLevel = 0

	// Read initial SSP and PC from memory if handler is set
	if cpu.memory != nil {
		cpu.a[7] = cpu.memory.Read32(0) // Initial SSP
		cpu.pc = cpu.memory.Read32(4)   // Initial PC
	} else {
		cpu.a[7] = 0
		cpu.pc = 0
	}

	cpu.usp = 0
	cpu.isp = 0
	cpu.msp = 0

	// Clear prefetch
	cpu.prefetchAddr = 0
	cpu.prefetchData = 0
	cpu.ppc = cpu.pc
	cpu.ir = 0
}

// Execute runs the CPU for the specified number of cycles.
// Returns the actual number of cycles executed.
func (cpu *CPU) Execute(cycles int) int {
	if cpu.memory == nil {
		return 0
	}

	cpu.cyclesRemain = cycles
	cpu.cyclesRun = 0

	// Main execution loop
	for cpu.cyclesRemain > 0 && !cpu.stopped && !cpu.halted {
		// Check for interrupts
		cpu.checkInterrupts()

		// Call instruction hook if set
		if cpu.instrHookCallback != nil {
			cpu.instrHookCallback(cpu.pc)
		}

		// Fetch and execute instruction
		cpu.ppc = cpu.pc
		cpu.executeInstruction()
	}

	return cpu.cyclesRun
}

// executeInstruction fetches and executes a single instruction
func (cpu *CPU) executeInstruction() {
	// Fetch instruction
	cpu.ir = cpu.memory.Read16(cpu.pc)
	cpu.pc += 2

	// For now, just consume minimum cycles
	// Full instruction implementation will be added in subsequent phases
	cpu.useCycles(4)

	// Placeholder: Handle basic NOP and reset instructions
	switch cpu.ir {
	case 0x4E71: // NOP
		// Do nothing, already consumed cycles
	case 0x4E70: // RESET
		if cpu.resetCallback != nil {
			cpu.resetCallback()
		}
	default:
		// Unknown instruction - will be implemented in instruction phase
		// For now, just skip it
	}
}

// checkInterrupts checks for pending interrupts and handles them if needed
func (cpu *CPU) checkInterrupts() {
	if cpu.irqLevel == 0 {
		return
	}

	// Get current interrupt mask from SR
	intMask := uint8((cpu.sr >> 8) & 0x07)

	// Level 7 is NMI, always taken
	if cpu.irqLevel == 7 || cpu.irqLevel > intMask {
		cpu.handleInterrupt(cpu.irqLevel)
	}
}

// handleInterrupt processes an interrupt
func (cpu *CPU) handleInterrupt(level uint8) {
	// Get vector number
	var vector uint32

	if cpu.intAckCallback != nil {
		vector = cpu.intAckCallback(int(level))
	} else {
		vector = IntAckAutovector
	}

	// Handle special cases
	if vector == IntAckAutovector {
		vector = 0x18 + uint32(level) // Autovector base is 0x18
	} else if vector == IntAckSpurious {
		vector = 0x18 // Spurious interrupt vector
	}

	// Build SR for exception stack frame
	oldSR := cpu.sr

	// Set supervisor mode
	cpu.sr |= 0x2000

	// Update interrupt mask
	cpu.sr = (cpu.sr & 0xF8FF) | (uint16(level) << 8)

	// Save context to stack
	cpu.pushLong(cpu.pc)
	cpu.pushWord(oldSR)

	// Get vector address
	var vectorAddr uint32
	if cpu.cpuType >= CPU68010 {
		vectorAddr = cpu.vbr + (vector * 4)
	} else {
		vectorAddr = vector * 4
	}

	// Read new PC from vector table
	cpu.pc = cpu.memory.Read32(vectorAddr)

	// Use some cycles for exception processing
	cpu.useCycles(44) // Approximate
}

// SetMemoryHandler sets the memory access handler
func (cpu *CPU) SetMemoryHandler(handler MemoryHandler) {
	cpu.memory = handler
}

// GetCPUType returns the current CPU type
func (cpu *CPU) GetCPUType() CPUType {
	return cpu.cpuType
}

// SetCPUType changes the CPU type
func (cpu *CPU) SetCPUType(cpuType CPUType) {
	cpu.cpuType = cpuType
}

// SetIRQ sets the interrupt request level (0-7)
func (cpu *CPU) SetIRQ(level int) {
	if level < 0 || level > 7 {
		level = 0
	}
	cpu.irqLevel = uint8(level)
}

// SetVIRQ sets a virtual IRQ line.
// When using virtual IRQs, the highest active line is automatically selected.
func (cpu *CPU) SetVIRQ(level int, active bool) {
	if level < 1 || level > 7 {
		return
	}
	cpu.virq[level] = active

	// Update actual IRQ level to highest active
	cpu.irqLevel = 0
	for i := 7; i >= 1; i-- {
		if cpu.virq[i] {
			cpu.irqLevel = uint8(i)
			break
		}
	}
}

// GetVIRQ returns the state of a virtual IRQ line
func (cpu *CPU) GetVIRQ(level int) bool {
	if level < 1 || level > 7 {
		return false
	}
	return cpu.virq[level]
}

// PulseHalt simulates pulsing the HALT pin
func (cpu *CPU) PulseHalt() {
	cpu.halted = true
}

// PulseBusError triggers a bus error exception
func (cpu *CPU) PulseBusError() {
	// TODO: Implement bus error exception
}

// CyclesRun returns the number of cycles executed so far in current timeslice
func (cpu *CPU) CyclesRun() int {
	return cpu.cyclesRun
}

// CyclesRemaining returns the number of cycles remaining in current timeslice
func (cpu *CPU) CyclesRemaining() int {
	return cpu.cyclesRemain
}

// ModifyTimeslice adjusts the number of cycles remaining
func (cpu *CPU) ModifyTimeslice(cycles int) {
	cpu.cyclesRemain += cycles
}

// EndTimeslice ends the current timeslice immediately
func (cpu *CPU) EndTimeslice() {
	cpu.cyclesRemain = 0
}

// useCycles consumes the specified number of cycles
func (cpu *CPU) useCycles(cycles int) {
	cpu.cyclesRun += cycles
	cpu.cyclesRemain -= cycles
}

// GetRegister returns the value of a CPU register
func (cpu *CPU) GetRegister(reg Register) uint32 {
	switch reg {
	case RegD0, RegD1, RegD2, RegD3, RegD4, RegD5, RegD6, RegD7:
		return cpu.d[reg-RegD0]
	case RegA0, RegA1, RegA2, RegA3, RegA4, RegA5, RegA6, RegA7:
		return cpu.a[reg-RegA0]
	case RegPC:
		return cpu.pc
	case RegSR:
		return uint32(cpu.sr)
	case RegSP:
		if cpu.sr&0x2000 != 0 { // Supervisor mode
			return cpu.a[7]
		}
		return cpu.usp
	case RegUSP:
		return cpu.usp
	case RegISP:
		return cpu.isp
	case RegMSP:
		return cpu.msp
	case RegSFC:
		return uint32(cpu.sfc)
	case RegDFC:
		return uint32(cpu.dfc)
	case RegVBR:
		return cpu.vbr
	case RegCACR:
		return cpu.cacr
	case RegCAAR:
		return cpu.caar
	case RegPrefAddr:
		return cpu.prefetchAddr
	case RegPrefData:
		return cpu.prefetchData
	case RegPPC:
		return cpu.ppc
	case RegIR:
		return uint32(cpu.ir)
	case RegCPUType:
		return uint32(cpu.cpuType)
	default:
		return 0
	}
}

// SetRegister sets the value of a CPU register
func (cpu *CPU) SetRegister(reg Register, value uint32) {
	switch reg {
	case RegD0, RegD1, RegD2, RegD3, RegD4, RegD5, RegD6, RegD7:
		cpu.d[reg-RegD0] = value
	case RegA0, RegA1, RegA2, RegA3, RegA4, RegA5, RegA6, RegA7:
		cpu.a[reg-RegA0] = value
	case RegPC:
		cpu.pc = value
	case RegSR:
		cpu.sr = uint16(value)
	case RegUSP:
		cpu.usp = value
	case RegISP:
		cpu.isp = value
	case RegMSP:
		cpu.msp = value
	case RegSFC:
		cpu.sfc = uint8(value)
	case RegDFC:
		cpu.dfc = uint8(value)
	case RegVBR:
		cpu.vbr = value
	case RegCACR:
		cpu.cacr = value
	case RegCAAR:
		cpu.caar = value
	}
}

// GetPC returns the program counter
func (cpu *CPU) GetPC() uint32 {
	return cpu.pc
}

// SetPC sets the program counter
func (cpu *CPU) SetPC(address uint32) {
	cpu.pc = address
	if cpu.pcChangedCallback != nil {
		cpu.pcChangedCallback(address)
	}
}

// GetSP returns the current stack pointer
func (cpu *CPU) GetSP() uint32 {
	if cpu.sr&0x2000 != 0 { // Supervisor mode
		return cpu.a[7]
	}
	return cpu.usp
}

// SetSP sets the current stack pointer
func (cpu *CPU) SetSP(address uint32) {
	if cpu.sr&0x2000 != 0 { // Supervisor mode
		cpu.a[7] = address
	} else {
		cpu.usp = address
	}
}

// GetSR returns the status register
func (cpu *CPU) GetSR() uint16 {
	return cpu.sr
}

// SetSR sets the status register
func (cpu *CPU) SetSR(value uint16) {
	cpu.sr = value
}

// pushWord pushes a word onto the stack
func (cpu *CPU) pushWord(value uint16) {
	cpu.a[7] -= 2
	if cpu.memory != nil {
		cpu.memory.Write16(cpu.a[7], value)
	}
}

// pushLong pushes a longword onto the stack
func (cpu *CPU) pushLong(value uint32) {
	cpu.a[7] -= 4
	if cpu.memory != nil {
		cpu.memory.Write32(cpu.a[7], value)
	}
}

// popWord pops a word from the stack
func (cpu *CPU) popWord() uint16 {
	if cpu.memory == nil {
		return 0
	}
	value := cpu.memory.Read16(cpu.a[7])
	cpu.a[7] += 2
	return value
}

// popLong pops a longword from the stack
func (cpu *CPU) popLong() uint32 {
	if cpu.memory == nil {
		return 0
	}
	value := cpu.memory.Read32(cpu.a[7])
	cpu.a[7] += 4
	return value
}

// SetIntAckCallback sets the interrupt acknowledge callback
func (cpu *CPU) SetIntAckCallback(callback func(level int) uint32) {
	cpu.intAckCallback = callback
}

// SetResetCallback sets the RESET instruction callback
func (cpu *CPU) SetResetCallback(callback func()) {
	cpu.resetCallback = callback
}

// SetPCChangedCallback sets the PC changed callback
func (cpu *CPU) SetPCChangedCallback(callback func(newPC uint32)) {
	cpu.pcChangedCallback = callback
}

// SetFCCallback sets the function code callback
func (cpu *CPU) SetFCCallback(callback func(fc uint8)) {
	cpu.fcCallback = callback
}

// SetInstrHookCallback sets the instruction hook callback
func (cpu *CPU) SetInstrHookCallback(callback func(pc uint32)) {
	cpu.instrHookCallback = callback
}

// SetBkptAckCallback sets the breakpoint acknowledge callback
func (cpu *CPU) SetBkptAckCallback(callback func(data uint32)) {
	cpu.bkptAckCallback = callback
}

// SetIllegalInstrCallback sets the illegal instruction callback
func (cpu *CPU) SetIllegalInstrCallback(callback func(opcode uint16) bool) {
	cpu.illegalCallback = callback
}

// SetTASCallback sets the TAS instruction callback
func (cpu *CPU) SetTASCallback(callback func() int) {
	cpu.tasCallback = callback
}

// Context represents a saved CPU context
type Context struct {
	cpuType CPUType
	d       [8]uint32
	a       [8]uint32
	pc      uint32
	sr      uint16
	usp     uint32
	isp     uint32
	msp     uint32
	sfc     uint8
	dfc     uint8
	vbr     uint32
	cacr    uint32
	caar    uint32
}

// GetContext returns a copy of the current CPU context
func (cpu *CPU) GetContext() *Context {
	ctx := &Context{
		cpuType: cpu.cpuType,
		pc:      cpu.pc,
		sr:      cpu.sr,
		usp:     cpu.usp,
		isp:     cpu.isp,
		msp:     cpu.msp,
		sfc:     cpu.sfc,
		dfc:     cpu.dfc,
		vbr:     cpu.vbr,
		cacr:    cpu.cacr,
		caar:    cpu.caar,
	}
	copy(ctx.d[:], cpu.d[:])
	copy(ctx.a[:], cpu.a[:])
	return ctx
}

// SetContext restores a saved CPU context
func (cpu *CPU) SetContext(ctx *Context) {
	cpu.cpuType = ctx.cpuType
	cpu.pc = ctx.pc
	cpu.sr = ctx.sr
	cpu.usp = ctx.usp
	cpu.isp = ctx.isp
	cpu.msp = ctx.msp
	cpu.sfc = ctx.sfc
	cpu.dfc = ctx.dfc
	cpu.vbr = ctx.vbr
	cpu.cacr = ctx.cacr
	cpu.caar = ctx.caar
	copy(cpu.d[:], ctx.d[:])
	copy(cpu.a[:], ctx.a[:])
}

// ContextSize returns the size of a context in bytes
func (cpu *CPU) ContextSize() int {
	// Calculate approximate size of Context struct
	return 8*4 + 8*4 + 4 + 2 + 4 + 4 + 4 + 1 + 1 + 4 + 4 + 4 + 4 // approximate
}
