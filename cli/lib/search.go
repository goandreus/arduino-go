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
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cli/cli/instance"
	"github.com/arduino/arduino-cli/commands/lib"
	rpc "github.com/arduino/arduino-cli/rpc/commands"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	semver "go.bug.st/relaxed-semver"
)

func initSearchCommand() *cobra.Command {
	searchCommand := &cobra.Command{
		Use:     "search [LIBRARY_NAME]",
		Short:   "Searchs for one or more libraries data.",
		Long:    "Search for one or more libraries data (case insensitive search).",
		Example: "  " + os.Args[0] + " lib search audio",
		Args:    cobra.ArbitraryArgs,
		Run:     runSearchCommand,
	}
	searchCommand.Flags().BoolVar(&searchFlags.namesOnly, "names", false, "Show library names only.")
	return searchCommand
}

var searchFlags struct {
	namesOnly bool // if true outputs lib names only.
}

func runSearchCommand(cmd *cobra.Command, args []string) {
	instance := instance.CreateInstaceIgnorePlatformIndexErrors()
	logrus.Info("Executing `arduino lib search`")
	searchResp, err := lib.LibrarySearch(context.Background(), &rpc.LibrarySearchReq{
		Instance: instance,
		Query:    (strings.Join(args, " ")),
	})
	if err != nil {
		feedback.Errorf("Error saerching for Library: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}

	feedback.PrintResult(result{
		results:   searchResp,
		namesOnly: searchFlags.namesOnly,
	})

	logrus.Info("Done")
}

// output from this command requires special formatting, let's create a dedicated
// feedback.Result implementation
type result struct {
	results   *rpc.LibrarySearchResp
	namesOnly bool
}

func (res result) Data() interface{} {
	if res.namesOnly {
		type LibName struct {
			Name string `json:"name,required"`
		}

		type NamesOnly struct {
			Libraries []LibName `json:"libraries,required"`
		}

		names := []LibName{}
		results := res.results.GetLibraries()
		for _, lsr := range results {
			names = append(names, LibName{lsr.Name})
		}

		return NamesOnly{
			names,
		}
	}

	return res.results
}

func (res result) String() string {
	results := res.results.GetLibraries()
	if len(results) == 0 {
		return "No libraries matching your search."
	}

	// get a sorted slice of results
	sort.Slice(results, func(i, j int) bool {
		return results[i].Name < results[j].Name
	})

	var out strings.Builder

	for _, lsr := range results {
		out.WriteString(fmt.Sprintf("Name: \"%s\"\n", lsr.Name))
		if res.namesOnly {
			continue
		}

		out.WriteString(fmt.Sprintf("  Author: %s\n", lsr.GetLatest().Author))
		out.WriteString(fmt.Sprintf("  Maintainer: %s\n", lsr.GetLatest().Maintainer))
		out.WriteString(fmt.Sprintf("  Sentence: %s\n", lsr.GetLatest().Sentence))
		out.WriteString(fmt.Sprintf("  Paragraph: %s\n", lsr.GetLatest().Paragraph))
		out.WriteString(fmt.Sprintf("  Website: %s\n", lsr.GetLatest().Website))
		out.WriteString(fmt.Sprintf("  Category: %s\n", lsr.GetLatest().Category))
		out.WriteString(fmt.Sprintf("  Architecture: %s\n", strings.Join(lsr.GetLatest().Architectures, ", ")))
		out.WriteString(fmt.Sprintf("  Types: %s\n", strings.Join(lsr.GetLatest().Types, ", ")))
		out.WriteString(fmt.Sprintf("  Versions: %s\n", strings.Replace(fmt.Sprint(versionsFromSearchedLibrary(lsr)), " ", ", ", -1)))
	}

	return fmt.Sprintf("%s", out.String())
}

func versionsFromSearchedLibrary(library *rpc.SearchedLibrary) []*semver.Version {
	res := []*semver.Version{}
	for str := range library.Releases {
		if v, err := semver.Parse(str); err == nil {
			res = append(res, v)
		}
	}
	sort.Sort(semver.List(res))
	return res
}
