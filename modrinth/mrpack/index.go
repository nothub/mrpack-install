package mrpack

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"github.com/nothub/mrpack-install/web/download"
	"io"
	"log"
	"net/url"
	"strings"
)

import modrinth "github.com/nothub/mrpack-install/modrinth/api"

type Index struct {
	Format  int    `json:"formatVersion"`
	Game    Game   `json:"game"`
	Version string `json:"versionId"`
	Name    string `json:"name"`
	Summary string `json:"summary"`
	Files   []File `json:"files"`
	Deps    Deps   `json:"dependencies"`
}

func (index *Index) ServerDls() []File {
	var dls []File
	for _, f := range index.Files {
		if f.Env.Server == modrinth.UnsupportedEnvSupport {
			continue
		}
		dls = append(dls, f)
	}
	return dls
}

type Game string

const (
	Minecraft Game = "minecraft"
)

type File struct {
	Path      string          `json:"path"`
	Hashes    modrinth.Hashes `json:"hashes"`
	Env       Env             `json:"env"`
	Downloads []string        `json:"downloads"` // array of HTTPS URLs
	FileSize  int             `json:"fileSize"`  // size in bytes
}

type Env struct {
	Client modrinth.EnvSupport `json:"client"`
	Server modrinth.EnvSupport `json:"server"`
}

type Deps struct {
	Minecraft string `json:"minecraft"`
	Fabric    string `json:"fabric-loader,omitempty"`
	Quilt     string `json:"quilt-loader,omitempty"`
	Forge     string `json:"forge,omitempty"`
	NeoForge  string `json:"neoforge,omitempty"`
}

func ReadIndex(zipFile string) (*Index, error) {
	r, err := zip.OpenReader(zipFile)
	if err != nil {
		return nil, err
	}
	defer func(r *zip.ReadCloser) {
		err := r.Close()
		if err != nil {
			log.Println(err)
		}
	}(r)

	var indexFile *zip.File
	for _, file := range r.File {
		if file.Name == "modrinth.index.json" {
			indexFile = file
			break
		}
	}
	if indexFile == nil {
		return nil, errors.New("no index file found")
	}

	fileReader, err := indexFile.Open()
	if err != nil {
		return nil, err
	}
	defer func(fileReader io.ReadCloser) {
		err := fileReader.Close()
		if err != nil {
			log.Println(err)
		}
	}(fileReader)

	var index Index
	err = json.NewDecoder(fileReader).Decode(&index)
	if err != nil {
		return nil, err
	}

	// https://github.com/modrinth/docs/issues/85 ¯\_(ツ)_/¯
	for i, file := range index.Files {
		if strings.Contains(file.Path, "\\") {
			index.Files[i].Path = strings.ReplaceAll(file.Path, "\\", "/")
			log.Printf("fixed file path: old=%q new=%q\n", file.Path, index.Files[i].Path)
		}
	}

	return &index, nil
}

func (index *Index) ServerDownloads(filter func(f File) bool) []*download.Download {
	var downloads []*download.Download
	projectIds := getProjectIds(index.Files)
	projects, projErr := modrinth.Client.GetProjects(projectIds)
	if projErr != nil {
		projects = nil
		log.Println("Error getting projects! You may see client mods appear.")
	}
	for _, file := range index.Files {
		if projects != nil {
			fid := getFileProjectId(file)
			project := findProjectById(projects, fid)
			if project != nil {
				if project.ServerSide == modrinth.UnsupportedEnvSupport {
					log.Printf("Skipped project %v\n", project.Title)
					continue
				}
			} else {
				log.Printf("Project at %v was not found online! Please verify this is not a client mod.\n", file.Path)
			}
		}

		if file.Env.Server == modrinth.UnsupportedEnvSupport {
			continue
		}

		if !filter(file) {
			continue
		}

		if len(file.Downloads) < 1 {
			log.Printf("No downloads for file: %s\n", file.Path)
			continue
		}

		downloads = append(downloads, &download.Download{
			Path:   file.Path,
			Urls:   file.Downloads,
			Hashes: file.Hashes,
		})
	}
	return downloads
}

func getFileProjectId(file File) string {
	if len(file.Downloads) <= 0 {
		log.Printf("No downloads for file: %s\n", file.Path)
		return ""
	}
	u, urlerr := url.Parse(file.Downloads[0])
	if urlerr != nil {
		log.Printf("Could not parse URL for file: %s\n", file.Path)
		return ""
	}
	// Extract the ID
	urlPath := u.Path
	urlSplit := strings.Split(urlPath, "/")
	// Assume data/{id}/...
	projectId := urlSplit[2]
	return projectId
}

func getProjectIds(files []File) []string {
	ids := make([]string, len(files))
	for i, file := range files {
		ids[i] = getFileProjectId(file)
	}
	return ids
}

func findProjectById(projects []*modrinth.Project, projectId string) *modrinth.Project {
	for _, project := range projects {
		if project.Id == projectId {
			return project
		}
	}
	return nil
}
