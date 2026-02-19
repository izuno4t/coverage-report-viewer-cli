package reportpath

import (
	"encoding/xml"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	mavenDefaultReportDir  = "${project.reporting.outputDirectory}/jacoco"
	mavenDefaultSiteDir    = "target/site"
	mavenDefaultBuildDir   = "target"
	defaultMavenReportPath = "target/site/jacoco/jacoco.xml"
	defaultGradlePath      = "build/reports/jacoco/test/jacocoTestReport.xml"
)

var errReportNotFound = errors.New("jacoco report xml not found")

// Detect finds a JaCoCo XML report path following REQUIREMENTS.md resolution order.
func Detect(cwd string) (string, error) {
	paths, err := DetectAll(cwd)
	if err != nil {
		return "", err
	}
	return paths[0], nil
}

// DetectAll finds JaCoCo XML report paths including maven multi-module projects.
func DetectAll(cwd string) ([]string, error) {
	pomPath := filepath.Join(cwd, "pom.xml")
	if fileExists(pomPath) {
		visited := map[string]struct{}{}
		detected := detectFromPOMTree(cwd, pomPath, visited)
		if len(detected) > 0 {
			return uniquePaths(detected), nil
		}
	}

	fallback := fallbackCandidates(cwd)
	if len(fallback) > 0 {
		return fallback, nil
	}

	return nil, fmt.Errorf("%w (tried: pom.xml, %s, %s)", errReportNotFound, defaultMavenReportPath, defaultGradlePath)
}

func detectFromPOMTree(cwd, pomPath string, visited map[string]struct{}) []string {
	absPom, err := filepath.Abs(pomPath)
	if err != nil {
		absPom = pomPath
	}
	if _, ok := visited[absPom]; ok {
		return nil
	}
	visited[absPom] = struct{}{}

	project, ok := parsePOM(pomPath)
	if !ok {
		return nil
	}

	paths := make([]string, 0)
	if detected, ok := detectFromProject(cwd, project); ok {
		paths = append(paths, detected)
	} else {
		paths = append(paths, fallbackCandidates(cwd)...)
	}

	for _, mod := range project.Modules {
		module := strings.TrimSpace(mod)
		if module == "" {
			continue
		}
		moduleDir := filepath.Clean(filepath.Join(cwd, module))
		modulePom := filepath.Join(moduleDir, "pom.xml")
		if fileExists(modulePom) {
			paths = append(paths, detectFromPOMTree(moduleDir, modulePom, visited)...)
			continue
		}
		paths = append(paths, fallbackCandidates(moduleDir)...)
	}

	return uniquePaths(paths)
}

func detectFromProject(cwd string, project pomProject) (string, bool) {
	plugins := make([]pomPlugin, 0, len(project.Build.Plugins)+len(project.Build.PluginManagement.Plugins))
	plugins = append(plugins, project.Build.Plugins...)
	plugins = append(plugins, project.Build.PluginManagement.Plugins...)

	for _, p := range plugins {
		if p.GroupID != "org.jacoco" || p.ArtifactID != "jacoco-maven-plugin" {
			continue
		}
		reportDir := resolveReportDir(p)
		reportDir = resolveMavenPlaceholders(reportDir)
		candidate := filepath.Join(cwd, reportDir, "jacoco.xml")
		if fileExists(candidate) {
			return candidate, true
		}
	}
	return "", false
}

func parsePOM(pomPath string) (pomProject, bool) {
	content, err := os.ReadFile(pomPath)
	if err != nil {
		return pomProject{}, false
	}
	var project pomProject
	if err := xml.Unmarshal(content, &project); err != nil {
		return pomProject{}, false
	}
	return project, true
}

func fallbackCandidates(cwd string) []string {
	found := make([]string, 0, 2)
	for _, rel := range []string{defaultMavenReportPath, defaultGradlePath} {
		candidate := filepath.Join(cwd, rel)
		if fileExists(candidate) {
			found = append(found, candidate)
		}
	}
	return found
}

func uniquePaths(paths []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(paths))
	for _, path := range paths {
		if _, ok := seen[path]; ok {
			continue
		}
		seen[path] = struct{}{}
		out = append(out, path)
	}
	return out
}

func resolveReportDir(plugin pomPlugin) string {
	for _, ex := range plugin.Executions {
		if !hasGoal(ex.Goals, "report") {
			continue
		}
		if ex.Configuration.OutputDirectory != "" {
			return ex.Configuration.OutputDirectory
		}
	}
	if plugin.Configuration.OutputDirectory != "" {
		return plugin.Configuration.OutputDirectory
	}
	return mavenDefaultReportDir
}

func resolveMavenPlaceholders(path string) string {
	replacer := strings.NewReplacer(
		"${project.reporting.outputDirectory}", mavenDefaultSiteDir,
		"${project.build.directory}", mavenDefaultBuildDir,
	)
	return replacer.Replace(path)
}

func hasGoal(goals []string, want string) bool {
	for _, g := range goals {
		if strings.TrimSpace(g) == want {
			return true
		}
	}
	return false
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}
