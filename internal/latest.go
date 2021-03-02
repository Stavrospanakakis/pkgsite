// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package internal

import (
	"fmt"

	"golang.org/x/mod/modfile"
	"golang.org/x/mod/semver"
)

// LatestModuleVersions describes the latest versions of a module. It also holds the
// go.mod file of the raw latest version, which establishes whether the module
// is deprecated, and what versions are retracted.
type LatestModuleVersions struct {
	ModulePath         string
	RawVersion         string        // ignoring retractions
	CookedVersion      string        // considering retractions
	GoodVersion        string        // successfully processed
	GoModFile          *modfile.File // of raw
	deprecated         bool
	deprecationComment string
}

func NewLatestModuleVersions(modulePath, raw, cooked, good string, modBytes []byte) (*LatestModuleVersions, error) {
	modFile, err := modfile.ParseLax(fmt.Sprintf("%s@%s/go.mod", modulePath, raw), modBytes, nil)
	if err != nil {
		return nil, err
	}

	dep, comment := isDeprecated(modFile)
	return &LatestModuleVersions{
		ModulePath:         modulePath,
		RawVersion:         raw,
		CookedVersion:      cooked,
		GoodVersion:        good,
		GoModFile:          modFile,
		deprecated:         dep,
		deprecationComment: comment,
	}, nil
}

// PopulateModuleInfo uses the LatestModuleVersions to populate fields of the given module.
func (li *LatestModuleVersions) PopulateModuleInfo(mi *ModuleInfo) {
	mi.Deprecated = li.deprecated
	mi.DeprecationComment = li.deprecationComment
	mi.Retracted, mi.RetractionRationale = IsRetracted(li.GoModFile, mi.Version)
}

// IsRetracted reports whether the go.mod file retracts the version.
// If so, it returns true along with the rationale for the retraction.
func IsRetracted(mf *modfile.File, resolvedVersion string) (bool, string) {
	for _, r := range mf.Retract {
		if semver.Compare(resolvedVersion, r.Low) >= 0 && semver.Compare(resolvedVersion, r.High) <= 0 {
			return true, r.Rationale
		}
	}
	return false, ""
}
