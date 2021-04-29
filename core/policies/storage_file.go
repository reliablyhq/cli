package policies

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/karrick/godirwalk"
)

const (
	fileExt = ".rego"
)

// fsstore represents a structure for storage of policies on local file system
type fsstore struct {
	// File System store
	basedir string // path of the base directory on local file system
}

// NewFSStore returns an empty store based on local file system
func NewFSStore(path string) Store {
	return &fsstore{
		basedir: path,
	}
}

func (fs *fsstore) ListPolicies() ([]string, error) {
	var names = make([]string, 0)
	_ = godirwalk.Walk(fs.basedir, &godirwalk.Options{
		Callback: func(osPathname string, de *godirwalk.Dirent) error {
			if strings.HasSuffix(osPathname, fileExt) {
				// when listing policies names,
				// we do not want to see the file extension
				// and the base directory prefix
				name := strings.TrimSuffix(osPathname, fileExt)
				name = strings.TrimPrefix(name, fs.basedir)
				if strings.HasPrefix(name, "/") {
					name = strings.TrimPrefix(name, "/")
				}
				names = append(names, name)
			}
			return nil
		},
		Unsorted: true,
	})

	return names, nil
}

func (fs *fsstore) GetPolicy(id string) ([]byte, error) {
	fpath := fs.PolicyPath(id)
	bs, err := os.ReadFile(fpath)
	if err != nil {
		return nil, ErrPolicyNotFound
	}
	return bs, nil
}

func (fs *fsstore) UpsertPolicy(id string, bs []byte) error {

	fpath := fs.PolicyPath(id)

	fpathParts := strings.Split(fpath, "/")
	if len(fpathParts) > 1 {
		parts := fpathParts[:len(fpathParts)-1]
		subfolders := filepath.Join(parts...)
		if strings.HasPrefix(fpath, "/") && !strings.HasPrefix(subfolders, "/") {
			subfolders = fmt.Sprintf("/%s", subfolders)
		}
		_ = os.MkdirAll(subfolders, 0700) // ensure to create sub-folders if not exist yet
	}

	err := os.WriteFile(fpath, bs, 0644)
	return err
}

func (fs *fsstore) HasPolicy(id string) bool {
	fpath := fs.PolicyPath(id)

	_, err := os.Stat(fpath)
	exists := !os.IsNotExist(err)
	return exists
}

func (fs *fsstore) PolicyPath(id string) string {
	fileWithExt := strings.ToLower(fmt.Sprintf("%s%s", id, fileExt))
	return filepath.Join(fs.basedir, fileWithExt)
}
