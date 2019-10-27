package go6502

import (
	"github.com/coopersimon/go6502/flags"
	"github.com/coopersimon/go6502/utils"
)

// A cpu method that gets data and the address used (if one was used).
type addrModeReadFn = func(cpu *CPU) (uint8, uint16)

// A cpu method that writes data back to the relevant location.
type addrModeWriteFn = func(cpu *CPU, addr uint16) uint8

// MemoryBus
type MemoryBus interface {
	Read(addr uint16) uint8
	Write(addr uint16, data uint8)
	Clock(cycles uint32) flags.Interrupt
}

// CPU: 6502 CPU
type CPU struct {
	// Registers
	acc uint8              // Accumulator
	x   uint8              // X Register
	y   uint8              // Y Register
	sp  uint8              // Stack Pointer
	pc  uint16             // Program Counter
	pf  flags.ProgramFlags // Program Flags

	// Memory
	memBus MemoryBus // 16-bit bus

	// Interrupts
	halt    bool
	intMask flags.Interrupt
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
	if cpu.intMask.Test(flags.NMI) {
		cpu.interruptRoutine(flags.NMI, 0xFFFA)
	} else if cpu.intMask.Test(flags.IRQ) && !cpu.pf.Test(flags.I) {
		cpu.interruptRoutine(flags.IRQ, 0xFFFE)
	}
}

func (cpu *CPU) interruptRoutine(intFlag flags.Interrupt, vector uint16) {
	cpu.intMask.Clear(intFlag)
	cpu.stackPush(utils.Hi(cpu.pc))
	cpu.stackPush(utils.Lo(cpu.pc))
	cpu.stackPush(uint8(cpu.pf))

	cpu.pf.Set(flags.I)
	cpu.pf.Clear(flags.B)

	var pcLo = cpu.memRead(vector)
	var pcHi = cpu.memRead(vector + 1)

	cpu.pc = utils.Make16(pcHi, pcLo)
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
		cpu.ora((*CPU).immediate)
	case 0x19:
		cpu.ora((*CPU).absoluteY)
	case 0x0A:
		cpu.ora((*CPU).absolute)
	case 0x1A:
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
		cpu.and((*CPU).immediate)
	case 0x39:
		cpu.and((*CPU).absoluteY)
	case 0x2A:
		cpu.and((*CPU).absolute)
	case 0x3A:
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
		cpu.eor((*CPU).immediate)
	case 0x59:
		cpu.eor((*CPU).absoluteY)
	case 0x4A:
		cpu.eor((*CPU).absolute)
	case 0x5A:
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
		cpu.adc((*CPU).immediate)
	case 0x79:
		cpu.adc((*CPU).absoluteY)
	case 0x6A:
		cpu.adc((*CPU).absolute)
	case 0x7A:
		cpu.adc((*CPU).absoluteX)

	case 0xE1:
		cpu.sbc((*CPU).indexedIndirect)
	case 0xF1:
		cpu.sbc((*CPU).indirectIndexed)
	case 0xE5:
		cpu.sbc((*CPU).zeroPage)
	case 0xF5:
		cpu.sbc((*CPU).zeroPageX)
	case 0xE9:
		cpu.sbc((*CPU).immediate)
	case 0xF9:
		cpu.sbc((*CPU).absoluteY)
	case 0xEA:
		cpu.sbc((*CPU).absolute)
	case 0xFA:
		cpu.sbc((*CPU).absoluteX)
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
	cpu.memWrite(utils.Make16(1, cpu.sp), data)
	cpu.sp--
}

// Pop a byte off from the stack.
func (cpu *CPU) stackPop() uint8 {
	cpu.sp++
	return cpu.memRead(utils.Make16(1, cpu.sp))
}

/*** Addressing modes ***/

// A
func (cpu *CPU) accRead() (uint8, uint16) {
	return cpu.acc, 0
}

// #vv
func (cpu *CPU) immediate() (uint8, uint16) {
	var data = cpu.fetch()
	return data, 0
}

// $xx
func (cpu *CPU) zeroPage() (uint8, uint16) {
	var addr = uint16(cpu.fetch())
	var data = cpu.memRead(addr)

	return data, addr
}

// $xx, X
func (cpu *CPU) zeroPageX() (uint8, uint16) {
	var addr = uint16(cpu.fetch() + cpu.x)
	var data = cpu.memRead(addr)

	return data, addr
}

