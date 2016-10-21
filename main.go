package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type SyncConfig struct {
	Mappings []Mapping
}

type Mapping struct {
	SourceRoot      string   `json:"srcRoot"`
	DestinationRoot string   `json:"dstRoot"`
	IncludedFiles   string   `json:"files"`
	IgnoredFiles    string   `json:"ignored"`
	CleanupPatterns []string `json:"cleanupPatterns"`
}

type RunConfig struct {
	ConfigFilePath string
	RunOnce        bool
	Verbose        bool
	NoClean        bool
	Tick           int
}

type WatchedFile struct {
	Path         string
	AbsolutePath string
	Destination  string
	LastChanged  time.Time
	Hash         string
	NeedsSync    bool
	Alive        bool
}

type FileWatchList map[string]*WatchedFile

var runConfig RunConfig
var watchList FileWatchList = make(FileWatchList, 0)

func main() {
	runConfig = parseCommandlineFlags()
	if runConfig.ConfigFilePath == "" {
		fmt.Print("\nError: No config file has been specified.")
		fmt.Print("\nUse -c flag to supply a valid config file or -h to see available options.")
		fmt.Print("\n\n*** sync has stopped ***")
		os.Exit(1)
	}
	filesyncConfig := loadConfigWithPath(getAbsPathRelativeToProcessDir(runConfig.ConfigFilePath))
	validateAndCreateDestRootPathes(filesyncConfig)
	if !runConfig.NoClean {
		doCleanup(filesyncConfig)
	}
	fmt.Println("Successfully started and running...")
	if runConfig.RunOnce {
		handleTick(filesyncConfig)
	} else {
		startMainLoop(filesyncConfig)
	}
}

func parseCommandlineFlags() (runConfig RunConfig) {
	flag.StringVar(&runConfig.ConfigFilePath, "c", "", "Set the path to the config file to use")
	flag.BoolVar(&runConfig.RunOnce, "o", false, "Set if the srcipt should only run once and exit aftewards")
	flag.BoolVar(&runConfig.Verbose, "v", false, "Enable logging to the terminal")
	flag.BoolVar(&runConfig.NoClean, "no-clean", false, "Set if no cleanup on startup shall be performed")
	flag.IntVar(&runConfig.Tick, "tick", 3000, "The intervall in ms between scans")
	flag.Parse()
	return runConfig
}

func validateAndCreateDestRootPathes(config *SyncConfig) {
	for _, mapping := range config.Mappings[:] {
		if !doesDirOrFileExist(mapping.DestinationRoot) {
			fmt.Printf("Destination root directory not found. Creating: %v \n", mapping.DestinationRoot)
			if err := os.MkdirAll(mapping.DestinationRoot, 0666); err != nil {
				logScanDirError(err)
			}
		}
	}
}

func getAbsPathRelativeToProcessDir(filePath string) string {
	if !filepath.IsAbs(filePath) {
		currentDir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
		return filepath.Join(currentDir, filePath)
	}
	return filePath
}

func startMainLoop(config *SyncConfig) {
	tickDuration := time.Duration(runConfig.Tick)
	for {
		handleTick(config)
		time.Sleep(tickDuration * time.Millisecond)
	}
}

func handleTick(config *SyncConfig) {
	markAndSweepWatchList()
	scanAndCoppyMappedDirs(config)
}

func markAndSweepWatchList() {
	for key, watchedFile := range watchList {
		if watchedFile.Alive {
			watchedFile.Alive = false
		} else {
			delete(watchList, key)
			err := os.Remove(watchedFile.Destination)
			if err == nil {
				logFileDelete(watchedFile)
			} else {
				logFileDeleteError(watchedFile, err)
			}
		}
	}
}

func deleteFile(filePath string) bool {
	err := os.Remove(filePath)
	if err != nil {
		return false
	}
	return true
}

func scanAndCoppyMappedDirs(config *SyncConfig) {
	for index, mapping := range config.Mappings[:] {
		err := buildFileWatchListForDirectory(index, mapping)
		if err != nil {
			panic(err)
		}
	}
	for _, watchedFile := range watchList {
		doCopy(watchedFile)
	}
}

