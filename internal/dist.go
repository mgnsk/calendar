package internal

import (
	"fmt"
	"hash/crc32"
	"io/fs"
)

// GetAssetLink returns an asset link with checksum or panics if not found.
func GetAssetLink(path string) string {
	sum, ok := checkSums[path]
	if !ok {
		panic(fmt.Sprintf("asset %s not found", path))
	}

	return fmt.Sprintf("/%s?crc=%d", path, sum)
}

// checkSums contains checksums for all files in DistFS.
var checkSums = map[string]uint32{}

func init() {
	if err := fs.WalkDir(DistFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		b, err := DistFS.ReadFile(path)
		if err != nil {
			return err
		}

		checkSums[path] = crc32.ChecksumIEEE(b)

		return nil
	}); err != nil {
		panic(err)
	}
}
