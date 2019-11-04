//
// This file is part of arduino-cli.
//
// Copyright 2018 ARDUINO SA (http://www.arduino.cc/)
//
// This software is released under the GNU General Public License version 3,
// which covers the main part of arduino-cli.
// The terms of this license can be found at:
// https://www.gnu.org/licenses/gpl-3.0.en.html
//
// You can be released from the requirements of the above licenses by purchasing
// a commercial license. Buying such a license is mandatory if you want to modify or
// otherwise use the software for commercial activities involving the Arduino
// software without disclosing the source code of your own applications. To purchase
// a commercial license, send an email to license@arduino.cc.
//

package commands

import (
	"fmt"
	"net/http"

	"github.com/arduino/arduino-cli/arduino/cores"
	"github.com/arduino/arduino-cli/arduino/cores/packagemanager"
	rpc "github.com/arduino/arduino-cli/rpc/commands"
)

// DownloadToolRelease downloads a ToolRelease
func DownloadToolRelease(pm *packagemanager.PackageManager, toolRelease *cores.ToolRelease,
	downloadCB DownloadProgressCB, downloaderHeaders http.Header) error {
	resp, err := pm.DownloadToolRelease(toolRelease, downloaderHeaders)
	if err != nil {
		return err
	}
	return Download(resp, toolRelease.String(), downloadCB)
}

// InstallToolRelease installs a ToolRelease
func InstallToolRelease(pm *packagemanager.PackageManager, toolRelease *cores.ToolRelease, taskCB TaskProgressCB) error {
	log := pm.Log.WithField("Tool", toolRelease)

	if toolRelease.IsInstalled() {
		log.Warn("Tool already installed")
		taskCB(&rpc.TaskProgress{Name: "Tool " + toolRelease.String() + " already installed", Completed: true})
		return nil
	}

	log.Info("Installing tool")
	taskCB(&rpc.TaskProgress{Name: "Installing " + toolRelease.String()})
	err := pm.InstallTool(toolRelease)
	if err != nil {
		log.WithError(err).Warn("Cannot install tool")
		return fmt.Errorf("installing tool %s: %s", toolRelease, err)
	}
	log.Info("Tool installed")
	taskCB(&rpc.TaskProgress{Message: toolRelease.String() + " installed", Completed: true})

	return nil
}
