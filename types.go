package main

type Mod struct {
	Slug string
}

type MavenMetadata struct {
	Versioning struct {
		Latest   string   `xml:"latest"`
		Release  string   `xml:"release"`
		Versions struct {
			Version []string `xml:"version"`
		} `xml:"versions"`
	} `xml:"versioning"`
}

type ModProvider interface {
	FetchMod(slug, mcVersion, loader string) (url, filename string, err error)
}
