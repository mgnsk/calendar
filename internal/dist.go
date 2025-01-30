package internal

import (
	"embed"
	"hash/crc32"
	"io/fs"
)

// DistFS contains the bundled assets for web page.
//
//go:embed dist/app.css
var DistFS embed.FS

// Checksums contains checksums for all files in DistFS.
var Checksums = map[string]uint32{}

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

		Checksums[path] = crc32.ChecksumIEEE(b)

		return nil
	}); err != nil {
		panic(err)
	}
}
