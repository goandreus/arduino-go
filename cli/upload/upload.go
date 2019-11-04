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

package upload

import (
	"context"
	"os"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cli/cli/instance"
	"github.com/arduino/arduino-cli/commands/upload"
	rpc "github.com/arduino/arduino-cli/rpc/commands"
	"github.com/arduino/go-paths-helper"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	fqbn       string
	port       string
	verbose    bool
	verify     bool
	importFile string
)

// NewCommand created a new `upload` command
func NewCommand() *cobra.Command {
	uploadCommand := &cobra.Command{
		Use:     "upload",
		Short:   "Upload Arduino sketches.",
		Long:    "Upload Arduino sketches.",
		Example: "  " + os.Args[0] + " upload /home/user/Arduino/MySketch",
		Args:    cobra.MaximumNArgs(1),
		Run:     run,
	}

	uploadCommand.Flags().StringVarP(&fqbn, "fqbn", "b", "", "Fully Qualified Board Name, e.g.: arduino:avr:uno")
	uploadCommand.Flags().StringVarP(&port, "port", "p", "", "Upload port, e.g.: COM10 or /dev/ttyACM0")
	uploadCommand.Flags().StringVarP(&importFile, "input", "i", "", "Input file to be uploaded.")
	uploadCommand.Flags().BoolVarP(&verify, "verify", "t", false, "Verify uploaded binary after the upload.")
	uploadCommand.Flags().BoolVarP(&verbose, "verbose", "v", false, "Optional, turns on verbose mode.")

	return uploadCommand
}

func run(command *cobra.Command, args []string) {
	instance := instance.CreateInstance()

	var path *paths.Path
	if len(args) > 0 {
		path = paths.New(args[0])
	}
	sketchPath := initSketchPath(path)

	_, err := upload.Upload(context.Background(), &rpc.UploadReq{
		Instance:   instance,
		Fqbn:       fqbn,
		SketchPath: sketchPath.String(),
		Port:       port,
		Verbose:    verbose,
		Verify:     verify,
		ImportFile: importFile,
	}, os.Stdout, os.Stderr)

	if err != nil {
		feedback.Errorf("Error during Upload: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}
}

// initSketchPath returns the current working directory
func initSketchPath(sketchPath *paths.Path) *paths.Path {
	if sketchPath != nil {
		return sketchPath
	}

	wd, err := paths.Getwd()
	if err != nil {
		feedback.Errorf("Couldn't get current working directory: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}
	logrus.Infof("Reading sketch from dir: %s", wd)
	return wd
}
