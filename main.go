package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/stefanoschrs/github-stars-cli/storage"
	"github.com/stefanoschrs/github-stars-cli/types"

	"github.com/urfave/cli/v2"
)

func getStarredReposPage(user string, pageNumber int) (repos []types.Repo, hasNext bool, err error) {
	baseUrl := "https://api.github.com"
	reqUrl := fmt.Sprintf("%s/users/%s/starred?page=%d", baseUrl, user, pageNumber)

	if os.Getenv("DEBUG") == "true" {
		log.Printf("Fetching page %d of %s..\n", pageNumber, user)
	}

	client := http.Client{}
	req, err := http.NewRequest(http.MethodGet, reqUrl, nil)
	if err != nil {
		return
	}
	res, err := client.Do(req)
	if err != nil {
		return
	}
	if res.Body != nil {
		defer res.Body.Close()
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, &repos)
	if err != nil {
		return
	}

	// TODO: Wait mechanism if limit reached
	if os.Getenv("DEBUG") == "true" {
		log.Printf("X-Ratelimit-Remaining: %s\tX-Ratelimit-Reset: %s\n",
			res.Header.Get("X-Ratelimit-Remaining"),
			res.Header.Get("X-Ratelimit-Reset"))
	}
	link := res.Header.Get("Link")
	for _, el := range strings.Split(link, ", ") {
		if strings.Contains(el, `rel="next"`) {
			hasNext = true
			break
		}
	}

	return
}

func getStarredRepos(db storage.DB, username string, languages []string) (repos []types.Repo, err error) {
	cachedRepos, err := db.GetUserRepos(username)
	if err != nil {
		return
	}

	if cachedRepos != nil {
		repos = *cachedRepos
	} else {
		pageNumber := 1
		for {
			pageRepos, hasNext, err2 := getStarredReposPage(username, pageNumber)
			if err2 != nil {
				err = err2
				return
			}

			repos = append(repos, pageRepos...)

			if !hasNext {
				break
			}
			pageNumber++
		}

		err = db.SaveUserRepos(username, repos)
		if err != nil {
			return
		}
	}

	if len(languages) > 0 {
		var filteredRepos []types.Repo
		for _, repo := range repos {
			for _, language := range languages {
				if strings.ToLower(repo.Language) == language {
					filteredRepos = append(filteredRepos, repo)
					break
				}
			}
		}
		repos = filteredRepos
	}

	return
}

func main() {
	const maxLineCharacters = 200

	db, err := storage.Init()
	if err != nil {
		log.Fatal("storage.Init", err)
	}
	defer db.Close()

	app := &cli.App{
		Name: "GitHub Starred Repositories",
		Commands: []*cli.Command{
			{
				Name:  "list",
				Usage: "list stars",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "username",
						Aliases:  []string{"u"},
						Usage:    "--username stefanoschrs",
						Required: true,
					},
					&cli.StringSliceFlag{
						Name:    "language",
						Aliases: []string{"l"},
						Usage:   "--language go --language javascript",
					},
				},
				Action: func(c *cli.Context) error {
					username := c.String("username")

					var languages []string
					for _, lan := range c.StringSlice("language") {
						languages = append(languages, strings.ToLower(lan))
					}

					repos, getErr := getStarredRepos(db, username, languages)
					if getErr != nil {
						return getErr
					}
					biggestNameLen := 0
					for _, repo := range repos {
						l := len(repo.Name)
						if l > biggestNameLen {
							biggestNameLen = l
						}
					}
					for _, repo := range repos {
						nameSpacing := strings.Repeat(" ", biggestNameLen-len(repo.Name))
						descriptionLimit := maxLineCharacters - len(repo.Name) - len(nameSpacing)
						if descriptionLimit > len(repo.Description) {
							descriptionLimit = len(repo.Description)
						}

						fmt.Printf("%s\t%s\n", repo.Name+nameSpacing, repo.Description[:descriptionLimit])
					}

					return nil
				},
			},
		},
	}

	err = app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
