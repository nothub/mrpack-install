package api

import (
	"errors"
	"net/url"
)

const apiVersion = "v2"

func (client *ApiClient) LabrinthInfo() (*LabrinthInfo, error) {
	u, err := url.Parse(client.BaseUrl)
	if err != nil {
		return nil, err
	}

	labrinthInfo := LabrinthInfo{}
	err = client.Http.GetJson(u.String(), &labrinthInfo, &Error{})
	if err != nil {
		return nil, err
	}

	return &labrinthInfo, nil
}

/* projects */

// GetProject https://docs.modrinth.com/api-spec/#tag/projects/operation/getProject
func (client *ApiClient) GetProject(id string) (*Project, error) {
	u, err := url.Parse(client.BaseUrl + apiVersion + "/project/" + id)
	if err != nil {
		return nil, err
	}

	project := Project{}
	err = client.Http.GetJson(u.String(), &project, &Error{})
	if err != nil {
		return nil, err
	}

	return &project, nil
}

// GetProjects https://docs.modrinth.com/api-spec/#tag/projects/operation/getProjects
func (client *ApiClient) GetProjects(ids []string) ([]*Project, error) {
	u, err := url.Parse(client.BaseUrl + apiVersion + "/projects")
	if err != nil {
		return nil, err
	}

	query := url.Values{}
	query.Add("ids", arrayAsParam(ids))
	u.RawQuery = query.Encode()

	var projects []*Project
	err = client.Http.GetJson(u.String(), &projects, &Error{})
	if err != nil {
		return nil, err
	}

	return projects, nil
}

// CheckProjectValidity https://docs.modrinth.com/api-spec/#tag/projects/operation/checkProjectValidity
func (client *ApiClient) CheckProjectValidity(id string) (*CheckResponse, error) {
	u, err := url.Parse(client.BaseUrl + apiVersion + "/project/" + id + "/check")
	if err != nil {
		return nil, err
	}

	var checkResponse CheckResponse
	err = client.Http.GetJson(u.String(), &checkResponse, &Error{})
	if err != nil {
		return nil, err
	}

	return &checkResponse, nil
}

// GetDependencies https://docs.modrinth.com/api-spec/#tag/projects/operation/getDependencies
func (client *ApiClient) GetDependencies(id string) (*Dependencies, error) {
	u, err := url.Parse(client.BaseUrl + apiVersion + "/project/" + id + "/dependencies")
	if err != nil {
		return nil, err
	}

	var dependencies Dependencies
	err = client.Http.GetJson(u.String(), &dependencies, &Error{})
	if err != nil {
		return nil, err
	}

	return &dependencies, nil
}

/* versions */

// GetProjectVersions https://docs.modrinth.com/api-spec/#tag/versions/operation/getProjectVersions
func (client *ApiClient) GetProjectVersions(id string, params *GetProjectVersionsParams) ([]*Version, error) {
	u, err := url.Parse(client.BaseUrl + apiVersion + "/project/" + id + "/version")
	if err != nil {
		return nil, err
	}

	if params != nil {
		query := url.Values{}
		if len(params.Loaders) > 0 {
			query.Add("loaders", arrayAsParam(params.Loaders))
		}
		if len(params.GameVersions) > 0 {
			query.Add("game_versions", arrayAsParam(params.GameVersions))
		}
		if params.FeaturedOnly {
			query.Add("featured", "true")
		}
		u.RawQuery = query.Encode()
	}

	var versions []*Version
	err = client.Http.GetJson(u.String(), &versions, &Error{})
	if err != nil {
		return nil, err
	}

	return versions, nil
}

type GetProjectVersionsParams struct {
	Loaders      []string
	GameVersions []string
	FeaturedOnly bool
}

// GetVersion https://docs.modrinth.com/api-spec/#tag/versions/operation/getVersion
func (client *ApiClient) GetVersion(id string) (*Version, error) {
	u, err := url.Parse(client.BaseUrl + apiVersion + "/version/" + id)
	if err != nil {
		return nil, err
	}

	var version Version
	err = client.Http.GetJson(u.String(), &version, &Error{})
	if err != nil {
		return nil, err
	}

	return &version, nil
}

// GetVersions https://docs.modrinth.com/api-spec/#tag/versions/operation/getVersions
func (client *ApiClient) GetVersions(ids []string) ([]*Version, error) {
	u, err := url.Parse(client.BaseUrl + apiVersion + "/versions")
	if err != nil {
		return nil, err
	}

	query := url.Values{}
	query.Add("ids", arrayAsParam(ids))
	u.RawQuery = query.Encode()

	var versions []*Version
	err = client.Http.GetJson(u.String(), &versions, &Error{})
	if err != nil {
		return nil, err
	}

	return versions, nil
}

/* version files */

// VersionFromHash https://docs.modrinth.com/api-spec/#tag/version-files/operation/versionFromHash
func (client *ApiClient) VersionFromHash(hash string, algorithm HashAlgo) (*Version, error) {
	u, err := url.Parse(client.BaseUrl + apiVersion + "/version_file/" + hash)
	if err != nil {
		return nil, err
	}

	query := url.Values{}
	query.Add("algorithm", string(algorithm))
	u.RawQuery = query.Encode()

	var version *Version
	err = client.Http.GetJson(u.String(), &version, &Error{})
	if err != nil {
		return nil, err
	}

	return version, nil
}

// GetLatestGameVersion https://docs.modrinth.com/api-spec/#tag/tags/operation/versionList
func (client *ApiClient) GetLatestGameVersion() (string, error) {
	u, err := url.Parse(client.BaseUrl + apiVersion + "/tag/game_version")
	if err != nil {
		return "", err
	}

	var gameVersions []*GameVersion
	err = client.Http.GetJson(u.String(), &gameVersions, &Error{})
	if err != nil {
		return "", err
	}

	for i := range gameVersions {
		if gameVersions[i].VersionType == ReleaseVersionType {
			return gameVersions[i].Version, nil
		}
	}

	return "", errors.New("no release candidate found")
}
