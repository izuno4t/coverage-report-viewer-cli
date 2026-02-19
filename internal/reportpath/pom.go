package reportpath

type pomProject struct {
	Modules []string `xml:"modules>module"`
	Build   pomBuild `xml:"build"`
}

type pomBuild struct {
	Plugins          []pomPlugin `xml:"plugins>plugin"`
	PluginManagement struct {
		Plugins []pomPlugin `xml:"plugins>plugin"`
	} `xml:"pluginManagement"`
}

type pomPlugin struct {
	GroupID       string           `xml:"groupId"`
	ArtifactID    string           `xml:"artifactId"`
	Configuration pomConfiguration `xml:"configuration"`
	Executions    []pomExecution   `xml:"executions>execution"`
}

type pomExecution struct {
	Goals         []string         `xml:"goals>goal"`
	Configuration pomConfiguration `xml:"configuration"`
}

type pomConfiguration struct {
	OutputDirectory string `xml:"outputDirectory"`
	DataFile        string `xml:"dataFile"`
}
