// This file is part of arduino-cli.
//
// Copyright 2019 ARDUINO SA (http://www.arduino.cc/)
//
// This software is released under the GNU General Public License version 3,
// which covers the main part of arduino-cli.
// The terms of this license can be found at:
// https://www.gnu.org/licenses/gpl-3.0.en.html
//
// You can be released from the requirements of the above licenses by purchasing
// a commercial license. Buying such a license is mandatory if you want to modify or
// otherwise use the software for commercial activities involving the Arduino
// software without disclosing the source code of your own applications. To purchase
// a commercial license, send an email to license@arduino.cc.

package main

import (
	"context"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	rpc "github.com/arduino/arduino-cli/rpc/commands"
	"google.golang.org/grpc"
)

var (
	dataDir string
)

// The main function implements an example workflow to show how to interact
// with the gRPC Api exposed by arduino-cli when running in daemon mode.
func main() {

	// Establish a connection with the gRPC server, started with the command:
	// arduino-cli daemon
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(100*time.Millisecond))
	if err != nil {
		log.Fatal("error connecting to arduino-cli rpc server, you can start it by running `arduino-cli daemon`")
	}
	defer conn.Close()

	// To avoid polluting an existing arduino-cli installation, the example
	// client uses a temp folder to keep cores, libraries and the likes.
	// You can point `dataDir` to a location that better fits your needs.
	dataDir, err = ioutil.TempDir("", "arduino-rpc-client")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dataDir)

	// Create an instance of the gRPC client.
	client := rpc.NewArduinoCoreClient(conn)

	// Now we can call various methods of the API...

	// `Version` can be called without any setup or init procedure.
	log.Println("calling Version")
	callVersion(client)

	// Before we can do anything with the CLI, an "instance" must be created.
	// We keep a reference to the created instance because we will need it to
	// run subsequent commands.
	log.Println("calling Init")
	instance := initInstance(client)

	// With a brand new instance, the first operation should always be updatating
	// the index.
	log.Println("calling UpdateIndex")
	callUpdateIndex(client, instance)

	// Let's search for a platform (also known as 'core') called 'samd'.
	log.Println("calling PlatformSearch(samd)")
	callPlatformSearch(client, instance)

	// Install arduino:samd@1.6.19
	log.Println("calling PlatformInstall(arduino:samd@1.6.19)")
	callPlatformInstall(client, instance)

	// Now list the installed platforms to double check previous installation
	// went right.
	log.Println("calling PlatformList()")
	callPlatformList(client, instance)

	// Upgrade the installed platform to the latest version.
	log.Println("calling PlatformUpgrade(arduino:samd)")
	callPlatformUpgrade(client, instance)

	// Query board details for a mkr1000
	log.Println("calling BoardDetails(arduino:samd:mkr1000)")
	callBoardsDetails(client, instance)

	// Attach a board to a sketch.
	// Uncomment if you do have an actual board connected.
	// log.Println("calling BoardAttach(serial:///dev/ttyACM0)")
	// callBoardAttach(client, instance)

	// Compile a sketch
	log.Println("calling Compile(arduino:samd:mkr1000, VERBOSE, hello.ino)")
	callCompile(client, instance)

	// Upload a sketch
	// Uncomment if you do have an actual board connected.
	// log.Println("calling Upload(arduino:samd:mkr1000, /dev/ttyACM0, VERBOSE, hello.ino)")
	// callUpload(client, instance)

	// List all boards
	log.Println("calling BoardListAll(mkr)")
	callListAll(client, instance)

	// List connected boards
	log.Println("calling BoardList()")
	callBoardList(client, instance)

	// Uninstall a platform
	log.Println("calling PlatformUninstall(arduino:samd)")
	callPlatformUnInstall(client, instance)

	// Update the Library index
	log.Println("calling UpdateLibrariesIndex()")
	callUpdateLibraryIndex(client, instance)

	// Download a library
	log.Println("calling LibraryDownload(WiFi101@0.15.2)")
	callLibDownload(client, instance)

	// Install a library
	log.Println("calling LibraryInstall(WiFi101@0.15.1)")
	callLibInstall(client, instance, "0.15.1")

	// Replace the previous version
	log.Println("calling LibraryInstall(WiFi101@0.15.2)")
	callLibInstall(client, instance, "0.15.2")

	// Upgrade all libs to latest
	log.Println("calling LibraryUpgradeAll()")
	callLibUpgradeAll(client, instance)

	// Search for a lib using the 'audio' keyword
	log.Println("calling LibrarySearch(audio)")
	callLibSearch(client, instance)

	// List installed libraries
	log.Println("calling LibraryList")
	callLibList(client, instance)

	// Uninstall a library
	log.Println("calling LibraryUninstall(WiFi101)")
	callLibUninstall(client, instance)
}

