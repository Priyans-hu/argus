package main

import "bytes"

// splitFront keeps a merge file's frontmatter out of the ARGUS:AUTO block.
// When generated starts with a frontmatter block, it is split off both
// inputs, and front is the existing file's block, or the generated one when
// the file has none. Otherwise nothing is split.
func splitFront(existing, generated []byte) (front, existingBody, generatedBody []byte) {
	genFront, genBody := cutFront(generated)
	if genFront == nil {
		return nil, existing, generated
	}
	if oldFront, oldBody := cutFront(existing); oldFront != nil {
		return oldFront, oldBody, genBody
	}
	return genFront, existing, genBody
}

// cutFront splits b after a leading frontmatter block: a "---" first line up
// to and including the next "---" line. front is nil when b has none.
func cutFront(b []byte) (front, rest []byte) {
	lines := bytes.SplitAfter(b, []byte("\n"))
	if !isFence(lines[0]) {
		return nil, b
	}
	n := len(lines[0])
	for _, line := range lines[1:] {
		n += len(line)
		if isFence(line) {
			return b[:n], b[n:]
		}
	}
	return nil, b
}

func isFence(line []byte) bool {
	return string(bytes.TrimRight(line, "\r\n")) == "---"
}
