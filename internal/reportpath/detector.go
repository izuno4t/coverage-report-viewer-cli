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
	pomPath := filepath.Join(cwd, "pom.xml")
	if fileExists(pomPath) {
		if detected, ok := detectFromPOM(cwd, pomPath); ok {
			return detected, nil
		}
	}

	for _, rel := range []string{defaultMavenReportPath, defaultGradlePath} {
		candidate := filepath.Join(cwd, rel)
		if fileExists(candidate) {
			return candidate, nil
		}
	}

	return "", fmt.Errorf("%w (tried: pom.xml, %s, %s)", errReportNotFound, defaultMavenReportPath, defaultGradlePath)
}

func detectFromPOM(cwd, pomPath string) (string, bool) {
	content, err := os.ReadFile(pomPath)
	if err != nil {
		return "", false
	}

	var project pomProject
	if err := xml.Unmarshal(content, &project); err != nil {
		return "", false
	}

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
