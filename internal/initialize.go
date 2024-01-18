// File: InitializeJitRepository.go
// Package: internal

// Program Description:
// This file handles the creation of a jit repository
// It validates the user's path, permission and ensures that the .jit directory
// Is cleaned up if any errors occurred during the creation process.

// Author: Martin Alemajoh
// Jit-VCS - v1.0.0
// Created on: January 16, 2024

package internal

import (
	"errors"
	"fmt"
	"jit/pkg/util"
	"log"
	"os"
	"path/filepath"
	"strconv"
)

var jitFileSystem = map[string]util.File{
	util.MAIN:      util.DataFile,
	util.HEAD:      util.DataFile,
	util.STAGE:     util.DataFile,
	util.CONFIG:    util.DataFile,
	util.LOGS:      util.Directory,
	util.INFO:      util.Directory,
	util.BRANCHES:  util.Directory,
	util.SNAPSHOTS: util.Directory,
	util.OBJECTS:   util.Directory,
}

// InitializeJitRepository initializes a new JIT repository based on the provided options.
//
// This function is responsible for setting up the entire structure of a JIT version control system.
// It handles various options for repository configuration, such as creating a bare repository,
// setting up in a separate directory, defining an initial branch, and more. The function ensures
// that the necessary files and directories are created and properly configured.
//
// Args:
//     options (map[string]any): A map containing configuration options for the repository.
//                               These options include flags for quiet mode, bare repository setup,
//                               separate directory setup, template, object format, initial branch, and permissions.
//     dir (string): The default directory where the repository is to be initialized if no separate directory is provided.
//
// Returns:
//     ok (bool): A boolean indicating whether the JIT repository initialization was successful.
//     err (error): An error object that captures any issues encountered during the initialization.
//                  If the process is successful, err will be nil.
//
// The function performs the following steps:
// 1. Parses and validates each option from the provided map (quiet mode, bare repository, etc.).
// 2. Determines the root directory for the repository, handling separate directory scenarios.
// 3. In the case of a separate directory, creates a symbolic link to it.
// 4. Creates the necessary directory structure and files for the repository.
// 5. Writes configuration settings to the repository's config file.
// 6. Sets up the initial branch for the repository.
//
// Usage:
//     options := map[string]any{"quiet": true, "bare": false, "separate-jit-dir": "/path/to/dir", "initial-branch": "main"}
//     ok, err := InitializeJitRepository(options, "/default/path")
//     if err != nil {
//         log.Fatalf("Failed to initialize JIT repository: %s", err)
//     }
//
// Note:
//     - The function provides extensive logging and error handling to ensure that any issues during
//       repository setup are clearly communicated and addressed.
//     - The repository initialization process is flexible, accommodating various configurations and setups.
//     - In case of any failure during the process, appropriate cleanup is performed to avoid partial setups.

