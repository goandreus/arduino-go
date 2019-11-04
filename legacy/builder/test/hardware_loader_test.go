/*
 * This file is part of Arduino Builder.
 *
 * Arduino Builder is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 2 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 51 Franklin St, Fifth Floor, Boston, MA  02110-1301  USA
 *
 * As a special exception, you may use this file as part of a free software
 * library without restriction.  Specifically, if other files instantiate
 * templates or use macros or inline functions from this file, or you compile
 * this file and link it with other files to produce an executable, this
 * file does not by itself cause the resulting executable to be covered by
 * the GNU General Public License.  This exception does not however
 * invalidate any other reasons why the executable file might be covered by
 * the GNU General Public License.
 *
 * Copyright 2015 Arduino LLC (http://www.arduino.cc/)
 */

package test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/arduino/arduino-cli/legacy/builder"
	"github.com/arduino/arduino-cli/legacy/builder/types"
	paths "github.com/arduino/go-paths-helper"
	"github.com/stretchr/testify/require"
)

func TestLoadHardware(t *testing.T) {
	ctx := &types.Context{
		HardwareDirs: paths.NewPathList("downloaded_hardware", filepath.Join("..", "hardware"), "hardware"),
	}

	commands := []types.Command{
		&builder.HardwareLoader{},
	}

	for _, command := range commands {
		err := command.Run(ctx)
		NoError(t, err)
	}

	packages := ctx.Hardware
	require.Equal(t, 2, len(packages))
	require.NotNil(t, packages["arduino"])
	require.Equal(t, 2, len(packages["arduino"].Platforms))

	require.Equal(t, "uno", packages["arduino"].Platforms["avr"].Releases[""].Boards["uno"].BoardID)
	require.Equal(t, "uno", packages["arduino"].Platforms["avr"].Releases[""].Boards["uno"].Properties.Get("_id"))

	require.Equal(t, "yun", packages["arduino"].Platforms["avr"].Releases[""].Boards["yun"].BoardID)
	require.Equal(t, "true", packages["arduino"].Platforms["avr"].Releases[""].Boards["yun"].Properties.Get("upload.wait_for_upload_port"))

	require.Equal(t, "{build.usb_flags}", packages["arduino"].Platforms["avr"].Releases[""].Boards["robotMotor"].Properties.Get("build.extra_flags"))

	require.Equal(t, "arduino_due_x", packages["arduino"].Platforms["sam"].Releases[""].Boards["arduino_due_x"].BoardID)

	require.Equal(t, "ATmega123", packages["arduino"].Platforms["avr"].Releases[""].Boards["diecimila"].Properties.Get("menu.cpu.atmega123"))

	avrPlatform := packages["arduino"].Platforms["avr"]
	require.Equal(t, "Arduino AVR Boards", avrPlatform.Releases[""].Properties.Get("name"))
	require.Equal(t, "-v", avrPlatform.Releases[""].Properties.Get("tools.avrdude.bootloader.params.verbose"))
	require.Equal(t, "/my/personal/avrdude", avrPlatform.Releases[""].Properties.Get("tools.avrdude.cmd.path"))

	require.Equal(t, "AVRISP mkII", avrPlatform.Releases[""].Programmers["avrispmkii"].Get("name"))

	//require.Equal(t, "{runtime.tools.ctags.path}", packages.Properties.Get("tools.ctags.path"])
	//require.Equal(t, "\"{cmd.path}\" -u --language-force=c++ -f - --c++-kinds=svpf --fields=KSTtzns --line-directives \"{source_file}\"", packages.Properties.Get("tools.ctags.pattern"])
	//require.Equal(t, "{runtime.tools.avrdude.path}", packages.Properties.Get("tools.avrdude.path"])
	//require.Equal(t, "-w -x c++ -E -CC", packages.Properties.Get("preproc.macros.flags"])
}

