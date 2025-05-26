// Command httprr-convert provides utilities for converting httprr files between
// compressed and uncompressed formats.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/tmc/langchaingo/internal/httprr"
)

func main() {
	var (
		compress   = flag.Bool("compress", false, "compress uncompressed .httprr files to .httprr.gz")
		decompress = flag.Bool("decompress", false, "decompress .httprr.gz files to .httprr")
		dir        = flag.String("dir", ".", "directory to process (default: current directory)")
		recursive  = flag.Bool("recursive", false, "process directories recursively")
	)
	flag.Parse()

	// Validate flags
	if *compress && *decompress {
		log.Fatal("Cannot specify both -compress and -decompress")
	}
	if !*compress && !*decompress {
		log.Fatal("Must specify either -compress or -decompress")
	}

	// Process directory
	if err := processDirectory(*dir, *compress, *recursive); err != nil {
		log.Fatalf("Error processing directory: %v", err)
	}

	action := "compressed"
	if *decompress {
		action = "decompressed"
	}
	fmt.Printf("Successfully %s httprr files in %s\n", action, *dir)
}

func processDirectory(dir string, compress, recursive bool) error {
	if recursive {
		return filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			return processFile(path, compress)
		})
	}

	return httprr.ConvertAllInDir(dir, compress)
}

func processFile(path string, compress bool) error {
	if compress {
		if filepath.Ext(path) == ".httprr" && filepath.Ext(filepath.Base(path)) != ".gz" {
			return httprr.CompressFile(path)
		}
	} else {
		if filepath.Ext(path) == ".gz" && filepath.Ext(filepath.Base(path)[:len(filepath.Base(path))-3]) == ".httprr" {
			return httprr.DecompressFile(path)
		}
	}
	return nil
}
