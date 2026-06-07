package network

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"git_cmd_impl_in_go/internal/core"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func fetchAndUnpack(url string, hash string) {
	// 1. Format the want line with the capability - we want the packfile data to be sent in side-band-64k format which allows the server to interleave progress messages and error messages with the packfile data. The want line should be in the format of "want <hash> side-band-64k\n" where <hash> is the hash of the commit we want to fetch from the remote repository.
	wantLine := fmt.Sprintf("want %s side-band-64k\n", hash)
	
	// 2. Calculate the exact length of the pkt-line (length of string + 4 bytes for the hex prefix itself)
	// say wantLine is 30 bytes long, then +4 = 34 bytes total, which is 0x22 in hex. %04x means we need to represent this length as a 4-byte hex string which will be like "0022" - extra 0 are added to make it 4 bytes long
	// Hence hexLen itself will be 4 bytes long (0022 in the example above) hence we add 4
	hexLen := fmt.Sprintf("%04x", len(wantLine)+4)
	
	// 3. Build the final request body
	// 0000 - flush packet to indicate the end of the want lines, and then "done\n" to indicate that we have finished sending our request. done\n is 5 bytes long, so we add 4 to get the total length of the pkt-line which will be 9 bytes (0009 in hex). Thats why we added "0009" before "done\n" in the request body.
	reqBody := hexLen + wantLine + "00000009done\n"

	// send post request to the remote repository's git-upload-pack endpoint with the request body we constructed. The URL for this endpoint is typically in the format of "http://<remote-repo-url>/git-upload-pack". We also need to set the Content-Type header to "application/x-git-upload-pack-request" to indicate that this is a Git upload-pack request.
	req, _ := http.NewRequest("POST", url+"/git-upload-pack", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/x-git-upload-pack-request")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching packfile: %s\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	packData := []byte{}
	offset := 0

	for offset < len(body) {
		if offset+4 > len(body) {
			break
		}

		// Read the next 4 bytes which will ALWAYS be a valid hex length now
		lengthHex := string(body[offset : offset+4])

		// Parse the hex string into an integer of base 16 and put it in a 64-bit integer. 32 here is just to check that the length can fit in 32 bits or not.
		// But parseInt still returns a 64-bit integer because that's the default size for integers in Go
		length64, err := strconv.ParseInt(lengthHex, 16, 32)
		if err != nil {
			// If we hit an error here, the server sent malformed data
			break
		}

		// Convert the length to an int
		// If length is 0, it's a flush packet (0000) which indicates the end of a section. We can skip it and continue to the next pkt-line.
		length := int(length64)
		if length == 0 { 
			// A flush packet (0000). Skip the 4 length bytes and continue.
			offset += 4
			continue
		}

		// Safety check to prevent out-of-bounds panics
		if offset+length > len(body) {
			break
		}

		// Extract the payload (ignoring the 4-byte header)
		pktData := body[offset+4 : offset+length]

		// Because we requested side-band-64k, the first byte of every payload 
		// is the channel identifier.
		if len(pktData) > 0 {
			if pktData[0] == 1 {
				// Channel 1: Packfile data. Trim the channel byte and append.
				packData = append(packData, pktData[1:]...)
			} else if pktData[0] == 2 {
				// Channel 2: Progress messages (e.g., "Compressing objects: 100%")
				fmt.Fprintf(os.Stderr, "Remote: %s", string(pktData[1:]))
			} else if pktData[0] == 3 {
				// Channel 3: Fatal error messages from the server
				fmt.Fprintf(os.Stderr, "Fatal: %s", string(pktData[1:]))
			}
		}

		// Move the offset exactly to the start of the next pkt-line
		offset += length
	}

	// We now have the complete, perfectly assembled raw packfile
	unpackPackfile(packData)
}

func unpackPackfile(data []byte) {
	// The packfile format starts with a 4-byte signature "PACK", followed by a 4-byte version number (which should be 2), and then a 4-byte big-endian integer indicating the number of objects in the packfile.
	if len(data) < 12 { return }
	// BigEndian means the most significant byte is at the smallest memory address. So when we read the 4 bytes for the number of objects, we need to interpret them as a big-endian integer to get the correct count of objects in the packfile.
	// Uint32 means we are reading an unsigned 32-bit integer, which can represent values from 0 to 4,294,967,295. This is sufficient for the number of objects in a packfile.
	// we are joining the 4 bytes from data[8] to data[11] (inclusive) to get the number of objects in the packfile to get the final count in 32-bit integer form. 
	// its like variable size encoding joining but here we already know that the number of objects is represented in exactly 4 bytes, so we can directly read those 4 bytes and convert them to an integer using BigEndian encoding.
	numObjects := binary.BigEndian.Uint32(data[8:12])
	offset := 12
	var pendingDeltas []PendingDelta

	for i := 0; i < int(numObjects); i++ {
		c := data[offset]
		// The first byte of the object entry contains both the object type and part of the size of the compressed data. 
		// format [ M | T | T | T | S | S | S | S ]
		// M is the most significant bit which indicates whether there are more bytes to read for the size (if M is 1, we need to read more bytes to get the full size; if M is 0, this byte alone gives us the full size).
		// TTT is the 3 bits that indicate the object type (e.g., commit, tree, blob, or delta).
		// SSSS are the 4 least significant bits that contribute to the size of the compressed data for this object.
		// if M is 1 then we keep reading the next byte and adding its 7 least significant bits to the size until we encounter a byte where M is 0.
		offset++
		objType := (c >> 4) & 7
		// 7 means 0b00000111 we shift right by 4 to get the 3 bits for the type and then mask with 7 to get the value of those 3 bits. 
		size := int(c & 15)
		shift := 4
		for c >= 0x80 {
			// 0x80 means 0b10000000, which checks if the most significant bit of the byte is set to 1. If it's 1, it means we need to read another byte to get more bits for the size. We keep reading bytes until we encounter a byte where the most significant bit is 0, which indicates that we have read all the bytes for the size for this object.
			c = data[offset]
			offset++
			size += int(c&0x7f) << shift
			// 0x7f means 0b01111111, which masks out the most significant bit of the byte to get the 7 bits that contribute to the size. We then shift these bits left by the appropriate amount (based on how many bytes we have read so far) and add them to the total size.
			shift += 7
		}

		var baseHash string
		// if the object type is 7, it means it's a delta object that depends on a base object. The packfile will include the hash of the base object
		// read the 20-byte hash of the base object from the packfile data and move the offset accordingly to point to the start of the compressed delta data for this object.
		if objType == 7 {
			baseHash = hex.EncodeToString(data[offset : offset+20])
			offset += 20
		}

		// read the compressed data for the object, which starts at the current offset and continues until we have read the entire compressed data for that object
		// Zlib will read till the end of the compressed data for the current object, and it will return the decompressed data.
		br := bytes.NewReader(data[offset:])
		r, err := zlib.NewReader(br)
		if err != nil { break }
		// decompressed data in bytes 
		decompressed, _ := io.ReadAll(r)
		r.Close()

		// br moved forward by the number of bytes read from the compressed data when zlib decompressed it and 
		consumed := len(data[offset:]) - br.Len()
		offset += consumed

		switch objType {
		case 1: core.WriteObject("commit", decompressed)
		case 2: core.WriteObject("tree", decompressed)
		case 3: core.WriteObject("blob", decompressed)
		case 7: pendingDeltas = append(pendingDeltas, PendingDelta{baseHash, decompressed})
		}
	}

	for len(pendingDeltas) > 0 {
		var nextPending []PendingDelta
		resolvedCount := 0
		for _, pd := range pendingDeltas {
			// get base type and base content in bytes
			baseType, baseContent, err := core.ReadObjectFull(pd.baseHash)
			if err != nil {
				// may be the base object is also a delta object which we have not resolved yet, so we will keep this pending delta in the nextPending list to try resolving it in the next iteration after we have resolved some of the other deltas.
				nextPending = append(nextPending, pd)
				continue
			}
			targetContent := applyDelta(baseContent, pd.deltaData)
			core.WriteObject(baseType, targetContent)
			resolvedCount++
		}
		// If in an entire pass we couldn't resolve any pending delta, it means there is a problem with the packfile (e.g., missing base objects), and we should break to avoid an infinite loop 
		if resolvedCount == 0 { break }
		pendingDeltas = nextPending
	}
}