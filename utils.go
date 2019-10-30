package go6502

// Make a 16-bit int from two 8 bit ints.
func Make16(hi, lo uint8) uint16 {
	return (uint16(hi) << 8) | uint16(lo)
}

// Get high byte of a 16-bit int.
func Hi(in uint16) uint8 {
	return uint8(in >> 8)
}

// Get high byte of a 16-bit int.
func Lo(in uint16) uint8 {
	return uint8(in & 0xFF)
}
