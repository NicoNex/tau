package vm

import (
	"fmt"
	"os"
	"path/filepath"
)

func lookup(taupath string) (string, error) {
	var paths []string
	taupath = filepath.Clean(taupath)

	if ext := filepath.Ext(taupath); ext != "" {
		paths = []string{taupath, filepath.Join("/", "lib", "tau", taupath)}
	} else {
		paths = []string{
			taupath + ".tau",
			taupath + ".tauc",
			filepath.Join("/", "lib", "tau", taupath+".tau"),
			filepath.Join("/", "lib", "tau", taupath+".tauc"),
		}
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	return "", fmt.Errorf("no module named %q", taupath)
}
