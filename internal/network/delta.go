package network

type PendingDelta struct {
	baseHash  string
	deltaData []byte
}

func readSize(data []byte) (int, []byte) {
	size := 0
	shift := 0
	for i, b := range data {
		size |= int(b&0x7f) << shift
		shift += 7
		if b&0x80 == 0 {
			return size, data[i+1:]
		}
	}
	return size, nil
}

func applyDelta(base []byte, delta []byte) []byte {
	_, delta = readSize(delta)
	targetSize, delta := readSize(delta)
	target := make([]byte, 0, targetSize)

	for len(delta) > 0 {
		cmd := delta[0]
		delta = delta[1:]
		if cmd&0x80 != 0 {
			var offset, size uint32
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
			size := int(cmd & 0x7f)
			target = append(target, delta[:size]...)
			delta = delta[size:]
		}
	}
	return target
}