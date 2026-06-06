package network

import (
	"io"
	"net/http"
	"strconv"
	"strings"
)

func discoverRefs(url string) string {
	resp, err := http.Get(url + "/info/refs?service=git-upload-pack")
	if err != nil || resp.StatusCode != 200 { return "" }
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	lines := parsePktLines(body)
	
	for _, line := range lines {
		if strings.Contains(line, "HEAD") {
			parts := strings.Split(line, " ")
			if len(parts) > 0 { return parts[0] }
		}
	}
	return ""
}

func parsePktLines(data []byte) []string {
	var lines []string
	offset := 0
	for offset < len(data) {
		if offset+4 > len(data) { break }
		lengthHex := string(data[offset : offset+4])
		offset += 4
		
		length64, err := strconv.ParseInt(lengthHex, 16, 32)
		if err != nil || length64 == 0 { continue }
		
		length := int(length64)
		if offset+length-4 > len(data) { break }
		
		lines = append(lines, string(data[offset:offset+length-4]))
		offset += length - 4
	}
	return lines
}