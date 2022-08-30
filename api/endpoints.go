package api

import (
	url2 "net/url"
)

const baseUrl = "https://api.modrinth.com/"
const apiVersion = "v2"
const apiUrl = baseUrl + apiVersion

func (client *Client) LabrinthInfo() (*LabrinthInfo, error) {
	url, err := url2.Parse(baseUrl)
	if err != nil {
		return nil, err
	}

	labrinthInfo := LabrinthInfo{}
	err = client.sendRequest("GET", url.String(), nil, &labrinthInfo)
	if err != nil {
		return nil, err
	}

	return &labrinthInfo, nil
}

/* projects */

// GetProject https://docs.modrinth.com/api-spec/#tag/projects/operation/getProject
func (client *Client) GetProject(id string) (*Project, error) {
	url, err := url2.Parse(apiUrl + "/project/" + id)
	if err != nil {
		return nil, err
	}

	project := Project{}
	err = client.sendRequest("GET", url.String(), nil, &project)
	if err != nil {
		return nil, err
	}

	return &project, nil
}

// GetProjects https://docs.modrinth.com/api-spec/#tag/projects/operation/getProjects
func (client *Client) GetProjects(ids []string) ([]*Project, error) {
	url, err := url2.Parse(apiUrl + "/projects")
	if err != nil {
		return nil, err
	}

	query := url2.Values{}
	query.Add("ids", arrayAsParam(ids))
	url.RawQuery = query.Encode()

	var projects []*Project
	err = client.sendRequest("GET", url.String(), nil, &projects)
	if err != nil {
		return nil, err
	}

	return projects, nil
}

// CheckProjectValidity https://docs.modrinth.com/api-spec/#tag/projects/operation/checkProjectValidity
func (client *Client) CheckProjectValidity(id string) (*CheckResponse, error) {
	url, err := url2.Parse(apiUrl + "/project/" + id + "/check")
	if err != nil {
		return nil, err
	}

	var checkResponse CheckResponse
	err = client.sendRequest("GET", url.String(), nil, &checkResponse)
	if err != nil {
		return nil, err
	}

	return &checkResponse, nil
}

// GetDependencies https://docs.modrinth.com/api-spec/#tag/projects/operation/getDependencies
func (client *Client) GetDependencies(id string) (*Dependencies, error) {
	url, err := url2.Parse(apiUrl + "/project/" + id + "/dependencies")
	if err != nil {
		return nil, err
	}

	var dependencies Dependencies
	err = client.sendRequest("GET", url.String(), nil, &dependencies)
	if err != nil {
		return nil, err
	}

	return &dependencies, nil
}

/* versions */

// GetProjectVersions https://docs.modrinth.com/api-spec/#tag/versions/operation/getProjectVersions
func (client *Client) GetProjectVersions(id string, params *GetProjectVersionsParams) ([]*Version, error) {
	url, err := url2.Parse(apiUrl + "/project/" + id + "/version")
	if err != nil {
		return nil, err
	}

	query := url2.Values{}
	if len(params.Loaders) > 0 {
		query.Add("loaders", arrayAsParam(params.Loaders))
	}
	if len(params.GameVersions) > 0 {
		query.Add("game_versions", arrayAsParam(params.GameVersions))
	}
	if params.FeaturedOnly {
		query.Add("featured", "true")
	}
	url.RawQuery = query.Encode()

	var versions []*Version
	err = client.sendRequest("GET", url.String(), nil, &versions)
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
func (client *Client) GetVersion(id string) (*Version, error) {
	url, err := url2.Parse(apiUrl + "/version/" + id)
	if err != nil {
		return nil, err
	}

	var version Version
	err = client.sendRequest("GET", url.String(), nil, &version)
	if err != nil {
		return nil, err
	}

	return &version, nil
}

// GetVersions https://docs.modrinth.com/api-spec/#tag/versions/operation/getVersions
func (client *Client) GetVersions(ids []string) ([]*Version, error) {
	url, err := url2.Parse(apiUrl + "/versions")
	if err != nil {
		return nil, err
	}

	query := url2.Values{}
	query.Add("ids", arrayAsParam(ids))
	url.RawQuery = query.Encode()

	var versions []*Version
	err = client.sendRequest("GET", url.String(), nil, &versions)
	if err != nil {
		return nil, err
	}

	return versions, nil
}

/* version files */

// VersionFromHash https://docs.modrinth.com/api-spec/#tag/version-files/operation/versionFromHash
func (client *Client) VersionFromHash(hash string, algorithm HashAlgo) (*Version, error) {
	url, err := url2.Parse(apiUrl + "/version_file/" + hash)
	if err != nil {
		return nil, err
	}

	query := url2.Values{}
	query.Add("algorithm", string(algorithm))
	url.RawQuery = query.Encode()

	var version *Version
	err = client.sendRequest("GET", url.String(), nil, &version)
	if err != nil {
		return nil, err
	}

	return version, nil
}
