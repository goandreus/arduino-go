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
	"os"

	"github.com/spf13/cobra"
)

// NewCommand created a new `sketch` command
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "sketch",
		Short:   "Arduino CLI Sketch Commands.",
		Long:    "Arduino CLI Sketch Commands.",
		Example: "  " + os.Args[0] + " sketch new MySketch",
	}

	cmd.AddCommand(initNewCommand())

	return cmd
}
