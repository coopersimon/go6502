package go6502

// A cpu method that gets the address.
type addrModeReadFn = func(cpu *CPU) uint16

// MemoryBus
type MemoryBus interface {
	Read(addr uint16) uint8
	Write(addr uint16, data uint8)
	Clock(cycles uint32) Interrupt
}

// CPU: 6502 CPU
type CPU struct {
	// Registers
	acc uint8        // Accumulator
	x   uint8        // X Register
	y   uint8        // Y Register
	sp  uint8        // Stack Pointer
	pc  uint16       // Program Counter
	pf  ProgramFlags // Program Flags

	// Memory
	memBus MemoryBus // 16-bit bus

	// Interrupts
	halt    bool
	intMask Interrupt
	cycles  uint32
}

func New(memoryBus MemoryBus) *CPU {
	return &CPU{
		0,
		0,
		0,
		0xFF,
		0xFFFE,
		0,

		memoryBus,

		false,
		0,
		0,
	}
}

// Step a single instruction and clock the memory bus.
func (cpu *CPU) Step() uint32 {
	if cpu.intMask != 0 {
		cpu.handleInterrupt()
	} else if !cpu.halt {
		cpu.executeInstruction()
	}

	var cycles = cpu.cycles
	cpu.cycles = 0
	cpu.intMask.Set(cpu.memBus.Clock(cycles))
	return cycles
}

/*** INTERNAL ***/

func (cpu *CPU) handleInterrupt() {
	if cpu.intMask.Test(NMI) {
		cpu.interruptRoutine(NMI, 0xFFFA)
	} else if cpu.intMask.Test(IRQ) && !cpu.pf.Test(I) {
		cpu.interruptRoutine(IRQ, 0xFFFE)
	}
}

func (cpu *CPU) interruptRoutine(intFlag Interrupt, vector uint16) {
	cpu.intMask.Clear(intFlag)
	cpu.stackPush(Hi(cpu.pc))
	cpu.stackPush(Lo(cpu.pc))
	cpu.stackPush(uint8(cpu.pf))

	cpu.pf.Set(I)

	var pcLo = cpu.memRead(vector)
	var pcHi = cpu.memRead(vector + 1)

	cpu.pc = Make16(pcHi, pcLo)
}

