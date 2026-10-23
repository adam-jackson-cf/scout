package clawpatch

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
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
	raw, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return errors.New("Clawpatch state config is missing; cannot exclude .scout/ from the inventory")
		}
		return fmt.Errorf("read Clawpatch state config: %w", err)
	}
	var config map[string]json.RawMessage
	if err := json.Unmarshal(raw, &config); err != nil {
		return fmt.Errorf("parse Clawpatch state config: %w", err)
	}
	var exclude []string
	if rawExclude, ok := config["exclude"]; ok {
		if err := json.Unmarshal(rawExclude, &exclude); err != nil {
			return fmt.Errorf("parse Clawpatch state config exclude: %w", err)
		}
	}
	for _, pattern := range exclude {
		if pattern == ScoutExcludePattern {
			return nil
		}
	}
	encoded, err := json.Marshal(append(exclude, ScoutExcludePattern))
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
