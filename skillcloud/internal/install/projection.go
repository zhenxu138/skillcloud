package install

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const ProjectionManifestFile = ".skillcloud-projection.json"

type ProjectionManifest struct {
	SourceID     string `json:"source_id"`
	Target       string `json:"target"`
	Scope        string `json:"scope"`
	Mode         string `json:"mode"`
	SourceCommit string `json:"source_commit,omitempty"`
}

func projectionManifestPath(dir string) string {
	info, err := os.Lstat(dir)
	if err == nil && info.Mode()&os.ModeSymlink != 0 {
		base := filepath.Base(dir)
		return filepath.Join(filepath.Dir(dir), "."+base+ProjectionManifestFile)
	}
	return filepath.Join(dir, ProjectionManifestFile)
}

func WriteProjectionManifest(dir string, manifest ProjectionManifest) error {
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(projectionManifestPath(dir), data, 0o644)
}

func ReadProjectionManifest(dir string) (ProjectionManifest, error) {
	data, err := os.ReadFile(projectionManifestPath(dir))
	if err != nil {
		return ProjectionManifest{}, err
	}
	var manifest ProjectionManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return ProjectionManifest{}, err
	}
	return manifest, nil
}

func CanRemoveProjection(dir string, expected ProjectionManifest) error {
	manifest, err := ReadProjectionManifest(dir)
	if err != nil {
		return fmt.Errorf("%s is not managed by skillcloud", dir)
	}
	if manifest.SourceID != expected.SourceID || manifest.Target != expected.Target || manifest.Scope != expected.Scope {
		return fmt.Errorf("%s projection manifest mismatch", dir)
	}
	return nil
}

func RemoveProjection(dir string) error {
	manifestPath := projectionManifestPath(dir)
	if err := os.RemoveAll(dir); err != nil {
		return err
	}
	if err := os.Remove(manifestPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