func (cpu *CPU) executeInstruction() {
	var instr = cpu.fetch()

	switch instr {
	case 0x00:
		cpu.brk()

	case 0x01:
		cpu.ora((*CPU).indexedIndirect)
	case 0x11:
		cpu.ora((*CPU).indirectIndexed)
	case 0x05:
		cpu.ora((*CPU).zeroPage)
	case 0x15:
		cpu.ora((*CPU).zeroPageX)
	case 0x09:
		cpu.ora(nil)
	case 0x19:
		cpu.ora((*CPU).absoluteY)
	case 0x0D:
		cpu.ora((*CPU).absolute)
	case 0x1D:
		cpu.ora((*CPU).absoluteX)

	case 0x21:
		cpu.and((*CPU).indexedIndirect)
	case 0x31:
		cpu.and((*CPU).indirectIndexed)
	case 0x25:
		cpu.and((*CPU).zeroPage)
	case 0x35:
		cpu.and((*CPU).zeroPageX)
	case 0x29:
		cpu.and(nil)
	case 0x39:
		cpu.and((*CPU).absoluteY)
	case 0x2D:
		cpu.and((*CPU).absolute)
	case 0x3D:
		cpu.and((*CPU).absoluteX)

	case 0x41:
		cpu.eor((*CPU).indexedIndirect)
	case 0x51:
		cpu.eor((*CPU).indirectIndexed)
	case 0x45:
		cpu.eor((*CPU).zeroPage)
	case 0x55:
		cpu.eor((*CPU).zeroPageX)
	case 0x49:
		cpu.eor(nil)
	case 0x59:
		cpu.eor((*CPU).absoluteY)
	case 0x4D:
		cpu.eor((*CPU).absolute)
	case 0x5D:
		cpu.eor((*CPU).absoluteX)

	case 0x61:
		cpu.adc((*CPU).indexedIndirect)
	case 0x71:
		cpu.adc((*CPU).indirectIndexed)
	case 0x65:
		cpu.adc((*CPU).zeroPage)
	case 0x75:
		cpu.adc((*CPU).zeroPageX)
	case 0x69:
		cpu.adc(nil)
	case 0x79:
		cpu.adc((*CPU).absoluteY)
	case 0x6D:
		cpu.adc((*CPU).absolute)
	case 0x7D:
		cpu.adc((*CPU).absoluteX)

	case 0x81:
		cpu.sta((*CPU).indexedIndirect)
	case 0x91:
		cpu.sta((*CPU).indirectIndexed)
	case 0x85:
		cpu.sta((*CPU).zeroPage)
	case 0x95:
		cpu.sta((*CPU).zeroPageX)
	case 0x99:
		cpu.sta((*CPU).absoluteY)
	case 0x8D:
		cpu.sta((*CPU).absolute)
	case 0x9D:
		cpu.sta((*CPU).absoluteX)

	case 0x84:
		cpu.sty((*CPU).zeroPage)
	case 0x94:
		cpu.sty((*CPU).zeroPageX)
	case 0x8C:
		cpu.sty((*CPU).absolute)

	case 0x86:
		cpu.stx((*CPU).zeroPage)
	case 0x96:
		cpu.stx((*CPU).zeroPageY)
	case 0x8E:
		cpu.stx((*CPU).absolute)

	case 0xA1:
		cpu.lda((*CPU).indexedIndirect)
	case 0xB1:
		cpu.lda((*CPU).indirectIndexed)
	case 0xA5:
		cpu.lda((*CPU).zeroPage)
	case 0xB5:
		cpu.lda((*CPU).zeroPageX)
	case 0xA9:
		cpu.lda(nil)
	case 0xB9:
		cpu.lda((*CPU).absoluteY)
	case 0xAD:
		cpu.lda((*CPU).absolute)
	case 0xBD:
		cpu.lda((*CPU).absoluteX)

	case 0xA0:
		cpu.ldy(nil)
	case 0xA4:
		cpu.ldy((*CPU).zeroPage)
	case 0xAC:
		cpu.ldy((*CPU).absolute)
	case 0xB4:
		cpu.ldy((*CPU).zeroPageX)
	case 0xBC:
		cpu.ldy((*CPU).absoluteX)

	case 0xA2:
		cpu.ldx(nil)
	case 0xA6:
		cpu.ldx((*CPU).zeroPage)
	case 0xAE:
		cpu.ldx((*CPU).absolute)
	case 0xB6:
		cpu.ldx((*CPU).zeroPageY)
	case 0xBE:
		cpu.ldx((*CPU).absoluteY)

	case 0xC0:
		cpu.cmp(cpu.y, nil)
	case 0xC4:
		cpu.cmp(cpu.y, (*CPU).zeroPage)
	case 0xCC:
		cpu.cmp(cpu.y, (*CPU).absolute)

	case 0xE0:
		cpu.cmp(cpu.x, nil)
	case 0xE4:
		cpu.cmp(cpu.x, (*CPU).zeroPage)
	case 0xEC:
		cpu.cmp(cpu.x, (*CPU).absolute)

	case 0xC1:
		cpu.cmp(cpu.acc, (*CPU).indexedIndirect)
	case 0xD1:
		cpu.cmp(cpu.acc, (*CPU).indirectIndexed)
	case 0xC5:
		cpu.cmp(cpu.acc, (*CPU).zeroPage)
	case 0xD5:
		cpu.cmp(cpu.acc, (*CPU).zeroPageX)
	case 0xC9:
		cpu.cmp(cpu.acc, nil)
	case 0xD9:
		cpu.cmp(cpu.acc, (*CPU).absoluteY)
	case 0xCD:
		cpu.cmp(cpu.acc, (*CPU).absolute)
	case 0xDD:
		cpu.cmp(cpu.acc, (*CPU).absoluteX)

	case 0xE1:
		cpu.sbc((*CPU).indexedIndirect)
	case 0xF1:
		cpu.sbc((*CPU).indirectIndexed)
	case 0xE5:
		cpu.sbc((*CPU).zeroPage)
	case 0xF5:
		cpu.sbc((*CPU).zeroPageX)
	case 0xE9:
		cpu.sbc(nil)
	case 0xF9:
		cpu.sbc((*CPU).absoluteY)
	case 0xED:
		cpu.sbc((*CPU).absolute)
	case 0xFD:
		cpu.sbc((*CPU).absoluteX)

	case 0x0A:
		cpu.asl(nil)
	case 0x06:
		cpu.asl((*CPU).zeroPage)
	case 0x16:
		cpu.asl((*CPU).zeroPageX)
	case 0x0E:
		cpu.asl((*CPU).absolute)
	case 0x1E:
		cpu.asl((*CPU).absoluteX)

	case 0x2A:
		cpu.rol(nil)
	case 0x26:
		cpu.rol((*CPU).zeroPage)
	case 0x36:
		cpu.rol((*CPU).zeroPageX)
	case 0x2E:
		cpu.rol((*CPU).absolute)
	case 0x3E:
		cpu.rol((*CPU).absoluteX)

	case 0x4A:
		cpu.lsr(nil)
	case 0x46:
		cpu.lsr((*CPU).zeroPage)
	case 0x56:
		cpu.lsr((*CPU).zeroPageX)
	case 0x4E:
		cpu.lsr((*CPU).absolute)
	case 0x5E:
		cpu.lsr((*CPU).absoluteX)

	case 0x6A:
		cpu.ror(nil)
	case 0x66:
		cpu.ror((*CPU).zeroPage)
	case 0x76:
		cpu.ror((*CPU).zeroPageX)
	case 0x6E:
		cpu.ror((*CPU).absolute)
	case 0x7E:
		cpu.ror((*CPU).absoluteX)

	case 0xC6:
		cpu.dec((*CPU).zeroPage)
	case 0xD6:
		cpu.dec((*CPU).zeroPageX)
	case 0xCE:
		cpu.dec((*CPU).absolute)
	case 0xDE:
		cpu.dec((*CPU).absoluteX)

	case 0xE6:
		cpu.inc((*CPU).zeroPage)
	case 0xF6:
		cpu.inc((*CPU).zeroPageX)
	case 0xEE:
		cpu.inc((*CPU).absolute)
	case 0xFE:
		cpu.inc((*CPU).absoluteX)

	case 0xCA:
		cpu.dex()
	case 0x88:
		cpu.dey()

	case 0xE8:
		cpu.inx()
	case 0xC8:
		cpu.iny()

	case 0x98: // TYA
		cpu.setNZ(cpu.y)
		cpu.acc = cpu.y
	case 0xA8: // TAY
		cpu.setNZ(cpu.acc)
		cpu.y = cpu.acc
	case 0x8A: // TXA
		cpu.setNZ(cpu.x)
		cpu.acc = cpu.x
	case 0x9A: // TXS
		cpu.setNZ(cpu.x)
		cpu.sp = cpu.x
	case 0xAA: // TAX
		cpu.setNZ(cpu.acc)
		cpu.x = cpu.acc
	case 0xBA: // TSX
		cpu.setNZ(cpu.sp)
		cpu.x = cpu.sp

	case 0x18:
		cpu.pf.Clear(C) // CLC
	case 0x38:
		cpu.pf.Set(C) // SEC
	case 0x58:
		cpu.pf.Clear(I) // CLI
	case 0x78:
		cpu.pf.Set(I) // SEI
	case 0xB8:
		cpu.pf.Clear(V) // CLV
	case 0xD8:
		cpu.pf.Clear(D) // CLD
	case 0xF8:
		cpu.pf.Set(D) //SED

	case 0x24:
		cpu.bit((*CPU).zeroPage)
	case 0x2C:
		cpu.bit((*CPU).absolute)

	case 0x08:
		cpu.stackPush(uint8(cpu.pf)) // PHP
	case 0x28:
		cpu.pf = ProgramFlags(cpu.stackPop()) // PLP
	case 0x48:
		cpu.stackPush(cpu.acc) // PHA
	case 0x68:
		cpu.acc = cpu.stackPop() // PLA

	case 0x10:
		cpu.branch(!cpu.pf.Test(N)) // BPL
	case 0x30:
		cpu.branch(cpu.pf.Test(N)) // BMI
	case 0x50:
		cpu.branch(!cpu.pf.Test(V)) // BVC
	case 0x70:
		cpu.branch(cpu.pf.Test(V)) // BVS
	case 0x90:
		cpu.branch(!cpu.pf.Test(C)) // BCC
	case 0xB0:
		cpu.branch(cpu.pf.Test(C)) // BCS
	case 0xD0:
		cpu.branch(!cpu.pf.Test(Z)) // BNE
	case 0xF0:
		cpu.branch(cpu.pf.Test(Z)) // BEQ

	case 0x20:
		cpu.jsr()
	case 0x4C:
		cpu.jmpAbsolute()
	case 0x6C:
		cpu.jmpIndirect()
	case 0x40:
		cpu.rti()
	case 0x60:
		cpu.rts()
	}
}