func InitializeJitRepository(options map[string]any, dir string) (ok bool, err error) {

	quiet, ok := options["quiet"].(bool)
	if !ok {
		log.Println("Quiet mode failed")
	}

	bare, ok := options["bare"].(bool)
	if !ok {
		log.Println("Bare option failed")
	}

	separateJitDir, ok := options["separate-jit-dir"].(string)
	if !ok {
		log.Println("Separate Jit Dir failed")
	}

	template, ok := options["template"].(string)
	if !ok {
		log.Println("Template option failed")
	}

	objectFormat, ok := options["object-format"].(string)
	if !ok {
		log.Println("Object format option failed")
	}

	initialBranch, ok := options["initial-branch"].(string)
	if !ok {
		log.Println("Initial Branch option failed")
	}

	directoryPerm, ok := options["perm"].(string)
	if !ok {
		log.Println("Permission option failed")
	}

	filePermission, convertErr := strconv.ParseUint(directoryPerm, 8, 32)
	if convertErr != nil {
		filePermission = 0755
	}

	var sepDir string
	var sepErr error

	if separateJitDir != "" {
		sepDir, sepErr = GetJitRootDir(separateJitDir)
		if sepErr != nil {
			return false, sepErr
		}
	}

	workingDir, wkDirErr := GetJitRootDir(dir)
	if wkDirErr != nil {
		return false, wkDirErr
	}

	if separateJitDir != "" {
		//Create a symbolic link
		createErr := os.Symlink(sepDir, filepath.Join(workingDir, util.JitDirName))
		if createErr != nil {
			return false, errors.New("start the terminal in administrative mode to use separate directory option")
		}

		if _, createJitDirErr := CreateJitDir(sepDir, true, bare, filePermission); createJitDirErr != nil {
			return false, createJitDirErr
		}
	} else {
		if _, createJitDirErr := CreateJitDir(workingDir, false, bare, filePermission); createJitDirErr != nil {
			return false, createJitDirErr
		}
	}

	//Write configuration
	config := map[string]string{
		"TEMPLATE":       template,
		"OBJECT-FORMAT":  objectFormat,
		"INITIAL-BRANCH": initialBranch,
	}

	finalJitDir := ConstructFinalJitDir(workingDir, sepDir, bare)

	if _, writeErr := WriteToConfigFile(config, finalJitDir); writeErr != nil {
		log.Println(writeErr)
	}

	//setup initial branch
	ok, setupErr := SetUpInitialBranch(finalJitDir, initialBranch)
	if setupErr != nil {
		log.Printf("encountered an error while creating a jit repository.")
		return false, setupErr
	}

	if !quiet {
		dirAbs, _ := filepath.Abs(workingDir)
		log.Printf("Successfully initialized a new jit repository -> %s", filepath.Join(dirAbs, util.JitDirName))
	}

	return true, nil

}

// ConstructFinalJitDir constructs the final directory path for the JIT repository.
//
// This function determines the final path where the JIT repository should be created or initialized.
// It considers whether a separate directory is specified for the repository and whether the repository
// is a bare repository. Based on these conditions, it constructs and returns the appropriate directory path.
//
// Args:
//
//	wkDir (string): The working directory where the repository might be created. This is used if
//	                no separate directory is specified.
//	sepDir (string): An optional separate directory path for the repository. If provided, the repository
//	                 will be created here instead of the working directory.
//	bare (bool): A flag indicating whether the repository is a bare repository. Bare repositories
//	             do not contain a working directory or .git subdirectory.
//
// Returns:
//
//	string: The final absolute path to the directory where the JIT repository will be created.
//
// The function follows this logic:
//  1. If a separate directory (sepDir) is provided, it returns sepDir as the final path.
//  2. If sepDir is not provided and the repository is bare, it returns the working directory (wkDir).
//  3. If sepDir is not provided and the repository is not bare, it returns the working directory
//     joined with the standard JIT directory name (util.JitDirName).
//
// Usage:
//
//	finalDir := ConstructFinalJitDir("/path/to/work/dir", "/path/to/sep/dir", false)
//	// finalDir will be "/path/to/sep/dir" in this case
//
// Note:
//   - The function does not create any directories; it merely constructs the path based on the given conditions.
//   - This function is essential for determining the correct location for repository initialization
//     and subsequent operations.
func ConstructFinalJitDir(wkDir string, sepDir string, bare bool) string {
	if sepDir != "" {
		return sepDir
	}
	if bare {
		return wkDir
	}
	return filepath.Join(wkDir, util.JitDirName)
}

