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
	"os"

	"github.com/spf13/cobra"
)

// NewCommand created a new `lib` command
func NewCommand() *cobra.Command {
	libCommand := &cobra.Command{
		Use:   "lib",
		Short: "Arduino commands about libraries.",
		Long:  "Arduino commands about libraries.",
		Example: "" +
			"  " + os.Args[0] + " lib install AudioZero\n" +
			"  " + os.Args[0] + " lib update-index",
	}

	libCommand.AddCommand(initDownloadCommand())
	libCommand.AddCommand(initInstallCommand())
	libCommand.AddCommand(initListCommand())
	libCommand.AddCommand(initSearchCommand())
	libCommand.AddCommand(initUninstallCommand())
	libCommand.AddCommand(initUpgradeCommand())
	libCommand.AddCommand(initUpdateIndexCommand())

	return libCommand
}
