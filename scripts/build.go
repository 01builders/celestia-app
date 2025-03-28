package main

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

/*

To run this script, you need to have the following tools installed:
- git
- go
- make

```go
go run scripts/build.go <branch_name> <version>
```

This will generate a tar.gz archive for the given branch name in the `internal/embedding` directory.

*/

// Architecture holds the GOOS and GOARCH values.
type Architecture struct {
	GOOS   string
	GOARCH string
}

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s <branch_name> <version>\n", os.Args[0])
		os.Exit(1)
	}
	branchName := os.Args[1]
	version := os.Args[2]

	// Set variables
	ceCelestiaRepo := "https://github.com/celestiaorg/celestia-app.git"
	binariesDir := "./binaries"
	targetDir := "./internal/embedding"
	repoDir := "celestia-app"

	// Map of architectures.
	architectures := map[string]Architecture{
		"darwin_arm64":  {GOOS: "darwin", GOARCH: "arm64"},
		"darwin_x86_64": {GOOS: "darwin", GOARCH: "amd64"},
		"linux_arm64":   {GOOS: "linux", GOARCH: "arm64"},
		"linux_x86_64":  {GOOS: "linux", GOARCH: "amd64"},
	}

	// Create necessary directories
	if err := os.MkdirAll(binariesDir, 0755); err != nil {
		log.Fatalf("Error creating downloads directory: %v", err)
	}
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		log.Fatalf("Error creating target directory: %v", err)
	}

	// Clone the repository if it does not exist.
	if _, err := os.Stat(repoDir); os.IsNotExist(err) {
		fmt.Printf("Cloning Celestia repository from branch: %s...\n", branchName)
		cmd := exec.Command("git", "clone", "-b", branchName, ceCelestiaRepo, repoDir)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			log.Fatalf("Error cloning repository: %v", err)
		}
	}

	// Change to repository directory.
	if err := os.Chdir(repoDir); err != nil {
		log.Fatalf("Error changing directory to %s: %v", repoDir, err)
	}

	// Print the current directory.
	fmt.Println("--------------------------------------------")
	fmt.Println("Current directory:")

	// Get and print the actual working directory
	if currentDir, err := os.Getwd(); err == nil {
		fmt.Println("Actual working directory:", currentDir)
	} else {
		fmt.Printf("Error getting working directory: %v\n", err)
	}
	fmt.Println("--------------------------------------------")
	fmt.Println("Cleaning build directory...")
	fmt.Println("--------------------------------------------")
	runCmd("make", "clean")

	// Iterate over each architecture.
	for archKey, arch := range architectures {
		fmt.Println("--------------------------------------------")
		fmt.Printf("Building Celestia binary for %s...\n", archKey)
		fmt.Printf("GOOS=%s, GOARCH=%s\n", arch.GOOS, arch.GOARCH)
		fmt.Println("--------------------------------------------")

		// Clean build directory
		runCmd("make", "clean")

		// Build binary by setting env variables.
		env := os.Environ()
		env = append(env, "GOOS="+arch.GOOS, "GOARCH="+arch.GOARCH, "CGO_ENABLED=0")
		buildCmd := exec.Command("make", "build")
		buildCmd.Env = env
		buildCmd.Stdout = os.Stdout
		buildCmd.Stderr = os.Stderr
		if err := buildCmd.Run(); err != nil {
			log.Fatalf("Error building binary for %s: %v", archKey, err)
		}

		// Define the binary name and paths.
		binarySrc := filepath.Join("build", "celestia-appd")
		binaryName := fmt.Sprintf("celestia-app-v%s_%s", version, archKey)
		binaryDst := filepath.Join("..", binariesDir, binaryName)

		// Copy binary to binaries directory.
		copyFile(binarySrc, binaryDst)

		// Change to binaries directory to package.
		absDownloadDir, err := filepath.Abs(filepath.Join("..", binariesDir))
		if err != nil {
			log.Fatalf("Error getting absolute path: %v", err)
		}
		if err := os.Chdir(absDownloadDir); err != nil {
			log.Fatalf("Error changing directory to binaries directory: %v", err)
		}

		// Create tar.gz archive.
		tarName := binaryName + ".tar.gz"
		if err := createTarGz(tarName, binaryName); err != nil {
			log.Fatalf("Error creating tar.gz archive: %v", err)
		}

		// Move the tarball to target directory.
		absTargetDir, err := filepath.Abs(filepath.Join("..", targetDir))
		if err != nil {
			log.Fatalf("Error getting absolute path for target directory: %v", err)
		}
		if err := os.MkdirAll(absTargetDir, 0755); err != nil {
			log.Fatalf("Error creating target directory: %v", err)
		}
		newTarPath := filepath.Join(absTargetDir, tarName)
		if err := os.Rename(tarName, newTarPath); err != nil {
			log.Fatalf("Error moving tar.gz to target directory: %v", err)
		}
		fmt.Printf("Tarball for %s moved to %s\n", archKey, newTarPath)

		// Remove the uncompressed binary from binaries.
		if err := os.Remove(binaryName); err != nil {
			log.Printf("Warning: could not remove file %s: %v", binaryName, err)
		}

		// Return to the repository directory for next iteration.
		repoAbsPath, err := filepath.Abs("..")
		if err != nil {
			log.Fatalf("Error getting repository path: %v", err)
		}
		if err := os.Chdir(filepath.Join(repoAbsPath, repoDir)); err != nil {
			log.Fatalf("Error returning to repository directory: %v", err)
		}
	}

	// Remove the binaries directory.
	if err := os.Remove(binariesDir); err != nil {
		log.Printf("Warning: could not remove directory %s: %v", repoDir, err)
	}

	fmt.Println("--------------------------------------------")
	fmt.Println("Build and packaging complete for all architectures!")
	fmt.Println("--------------------------------------------")
}

