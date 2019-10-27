package flags

// Interrupt values.
type Interrupt uint8

// Interrupt constants
const (
	IRQ Interrupt = 1 << iota
	NMI Interrupt = 1 << iota
)

func (i Interrupt) Test(flags Interrupt) bool {
	return (i & flags) == flags
}

func (i *Interrupt) Set(flags Interrupt) {
	*i |= flags
}

func (i *Interrupt) Clear(flags Interrupt) {
	*i &^= flags
}

// 6502 Program Flags
type ProgramFlags uint8

// Flag constants
const (
	C ProgramFlags = 1 << 0
	Z ProgramFlags = 1 << 1
	I ProgramFlags = 1 << 2
	D ProgramFlags = 1 << 3
	B ProgramFlags = 1 << 4
	V ProgramFlags = 1 << 6
	N ProgramFlags = 1 << 7
)

// Check if one or more flags are set.
func (pf ProgramFlags) Test(flags ProgramFlags) bool {
	return (pf & flags) == flags
}

// Set the flags specified.
func (pf *ProgramFlags) Set(flags ProgramFlags) {
	*pf |= flags
}

// Clear the flags specified.
func (pf *ProgramFlags) Clear(flags ProgramFlags) {
	*pf &^= flags
}

// Set or clear the flag based on the condition.
func (pf *ProgramFlags) SetIf(flags ProgramFlags, cond bool) {
	if cond {
		pf.Set(flags)
	} else {
		pf.Clear(flags)
	}
}