func TestLoadHardwareMixingUserHardwareFolder(t *testing.T) {
	ctx := &types.Context{
		HardwareDirs: paths.NewPathList("downloaded_hardware", filepath.Join("..", "hardware"), "hardware", "user_hardware"),
	}

	commands := []types.Command{
		&builder.AddAdditionalEntriesToContext{},
		&builder.HardwareLoader{},
		&builder.PlatformKeysRewriteLoader{},
		&builder.RewriteHardwareKeys{},
	}

	for _, command := range commands {
		err := command.Run(ctx)
		NoError(t, err)
	}

	packages := ctx.Hardware

	if runtime.GOOS == "windows" {
		//a package is a symlink, and windows does not support them
		require.Equal(t, 3, len(packages))
	} else {
		require.Equal(t, 4, len(packages))
	}

	require.NotNil(t, packages["arduino"])
	require.Equal(t, 2, len(packages["arduino"].Platforms))

	require.Equal(t, "uno", packages["arduino"].Platforms["avr"].Releases[""].Boards["uno"].BoardID)
	require.Equal(t, "uno", packages["arduino"].Platforms["avr"].Releases[""].Boards["uno"].Properties.Get("_id"))

	require.Equal(t, "yun", packages["arduino"].Platforms["avr"].Releases[""].Boards["yun"].BoardID)
	require.Equal(t, "true", packages["arduino"].Platforms["avr"].Releases[""].Boards["yun"].Properties.Get("upload.wait_for_upload_port"))

	require.Equal(t, "{build.usb_flags}", packages["arduino"].Platforms["avr"].Releases[""].Boards["robotMotor"].Properties.Get("build.extra_flags"))

	require.Equal(t, "arduino_due_x", packages["arduino"].Platforms["sam"].Releases[""].Boards["arduino_due_x"].BoardID)

	avrPlatform := packages["arduino"].Platforms["avr"].Releases[""]
	require.Equal(t, "Arduino AVR Boards", avrPlatform.Properties.Get("name"))
	require.Equal(t, "-v", avrPlatform.Properties.Get("tools.avrdude.bootloader.params.verbose"))
	require.Equal(t, "/my/personal/avrdude", avrPlatform.Properties.Get("tools.avrdude.cmd.path"))

	require.Equal(t, "AVRISP mkII", avrPlatform.Programmers["avrispmkii"].Get("name"))

	require.Equal(t, "-w -x c++ -M -MG -MP", avrPlatform.Properties.Get("preproc.includes.flags"))
	require.Equal(t, "-w -x c++ -E -CC", avrPlatform.Properties.Get("preproc.macros.flags"))
	require.Equal(t, "\"{compiler.path}{compiler.cpp.cmd}\" {compiler.cpp.flags} {preproc.includes.flags} -mmcu={build.mcu} -DF_CPU={build.f_cpu} -DARDUINO={runtime.ide.version} -DARDUINO_{build.board} -DARDUINO_ARCH_{build.arch} {compiler.cpp.extra_flags} {build.extra_flags} {includes} \"{source_file}\"", avrPlatform.Properties.Get("recipe.preproc.includes"))
	require.False(t, avrPlatform.Properties.ContainsKey("preproc.macros.compatibility_flags"))

	require.NotNil(t, packages["my_avr_platform"])
	myAVRPlatform := packages["my_avr_platform"]
	//require.Equal(t, "hello world", myAVRPlatform.Properties.Get("example"))
	myAVRPlatformAvrArch := myAVRPlatform.Platforms["avr"].Releases[""]
	require.Equal(t, "custom_yun", myAVRPlatformAvrArch.Boards["custom_yun"].BoardID)

	require.False(t, myAVRPlatformAvrArch.Properties.ContainsKey("preproc.includes.flags"))

	//require.Equal(t, "{runtime.tools.ctags.path}", packages.Properties.Get("tools.ctags.path"))
	//require.Equal(t, "\"{cmd.path}\" -u --language-force=c++ -f - --c++-kinds=svpf --fields=KSTtzns --line-directives \"{source_file}\"", packages.Properties.Get("tools.ctags.pattern"))
	//require.Equal(t, "{runtime.tools.avrdude.path}", packages.Properties.Get("tools.avrdude.path"))
	//require.Equal(t, "-w -x c++ -E -CC", packages.Properties.Get("preproc.macros.flags"))

	if runtime.GOOS != "windows" {
		require.NotNil(t, packages["my_symlinked_avr_platform"])
		require.NotNil(t, packages["my_symlinked_avr_platform"].Platforms["avr"])
	}
}

