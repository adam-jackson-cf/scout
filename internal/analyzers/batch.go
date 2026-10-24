package analyzers

import (
	"os"
	"path/filepath"

	"scout/internal/domain"
)

// maxBatchArgBytes bounds the path bytes placed on one scanner command line,
// far below the platform argv limit.
const maxBatchArgBytes = 96 << 10

// Batches splits files, preserving order, into groups that one scanner
// invocation may safely scan under the spec's declared input limits and an
// argv-size cap. Batching is purely an execution detail: Run still accounts
// coverage per file.
func (s Spec) Batches(root string, files []domain.File) [][]domain.File {
	var batches [][]domain.File
	var current []domain.File
	var bytes int64
	var argBytes, manifestBytes int
	for _, file := range files {
		// An unstatable file contributes no scanned bytes: Run rejects it as
		// unreadable with its own incomplete coverage row, so size 0 cannot
		// push a batch past MaxTotalBytes.
		size := int64(0)
		if info, err := os.Stat(filepath.Join(root, file.Path)); err == nil {
			size = info.Size()
		}
		args := 2*len(file.Path) + len(root) + 2
		manifest := 2*len(file.Path) + 320
		if len(current) > 0 && (len(current) >= s.Inputs.MaxFiles ||
			bytes+size > s.Inputs.MaxTotalBytes ||
			argBytes+args > maxBatchArgBytes ||
			manifestBytes+manifest > s.Inputs.MaxManifestBytes/2) {
			batches = append(batches, current)
			current, bytes, argBytes, manifestBytes = nil, 0, 0, 0
		}
		current = append(current, file)
		bytes += size
		argBytes += args
		manifestBytes += manifest
	}
	if len(current) > 0 {
		batches = append(batches, current)
	}
	return batches
}
