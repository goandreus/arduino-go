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

package sketch

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/spf13/cobra"
)

func initNewCommand() *cobra.Command {
	newCommand := &cobra.Command{
		Use:     "new",
		Short:   "Create a new Sketch",
		Long:    "Create a new Sketch",
		Example: "  " + os.Args[0] + " sketch new MultiBlinker",
		Args:    cobra.ExactArgs(1),
		Run:     runNewCommand,
	}
	return newCommand
}

var emptySketch = []byte(`
void setup() {
}

void loop() {
}
`)

func runNewCommand(cmd *cobra.Command, args []string) {
	// Trim to avoid issues if user creates a sketch adding the .ino extesion to the name
	trimmedSketchName := strings.TrimSuffix(args[0], ".ino")
	sketchDir, err := filepath.Abs(trimmedSketchName)
	if err != nil {
		feedback.Errorf("Error creating sketch: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}
	if err := os.MkdirAll(sketchDir, os.FileMode(0755)); err != nil {
		feedback.Errorf("Could not create sketch directory: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}
	sketchName := filepath.Base(sketchDir)
	sketchFile := filepath.Join(sketchDir, sketchName+".ino")
	if err := ioutil.WriteFile(sketchFile, emptySketch, os.FileMode(0644)); err != nil {
		feedback.Errorf("Error creating sketch: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}

	feedback.Print("Sketch created in: " + sketchDir)
}