func doCopy(watchedFile *WatchedFile) {
	if watchedFile.NeedsSync {
		if doesDirOrFileExist(watchedFile.AbsolutePath) {
			logFileCopy(watchedFile)
			if watchedFile.AbsolutePath != "" && watchedFile.Destination != "" {
				os.MkdirAll(filepath.Dir(watchedFile.Destination), 0666)
				err := copyFile(watchedFile.AbsolutePath, watchedFile.Destination)
				if err != nil {
					logScanDirError(err)
				}
				watchedFile.NeedsSync = false
			}
		} else {
			logFileNotFound(watchedFile)
		}
	}

}

func logFileCopy(watchedFile *WatchedFile) {
	if runConfig.Verbose {
		fmt.Printf("Copy from: %v \n", watchedFile.AbsolutePath)
		fmt.Printf("       To: %v\n", watchedFile.Destination)
	}
}

func logFileNotFound(watchedFile *WatchedFile) {
	if runConfig.Verbose {
		fmt.Printf("### Error could not open/find: %v\n", watchedFile.AbsolutePath)
	}
}

func logFileDelete(watchedFile *WatchedFile) {
	if runConfig.Verbose {
		fmt.Printf("  Deleted:%v\n", watchedFile.Destination)
	}
}

func logFileDeleteError(watchedFile *WatchedFile, err error) {
	fmt.Printf("### Error deleting: %v\n%v\n", watchedFile.Destination, err)
}

func logScanDirError(err error) {
	fmt.Printf("### Error: %v\n", err)
}

func loadConfigWithPath(path string) *SyncConfig {
	fmt.Println("Loading config: ", path)
	configFile := readFile(path)
	if configFile != nil {
		config := &SyncConfig{}
		err := json.Unmarshal(configFile, config)
		if err != nil {
			fmt.Println("Error: Could not read config file. Check if valid JSON.")
			fmt.Println("*** sync has stopped ***")
			os.Exit(1)
		}
		return config
	}
	return nil
}

func readFile(fileName string) []byte {
	file, err := ioutil.ReadFile(fileName)
	if err != nil {
		fmt.Println("Could not read file", fileName)
		panic(err)
	}
	return file
}

func doesDirOrFileExist(path string) bool {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false
		} else {
			return true
		}
	}
	return true
}

func buildSourceDirNotFoundErr(index int, mapping Mapping) error {
	m := fmt.Sprintf("Source dir could not be found.\nMapping index: %v\nSource path: %v", index, mapping.SourceRoot)
	return errors.New(m)
}

func buildFileWatchListForDirectory(index int, mapping Mapping) error {
	if !doesDirOrFileExist(getAbsPathRelativeToProcessDir(mapping.SourceRoot)) {
		return buildSourceDirNotFoundErr(index, mapping)
	}
	callback := func(path string, fileInfo os.FileInfo, err error) error {
		if !fileInfo.IsDir() {
			absoluteSrcFilepath, _ := filepath.Abs(path)
			pathRelativeToSource, _ := filepath.Rel(mapping.SourceRoot, path)
			destinationFilepath := filepath.Join(mapping.DestinationRoot, pathRelativeToSource)
			key := fmt.Sprintf("%v-%v", index, destinationFilepath)
			if fileShouldBeSynced(absoluteSrcFilepath, mapping) {
				//check if file already was discovered and is in the watchList
				watcher, ok := watchList[key]
				if !ok {
					watcher = &WatchedFile{}
					watcher.Path = path
					watcher.AbsolutePath = absoluteSrcFilepath
					watcher.Destination = destinationFilepath
				}
				fileHash := generateSha256HashForFile(absoluteSrcFilepath)
				if watcher.Hash != fileHash {
					watcher.Hash = fileHash
					watcher.NeedsSync = true
				}
				watcher.LastChanged = fileInfo.ModTime()
				watcher.Alive = true
				watchList[key] = watcher
			}
		}
		return nil
	}
	return filepath.Walk(mapping.SourceRoot, callback)
}

func fileShouldBeSynced(fileName string, mapping Mapping) bool {
	if isDotFile(fileName) {
		return false
	}
	if matchesIgnoredFileConfig(fileName, mapping) {
		return false
	} else if matchesIncludedFileConfig(fileName, mapping) {
		return true
	} else {
		return false
	}
}