func callVersion(client rpc.ArduinoCoreClient) {
	versionResp, err := client.Version(context.Background(), &rpc.VersionReq{})
	if err != nil {
		log.Fatalf("Error getting version: %s", err)
	}

	log.Printf("arduino-cli version: %v", versionResp.GetVersion())
}

func initInstance(client rpc.ArduinoCoreClient) *rpc.Instance {
	// The configuration for this example client only contains the path to
	// the data folder.
	initRespStream, err := client.Init(context.Background(), &rpc.InitReq{
		Configuration: &rpc.Configuration{
			DataDir:       dataDir,
			SketchbookDir: filepath.Join(dataDir, "sketchbook"),
			DownloadsDir:  filepath.Join(dataDir, "staging"),
		},
	})
	if err != nil {
		log.Fatalf("Error creating server instance: %s", err)

	}

	var instance *rpc.Instance
	// Loop and consume the server stream until all the setup procedures are done.
	for {
		initResp, err := initRespStream.Recv()
		// The server is done.
		if err == io.EOF {
			break
		}

		// There was an error.
		if err != nil {
			log.Fatalf("Init error: %s", err)
		}

		// The server sent us a valid instance, let's print its ID.
		if initResp.GetInstance() != nil {
			instance = initResp.GetInstance()
			log.Printf("Got a new instance with ID: %v", instance.GetId())
		}

		// When a download is ongoing, log the progress
		if initResp.GetDownloadProgress() != nil {
			log.Printf("DOWNLOAD: %s", initResp.GetDownloadProgress())
		}

		// When an overall task is ongoing, log the progress
		if initResp.GetTaskProgress() != nil {
			log.Printf("TASK: %s", initResp.GetTaskProgress())
		}
	}

	return instance
}

func callUpdateIndex(client rpc.ArduinoCoreClient, instance *rpc.Instance) {
	uiRespStream, err := client.UpdateIndex(context.Background(), &rpc.UpdateIndexReq{
		Instance: instance,
	})
	if err != nil {
		log.Fatalf("Error updating index: %s", err)
	}

	// Loop and consume the server stream until all the operations are done.
	for {
		uiResp, err := uiRespStream.Recv()

		// the server is done
		if err == io.EOF {
			log.Print("Update index done")
			break
		}

		// there was an error
		if err != nil {
			log.Fatalf("Update error: %s", err)
		}

		// operations in progress
		if uiResp.GetDownloadProgress() != nil {
			log.Printf("DOWNLOAD: %s", uiResp.GetDownloadProgress())
		}
	}
}

func callPlatformSearch(client rpc.ArduinoCoreClient, instance *rpc.Instance) {
	searchResp, err := client.PlatformSearch(context.Background(), &rpc.PlatformSearchReq{
		Instance:   instance,
		SearchArgs: "samd",
	})

	if err != nil {
		log.Fatalf("Search error: %s", err)
	}

	platforms := searchResp.GetSearchOutput()
	for _, plat := range platforms {
		// We only print ID and version of the platforms found but you can look
		// at the definition for the rpc.Platform struct for more fields.
		log.Printf("Search result: %+v - %+v", plat.GetID(), plat.GetLatest())
	}
}

func callPlatformInstall(client rpc.ArduinoCoreClient, instance *rpc.Instance) {
	installRespStream, err := client.PlatformInstall(context.Background(),
		&rpc.PlatformInstallReq{
			Instance:        instance,
			PlatformPackage: "arduino",
			Architecture:    "samd",
			Version:         "1.6.19",
		})

	if err != nil {
		log.Fatalf("Error installing platform: %s", err)
	}

	// Loop and consume the server stream until all the operations are done.
	for {
		installResp, err := installRespStream.Recv()

		// The server is done.
		if err == io.EOF {
			log.Printf("Install done")
			break
		}

		// There was an error.
		if err != nil {
			log.Fatalf("Install error: %s", err)
		}

		// When a download is ongoing, log the progress
		if installResp.GetProgress() != nil {
			log.Printf("DOWNLOAD: %s", installResp.GetProgress())
		}

		// When an overall task is ongoing, log the progress
		if installResp.GetTaskProgress() != nil {
			log.Printf("TASK: %s", installResp.GetTaskProgress())
		}
	}
}