func TestLoadHardwareWithBoardManagerFolderStructure(t *testing.T) {
	ctx := &types.Context{
		HardwareDirs: paths.NewPathList("downloaded_board_manager_stuff"),
	}

	commands := []types.Command{
		&builder.HardwareLoader{},
	}

	for _, command := range commands {
		err := command.Run(ctx)
		NoError(t, err)
	}

	packages := ctx.Hardware
	require.Equal(t, 3, len(packages))
	require.NotNil(t, packages["arduino"])
	require.Equal(t, 1, len(packages["arduino"].Platforms))
	require.NotNil(t, packages["RedBearLab"])
	require.Equal(t, 1, len(packages["RedBearLab"].Platforms))
	require.NotNil(t, packages["RFduino"])
	require.Equal(t, 0, len(packages["RFduino"].Platforms))

	samdPlatform := packages["arduino"].Platforms["samd"].Releases["1.6.5"]
	require.Equal(t, 3, len(samdPlatform.Boards))

	require.Equal(t, "arduino_zero_edbg", samdPlatform.Boards["arduino_zero_edbg"].BoardID)
	require.Equal(t, "arduino_zero_edbg", samdPlatform.Boards["arduino_zero_edbg"].Properties.Get("_id"))

	require.Equal(t, "arduino_zero", samdPlatform.Boards["arduino_zero_native"].Properties.Get("build.variant"))
	require.Equal(t, "-D__SAMD21G18A__ {build.usb_flags}", samdPlatform.Boards["arduino_zero_native"].Properties.Get("build.extra_flags"))

	require.Equal(t, "Arduino SAMD (32-bits ARM Cortex-M0+) Boards", samdPlatform.Properties.Get("name"))
	require.Equal(t, "-d3", samdPlatform.Properties.Get("tools.openocd.erase.params.verbose"))

	require.Equal(t, 3, len(samdPlatform.Programmers))

	require.Equal(t, "Atmel EDBG", samdPlatform.Programmers["edbg"].Get("name"))
	require.Equal(t, "openocd", samdPlatform.Programmers["edbg"].Get("program.tool"))

	avrRedBearPlatform := packages["RedBearLab"].Platforms["avr"].Releases["1.0.0"]
	require.Equal(t, 3, len(avrRedBearPlatform.Boards))

	require.Equal(t, "blend", avrRedBearPlatform.Boards["blend"].BoardID)
	require.Equal(t, "blend", avrRedBearPlatform.Boards["blend"].Properties.Get("_id"))
	require.Equal(t, "arduino:arduino", avrRedBearPlatform.Boards["blend"].Properties.Get("build.core"))
}

func TestLoadLotsOfHardware(t *testing.T) {
	ctx := &types.Context{
		HardwareDirs: paths.NewPathList("downloaded_board_manager_stuff", "downloaded_hardware", filepath.Join("..", "hardware"), "hardware", "user_hardware"),
	}

	commands := []types.Command{
		&builder.HardwareLoader{},
	}

	for _, command := range commands {
		err := command.Run(ctx)
		NoError(t, err)
	}

	packages := ctx.Hardware

	if runtime.GOOS == "windows" {
		//a package is a symlink, and windows does not support them
		require.Equal(t, 5, len(packages))
	} else {
		require.Equal(t, 6, len(packages))
	}

	require.NotNil(t, packages["arduino"])
	require.NotNil(t, packages["my_avr_platform"])

	require.Equal(t, 3, len(packages["arduino"].Platforms))
	require.Equal(t, 20, len(packages["arduino"].Platforms["avr"].Releases[""].Boards))
	require.Equal(t, 2, len(packages["arduino"].Platforms["sam"].Releases[""].Boards))
	require.Equal(t, 3, len(packages["arduino"].Platforms["samd"].Releases["1.6.5"].Boards))

	require.Equal(t, 1, len(packages["my_avr_platform"].Platforms))
	require.Equal(t, 2, len(packages["my_avr_platform"].Platforms["avr"].Releases[""].Boards))

	if runtime.GOOS != "windows" {
		require.Equal(t, 1, len(packages["my_symlinked_avr_platform"].Platforms))
		require.Equal(t, 2, len(packages["my_symlinked_avr_platform"].Platforms["avr"].Releases[""].Boards))
	}
}
