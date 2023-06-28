package main

import (
	"context"
	"encoding/json"
	"github.com/google/go-github/v42/github"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v6"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v6/git"
	"github.com/mitchellh/mapstructure"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"
	"github.com/xanzy/go-gitlab"
	"golang.org/x/oauth2"
	"gopkg.in/alecthomas/kingpin.v2"
	"io/ioutil"
	"net/http"
	"strings"
)

const composerPath = "composer.json"

type source struct {
	SourceType  string   `json:"sourceType"`
	SourceIdent string   `json:"sourceIdent"`
	SourceAuth  string   `json:"sourceAuth"`
	Exclude     []string `json:"exclude"`
}
type repository struct {
	Name    string      `json:"name"`
	Type    string      `json:"type"`
	URL     string      `json:"url"`
	Options interface{} `json:"options"`
}

var (
	inputFile  = kingpin.Flag("input", "Input file with basic satis.json configuration").Default("input.json").String()
	outputFile = kingpin.Flag("output", "Output file - where to save generated result").Default("satis.json").String()
)

func main() {
	log.AddFlags(kingpin.CommandLine)
	kingpin.HelpFlag.Short('h')
	kingpin.Version(version.Version)
	kingpin.Parse()

	output := parseSources()

	file, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		log.Fatalf("could not prepare output file: %s", err.Error())
	}
	err = ioutil.WriteFile(*outputFile, file, 0644)
	if err != nil {
		log.Fatalf("could not write file %s: %s", *outputFile, err.Error())
	}
}

func parseSources() interface{} {
	output, sources := parseInput()
	var repositories []repository
	if output["repositories"] != nil && len(output["repositories"].([]interface{})) > 0 {
		switch t := output["repositories"].(type) {
		case []interface{}:
			for _, v := range t {
				var repo repository
				err := mapstructure.Decode(v, &repo)
				if err != nil {
					log.Errorf("could not map `%+v` repository struct, skipping", v)
					continue
				}
				repositories = append(repositories, repo)
			}
		}
	}

	for _, source := range sources {
		var repos []repository
		switch source.SourceType {
		case "gitlab":
			repos = parseGitlab(source)
		case "github":
			repos = parseGithub(source)
		case "azdo":
			repos = parseAzDO(source)
		default:
			log.Errorf("could not parse source %s with unknown type %s", source.SourceIdent, source.SourceType)
			continue
		}
		repositories = append(repositories, repos...)
	}
	output["repositories"] = repositories
	return output
}

func parseAzDO(s source) (repositories []repository) {
	log.Infof("downloading Azure DevOps projects for %s", s.SourceIdent)
	project, ctx, client, err := getAzDOClient(s)
	repos, err := client.GetRepositories(ctx, git.GetRepositoriesArgs{Project: &project})
	if err != nil {
		log.Errorf("could not fetch azdo repositories %s", err.Error())
	}

	for _, repo := range *repos {
		repoInfo := prepareAzDORepo(ctx, client, repo)
		if repoInfo == nil || isExcluded(s.Exclude, repoInfo.Name) {
			continue
		}

		repositories = append(
			repositories,
			*repoInfo,
		)
	}

	return repositories
}

func getAzDOClient(s source) (string, context.Context, git.Client, error) {
	idents := strings.Split(s.SourceIdent, "/")
	org := strings.Join(idents[:len(idents)-1], "/")
	project := idents[len(idents)-1]

	connection := azuredevops.NewPatConnection(org, s.SourceAuth)

	ctx := context.Background()

	client, err := git.NewClient(ctx, connection)
	if err != nil {
		log.Fatal(err)
	}
	return project, ctx, client, err
}

func prepareAzDORepo(ctx context.Context, client git.Client, repo git.GitRepository) *repository {
	//Cannot check for disabled repository - `isDisabled` attribute is available from API version 7.1 which is currently in preview.
	file, err := client.GetItem(ctx, git.GetItemArgs{
		RepositoryId: gitlab.String(repo.Id.String()),
		Path:         gitlab.String("/" + composerPath),
	})
	if file == nil || err != nil {
		return nil
	}

	repoInfo := repository{
		Name: *repo.Name,
		Type: "git",
		URL:  *repo.SshUrl,
	}
	return &repoInfo
}

