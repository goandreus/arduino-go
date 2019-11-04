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

package cores

import (
	"fmt"
	"strings"

	"github.com/arduino/go-properties-orderedmap"
)

// Board represents a board loaded from an installed platform
type Board struct {
	BoardID         string
	Properties      *properties.Map  `json:"-"`
	PlatformRelease *PlatformRelease `json:"-"`
}

// HasUsbID returns true if the board match the usb vid and pid parameters
func (b *Board) HasUsbID(reqVid, reqPid string) bool {
	vids := b.Properties.SubTree("vid")
	pids := b.Properties.SubTree("pid")
	for id, vid := range vids.AsMap() {
		if pid, ok := pids.GetOk(id); ok {
			if strings.EqualFold(vid, reqVid) && strings.EqualFold(pid, reqPid) {
				return true
			}
		}
	}
	return false
}

// Name returns the board name as defined in boards.txt properties
func (b *Board) Name() string {
	return b.Properties.Get("name")
}

// FQBN return the Fully-Qualified-Board-Name for the default configuration of this board
func (b *Board) FQBN() string {
	platform := b.PlatformRelease.Platform
	return platform.Package.Name + ":" + platform.Architecture + ":" + b.BoardID
}

func (b *Board) String() string {
	return b.FQBN()
}

// GetConfigOptions returns an OrderedMap of configuration options for this board.
// The returned map will have key and value as option id and option name, respectively.
func (b *Board) GetConfigOptions() *properties.Map {
	res := properties.NewMap()
	menu := b.Properties.SubTree("menu")
	for _, option := range menu.FirstLevelKeys() {
		res.Set(option, b.PlatformRelease.Menus.Get(option))
	}
	return res
}

// GetConfigOptionValues returns an OrderedMap of possible values for a specific configuratio options
// for this board. The returned map will have key and value as option value and option value name,
// respectively.
func (b *Board) GetConfigOptionValues(option string) *properties.Map {
	res := properties.NewMap()
	menu := b.Properties.SubTree("menu").SubTree(option)
	for _, value := range menu.FirstLevelKeys() {
		if label, ok := menu.GetOk(value); ok {
			res.Set(value, label)
		}
	}
	return res
}

// GetBuildProperties returns the build properties and the build
// platform for the Board with the configuration passed as parameter.
func (b *Board) GetBuildProperties(userConfigs *properties.Map) (*properties.Map, error) {
	// Clone user configs because they are destroyed during iteration
	userConfigs = userConfigs.Clone()

	// Start with board's base properties
	buildProperties := b.Properties.Clone()

	// Add all sub-configurations one by one (a config is: option=value)
	menu := b.Properties.SubTree("menu")
	for _, option := range menu.FirstLevelKeys() {
		optionMenu := menu.SubTree(option)
		userValue, haveUserValue := userConfigs.GetOk(option)
		if haveUserValue {
			userConfigs.Remove(option)
			if !optionMenu.ContainsKey(userValue) {
				return nil, fmt.Errorf("invalid value '%s' for option '%s'", userValue, option)
			}
		} else {
			// apply default
			userValue = optionMenu.FirstLevelKeys()[0]
		}

		optionsConf := optionMenu.SubTree(userValue)
		buildProperties.Merge(optionsConf)
	}

	// Check for residual invalid options...
	for _, invalidOption := range userConfigs.Keys() {
		if invalidOption == "" {
			return nil, fmt.Errorf("invalid empty option found")
		}
		return nil, fmt.Errorf("invalid option '%s'", invalidOption)
	}

	return buildProperties, nil
}

// GeneratePropertiesForConfiguration returns the board properties for a particular
// configuration. The parameter is the latest part of the FQBN, for example if
// the full FQBN is "arduino:avr:mega:cpu=atmega2560" the config part must be
// "cpu=atmega2560".
// FIXME: deprecated, use GetBuildProperties instead
func (b *Board) GeneratePropertiesForConfiguration(config string) (*properties.Map, error) {
	fqbn, err := ParseFQBN(b.String() + ":" + config)
	if err != nil {
		return nil, fmt.Errorf("parsing fqbn: %s", err)
	}
	return b.GetBuildProperties(fqbn.Configs)
}
