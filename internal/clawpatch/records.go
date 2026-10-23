package clawpatch

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"scout/internal/domain"
)

// StaleStatus is the status Clawpatch writes on every record the latest map
// did not emit; it never deletes records (third_party/clawpatch/src/app.ts:139-147).
const StaleStatus = "skipped"

// LoadFeatures imports the active schema-v1 records written by the pinned
// mapper: the latest map's output, with stale (`skipped`) records excluded so
// ownership never reads them (design §4). RawOriginal retains the exact
// canonical record bytes.
func LoadFeatures(stateDir string) ([]domain.Feature, error) {
	return loadActiveFeatures(stateDir)
}

func loadActiveFeatures(stateDir string) ([]domain.Feature, error) {
	all, err := loadFeatures(stateDir)
	if err != nil {
		return nil, err
	}
	active := all[:0]
	for _, feature := range all {
		if feature.Status != StaleStatus {
			active = append(active, feature)
		}
	}
	return active, nil
}

func loadFeatures(stateDir string) ([]domain.Feature, error) {
	featuresDir := filepath.Join(stateDir, "features")
	entries, err := os.ReadDir(featuresDir)
	if err != nil {
		if errorsIsNotExist(err) {
			return []domain.Feature{}, nil
		}
		return nil, fmt.Errorf("read Clawpatch feature records: %w", err)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	features := make([]domain.Feature, 0, len(entries))
	seen := make(map[string]struct{}, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			return nil, fmt.Errorf("inspect Clawpatch feature record %q: %w", entry.Name(), err)
		}
		if !info.Mode().IsRegular() {
			return nil, fmt.Errorf("Clawpatch feature record is not a regular file: %q", entry.Name())
		}
		raw, err := os.ReadFile(filepath.Join(featuresDir, entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("read Clawpatch feature record %q: %w", entry.Name(), err)
		}
		feature, err := decodeFeature(raw)
		if err != nil {
			return nil, fmt.Errorf("decode Clawpatch feature record %q: %w", entry.Name(), err)
		}
		if _, duplicate := seen[feature.FeatureID]; duplicate {
			return nil, fmt.Errorf("duplicate canonical Clawpatch feature ID %q", feature.FeatureID)
		}
		seen[feature.FeatureID] = struct{}{}
		features = append(features, feature)
	}
	return features, nil
}

// DecodeFeature validates one schema-v1 canonical record without reconstructing
// its ID or replacing its source JSON.
func DecodeFeature(raw []byte) (domain.Feature, error) {
	return decodeFeature(raw)
}

func decodeFeature(raw []byte) (domain.Feature, error) {
	if !json.Valid(raw) {
		return domain.Feature{}, fmt.Errorf("record is not valid JSON")
	}
	var feature domain.Feature
	if err := json.Unmarshal(raw, &feature); err != nil {
		return domain.Feature{}, err
	}
	feature.RawOriginal = append(json.RawMessage(nil), raw...)
	if err := feature.Validate(); err != nil {
		return domain.Feature{}, err
	}
	return feature, nil
}

func errorsIsNotExist(err error) bool {
	return errors.Is(err, fs.ErrNotExist)
}
