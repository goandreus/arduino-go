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

package lib

import (
	"fmt"
	"net/http"

	"github.com/arduino/arduino-cli/arduino/libraries/librariesmanager"
	"github.com/arduino/arduino-cli/commands"
)

// LibraryUpgradeAll upgrades all the available libraries
func LibraryUpgradeAll(instanceID int32, downloadCB commands.DownloadProgressCB,
	taskCB commands.TaskProgressCB, headers http.Header) error {
	// get the library manager
	lm := commands.GetLibraryManager(instanceID)

	if err := upgrade(lm, listLibraries(lm, true, true), downloadCB, taskCB, headers); err != nil {
		return err
	}

	if _, err := commands.Rescan(instanceID); err != nil {
		return fmt.Errorf("rescanning libraries: %s", err)
	}

	return nil
}

// LibraryUpgrade upgrades only the given libraries
func LibraryUpgrade(instanceID int32, libraryNames []string, downloadCB commands.DownloadProgressCB,
	taskCB commands.TaskProgressCB, headers http.Header) error {
	// get the library manager
	lm := commands.GetLibraryManager(instanceID)

	// get the libs to upgrade
	libs := filterByName(listLibraries(lm, true, true), libraryNames)

	// do it
	return upgrade(lm, libs, downloadCB, taskCB, headers)
}

func upgrade(lm *librariesmanager.LibrariesManager, libs []*installedLib, downloadCB commands.DownloadProgressCB,
	taskCB commands.TaskProgressCB, downloaderHeaders http.Header) error {

	// Go through the list and download them

	for _, lib := range libs {
		if err := downloadLibrary(lm, lib.Available, downloadCB, taskCB, downloaderHeaders); err != nil {
			return err
		}
	}

	// Go through the list and install them
	for _, lib := range libs {
		if err := installLibrary(lm, lib.Available, taskCB); err != nil {
			return err
		}
	}

	return nil
}

func filterByName(libs []*installedLib, names []string) []*installedLib {
	// put the names in a map to ease lookup
	queryMap := make(map[string]struct{})
	for _, name := range names {
		queryMap[name] = struct{}{}
	}

	ret := []*installedLib{}
	for _, lib := range libs {
		// skip if library name wasn't in the query
		if _, found := queryMap[lib.Library.Name]; found {
			ret = append(ret, lib)
		}
	}

	return ret
}
