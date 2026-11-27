package proof

import (
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/BurntSushi/toml"
)

type Exception struct {
	Rule       string `toml:"rule"`
	Path       string `toml:"path"`
	Reason     string `toml:"reason"`
	ApprovedBy string `toml:"approved_by"`
	Check      string `toml:"check"`
}

type Scope struct {
	Mode    string   `toml:"mode"`
	BaseRef string   `toml:"base_ref"`
	Paths   []string `toml:"paths"`
}

type File struct {
	Scope      Scope       `toml:"scope"`
	Exceptions []Exception `toml:"exception"`
	raw        map[string]toml.Primitive
	meta       toml.MetaData
}

func Load(file string) (*File, error) {
	data, err := os.ReadFile(file)
	if errors.Is(err, os.ErrNotExist) {
		return &File{}, nil
	}
	if err != nil {
		return nil, err
	}
	var f File
	meta, err := toml.Decode(string(data), &f)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", file, err)
	}
	var raw map[string]toml.Primitive
	if _, err := toml.Decode(string(data), &raw); err != nil {
		return nil, fmt.Errorf("%s: %w", file, err)
	}
	f.raw, f.meta = raw, meta
	return &f, nil
}

func (f *File) Section(name string, v any) (bool, error) {
	p, ok := f.raw[name]
	if !ok {
		return false, nil
	}
	return true, f.meta.PrimitiveDecode(p, v)
}

func (f *File) Excepted(rule, file, check string) bool {
	for _, e := range f.Exceptions {
		if e.Rule != rule || (e.Check != "" && e.Check != check) {
			continue
		}
		if ok, _ := path.Match(e.Path, file); ok || matchPrefix(e.Path, file) {
			return true
		}
	}
	return false
}

func matchPrefix(pattern, file string) bool {
	if len(pattern) < 3 || pattern[len(pattern)-3:] != "/**" {
		return false
	}
	dir := pattern[:len(pattern)-3]
	return len(file) > len(dir) && file[:len(dir)+1] == dir+"/"
}