func isDotFile(path string) bool {
	_, fileName := filepath.Split(path)
	if strings.HasPrefix(fileName, ".") {
		return true
	}
	return false
}

func matchesIncludedFileConfig(fileName string, mapping Mapping) bool {
	if mapping.IncludedFiles == "" {
		return true
	}
	return patternMatchtesFilepath(mapping.IncludedFiles, fileName)
}

func matchesIgnoredFileConfig(fileName string, mapping Mapping) bool {
	if mapping.IgnoredFiles == "" {
		return false
	}
	return patternMatchtesFilepath(mapping.IgnoredFiles, fileName)
}

func generateSha256HashForFile(file string) string {
	hasher := sha256.New()
	hasher.Write(readFile(file))
	return base64.URLEncoding.EncodeToString(hasher.Sum(nil))
}

func patternMatchtesFilepath(pattern string, fileName string) bool {
	parts := strings.Split(pattern, "/")
	// replace windows backslashes for pattern matching to work
	fileName = strings.Replace(fileName, "\\", "/", -1)
	if len(parts) > 1 {
		if containsExactDirPattern(parts) {
			if strings.Contains(fileName, getPatternWithoutWildcards(parts)) {
				if isFileWildCardPattern(parts[len(parts)-1]) {
					_, fileName = filepath.Split(fileName)
					result, _ := filepath.Match(parts[len(parts)-1], fileName)
					return result
				}
				return true
			}
		} else {
			// if dir wildcard extract and only use filename for match
			if isDirWildCardPattern(pattern) {
				pattern = parts[len(parts)-1]
				_, fileName = filepath.Split(fileName)
			}
			result, _ := filepath.Match(pattern, fileName)
			return result
		}
	}
	result, _ := filepath.Match(pattern, fileName)
	return result
}

func containsExactDirPattern(patterns []string) bool {
	containsExactDirs := false
	for _, pattern := range patterns[:] {
		if !isDirWildCardPattern(pattern) {
			if !isFileWildCardPattern(pattern) {
				containsExactDirs = true
			}
		}
	}
	return containsExactDirs
}

func isDirWildCardPattern(pattern string) bool {
	return strings.HasPrefix(pattern, "**")
}

func isFileWildCardPattern(pattern string) bool {
	return strings.HasPrefix(pattern, "*.")
}

func getPatternWithoutWildcards(patterns []string) string {
	result := make([]string, 0)
	for _, pattern := range patterns[:] {
		if !isDirWildCardPattern(pattern) {
			if !isFileWildCardPattern(pattern) {
				result = append(result, pattern)
			}
		}
	}
	return strings.Join(result, "/")
}

var cleanupList []string = make([]string, 0)

func doCleanup(syncConfig *SyncConfig) {
	for _, mapping := range syncConfig.Mappings[:] {
		buildDirectoryCleanupList(mapping)
	}
	for _, cleanupDir := range cleanupList {
		err := os.RemoveAll(cleanupDir)
		if err != nil {
			fmt.Printf("#Error could not delete: %v\n%v\n", cleanupDir, err)
		}
		if runConfig.Verbose {
			fmt.Println("  Cleaned:", cleanupDir)
		}
	}
}

func buildDirectoryCleanupList(mapping Mapping) error {
	callback := func(path string, fileInfo os.FileInfo, err error) error {
		if fileInfo.IsDir() {
			for _, pattern := range mapping.CleanupPatterns[:] {
				if len(pattern) > 0 && patternMatchtesFilepath(pattern, path) {
					cleanupList = append(cleanupList, path)
				}
			}
		}
		return nil
	}
	return filepath.Walk(mapping.DestinationRoot, callback)
}

func copyFile(sourcePath, destinationPath string) error {
	if doesDirOrFileExist(sourcePath) {
		sourceFile, err := os.Open(sourcePath)
		if err != nil {
			return err
		}
		defer sourceFile.Close()

		destinationFile, err := os.Create(destinationPath)
		if err != nil {
			return err
		}
		defer destinationFile.Close()

		_, err = io.Copy(destinationFile, sourceFile)
		if err != nil {
			return err
		}
	}

	return nil
}
