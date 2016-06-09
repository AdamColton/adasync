// +build linux

package collection

import (
	"github.com/adamcolton/err"
	"os"
	"path/filepath"
	"strings"
)

var DefaultIgnore = ""

func osIgnore(path string) bool {
	return false
}

// full returns all drives
func full() []string {
	return []string{"/home/"}
}

// quick returns only new drives that have appeared since last scan
func quick() []string {
	return findMedia([]string{"/media/"})
}

func findMedia(paths []string) []string {
	lp := linuxPaths{}
	for _, path := range paths {
		filepath.Walk(string(path), lp.getMedia)
	}
	media = lp.all
	return lp.nu
}

type linuxPaths struct {
	all []string
	nu  []string
}

var media = []string{}

func (lp *linuxPaths) getMedia(subPath string, _ os.FileInfo, _ error) error {
	segments := strings.Split(subPath, "/")
	if len(segments) == 4 {
		if stat, e := fs.Stat(subPath); err.Log(e) && stat.IsDir() {
			lp.all = append(lp.all, subPath)
			isNew := true
			for _, path := range media {
				if path == subPath {
					isNew = false
					break
				}
			}
			if isNew {
				lp.nu = append(lp.nu, subPath)
			}
			return filepath.SkipDir
		}
	}
	return nil
}
