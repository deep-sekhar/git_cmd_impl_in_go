package network

type PendingDelta struct {
	baseHash  string
	deltaData []byte
}

func readSize(data []byte) (int, []byte) {
	size := 0
	shift := 0
	// Git uses a variable-length encoding for sizes in delta objects. Each byte contributes 7 bits to the size, and the most significant bit of each byte indicates whether there are more bytes to read (1) or if it's the last byte (0). We read bytes until we encounter a byte with the most significant bit set to 0, which indicates the end of the size encoding. We also need to keep track of how many bytes we have consumed from the input data so that we can return the remaining data after reading the size.
	// example: if the size is 300, it would be encoded as two bytes: 0b10111100 (0xBC) and 0b00000001 (0x01). The first byte has the most significant bit set to 1, indicating that there is another byte to read. The second byte has the most significant bit set to 0, indicating that it's the last byte. To decode this, we take the first byte (0xBC) and mask out the most significant bit to get 0x3C (60 in decimal), then we take the second byte (0x01) and shift it left by 7 bits to get 0x80 (128 in decimal). Finally, we add these two values together to get the original size of 300.
	for i, b := range data {
		size |= int(b&0x7f) << shift
		// 0x7f means 0b01111111, which masks out the most significant bit of the byte to get the 7 bits that contribute to the size.
		shift += 7
		if b&0x80 == 0 {
			// 0x80 means 0b10000000, which checks if the most significant bit of the byte is set to 1. If it's 0, it means this is the last byte of the size encoding, and we can stop reading further bytes.
			// delta[i+1:] gives us the remaining data after consuming the bytes for the size encoding.
			return size, data[i+1:]
		}
	}
	return size, nil
}

func applyDelta(base []byte, delta []byte) []byte {
	// the first part of the delta data contains the size of the base object and the size of the target object, both encoded using the variable-length encoding we described in the readSize function.
	_, delta = readSize(delta)
	targetSize, delta := readSize(delta)
	target := make([]byte, 0, targetSize)

	for len(delta) > 0 {
		cmd := delta[0]
		delta = delta[1:]
		// command is 1 byte - first byte tells us whether it's a copy command (if the most significant bit is 1) or an insert command (if the most significant bit is 0). For copy commands, the remaining 7 bits of the command byte indicate which parts of the base object to copy. For insert commands, the remaining 7 bits indicate how many bytes of new data to insert into the target object.
		if cmd&0x80 != 0 {
			// offset -> from where in the base object to copy, size -> how many bytes to copy from the base object.
			var offset, size uint32
			// 0x01 mean 0b00000001, 0x02 means 0b00000010 etc
			// if bit 0 set -> means the next byte in the delta contributes to the offset (least significant 8 bits)
			// if bit 1 set -> means the next byte in the delta contributes to the offset (bit 8 to 15)
			// if bit 2 set -> means the next byte in the delta contributes to the offset (bit 16 to 23)
			// if bit 3 set -> means the next byte in the delta contributes to the offset (most significant bits from bit 24 to 31)
			// similarly for size, if bit 4 set -> means the next byte in the delta contributes to the size (least significant 8 bits)
			// if bit 5 set -> means the next byte in the delta contributes to the size (bit 8 to 15)
			// if bit 6 set -> means the next byte in the delta contributes to the size (bit 16 to 23)
			// if bit 7 set -> means the next byte in the delta contributes to the size (most significant bits from bit 24 to 31)
			if cmd&0x01 != 0 { offset |= uint32(delta[0]); delta = delta[1:] }
			if cmd&0x02 != 0 { offset |= uint32(delta[0]) << 8; delta = delta[1:] }
			if cmd&0x04 != 0 { offset |= uint32(delta[0]) << 16; delta = delta[1:] }
			if cmd&0x08 != 0 { offset |= uint32(delta[0]) << 24; delta = delta[1:] }
			if cmd&0x10 != 0 { size |= uint32(delta[0]); delta = delta[1:] }
			if cmd&0x20 != 0 { size |= uint32(delta[0]) << 8; delta = delta[1:] }
			if cmd&0x40 != 0 { size |= uint32(delta[0]) << 16; delta = delta[1:] }
			if size == 0 { size = 0x10000 }
			target = append(target, base[offset:offset+size]...)
		} else {
			// read the 7 low bits to get the size of the new data to insert, and then read that many bytes from the delta data and append it to the target object.
			size := int(cmd & 0x7f)
			target = append(target, delta[:size]...)
			// progress the delta data by the size of the inserted data to move to the next command in the delta.
			delta = delta[size:]
		}
	}
	return target
}