func callPlatformList(client rpc.ArduinoCoreClient, instance *rpc.Instance) {
	listResp, err := client.PlatformList(context.Background(),
		&rpc.PlatformListReq{Instance: instance})

	if err != nil {
		log.Fatalf("List error: %s", err)
	}

	for _, plat := range listResp.GetInstalledPlatform() {
		// We only print ID and version of the installed platforms but you can look
		// at the definition for the rpc.Platform struct for more fields.
		log.Printf("Installed platform: %s - %s", plat.GetID(), plat.GetInstalled())
	}
}

func callPlatformUpgrade(client rpc.ArduinoCoreClient, instance *rpc.Instance) {
	upgradeRespStream, err := client.PlatformUpgrade(context.Background(),
		&rpc.PlatformUpgradeReq{
			Instance:        instance,
			PlatformPackage: "arduino",
			Architecture:    "samd",
		})

	if err != nil {
		log.Fatalf("Error upgrading platform: %s", err)
	}

	// Loop and consume the server stream until all the operations are done.
	for {
		upgradeResp, err := upgradeRespStream.Recv()

		// The server is done.
		if err == io.EOF {
			log.Printf("Upgrade done")
			break
		}

		// There was an error.
		if err != nil {
			log.Fatalf("Upgrade error: %s", err)
		}

		// When a download is ongoing, log the progress
		if upgradeResp.GetProgress() != nil {
			log.Printf("DOWNLOAD: %s", upgradeResp.GetProgress())
		}

		// When an overall task is ongoing, log the progress
		if upgradeResp.GetTaskProgress() != nil {
			log.Printf("TASK: %s", upgradeResp.GetTaskProgress())
		}
	}
}

func callBoardsDetails(client rpc.ArduinoCoreClient, instance *rpc.Instance) {
	details, err := client.BoardDetails(context.Background(),
		&rpc.BoardDetailsReq{
			Instance: instance,
			Fqbn:     "arduino:samd:mkr1000",
		})

	if err != nil {
		log.Fatalf("Error getting board data: %s\n", err)
	}

	log.Printf("Board details for %s", details.GetName())
	log.Printf("Required tools: %s", details.GetRequiredTools())
	log.Printf("Config options: %s", details.GetConfigOptions())
}

func callBoardAttach(client rpc.ArduinoCoreClient, instance *rpc.Instance) {
	currDir, _ := os.Getwd()
	boardattachresp, err := client.BoardAttach(context.Background(),
		&rpc.BoardAttachReq{
			Instance:   instance,
			BoardUri:   "/dev/ttyACM0",
			SketchPath: filepath.Join(currDir, "hello.ino"),
		})

	if err != nil {
		log.Fatalf("Attach error: %s", err)
	}

	// Loop and consume the server stream until all the operations are done.
	for {
		attachResp, err := boardattachresp.Recv()

		// The server is done.
		if err == io.EOF {
			log.Print("Attach done")
			break
		}

		// There was an error.
		if err != nil {
			log.Fatalf("Attach error: %s\n", err)
		}

		// When an overall task is ongoing, log the progress
		if attachResp.GetTaskProgress() != nil {
			log.Printf("TASK: %s", attachResp.GetTaskProgress())
		}
	}
}

func callCompile(client rpc.ArduinoCoreClient, instance *rpc.Instance) {
	currDir, _ := os.Getwd()
	compRespStream, err := client.Compile(context.Background(),
		&rpc.CompileReq{
			Instance:   instance,
			Fqbn:       "arduino:samd:mkr1000",
			SketchPath: filepath.Join(currDir, "hello.ino"),
			Verbose:    true,
		})

	if err != nil {
		log.Fatalf("Compile error: %s\n", err)
	}

	// Loop and consume the server stream until all the operations are done.
	for {
		compResp, err := compRespStream.Recv()

		// The server is done.
		if err == io.EOF {
			log.Print("Compilation done")
			break
		}

		// There was an error.
		if err != nil {
			log.Fatalf("Compile error: %s\n", err)
		}

		// When an operation is ongoing you can get its output
		if resp := compResp.GetOutStream(); resp != nil {
			log.Printf("STDOUT: %s", resp)
		}
		if resperr := compResp.GetErrStream(); resperr != nil {
			log.Printf("STDERR: %s", resperr)
		}
	}
}

