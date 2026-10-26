package budget

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// CharsPerToken is the rough English/code average used for estimates.
const CharsPerToken = 4

// CountTokens estimates the tokens in text.
func CountTokens(text []byte) int {
	return (len(text) + CharsPerToken - 1) / CharsPerToken
}

// FileCount is the estimated size of one file.
type FileCount struct {
	Path   string
	Tokens int
}

// CountPaths walks files and directories and estimates tokens per text file.
// It skips hidden directories (like .git) and files that look binary.
func CountPaths(paths []string) ([]FileCount, error) {
	var out []FileCount
	for _, root := range paths {
		err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				if p != root && strings.HasPrefix(d.Name(), ".") {
					return filepath.SkipDir
				}
				return nil
			}
			data, err := os.ReadFile(p)
			if err != nil {
				return err
			}
			if bytes.IndexByte(data, 0) >= 0 {
				return nil
			}
			out = append(out, FileCount{Path: p, Tokens: CountTokens(data)})
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}
