package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"github.com/nothub/mrpack-install/cmd"
	"github.com/spf13/cobra"
	"log"
	"net/http"
	"os"
	"text/template"
)

//go:embed readme.tmpl usage.tmpl
var fs embed.FS

func init() {
	log.SetFlags(0)
}

func main() {

	var data struct {
		Entries  []command
		Contribs []ghUser
	}

	// The project's own commands get a section, cobra's help and completion
	// commands don't. Execute adds those two, so add them by hand for the
	// root help to list them as the binary does.
	subcommands := cmd.RootCmd.Commands()
	cmd.RootCmd.InitDefaultHelpCmd()
	cmd.RootCmd.InitDefaultCompletionCmd()
	data.Entries = append(data.Entries, helpOf("root", cmd.RootCmd))
	for _, sc := range subcommands {
		data.Entries = append(data.Entries, helpOf(sc.Name(), sc))
	}

	// fetch contribs from gh api
	var err error
	data.Contribs, err = getContribs("nothub", "mrpack-install")
	if err != nil {
		log.Fatalln(err.Error())
	}
	data.Contribs = filterBots(data.Contribs)

	render("readme.tmpl", "README.md", data)
	render("usage.tmpl", "USAGE.md", data)
}

// render executes the embedded template name with data into the file out,
// relative to the working directory.
func render(name string, out string, data any) {
	tmpl, err := template.ParseFS(fs, name)
	if err != nil {
		log.Fatalln(err.Error())
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		log.Fatalln(err.Error())
	}
	err = os.WriteFile(out, buf.Bytes(), 0644)
	if err != nil {
		log.Fatalln(err.Error())
	}
}

type command struct {
	Name string
	Help string
}

func helpOf(name string, c *cobra.Command) command {
	c.InitDefaultHelpFlag()
	var buf bytes.Buffer
	c.SetOut(&buf)
	err := c.Help()
	if err != nil {
		log.Fatalln(err.Error())
	}
	return command{
		Name: name,
		Help: buf.String(),
	}
}

type ghUser struct {
	Login         string `json:"login"`
	Name          string `json:"name"`
	HtmlUrl       string `json:"html_url"`
	AvatarUrl     string `json:"avatar_url"`
	Contributions int    `json:"contributions"`
}

func getContribs(owner string, repo string) (contribs []ghUser, err error) {

	err = ghApi(fmt.Sprintf("/repos/%s/%s/contributors", owner, repo), &contribs)
	if err != nil {
		return nil, err
	}

	// 'contributors' endpoint is not enough, it does not provide names...
	// we have to fetch them from the 'users' endpoint:
	for i, contrib := range contribs {

		err = ghApi(fmt.Sprintf("/users/%s", contrib.Login), &contrib)
		if err != nil {
			return nil, err
		}

		if contrib.Name == "" {
			contrib.Name = contrib.Login
		}

		contribs[i] = contrib
	}

	return contribs, nil
}

func ghApi(endpoint string, target interface{}) error {

	req, err := http.NewRequest(http.MethodGet, "https://api.github.com"+endpoint, nil)
	if err != nil {
		return err
	}

	req.Header.Add("Accept", "application/vnd.github+json")
	req.Header.Add("X-GitHub-Api-Version", "2022-11-28")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	return json.NewDecoder(res.Body).Decode(target)
}

func filterBots(list []ghUser) []ghUser {
	result := make([]ghUser, 0, len(list))

	for _, usr := range list {
		if usr.Login != "dependabot[bot]" && usr.Login != "renovate[bot]" {
			result = append(result, usr)
		}
	}

	return result
}
