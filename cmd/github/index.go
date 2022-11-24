package github

import "time"

type IndexYaml struct {
	APIVersion string    `yaml:"apiVersion"`
	Entries    Entries   `yaml:"entries"`
	Generated  time.Time `yaml:"generated"`
}
type FactorycubeEdge struct {
	APIVersion  string    `yaml:"apiVersion"`
	AppVersion  string    `yaml:"appVersion"`
	Created     time.Time `yaml:"created"`
	Description string    `yaml:"description"`
	Digest      string    `yaml:"digest"`
	Home        string    `yaml:"home"`
	Name        string    `yaml:"name"`
	Sources     []string  `yaml:"sources"`
	Type        string    `yaml:"type"`
	Urls        []string  `yaml:"urls"`
	Version     string    `yaml:"version"`
}
type Dependencies struct {
	Condition  string `yaml:"condition"`
	Name       string `yaml:"name"`
	Repository string `yaml:"repository"`
	Version    string `yaml:"version"`
}
type FactorycubeServer struct {
	APIVersion   string         `yaml:"apiVersion"`
	AppVersion   string         `yaml:"appVersion"`
	Created      time.Time      `yaml:"created"`
	Dependencies []Dependencies `yaml:"dependencies"`
	Description  string         `yaml:"description"`
	Digest       string         `yaml:"digest"`
	Home         string         `yaml:"home"`
	Name         string         `yaml:"name"`
	Sources      []string       `yaml:"sources"`
	Type         string         `yaml:"type"`
	Urls         []string       `yaml:"urls"`
	Version      string         `yaml:"version"`
}
type UnitedManufacturingHub struct {
	APIVersion   string         `yaml:"apiVersion"`
	AppVersion   string         `yaml:"appVersion"`
	Created      time.Time      `yaml:"created"`
	Dependencies []Dependencies `yaml:"dependencies"`
	Description  string         `yaml:"description"`
	Digest       string         `yaml:"digest"`
	Home         string         `yaml:"home"`
	Icon         string         `yaml:"icon"`
	Name         string         `yaml:"name"`
	Sources      []string       `yaml:"sources"`
	Type         string         `yaml:"type"`
	Urls         []string       `yaml:"urls"`
	Version      string         `yaml:"version"`
}
type Entries struct {
	FactorycubeEdge        []FactorycubeEdge        `yaml:"factorycube-edge"`
	FactorycubeServer      []FactorycubeServer      `yaml:"factorycube-server"`
	UnitedManufacturingHub []UnitedManufacturingHub `yaml:"united-manufacturing-hub"`
}