/*** Basic Memory ***/

// Read from the bus.
func (cpu *CPU) memRead(addr uint16) uint8 {
	cpu.cycles++
	return cpu.memBus.Read(addr)
}

// Write to the bus.
func (cpu *CPU) memWrite(addr uint16, data uint8) {
	cpu.cycles++
	cpu.memBus.Write(addr, data)
}

// Fetch the next byte from the program counter.
func (cpu *CPU) fetch() uint8 {
	var data = cpu.memRead(cpu.pc)
	cpu.pc++
	return data
}

// Push a byte onto the stack.
func (cpu *CPU) stackPush(data uint8) {
	cpu.memWrite(Make16(1, cpu.sp), data)
	cpu.sp--
}

// Pop a byte off from the stack.
func (cpu *CPU) stackPop() uint8 {
	cpu.sp++
	return cpu.memRead(Make16(1, cpu.sp))
}

/*** Addressing modes ***/

// $xx
func (cpu *CPU) zeroPage() uint16 {
	return uint16(cpu.fetch())
}

// $xx, X
func (cpu *CPU) zeroPageX() uint16 {
	return uint16(cpu.fetch() + cpu.x)
}

// $xx, Y
func (cpu *CPU) zeroPageY() uint16 {
	return uint16(cpu.fetch() + cpu.y)
}

