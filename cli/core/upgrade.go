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
	"os"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cli/cli/globals"
	"github.com/arduino/arduino-cli/cli/instance"
	"github.com/arduino/arduino-cli/cli/output"
	"github.com/arduino/arduino-cli/commands/core"
	rpc "github.com/arduino/arduino-cli/rpc/commands"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func initUpgradeCommand() *cobra.Command {
	upgradeCommand := &cobra.Command{
		Use:   "upgrade [PACKAGER:ARCH] ...",
		Short: "Upgrades one or all installed platforms to the latest version.",
		Long:  "Upgrades one or all installed platforms to the latest version.",
		Example: "" +
			"  # upgrade everything to the latest version\n" +
			"  " + os.Args[0] + " core upgrade\n\n" +
			"  # upgrade arduino:samd to the latest version\n" +
			"  " + os.Args[0] + " core upgrade arduino:samd",
		Run: runUpgradeCommand,
	}
	return upgradeCommand
}

func runUpgradeCommand(cmd *cobra.Command, args []string) {
	instance := instance.CreateInstance()
	logrus.Info("Executing `arduino core upgrade`")

	// if no platform was passed, upgrade allthethings
	if len(args) == 0 {
		targets, err := core.GetPlatforms(instance.Id, true)
		if err != nil {
			feedback.Errorf("Error retrieving core list: %v", err)
			os.Exit(errorcodes.ErrGeneric)
		}

		if len(targets) == 0 {
			feedback.Print("All the cores are already at the latest version")
			return
		}

		for _, t := range targets {
			args = append(args, t.Platform.String())
		}
	}

	// proceed upgrading, if anything is upgradable
	exitErr := false
	platformsRefs, err := globals.ParseReferenceArgs(args, true)
	if err != nil {
		feedback.Errorf("Invalid argument passed: %v", err)
		os.Exit(errorcodes.ErrBadArgument)
	}

	for i, platformRef := range platformsRefs {
		if platformRef.Version != "" {
			feedback.Error(("Invalid item " + args[i]))
			exitErr = true
			continue
		}

		r := &rpc.PlatformUpgradeReq{
			Instance:        instance,
			PlatformPackage: platformRef.PackageName,
			Architecture:    platformRef.Architecture,
		}

		_, err := core.PlatformUpgrade(context.Background(), r, output.ProgressBar(), output.TaskProgress(), globals.NewHTTPClientHeader())
		if err == core.ErrAlreadyLatest {
			feedback.Printf("Platform %s is already at the latest version", platformRef)
		} else if err != nil {
			feedback.Errorf("Error during upgrade: %v", err)
			os.Exit(errorcodes.ErrGeneric)
		}
	}

	if exitErr {
		os.Exit(errorcodes.ErrBadArgument)
	}
}
