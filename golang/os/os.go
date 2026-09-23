package mos

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type IOS struct {
	Dir    string
	Find   string // command to the inside dir -> /dir
	Target string // folder created after found dir
}

func (o *IOS) clean(parent, clear string) (string, bool) {
	_, chd, f := strings.Cut(parent, clear)
	return chd, f
}

func (o *IOS) SearchAndCreate(filename, hash string) error {
	return filepath.WalkDir(o.Dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}

		chd, ok := o.clean(path, o.Dir)
		if ok {
			log.Printf("[entering %s]", chd)
		}

		cd := filepath.Clean(filepath.Join(o.Dir, o.Find))
		if filepath.Clean(path) != cd {
			return nil
		}

		log.Println("found target dir")
		fnx := filepath.Join(path, filename+"_"+hash)
		if err := os.MkdirAll(fnx, 0755); err != nil {
			return err
		}
		o.Target = fnx // fixed: no more path+fnx smash
		return filepath.SkipAll
	})
}

func (o *IOS) FindAndCreate(filename, hash string) error {
	cd := filepath.Clean(filepath.Join(o.Dir, o.Find))

	info, err := os.Stat(cd)
	if err != nil {
		return fmt.Errorf("[target dir not found: %s: %w]", cd, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("[%s exists but is not a directory]", cd)
	}

	log.Println("found target dir:", cd)
	fnx := filepath.Join(cd, filename+"_"+hash)

	if err := os.MkdirAll(fnx, 0755); err != nil {
		return err
	}
	o.Target = fnx
	return nil
}