func parseGithub(s source) (repositories []repository) {
	log.Infof("downloading github projects for organization %s", s.SourceIdent)
	ctx, client := getGithubClient(s)

	options := github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{
			Page:    1,
			PerPage: 100,
		},
	}
	for {
		repos, response, err := client.Repositories.ListByOrg(ctx, s.SourceIdent, &options)
		if err != nil {
			log.Errorf("could not fetch github response %s", err.Error())
		}
		for _, repo := range repos {
			repoInfo := prepareGithubRepo(ctx, client, s, repo)
			if repoInfo == nil || isExcluded(s.Exclude, repoInfo.Name) {
				continue
			}
			repositories = append(
				repositories,
				*repoInfo,
			)
		}
		if response.NextPage == 0 {
			break
		}
		options.Page++
	}
	return repositories
}

func getGithubClient(s source) (context.Context, *github.Client) {
	var tc *http.Client
	ctx := context.Background()
	if s.SourceAuth != "" {
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: s.SourceAuth},
		)
		tc = oauth2.NewClient(ctx, ts)
	}

	client := github.NewClient(tc)
	return ctx, client
}

func prepareGithubRepo(ctx context.Context, client *github.Client, s source, repo *github.Repository) *repository {
	if repo.GetArchived() {
		return nil
	}
	file, _, _, err := client.Repositories.GetContents(ctx, s.SourceIdent, repo.GetName(), composerPath, &github.RepositoryContentGetOptions{})
	if file == nil || err != nil {
		return nil
	}
	repoInfo := repository{
		Name: repo.GetName(),
		Type: "git",
		URL:  repo.GetSSHURL(),
	}
	return &repoInfo
}

func parseGitlab(s source) (repositories []repository) {
	log.Infof("downloading gitlab projects for group %s", s.SourceIdent)
	client, err := gitlab.NewClient(s.SourceAuth)
	if err != nil {
		log.Error(err)
		return repositories
	}

	options := gitlab.ListGroupProjectsOptions{
		ListOptions: gitlab.ListOptions{
			Page:    1,
			PerPage: 100,
		},
		IncludeSubgroups: gitlab.Bool(true),
	}

	for {
		groupProjects, response, _ := client.Groups.ListGroupProjects(s.SourceIdent, &options)
		for _, project := range groupProjects {
			repoInfo := prepareGitlabRepo(project, client)
			if repoInfo == nil || isExcluded(s.Exclude, repoInfo.Name) {
				continue
			}
			repositories = append(
				repositories,
				*repoInfo,
			)
		}

		if response.CurrentPage == response.TotalPages {
			break
		}
		options.Page++
	}

	return repositories
}

func prepareGitlabRepo(project *gitlab.Project, client *gitlab.Client) *repository {
	if project.Archived {
		return nil
	}
	file, _, err := client.RepositoryFiles.GetFile(project.ID, composerPath, &gitlab.GetFileOptions{Ref: &project.DefaultBranch})
	if file == nil || err != nil {
		return nil
	}
	repoInfo := repository{
		Name: project.Name,
		Type: "git",
		URL:  project.SSHURLToRepo,
	}
	return &repoInfo
}

func parseInput() (map[string]interface{}, []source) {
	file, _ := ioutil.ReadFile(*inputFile)

	var configFile interface{}
	var sources []source

	err := json.Unmarshal(file, &configFile)
	if err != nil {
		log.Fatal(err)
	}
	mapConfig := configFile.(map[string]interface{})
	if _, ok := mapConfig["name"]; ok && len(strings.TrimSpace(mapConfig["name"].(string))) == 0 {
		log.Error("input file does not contain `name` attribute")
	}
	if _, ok := mapConfig["homepage"]; ok && len(strings.TrimSpace(mapConfig["homepage"].(string))) == 0 {
		log.Error("input file does not contain `homepage` attribute")
	}

	if _, ok := mapConfig["sources"]; ok && len(mapConfig["sources"].([]interface{})) == 0 {
		log.Fatal("`sources` attribute has to be specified in the input file")
	}
	err = mapstructure.Decode(mapConfig["sources"], &sources)
	if err != nil {
		log.Fatalf("could not map `sources` attribute to struct, check validity of input file")
	}
	delete(mapConfig, "sources")

	if _, ok := mapConfig["repositories"]; !ok || len(mapConfig["repositories"].([]interface{})) == 0 {
		mapConfig["repositories"] = nil
	}

	return mapConfig, sources
}

func isExcluded(l []string, n string) bool {
	for _, a := range l {
		if a == n {
			return true
		}
	}
	return false
}
