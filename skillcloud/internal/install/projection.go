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

func WriteProjectionManifest(dir string, manifest ProjectionManifest) error {
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(filepath.Join(dir, ProjectionManifestFile), data, 0o644)
}

func ReadProjectionManifest(dir string) (ProjectionManifest, error) {
	data, err := os.ReadFile(filepath.Join(dir, ProjectionManifestFile))
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