func callUpload(client rpc.ArduinoCoreClient, instance *rpc.Instance) {
	currDir, _ := os.Getwd()
	uplRespStream, err := client.Upload(context.Background(),
		&rpc.UploadReq{
			Instance:   instance,
			Fqbn:       "arduino:samd:mkr1000",
			SketchPath: filepath.Join(currDir, "hello.ino"),
			Port:       "/dev/ttyACM0",
			Verbose:    true,
		})

	if err != nil {
		log.Fatalf("Upload error: %s\n", err)
	}

	for {
		uplResp, err := uplRespStream.Recv()
		if err == io.EOF {
			log.Printf("Upload done")
			break
		}

		if err != nil {
			log.Fatalf("Upload error: %s", err)
			break
		}

		// When an operation is ongoing you can get its output
		if resp := uplResp.GetOutStream(); resp != nil {
			log.Printf("STDOUT: %s", resp)
		}
		if resperr := uplResp.GetErrStream(); resperr != nil {
			log.Printf("STDERR: %s", resperr)
		}
	}
}

func callListAll(client rpc.ArduinoCoreClient, instance *rpc.Instance) {
	boardListAllResp, err := client.BoardListAll(context.Background(),
		&rpc.BoardListAllReq{
			Instance:   instance,
			SearchArgs: []string{"mkr"},
		})

	if err != nil {
		log.Fatalf("Board list-all error: %s", err)
	}

	for _, board := range boardListAllResp.GetBoards() {
		log.Printf("%s: %s", board.GetName(), board.GetFQBN())
	}
}

func callBoardList(client rpc.ArduinoCoreClient, instance *rpc.Instance) {
	boardListResp, err := client.BoardList(context.Background(),
		&rpc.BoardListReq{Instance: instance})

	if err != nil {
		log.Fatalf("Board list error: %s\n", err)
	}

	for _, port := range boardListResp.GetPorts() {
		log.Printf("port: %s, boards: %+v\n", port.GetAddress(), port.GetBoards())
	}
}

func callPlatformUnInstall(client rpc.ArduinoCoreClient, instance *rpc.Instance) {
	uninstallRespStream, err := client.PlatformUninstall(context.Background(),
		&rpc.PlatformUninstallReq{
			Instance:        instance,
			PlatformPackage: "arduino",
			Architecture:    "samd",
		})

	if err != nil {
		log.Fatalf("Uninstall error: %s", err)
	}

	// Loop and consume the server stream until all the operations are done.
	for {
		uninstallResp, err := uninstallRespStream.Recv()
		if err == io.EOF {
			log.Print("Uninstall done")
			break
		}

		if err != nil {
			log.Fatalf("Uninstall error: %s\n", err)
		}

		if uninstallResp.GetTaskProgress() != nil {
			log.Printf("TASK: %s\n", uninstallResp.GetTaskProgress())
		}
	}
}

func callUpdateLibraryIndex(client rpc.ArduinoCoreClient, instance *rpc.Instance) {
	libIdxUpdateStream, err := client.UpdateLibrariesIndex(context.Background(),
		&rpc.UpdateLibrariesIndexReq{Instance: instance})

	if err != nil {
		log.Fatalf("Error updating libraries index: %s\n", err)
	}

	// Loop and consume the server stream until all the operations are done.
	for {
		resp, err := libIdxUpdateStream.Recv()
		if err == io.EOF {
			log.Print("Library index update done")
			break
		}

		if err != nil {
			log.Fatalf("Error updating libraries index: %s", err)
		}

		if resp.GetDownloadProgress() != nil {
			log.Printf("DOWNLOAD: %s", resp.GetDownloadProgress())
		}
	}
}

