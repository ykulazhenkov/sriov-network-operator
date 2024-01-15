package types

// the file contains public types of the package

import (
	"github.com/coreos/go-systemd/v22/unit"
)

var (
	// Remove run condition form the service
	ConditionOpt = &unit.UnitOption{
		Section: "Unit",
		Name:    "ConditionPathExists",
		Value:   "!/etc/ignition-machine-config-encapsulated.json",
	}
)

// Service contains information about systemd service
type Service struct {
	Name    string
	Path    string
	Content string
}

// ServiceManifestFile service manifest file structure
type ServiceManifestFile struct {
	Name     string
	Contents string
}

// ScriptManifestFile script manifest file structure
type ScriptManifestFile struct {
	Path     string
	Contents struct {
		Inline string
	}
}
