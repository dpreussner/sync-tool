package main

import "testing"

func TestIsDirWildcardPattern(t *testing.T) {
	if !isDirWildCardPattern("**") {
		t.Errorf("Should match as dir wildcard pattern '%v'", "**")
	}
	if isDirWildCardPattern("*") {
		t.Errorf("Should not match as dir wildcard pattern '%v'", "*")
	}
	if isDirWildCardPattern("test") {
		t.Errorf("Should not match as dir wildcard pattern '%v'", "test")
	}
	if isDirWildCardPattern("") {
		t.Errorf("Should not match as dir wildcard pattern '%v'", "")
	}
}

func TestIsFileWildcardPattern(t *testing.T) {
	if !isFileWildCardPattern("*.") {
		t.Errorf("Should match as file wildcard pattern '%v'", "**")
	}
	if !isFileWildCardPattern("*.*") {
		t.Errorf("Should match as file wildcard pattern '%v'", "*.*")
	}
	if !isFileWildCardPattern("*.test") {
		t.Errorf("Should match as file wildcard pattern '%v'", "*.test")
	}
	if isFileWildCardPattern("**") {
		t.Errorf("Should not match as file wildcard pattern '%v'", "**")
	}
	if isFileWildCardPattern("test") {
		t.Errorf("Should not match as file wildcard pattern '%v'", "test")
	}
	if isFileWildCardPattern("") {
		t.Errorf("Should not match as file wildcard pattern '%v'", "")
	}
}

func TestContainsExactDirPattern(t *testing.T) {
	var pattern []string

	pattern = []string{"**", "test", "*.*"}
	if !containsExactDirPattern(pattern) {
		t.Errorf("Should match pattern  %v", pattern)
	}

	pattern = []string{"**", "*.*"}
	if containsExactDirPattern(pattern) {
		t.Errorf("Should not match pattern %v", pattern)
	}
}

func TestGetPatternWithoutWildcards(t *testing.T) {
	var pattern []string

	pattern = []string{"**", "test", "*.*"}
	if getPatternWithoutWildcards(pattern) != "test" {
		t.Errorf("Should match pattern '%v' %v", "test", getPatternWithoutWildcards(pattern))
	}

	pattern = []string{"**", "test", "abc", "*.*"}
	if getPatternWithoutWildcards(pattern) != "test/abc" {
		t.Errorf("Should match pattern '%v' %v", "test/abc", getPatternWithoutWildcards(pattern))
	}

	pattern = []string{"**", "test", "abc", "test.css"}
	if getPatternWithoutWildcards(pattern) != "test/abc/test.css" {
		t.Errorf("Should match pattern '%v' %v", "test/abc/test.css", getPatternWithoutWildcards(pattern))
	}

	pattern = []string{"**", "test.css"}
	if getPatternWithoutWildcards(pattern) != "test.css" {
		t.Errorf("Should match pattern '%v' %v", "test.css", getPatternWithoutWildcards(pattern))
	}
}

