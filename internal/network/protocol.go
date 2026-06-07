package network

import (
	"io"
	"net/http"
	"strconv"
	"strings"
)

func discoverRefs(url string) string {
	// send a req to get the refs from the remote repository. The URL for this is typically in the format of "http://<remote-repo-url>/info/refs?service=git-upload-pack". This endpoint returns a list of refs in the remote repository along with their corresponding hashes. We need to parse this response to find the hash of the HEAD ref, which points to the latest commit in the default branch (usually main).
	// other refs will be in the format of <hash> refs/heads/<branch-name> for branches and <hash> refs/tags/<tag-name> for tags. These <hash> contain the hash of the latest commit for branches and the hash of the tag object for tags.
	resp, err := http.Get(url + "/info/refs?service=git-upload-pack")
	if err != nil || resp.StatusCode != 200 { return "" }
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	lines := parsePktLines(body)
	
	for _, line := range lines {
		if strings.Contains(line, "HEAD") {
			// format of this line : <hash> HEAD ref:refs/heads/<branch-name>
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