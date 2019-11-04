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

package libraries

import (
	"sort"
)

// List is a list of Libraries
type List []*Library

// Contains check if a lib is contained in the list
func (list *List) Contains(lib *Library) bool {
	for _, l := range *list {
		if l == lib {
			return true
		}
	}
	return false
}

// Add appends all libraries passed as parameter in the list
func (list *List) Add(libs ...*Library) {
	for _, lib := range libs {
		*list = append(*list, lib)
	}
}

// FindByName returns the first library in the list that match
// the specified name or nil if not found
func (list *List) FindByName(name string) *Library {
	for _, lib := range *list {
		if lib.Name == name {
			return lib
		}
	}
	return nil
}

// SortByArchitecturePriority sorts the libraries in descending order using
// the Arduino lib priority ordering (the first has the higher priority)
func (list *List) SortByArchitecturePriority(arch string) {
	sort.Slice(*list, func(i, j int) bool {
		a, b := (*list)[i], (*list)[j]
		return a.PriorityForArchitecture(arch) > b.PriorityForArchitecture(arch)
	})
}

/*
// HasHigherPriority returns true if library x has higher priority compared to library
// y for the given header and architecture.
func HasHigherPriority(libX, libY *Library, header string, arch string) bool {
	//return computePriority(libX, header, arch) > computePriority(libY, header, arch)
	header = strings.TrimSuffix(header, filepath.Ext(header))

	simplify := func(name string) string {
		name = utils.SanitizeName(name)
		name = strings.ToLower(name)
		return name
	}
	header = simplify(header)
	nameX := simplify(libX.Name)
	nameY := simplify(libY.Name)

	compareLocations := func() bool {
		// XXX: priority inversion case.
		if libX.Location < libY.Location {
			return true
		}
		return false
	}

	checks := []func(name, header string) bool{
		func(name, header string) bool { return name == header },
		func(name, header string) bool { return name == header+"-master" },
		strings.HasPrefix,
		strings.HasSuffix,
		strings.Contains,
	}
	// Run all checks to sort priorities based on library name
	// If both library match the same name check, then fallback to
	// compare locations
	for _, check := range checks {
		x := check(nameX, header)
		y := check(nameY, header)
		if x && y {
			return compareLocations()
		}
		if x {
			return true
		}
		if y {
			return false
		}
	}

	return compareLocations()
}
*/