// GetJitRootDir determines the root directory for a JIT repository.
//
// This function is used to ascertain the root directory of the JIT repository. It either validates
// and returns the provided directory path or, if no path is provided, it returns the current working
// directory of the program. This function ensures that the directory exists, is indeed a directory,
// and that the program has write permissions to it.
//
// Args:
//
//	dirPath (string): The path to the directory to be used as the JIT repository root.
//	                  If empty, the current working directory is used.
//
// Returns:
//
//	dir (string): The absolute path to the JIT repository root directory.
//	err (error): An error object that captures any issues encountered during the process.
//	             If the directory is valid and accessible, err will be nil.
//
// The function performs the following steps:
//  1. If a dirPath is provided, it validates the path using ValidateDirPath and checks write permissions
//     with CheckWritePermission.
//  2. If dirPath is not provided (empty string), the function fetches and returns the current working directory.
//  3. If any error is encountered during validation, permission checks, or fetching the current directory,
//     the error is returned.
//
// Usage:
//
//	rootDir, err := GetJitRootDir("/path/to/proposed/root")
//	if err != nil {
//	    log.Fatalf("Failed to determine JIT root directory: %s", err)
//	}
//
// Note:
//   - The function is a utility to ensure the selected directory for the JIT repository is valid and writable.
//   - It encapsulates directory validation and permission checks, centralizing these checks for reuse.
func GetJitRootDir(dirPath string) (dir string, err error) {
	if dirPath != "" {
		validPathErr := ValidateDirPath(dirPath)
		if validPathErr != nil {
			return "", validPathErr
		}

		permErr := CheckWritePermission(dirPath)
		if permErr != nil {
			return "", permErr
		}

		return dirPath, nil

	} else {
		curDir, curErr := os.Getwd()
		if curErr != nil {
			return "", curErr
		}
		return curDir, nil
	}

}

// CheckWritePermission verifies if the current user has write permissions in the specified directory.
//
// This function is used to check if the program can create files in the provided directory path.
// It attempts to create a temporary file in the directory and evaluates if this operation is successful
// to determine write access. This is a common method to verify write permissions in a directory without
// altering any existing files or data.
//
// Args:
//
//	currentDir (string): The directory path where write permissions are being checked.
//
// Returns:
//
//	err (error): An error object that indicates the lack of write permissions.
//	             If the function can successfully create and delete a temporary file in the directory,
//	             it returns nil, implying write permission exists. If not, it returns an error with a
//	             descriptive message.
//
// The function performs the following steps:
//  1. It attempts to create a temporary file in the specified directory.
//  2. If the file creation fails, it returns an error, indicating lack of write permissions.
//  3. If the file is created successfully, it closes and removes the file, then returns nil, indicating
//     write permissions are available.
//
// Usage:
//
//	err := CheckWritePermission("/path/to/directory")
//	if err != nil {
//	    log.Printf("No write permissions in the directory: %s", err)
//	}
//
// Note:
//   - The function employs a temporary file to test write access, ensuring no disruption to existing data.
//   - The temporary file is removed immediately after creation to avoid leaving residual files in the file system.
//   - Errors encountered during file closure or removal are logged for informational purposes.
func CheckWritePermission(currentDir string) (err error) {
	//check to see if user has write permission
	file, tempErr := os.CreateTemp(currentDir, "test")
	if tempErr != nil {
		errMsg := fmt.Sprintf("you don't have write permissions here -> %s", currentDir)
		return errors.New(errMsg)
	}
	defer func(name string) {
		removeErr := os.Remove(name)
		if removeErr != nil {
			log.Printf("Error removing temporary file: %v", removeErr)
		}
	}(file.Name())

	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			log.Printf("Error closing temporary file: %v", closeErr)
		}
	}()

	return nil
}

// ValidateDirPath checks if the provided path is a valid directory.
//
// This function validates whether the given path exists and is indeed a directory. It's
// commonly used in situations where a directory path is required, ensuring that the provided
// path meets these criteria before proceeding with operations that assume a directory at that path.
//
// Args:
//
//	currentDir (string): The path to be validated as a directory.
//
// Returns:
//
//	err (error): An error object that captures any issues encountered during validation.
//	             If the path is a valid directory, err will be nil. If the path does not exist,
//	             or is not a directory, the function returns an error with a descriptive message.
//
// The function performs the following steps:
// 1. It uses os.Stat to obtain file information about the path.
// 2. If os.Stat returns an error (e.g., if the path does not exist), this error is returned.
// 3. If os.Stat indicates the path is not a directory (using the IsDir method), an error is returned.
//
// Usage:
//
//	err := ValidateDirPath("/path/to/validate")
//	if err != nil {
//	    log.Printf("Invalid directory path: %s", err)
//	}
//
// Note:
//   - This function is particularly useful before performing directory-specific operations,
//     such as reading contents, creating subdirectories, etc.
//   - It ensures that operations intended for directories are not mistakenly performed on files.
func ValidateDirPath(currentDir string) (err error) {
	//check to make sure the given path is a directory
	info, pathErr := os.Stat(currentDir)
	if pathErr != nil {
		return pathErr
	}
	if !info.IsDir() {
		errMsg := fmt.Sprintf("%s is not a directory", currentDir)
		return errors.New(errMsg)
	}

	return nil
}

