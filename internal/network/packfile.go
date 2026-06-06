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
	// 1. Format the want line with the capability
	wantLine := fmt.Sprintf("want %s side-band-64k\n", hash)
	
	// 2. Calculate the exact length of the pkt-line (length of string + 4 bytes for the hex prefix itself)
	hexLen := fmt.Sprintf("%04x", len(wantLine)+4)
	
	// 3. Build the final request body
	reqBody := hexLen + wantLine + "00000009done\n"

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

		// Parse the hex string into an integer
		length64, err := strconv.ParseInt(lengthHex, 16, 32)
		if err != nil {
			// If we hit an error here, the server sent malformed data
			break
		}

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
				// fmt.Fprintf(os.Stderr, "Remote: %s", string(pktData[1:]))
			} else if pktData[0] == 3 {
				// Channel 3: Fatal error messages from the server
				// fmt.Fprintf(os.Stderr, "Fatal: %s", string(pktData[1:]))
			}
		}

		// Move the offset exactly to the start of the next pkt-line
		offset += length
	}

	// We now have the complete, perfectly assembled raw packfile
	unpackPackfile(packData)
}

func unpackPackfile(data []byte) {
	if len(data) < 12 { return }
	numObjects := binary.BigEndian.Uint32(data[8:12])
	offset := 12
	var pendingDeltas []PendingDelta

	for i := 0; i < int(numObjects); i++ {
		c := data[offset]
		offset++
		objType := (c >> 4) & 7
		size := int(c & 15)
		shift := 4
		for c >= 0x80 {
			c = data[offset]
			offset++
			size += int(c&0x7f) << shift
			shift += 7
		}

		var baseHash string
		if objType == 7 {
			baseHash = hex.EncodeToString(data[offset : offset+20])
			offset += 20
		}

		br := bytes.NewReader(data[offset:])
		r, err := zlib.NewReader(br)
		if err != nil { break }
		decompressed, _ := io.ReadAll(r)
		r.Close()

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
			baseType, baseContent, err := core.ReadObjectFull(pd.baseHash)
			if err != nil {
				nextPending = append(nextPending, pd)
				continue
			}
			targetContent := applyDelta(baseContent, pd.deltaData)
			core.WriteObject(baseType, targetContent)
			resolvedCount++
		}
		if resolvedCount == 0 { break }
		pendingDeltas = nextPending
	}
}