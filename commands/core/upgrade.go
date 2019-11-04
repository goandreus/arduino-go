/*
 * This file is part of arduino-cli.
 *
 * Copyright 2018 ARDUINO SA (http://www.arduino.cc/)
 *
 * This software is released under the GNU General Public License version 3,
 * which covers the main part of arduino-cli.
 * The terms of this license can be found at:
 * https://www.gnu.org/licenses/gpl-3.0.en.html
 *
 * You can be released from the requirements of the above licenses by purchasing
 * a commercial license. Buying such a license is mandatory if you want to modify or
 * otherwise use the software for commercial activities involving the Arduino
 * software without disclosing the source code of your own applications. To purchase
 * a commercial license, send an email to license@arduino.cc.
 */

package core

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/arduino/arduino-cli/arduino/cores/packagemanager"
	"github.com/arduino/arduino-cli/commands"
	rpc "github.com/arduino/arduino-cli/rpc/commands"
)

var (
	// ErrAlreadyLatest is returned when an upgrade is not possible because
	// already at latest version.
	ErrAlreadyLatest = errors.New("platform already at latest version")
)

// PlatformUpgrade FIXMEDOC
func PlatformUpgrade(ctx context.Context, req *rpc.PlatformUpgradeReq,
	downloadCB commands.DownloadProgressCB, taskCB commands.TaskProgressCB, downloaderHeaders http.Header) (*rpc.PlatformUpgradeResp, error) {

	pm := commands.GetPackageManager(req.GetInstance().GetId())
	if pm == nil {
		return nil, errors.New("invalid instance")
	}

	// Extract all PlatformReference to platforms that have updates
	ref := &packagemanager.PlatformReference{
		Package:              req.PlatformPackage,
		PlatformArchitecture: req.Architecture,
	}
	if err := upgradePlatform(pm, ref, downloadCB, taskCB, downloaderHeaders); err != nil {
		return nil, err
	}

	if _, err := commands.Rescan(req.GetInstance().GetId()); err != nil {
		return nil, err
	}

	return &rpc.PlatformUpgradeResp{}, nil
}

func upgradePlatform(pm *packagemanager.PackageManager, platformRef *packagemanager.PlatformReference,
	downloadCB commands.DownloadProgressCB, taskCB commands.TaskProgressCB, downloaderHeaders http.Header) error {
	if platformRef.PlatformVersion != nil {
		return fmt.Errorf("upgrade doesn't accept parameters with version")
	}

	// Search the latest version for all specified platforms
	toInstallRefs := []*packagemanager.PlatformReference{}
	platform := pm.FindPlatform(platformRef)
	if platform == nil {
		return fmt.Errorf("platform %s not found", platformRef)
	}
	installed := pm.GetInstalledPlatformRelease(platform)
	if installed == nil {
		return fmt.Errorf("platform %s is not installed", platformRef)
	}
	latest := platform.GetLatestRelease()
	if !latest.Version.GreaterThan(installed.Version) {
		return ErrAlreadyLatest
	}
	platformRef.PlatformVersion = latest.Version
	toInstallRefs = append(toInstallRefs, platformRef)

	for _, platformRef := range toInstallRefs {
		platform, tools, err := pm.FindPlatformReleaseDependencies(platformRef)
		if err != nil {
			return fmt.Errorf("platform %s is not installed", platformRef)
		}
		err = installPlatform(pm, platform, tools, downloadCB, taskCB, downloaderHeaders)
		if err != nil {
			return err
		}
	}
	return nil
}