// CreateJitDir creates the necessary files and directories for a JIT repository.
//
// This function is a key part of setting up a JIT repository. It is responsible for creating
// the structure of the repository, including various files and directories as defined in the
// jitFileSystem map. The function takes into account whether the repository is a bare repository
// and whether it is being created in a separate directory.
//
// Args:
//
//	wkDir (string): The working directory where the repository is being created.
//	sepDir (bool): A flag indicating whether the repository is being created in a separate directory.
//	bare (bool): A flag indicating whether the repository is a bare repository.
//	filePermission (uint64): The file permissions to use when creating new files and directories.
//
// Returns:
//
//	ok (bool): A boolean indicating whether the repository creation was successful.
//	err (error): An error object that captures any issues encountered during the creation process.
//	             If the creation is successful, err will be nil.
//
// The function performs the following steps:
// 1. If the repository is not bare and not in a separate directory, it creates the root ".jit" directory.
// 2. It then iterates over the jitFileSystem map, creating each file and directory specified therein.
//   - For each file, it uses os.Create to create the file, immediately closing it afterward.
//   - For each directory, it uses os.MkdirAll to ensure the directory and all necessary parent directories are created.
//
// 3. If any error occurs during the creation process, it triggers a cleanup and returns an error.
//
// Usage:
//
//	ok, err := CreateJitDir("/path/to/work/dir", false, false, 0755)
//	if err != nil {
//	    log.Fatalf("Failed to create JIT directory: %s", err)
//	}
//
// Note:
//   - The function is careful to close all file resources it opens, using deferred Close calls.
//   - File and directory creation is logged, and any errors encountered halt the process and trigger cleanup.
//   - The behavior of the function changes based on the `bare` and `sepDir` flags,
//     accommodating different repository setups.
func CreateJitDir(wkDir string, sepDir bool, bare bool, filePermission uint64) (ok bool, err error) {

	if sepDir == false && bare == false {
		//Creat the root ".jit" directory if it's not a bare repo
		if mkErr := os.Mkdir(filepath.Join(wkDir, util.JitDirName), os.FileMode(filePermission)); mkErr != nil {
			errMsg := fmt.Sprintf(" %s already contains a jit repository. change the current directory or remove the .jit from current directory.", wkDir)
			return false, errors.New(errMsg)
		}
		wkDir = filepath.Join(wkDir, util.JitDirName) // Create repository in .jit directory

	}

	for k, v := range jitFileSystem {
		if v == util.DataFile {
			file, createErr := os.Create(filepath.Join(wkDir, k))
			if createErr != nil {
				log.Println(createErr)
				break
			}
			// Close the file as soon as you're done
			if closeErr := file.Close(); closeErr != nil {
				log.Println(closeErr)
				break
			}
		}
		if v == util.Directory {
			if createErr := os.MkdirAll(filepath.Join(wkDir, k), util.DefaultFilePerm); createErr != nil {
				log.Println(createErr)
				break
			}
		}
	}

	return true, nil
}