// runCmd executes a command with given arguments, passing through stdout/stderr.
func runCmd(name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatalf("Error running command %s %v: %v", name, args, err)
	}
}

// copyFile copies a file from src to dst while preserving mode.
func copyFile(src, dst string) {
	srcFile, err := os.Open(src)
	if err != nil {
		log.Fatalf("Error opening source file %s: %v", src, err)
	}
	defer srcFile.Close()

	// Get source file mode
	srcInfo, err := srcFile.Stat()
	if err != nil {
		log.Fatalf("Error getting source file info: %v", err)
	}

	// Create destination with same permissions
	dstFile, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		log.Fatalf("Error creating destination file %s: %v", dst, err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		log.Fatalf("Error copying from %s to %s: %v", src, dst, err)
	}
}

// createTarGz creates a tar.gz archive with the given tarName containing the fileOrDir.
func createTarGz(tarName, fileOrDir string) error {
	tarFile, err := os.Create(tarName)
	if err != nil {
		return fmt.Errorf("creating tar file: %w", err)
	}
	defer tarFile.Close()

	gw := gzip.NewWriter(tarFile)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	info, err := os.Stat(fileOrDir)
	if err != nil {
		return fmt.Errorf("stat %s: %w", fileOrDir, err)
	}

	// Open file if it's not a directory.
	if !info.IsDir() {
		return addFileToTar(tw, fileOrDir, info)
	}

	// Otherwise, iterate over directory entries.
	return filepath.Walk(fileOrDir, func(file string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		return addFileToTar(tw, file, fi)
	})
}

// addFileToTar adds an individual file to the tar writer.
func addFileToTar(tw *tar.Writer, file string, fi os.FileInfo) error {
	header, err := tar.FileInfoHeader(fi, file)
	if err != nil {
		return fmt.Errorf("creating tar header for %s: %w", file, err)
	}

	// Ensure the header has the proper name.
	header.Name = file

	if err := tw.WriteHeader(header); err != nil {
		return fmt.Errorf("writing tar header for %s: %w", file, err)
	}

	// If not a regular file, nothing more to do.
	if !fi.Mode().IsRegular() {
		return nil
	}

	f, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("opening file %s: %w", file, err)
	}
	defer f.Close()

	if _, err := io.Copy(tw, f); err != nil {
		return fmt.Errorf("copying file %s to tar: %w", file, err)
	}
	return nil
}