// $xxxx
func (cpu *CPU) absolute() uint16 {
	var addrLo = cpu.fetch()
	var addrHi = cpu.fetch()

	return Make16(addrHi, addrLo)
}

// $xxxx, X
func (cpu *CPU) absoluteX() uint16 {
	var addrLo = cpu.fetch()
	var addrHi = cpu.fetch()

	return Make16(addrHi, addrLo) + uint16(cpu.x)
}

// $xxxx, Y
func (cpu *CPU) absoluteY() uint16 {
	var addrLo = cpu.fetch()
	var addrHi = cpu.fetch()

	return Make16(addrHi, addrLo) + uint16(cpu.y)
}

// ($xx, X)
func (cpu *CPU) indexedIndirect() uint16 {
	var target = uint16(cpu.fetch() + cpu.x)

	var addrLo = cpu.memRead(target)
	var addrHi = cpu.memRead(target + 1)

	return Make16(addrHi, addrLo)
}

// ($xx), Y
func (cpu *CPU) indirectIndexed() uint16 {
	var target = uint16(cpu.fetch())

	var addrLo = cpu.memRead(target)
	var addrHi = cpu.memRead(target + 1)

	return Make16(addrHi, addrLo) + uint16(cpu.y)
}

// Addressing modes

/*** Instructions ***/

/*** Arithmetic ***/

func (cpu *CPU) adc(addrMode addrModeReadFn) {
	data, _ := cpu.dataAddr(addrMode)

	if cpu.pf.Test(D) {
		// Decimal
	} else {
		cpu.binaryArithmetic(data)
	}
}