// WriteToConfigFile writes configuration key-value pairs to a configuration file in the JIT repository.
//
// This function is responsible for appending configuration settings to a config file
// within the specified JIT repository. It takes a map of configuration key-value pairs (config)
// and the directory of the JIT repository (jitDir) as arguments.
//
// Args:
//
//	config (map[string]string): A map containing configuration keys and their corresponding values.
//	jitDir (string): The directory where the JIT repository's config file is located.
//
// Returns:
//
//	ok (bool): A boolean value indicating whether the write operation was successful.
//	err (error): An error object that captures any issues encountered during the write operation.
//	             If the write operation is successful, err will be nil.
//
// The function performs the following steps:
//  1. It constructs the full path of the config file within the JIT repository.
//  2. It opens (or creates if not existent) the config file with append mode and default file permissions.
//  3. It iterates over each key-value pair in the provided configuration map and writes them to the
//     config file in the format "key=value\n".
//
// Usage:
//
//	config := map[string]string{"TEMPLATE": "templatePath", "INITIAL-BRANCH": "main"}
//	ok, err := WriteToConfigFile(config, "/path/to/jit/repo")
//	if err != nil {
//	    log.Fatalf("Failed to write to config file: %s", err)
//	}
//
// Note:
//   - The function uses `os.OpenFile` with the `os.O_APPEND|os.O_CREATE` flags, ensuring that
//     the config file is created if it does not exist, and existing content is not overwritten.
//   - Proper error handling is implemented to catch and return errors encountered during file
//     operations.
//   - The function ensures file resources are properly closed using deferred Close calls.
//   - Any write errors encountered during the iteration over the configuration map are log but
//     do not interrupt the process. This behavior could be modified based on requirements.
func WriteToConfigFile(config map[string]string, jitDir string) (ok bool, err error) {

	configFile := filepath.Join(jitDir, util.CONFIG)
	f, openErr := os.OpenFile(configFile, os.O_APPEND|os.O_CREATE, util.DefaultFilePerm)
	defer func() {
		_ = f.Close()
	}()
	if openErr != nil {
		return false, openErr
	}

	for k, v := range config {
		line := fmt.Sprintf("%s=%s\n", k, v)
		if _, writeErr := f.Write([]byte(line)); writeErr != nil {
			log.Println(writeErr)
		}
	}

	return true, nil
}

// SetUpInitialBranch sets up the initial branch for a JIT repository.
//
// This function is responsible for creating the initial branch in the JIT repository and
// updating the HEAD file to point to this new branch. It takes the directory of the JIT
// repository (jitDir) and the name of the initial branch (initialBranch) as arguments.
//
// Args:
//
//	jitDir (string): The directory where the JIT repository is located.
//	initialBranch (string): The name of the initial branch to set up in the repository.
//
// Returns:
//
//	ok (bool): A boolean value indicating whether the initial branch setup was successful.
//	err (error): An error object that captures any issues encountered during the setup process.
//	             If the setup is successful, err will be nil.
//
// The function performs the following steps:
//  1. It constructs the path for the new branch file within the JIT repository.
//  2. It opens (or creates if not existent) the branch file with append mode and default file permissions.
//  3. It also constructs the path for the HEAD file within the JIT repository.
//  4. It opens (or creates if not existent) the HEAD file with read-write mode, creates the file if it does
//     not exist, and truncates it if it does.
//  5. It then writes the path of the branch file into the HEAD file, effectively pointing HEAD to the new branch.
//
// Usage:
//
//	ok, err := SetUpInitialBranch("/path/to/jit/repo", "main")
//	if err != nil {
//	    log.Fatalf("Failed to set up initial branch: %s", err)
//	}
//
// Note:
//   - The function uses `os.OpenFile` to handle file operations which ensures that the files
//     are created if they do not exist.
//   - Proper error handling is implemented to catch and return errors encountered during file
//     operations.
//   - The function ensures file resources are properly closed using deferred Close calls.
func SetUpInitialBranch(jitDir string, initialBranch string) (ok bool, err error) {

	branchPath := filepath.Join(jitDir, util.BRANCHES, initialBranch)
	bf, openBranchErr := os.OpenFile(branchPath, os.O_APPEND|os.O_CREATE, util.DefaultFilePerm)
	defer func() {
		_ = bf.Close()
	}()
	if openBranchErr != nil {
		return false, openBranchErr
	}

	headPath := filepath.Join(jitDir, util.HEAD)
	hf, openHeadErr := os.OpenFile(headPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, util.DefaultFilePerm)
	defer func() {
		_ = hf.Close()
	}()
	if openHeadErr != nil {
		return false, openHeadErr
	}

	if _, writeErr := hf.WriteString(bf.Name()); writeErr != nil {
		return false, writeErr
	}

	return true, nil
}
