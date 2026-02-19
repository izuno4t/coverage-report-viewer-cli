package jacoco

type xmlReport struct {
	Name     string       `xml:"name,attr"`
	Packages []xmlPackage `xml:"package"`
	Counters []xmlCounter `xml:"counter"`
}

type xmlPackage struct {
	Name     string       `xml:"name,attr"`
	Classes  []xmlClass   `xml:"class"`
	Counters []xmlCounter `xml:"counter"`
}

type xmlClass struct {
	Name           string       `xml:"name,attr"`
	SourceFileName string       `xml:"sourcefilename,attr"`
	Methods        []xmlMethod  `xml:"method"`
	Counters       []xmlCounter `xml:"counter"`
}

type xmlMethod struct {
	Name     string       `xml:"name,attr"`
	Desc     string       `xml:"desc,attr"`
	Line     int          `xml:"line,attr"`
	Counters []xmlCounter `xml:"counter"`
}

type xmlCounter struct {
	Type    string `xml:"type,attr"`
	Missed  int    `xml:"missed,attr"`
	Covered int    `xml:"covered,attr"`
}

type xmlCoberturaCoverage struct {
	Packages []xmlCoberturaPackage `xml:"packages>package"`
}

type xmlCoberturaPackage struct {
	Name    string              `xml:"name,attr"`
	Classes []xmlCoberturaClass `xml:"classes>class"`
}

type xmlCoberturaClass struct {
	Name    string                 `xml:"name,attr"`
	File    string                 `xml:"filename,attr"`
	Methods []xmlCoberturaMethod   `xml:"methods>method"`
	Lines   []xmlCoberturaLineNode `xml:"lines>line"`
}

type xmlCoberturaMethod struct {
	Name      string                 `xml:"name,attr"`
	Signature string                 `xml:"signature,attr"`
	Lines     []xmlCoberturaLineNode `xml:"lines>line"`
}

type xmlCoberturaLineNode struct {
	Number            int    `xml:"number,attr"`
	Hits              int    `xml:"hits,attr"`
	Branch            string `xml:"branch,attr"`
	ConditionCoverage string `xml:"condition-coverage,attr"`
}