func (cpu *CPU) sbc(addrMode addrModeReadFn) {
	data, _ := cpu.dataAddr(addrMode)

	if cpu.pf.Test(D) {
		// Decimal
	} else {
		cpu.binaryArithmetic(^data)
	}
}

func (cpu *CPU) inc(addrMode addrModeReadFn) {
	data, addr := cpu.dataAddr(addrMode)

	cpu.memWrite(addr, data+1)
}

func (cpu *CPU) dec(addrMode addrModeReadFn) {
	data, addr := cpu.dataAddr(addrMode)

	cpu.memWrite(addr, data-1)
}

func (cpu *CPU) inx() {
	cpu.x++
}

func (cpu *CPU) dex() {
	cpu.x--
}

func (cpu *CPU) iny() {
	cpu.y++
}

func (cpu *CPU) dey() {
	cpu.y--
}

/*** Bitwise ***/

func (cpu *CPU) ora(addrMode addrModeReadFn) {
	data, _ := cpu.dataAddr(addrMode)
	cpu.acc |= data
	cpu.setNZ(cpu.acc)
}

func (cpu *CPU) and(addrMode addrModeReadFn) {
	data, _ := cpu.dataAddr(addrMode)
	cpu.acc &= data
	cpu.setNZ(cpu.acc)
}

func (cpu *CPU) eor(addrMode addrModeReadFn) {
	data, _ := cpu.dataAddr(addrMode)
	cpu.acc ^= data
	cpu.setNZ(cpu.acc)
}

func (cpu *CPU) asl(addrMode addrModeReadFn) {
	const highBit = 1 << 7

	var data uint8
	var addr uint16
	if addrMode == nil {
		data, addr = cpu.dataAddr(addrMode)
	} else {
		data = cpu.acc
	}

	cpu.pf.SetIf(C, (data&highBit) != 0)
	cpu.setNZ(data)

	if addrMode == nil {
		cpu.memWrite(addr, data<<1)
	} else {
		cpu.acc = data << 1
	}
}

func (cpu *CPU) lsr(addrMode addrModeReadFn) {
	const lowBit = 1 << 0

	var data uint8
	var addr uint16
	if addrMode == nil {
		data, addr = cpu.dataAddr(addrMode)
	} else {
		data = cpu.acc
	}

	cpu.pf.SetIf(C, (data&lowBit) != 0)
	cpu.setNZ(data)

	if addrMode == nil {
		cpu.memWrite(addr, data>>1)
	} else {
		cpu.acc = data >> 1
	}
}

func (cpu *CPU) rol(addrMode addrModeReadFn) {
	const highBit = 1 << 7

	var data uint8
	var addr uint16
	if addrMode == nil {
		data, addr = cpu.dataAddr(addrMode)
	} else {
		data = cpu.acc
	}

	var carry = uint8(cpu.pf & C)
	var result = (data << 1) | carry

	cpu.pf.SetIf(C, (data&highBit) != 0)
	cpu.setNZ(data)

	if addrMode == nil {
		cpu.memWrite(addr, result)
	} else {
		cpu.acc = result
	}
}

func (cpu *CPU) ror(addrMode addrModeReadFn) {
	const lowBit = 1 << 0

	var data uint8
	var addr uint16
	if addrMode == nil {
		data, addr = cpu.dataAddr(addrMode)
	} else {
		data = cpu.acc
	}

	var carry = uint8(cpu.pf&C) << 7
	var result = (data >> 1) | carry

	cpu.pf.SetIf(C, (data&lowBit) != 0)
	cpu.setNZ(data)

	if addrMode == nil {
		cpu.memWrite(addr, result)
	} else {
		cpu.acc = result
	}
}

/*** Data moving ***/

func (cpu *CPU) sta(addrMode addrModeReadFn) {
	addr := addrMode(cpu)
	cpu.memWrite(addr, cpu.acc)
}

func (cpu *CPU) sty(addrMode addrModeReadFn) {
	addr := addrMode(cpu)
	cpu.memWrite(addr, cpu.y)
}

