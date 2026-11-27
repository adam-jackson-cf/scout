package clawpatch

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
)

// ScoutExcludePattern keeps Scout's own state (.scout/: SQLite, artifacts and
// prior Clawpatch feature JSONs) out of Clawpatch's inventory. The agent
// mapper lists files with `git ls-files --cached --others --exclude-standard`
// and its skip list does not cover .scout/, so without this an adopter that
// does not gitignore .scout/ feeds Scout's state into the agent-map prompt.
// Clawpatch applies the config `exclude` globs to that inventory.
const ScoutExcludePattern = ".scout/**"

// ensureScoutExcluded makes the Scout-owned Clawpatch state config exclude
// .scout/**. The config sits in the state directory, which Clawpatch always
// discovers first because Scout passes --state-dir. Unknown config fields are
// preserved. External (adopter-owned) state is never rewritten.
func (c *Client) ensureScoutExcluded() error {
	if c.externalState {
		return nil
	}
	path := filepath.Join(c.stateDir, "config.json")
	config, err := readStateConfig(path)
	if err != nil {
		return err
	}
	exclude, err := stateConfigExclude(config)
	if err != nil {
		return err
	}
	if slices.Contains(exclude, ScoutExcludePattern) {
		return nil
	}
	return writeStateConfigExclude(path, config, append(exclude, ScoutExcludePattern))
}

func readStateConfig(path string) (map[string]json.RawMessage, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, errors.New("clawpatch state config is missing; cannot exclude .scout/ from the inventory")
		}
		return nil, fmt.Errorf("read Clawpatch state config: %w", err)
	}
	var config map[string]json.RawMessage
	if err := json.Unmarshal(raw, &config); err != nil {
		return nil, fmt.Errorf("parse Clawpatch state config: %w", err)
	}
	return config, nil
}

func stateConfigExclude(config map[string]json.RawMessage) ([]string, error) {
	var exclude []string
	if rawExclude, ok := config["exclude"]; ok {
		if err := json.Unmarshal(rawExclude, &exclude); err != nil {
			return nil, fmt.Errorf("parse Clawpatch state config exclude: %w", err)
		}
	}
	return exclude, nil
}

func writeStateConfigExclude(path string, config map[string]json.RawMessage, exclude []string) error {
	encoded, err := json.Marshal(exclude)
	if err != nil {
		return err
	}
	config["exclude"] = encoded
	updated, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	temporary := path + ".tmp"
	if err := os.WriteFile(temporary, append(updated, '\n'), 0o600); err != nil {
		return fmt.Errorf("write Clawpatch state config: %w", err)
	}
	if err := os.Rename(temporary, path); err != nil {
		return fmt.Errorf("replace Clawpatch state config: %w", err)
	}
	return nil
}