// $xx, Y
func (cpu *CPU) zeroPageY() (uint8, uint16) {
	var addr = uint16(cpu.fetch() + cpu.y)
	var data = cpu.memRead(addr)

	return data, addr
}

// $xxxx
func (cpu *CPU) absolute() (uint8, uint16) {
	var addrLo = cpu.fetch()
	var addrHi = cpu.fetch()

	var addr = utils.Make16(addrHi, addrLo)
	var data = cpu.memRead(addr)

	return data, addr
}

// $xxxx, X
func (cpu *CPU) absoluteX() (uint8, uint16) {
	var addrLo = cpu.fetch()
	var addrHi = cpu.fetch()

	var addr = utils.Make16(addrHi, addrLo) + uint16(cpu.x)
	var data = cpu.memRead(addr)

	return data, addr
}

// $xxxx, Y
func (cpu *CPU) absoluteY() (uint8, uint16) {
	var addrLo = cpu.fetch()
	var addrHi = cpu.fetch()

	var addr = utils.Make16(addrHi, addrLo) + uint16(cpu.y)
	var data = cpu.memRead(addr)

	return data, addr
}

// ($xx, X)
func (cpu *CPU) indexedIndirect() (uint8, uint16) {
	var target = uint16(cpu.fetch() + cpu.x)

	var addrLo = cpu.memRead(target)
	var addrHi = cpu.memRead(target + 1)

	var addr = utils.Make16(addrHi, addrLo)
	var data = cpu.memRead(addr)

	return data, addr
}

// ($xx), Y
func (cpu *CPU) indirectIndexed() (uint8, uint16) {
	var target = uint16(cpu.fetch())

	var addrLo = cpu.memRead(target)
	var addrHi = cpu.memRead(target + 1)

	var addr = utils.Make16(addrHi, addrLo) + uint16(cpu.y)
	var data = cpu.memRead(addr)

	return data, addr
}

// Addressing modes

/*** Instructions ***/

/*** Arithmetic ***/

func (cpu *CPU) adc(dataMode addrModeReadFn) {
	data, _ := dataMode(cpu)

	if cpu.pf.Test(flags.D) {
		// Decimal
	} else {
		cpu.binaryArithmetic(data)
	}
}

func (cpu *CPU) sbc(dataMode addrModeReadFn) {
	data, _ := dataMode(cpu)

	if cpu.pf.Test(flags.D) {
		// Decimal
	} else {
		cpu.binaryArithmetic(^data)
	}
}

func (cpu *CPU) ora(dataMode addrModeReadFn) {
	data, _ := dataMode(cpu)
	cpu.acc |= data
	cpu.setNZ(cpu.acc)
}

func (cpu *CPU) and(dataMode addrModeReadFn) {
	data, _ := dataMode(cpu)
	cpu.acc &= data
	cpu.setNZ(cpu.acc)
}

func (cpu *CPU) eor(dataMode addrModeReadFn) {
	data, _ := dataMode(cpu)
	cpu.acc ^= data
	cpu.setNZ(cpu.acc)
}

/*** MISC ***/

func (cpu *CPU) brk() {
	cpu.pf.Set(flags.B)
	cpu.interruptRoutine(flags.IRQ, 0xFFFE)
}

/*** Instruction Helpers ***/

func (cpu *CPU) setNZ(data uint8) {
	const signBit = 1 << 7
	cpu.pf.SetIf(flags.N, (data&signBit) != 0)
	cpu.pf.SetIf(flags.Z, data == 0)
}

func (cpu *CPU) binaryArithmetic(data uint8) {
	const signBit = 1 << 7
	const carryBit = 1 << 8

	var carry = uint16(cpu.pf & flags.C)
	var result = uint16(cpu.acc) + uint16(data) + carry
	var finalResult = utils.Lo(result)

	cpu.pf.SetIf(flags.N, (finalResult&signBit) != 0)
	cpu.pf.SetIf(flags.V, ^((cpu.acc^data)&(cpu.acc^finalResult)) == signBit)
	cpu.pf.SetIf(flags.Z, finalResult == 0)
	cpu.pf.SetIf(flags.C, (result&carryBit) != 0)

	cpu.acc = finalResult
}