func TestPatternMatchtesFilepath(t *testing.T) {
	if !patternMatchtesFilepath("**/*.*", "test/config.json") {
		t.Errorf("Did not match as expected '%v' '%v'", "**/*.*", "test/config.json")
	}
	if !patternMatchtesFilepath("**/*.*", "test/test/config.json") {
		t.Errorf("Did not match as expected '%v' '%v'", "**/*.*", "test/test/config.json")
	}
	if !patternMatchtesFilepath("**/*.*", "config.json") {
		t.Errorf("Did not match as expected '%v' '%v'", "**/*.*", "config.json")
	}
	if !patternMatchtesFilepath("*.*", "config.json") {
		t.Errorf("Did not match as expected '%v' '%v'", "*.*", "config.json")
	}
	if !patternMatchtesFilepath("*.json", "config.json") {
		t.Errorf("Did not match as expected '%v' '%v'", "*.json", "config.json")
	}
	if !patternMatchtesFilepath("config.json", "config.json") {
		t.Errorf("Did not match as expected '%v' '%v'", "config.json", "config.json")
	}
	if !patternMatchtesFilepath("**/config.json", "G:/config.json") {
		t.Errorf("Did not match as expected '%v' '%v'", "**/config.json", "G:/config.json")
	}
	if !patternMatchtesFilepath("*.css", "config.test.css") {
		t.Errorf("Did not match as expected '%v' '%v'", "*.css", "config.test.css")
	}
	if !patternMatchtesFilepath("*.css", "aaa/config.tpl.css") {
		t.Errorf("Did not match as expected '%v' '%v'", "*.css", "aaa/config.tpl.css")
	}
	if !patternMatchtesFilepath("**/*.css", "G:/aaa/config.test.css") {
		t.Errorf("Did not match as expected '%v' '%v'", "**/*.css", "config.test.css")
	}

	if !patternMatchtesFilepath("modules/samco/css", "modules/samco/css/test.css") {
		t.Errorf("Did not match as expected '%v' '%v'", "modules/samco/css", "modules/samco/css/test.css")
	}

	if !patternMatchtesFilepath("modules/samco/css/test.css", "modules/samco/css/test.css") {
		t.Errorf("Did not match as expected '%v' '%v'", "modules/samco/css/test.css", "modules/samco/css/test.css")
	}

	if !patternMatchtesFilepath("modules/samco/css/*.*", "modules/samco/css/test.css") {
		t.Errorf("Did not match as expected '%v' '%v'", "modules/samco/css/*.*", "modules/samco/css/test.css")
	}

	if patternMatchtesFilepath("modules/samco/css/*.txt", "modules/samco/css/test.css") {
		t.Errorf("Did not match as expected '%v' '%v'", "modules/samco/css/*.txt", "modules/samco/css/test.css")
	}

	if patternMatchtesFilepath("**/*.*", "config") {
		t.Error("Did not match config with filepath as expected")
	}
	if patternMatchtesFilepath("**/*.*", "") {
		t.Error("Did not match config with filepath as expected")
	}
	if patternMatchtesFilepath("*.*", "") {
		t.Error("Did not match config with filepath as expected")
	}
}

func TestConfig(t *testing.T) {
	config := loadConfigWithPath("test/config/filesync.json")
	if config == nil {
		t.Error("Failed to load the test config")
	}
	//check config parsing
	if config.Mappings == nil {
		t.Error("Mappings should not be empty")
	}
	if len(config.Mappings) != 4 {
		t.Error("Not all mappings have been parsed")
	}
	//test mapping
	mapping := config.Mappings[0]
	if mapping.SourceRoot != "src/main/resources" {
		t.Error("SourceRoot was not parsed correctly")
	}
	if mapping.DestinationRoot != "G:/domains/test-app/apps/staging/frontendapp/frontendapp/etv-webapp/test-app/WEB-INF/classes" {
		t.Error("DestinationRoot was not parsed correctly")
	}
	if mapping.IncludedFiles != "**/*.xml" {
		t.Error("IncludedFiles was not parsed correctly")
	}
	if mapping.IgnoredFiles != "" {
		t.Error("IgnoredFiles was not parsed correctly")
	}
	if len(mapping.CleanupPatterns) != 1 {
		t.Error("CleanupPatterns was not parsed correctly")
	}
	cleanupPattern := mapping.CleanupPatterns[0]
	if cleanupPattern != "**/test" {
		t.Error("CleanupPatterns was not parsed correctly")
	}
}

func TestFileShouldBeSynced(t *testing.T) {
	config := loadConfigWithPath("test/config/filesync.json")

	if !fileShouldBeSynced("G:/aaa/t.tpl.xml", config.Mappings[0]) {
		t.Error("File should be synced")
	}

	if fileShouldBeSynced("G:/aaa/t.tpl.css", config.Mappings[0]) {
		t.Error("File should be synced")
	}

	if !fileShouldBeSynced("G:/aaa/t.tpl.txt", config.Mappings[2]) {
		t.Error("File should be synced")
	}

	if fileShouldBeSynced("G:/aaa/t.tpl.css", config.Mappings[2]) {
		t.Error("File should not be synced")
	}

	if fileShouldBeSynced("G:\\aaa\\t.tpl.css", config.Mappings[2]) {
		t.Error("File should not be synced")
	}

	if fileShouldBeSynced("G:\\aaa\\.git", config.Mappings[2]) {
		t.Error("Dot file should not be synced")
	}
}

func TestDirExists(t *testing.T) {
	if doesDirOrFileExist("bli") {
		t.Error("Path does not exist. Should not return true")
	}
	if doesDirOrFileExist("/bli/bla") {
		t.Error("Path does not exist. Should not return true")
	}
	if doesDirOrFileExist("bli/bla.txt") {
		t.Error("Path does not exist. Should not return true")
	}

	if !doesDirOrFileExist("test/config/filesync.json") {
		t.Error("Path does not exist. Should  return true")
	}

	if !doesDirOrFileExist("test/config/") {
		t.Error("Path does exist. Should return true")
	}

}
