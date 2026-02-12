package filesdownload

import (
	"encoding/json"
	"io"
	"net/http"
)

const DEFAULT_VERSION_MANIFEST_URL = "https://piston-meta.mojang.com/mc/game/version_manifest_v2.json"
const FALLBACK_VERSION_MANIFEST_URL = "https://piston-meta.mojang.com/mc/game/version_manifest.json"

type ServerType int

const (
	Vanilla ServerType = iota
	// Fabric
	// Quilt
	// Forge
	// NeoForge
	// Paper
	// Folia
	// Above to be implemented
)

type MojangVersionManifest struct {
	LatestVersions struct {
		Release string
		Snapshot string
	}

	Versions []struct{
		Id string
		VersionType string `json:"type"`
		ReleaseTime string
		Sha1 string
	}
}

func (manifest MojangVersionManifest) Populate(url string) error {
	manifestData, err := http.Get(url)
	if err != nil { return err }
	
	defer manifestData.Body.Close()

	manifestBody, err := io.ReadAll(manifestData.Body)
	if err != nil { return err }
	
	err = json.Unmarshal(manifestBody, &manifest)
	if err != nil { return err }
	
	return nil
}

func DownloadServer(path string, serverType ServerType, version string) error {
	return nil
}