func callLibDownload(client rpc.ArduinoCoreClient, instance *rpc.Instance) {
	downloadRespStream, err := client.LibraryDownload(context.Background(),
		&rpc.LibraryDownloadReq{
			Instance: instance,
			Name:     "WiFi101",
			Version:  "0.15.2",
		})

	if err != nil {
		log.Fatalf("Error downloading library: %s", err)
	}

	// Loop and consume the server stream until all the operations are done.
	for {
		downloadResp, err := downloadRespStream.Recv()
		if err == io.EOF {
			log.Print("Lib download done")
			break
		}

		if err != nil {
			log.Fatalf("Download error: %s", err)
		}

		if downloadResp.GetProgress() != nil {
			log.Printf("DOWNLOAD: %s", downloadResp.GetProgress())
		}
	}
}

func callLibInstall(client rpc.ArduinoCoreClient, instance *rpc.Instance, version string) {
	installRespStream, err := client.LibraryInstall(context.Background(),
		&rpc.LibraryInstallReq{
			Instance: instance,
			Name:     "WiFi101",
			Version:  version,
		})

	if err != nil {
		log.Fatalf("Error installing library: %s", err)
	}

	for {
		installResp, err := installRespStream.Recv()
		if err == io.EOF {
			log.Print("Lib install done")
			break
		}

		if err != nil {
			log.Fatalf("Install error: %s", err)
		}

		if installResp.GetProgress() != nil {
			log.Printf("DOWNLOAD: %s\n", installResp.GetProgress())
		}
		if installResp.GetTaskProgress() != nil {
			log.Printf("TASK: %s\n", installResp.GetTaskProgress())
		}
	}
}

func callLibUpgradeAll(client rpc.ArduinoCoreClient, instance *rpc.Instance) {
	libUpgradeAllRespStream, err := client.LibraryUpgradeAll(context.Background(),
		&rpc.LibraryUpgradeAllReq{
			Instance: instance,
		})

	if err != nil {
		log.Fatalf("Error upgrading all: %s\n", err)
	}

	for {
		resp, err := libUpgradeAllRespStream.Recv()
		if err == io.EOF {
			log.Printf("Lib upgrade all done")
			break
		}

		if err != nil {
			log.Fatalf("Upgrading error: %s", err)
		}

		if resp.GetProgress() != nil {
			log.Printf("DOWNLOAD: %s\n", resp.GetProgress())
		}
		if resp.GetTaskProgress() != nil {
			log.Printf("TASK: %s\n", resp.GetTaskProgress())
		}
	}
}

func callLibSearch(client rpc.ArduinoCoreClient, instance *rpc.Instance) {
	libSearchResp, err := client.LibrarySearch(context.Background(),
		&rpc.LibrarySearchReq{
			Instance: instance,
			Query:    "audio",
		})

	if err != nil {
		log.Fatalf("Error searching for library: %s", err)
	}

	for _, res := range libSearchResp.GetLibraries() {
		log.Printf("Result: %s - %s", res.GetName(), res.GetLatest().GetVersion())
	}
}

func callLibList(client rpc.ArduinoCoreClient, instance *rpc.Instance) {
	libLstResp, err := client.LibraryList(context.Background(),
		&rpc.LibraryListReq{
			Instance:  instance,
			All:       false,
			Updatable: false,
		})

	if err != nil {
		log.Fatalf("Error List Library: %s", err)
	}

	for _, res := range libLstResp.GetInstalledLibrary() {
		log.Printf("%s - %s", res.GetLibrary().GetName(), res.GetLibrary().GetVersion())
	}
}

func callLibUninstall(client rpc.ArduinoCoreClient, instance *rpc.Instance) {
	libUninstallRespStream, err := client.LibraryUninstall(context.Background(),
		&rpc.LibraryUninstallReq{
			Instance: instance,
			Name:     "WiFi101",
		})

	if err != nil {
		log.Fatalf("Error uninstalling: %s", err)
	}

	for {
		uninstallResp, err := libUninstallRespStream.Recv()
		if err == io.EOF {
			log.Printf("Lib uninstall done")
			break
		}

		if err != nil {
			log.Fatalf("Uninstall error: %s", err)
		}

		if uninstallResp.GetTaskProgress() != nil {
			log.Printf("TASK: %s", uninstallResp.GetTaskProgress())
		}
	}
}