func (cpu *CPU) stx(addrMode addrModeReadFn) {
	addr := addrMode(cpu)
	cpu.memWrite(addr, cpu.x)
}

func (cpu *CPU) lda(addrMode addrModeReadFn) {
	data, _ := cpu.dataAddr(addrMode)
	cpu.setNZ(data)
	cpu.acc = data
}

func (cpu *CPU) ldy(addrMode addrModeReadFn) {
	data, _ := cpu.dataAddr(addrMode)
	cpu.setNZ(data)
	cpu.y = data
}

func (cpu *CPU) ldx(addrMode addrModeReadFn) {
	data, _ := cpu.dataAddr(addrMode)
	cpu.setNZ(data)
	cpu.x = data
}

/*** Flags ***/

func (cpu *CPU) cmp(reg uint8, addrMode addrModeReadFn) {
	data, _ := cpu.dataAddr(addrMode)

	cpu.setNZ(reg - data)
	cpu.pf.SetIf(C, reg >= data)
}

func (cpu *CPU) bit(addrMode addrModeReadFn) {
	const signBit = 1 << 7
	const overflowBit = 1 << 6

	data, _ := cpu.dataAddr(addrMode)

	cpu.pf.SetIf(N, (data&signBit) != 0)
	cpu.pf.SetIf(V, (data&overflowBit) != 0)
	cpu.pf.SetIf(Z, (cpu.acc&data) == 0)
}

/*** Branches ***/

func (cpu *CPU) branch(cond bool) {
	data := int16(int8(cpu.fetch()))

	if cond {
		cpu.pc += uint16(data)
	}
}

func (cpu *CPU) jmpAbsolute() {
	cpu.pc = cpu.absolute()
}

func (cpu *CPU) jmpIndirect() {
	var addr = cpu.absolute()

	var pcLo = cpu.memRead(addr)
	var pcHi = cpu.memRead(Make16(Hi(addr), Lo(addr)+1))

	cpu.pc = Make16(pcHi, pcLo)
}

func (cpu *CPU) jsr() {
	var addr = cpu.absolute()

	var storePC = cpu.pc - 1
	cpu.stackPush(Hi(storePC))
	cpu.stackPush(Lo(storePC))

	cpu.pc = addr
}

func (cpu *CPU) rts() {
	var cpuLo = cpu.stackPop()
	var cpuHi = cpu.stackPop()

	cpu.pc = Make16(cpuHi, cpuLo) + 1
}

func (cpu *CPU) rti() {
	cpu.pf = ProgramFlags(cpu.stackPop())

	var cpuLo = cpu.stackPop()
	var cpuHi = cpu.stackPop()

	cpu.pc = Make16(cpuHi, cpuLo)
}

/*** MISC ***/

func (cpu *CPU) brk() {
	cpu.pf.Set(B)
	cpu.interruptRoutine(IRQ, 0xFFFE)
	cpu.pf.Clear(B)
}

/*** Instruction Helpers ***/

func (cpu *CPU) setNZ(data uint8) {
	const signBit = 1 << 7
	cpu.pf.SetIf(N, (data&signBit) != 0)
	cpu.pf.SetIf(Z, data == 0)
}

func (cpu *CPU) binaryArithmetic(data uint8) {
	const signBit = 1 << 7
	const carryBit = 1 << 8

	var carry = uint16(cpu.pf & C)
	var result = uint16(cpu.acc) + uint16(data) + carry
	var finalResult = Lo(result)

	cpu.pf.SetIf(N, (finalResult&signBit) != 0)
	cpu.pf.SetIf(V, ^((cpu.acc^data)&(cpu.acc^finalResult)) == signBit)
	cpu.pf.SetIf(Z, finalResult == 0)
	cpu.pf.SetIf(C, (result&carryBit) != 0)

	cpu.acc = finalResult
}

// Resolve an address and load the data.
func (cpu *CPU) dataAddr(addrMode addrModeReadFn) (data uint8, addr uint16) {
	if addrMode == nil {
		addr = 0
		data = cpu.fetch()
	} else {
		addr = addrMode(cpu)
		data = cpu.memRead(addr)
	}

	return
}
