package maven

import (
	"hub.lol/mrpack-install/web"
)

type Metadata struct {
	GroupId    string `xml:"groupId"`
	ArtifactId string `xml:"artifactId"`
	Versioning struct {
		Latest      string   `xml:"latest"`
		Release     string   `xml:"release"`
		Versions    []string `xml:"versions>version"`
		LastUpdated string   `xml:"lastUpdated"` // TODO: use Time type
	} `xml:"versioning"`
}

func FetchMetadata(url string) (*Metadata, error) {
	m := &Metadata{}
	err := web.DefaultClient.GetXml(url, m, nil)
	if err != nil {
		return nil, err
	}
	return m, nil
}
