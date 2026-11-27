package codeintegrity

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

type ruffSelect struct {
	Select       []string `toml:"select"`
	ExtendSelect []string `toml:"extend-select"`
}

type ruffConfig struct {
	ruffSelect
	Lint ruffSelect `toml:"lint"`
}

type pyproject struct {
	Tool struct {
		Ruff ruffConfig `toml:"ruff"`
	} `toml:"tool"`
}

func docstringsRequired(root string) (bool, error) {
	for _, name := range []string{"ruff.toml", ".ruff.toml"} {
		var cfg ruffConfig
		found, err := decodeIfPresent(filepath.Join(root, name), &cfg)
		if err != nil || found {
			return selectsDocstrings(cfg), err
		}
	}
	var py pyproject
	if _, err := decodeIfPresent(filepath.Join(root, "pyproject.toml"), &py); err != nil {
		return false, err
	}
	return selectsDocstrings(py.Tool.Ruff), nil
}

func decodeIfPresent(file string, v any) (bool, error) {
	data, err := os.ReadFile(file)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if _, err := toml.Decode(string(data), v); err != nil {
		return false, fmt.Errorf("%s: %w", file, err)
	}
	return true, nil
}

func selectsDocstrings(cfg ruffConfig) bool {
	for _, list := range [][]string{cfg.Select, cfg.ExtendSelect, cfg.Lint.Select, cfg.Lint.ExtendSelect} {
		for _, code := range list {
			if code == "D" || code == "ALL" || strings.HasPrefix(code, "D1") {
				return true
			}
		}
	}
	return false
}
