package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/google/go-github/v42/github"
	"github.com/gookit/color"
	"github.com/pkg/browser"
	"golang.org/x/oauth2"
)

func main() {
	ctx := context.Background()
	at := os.Getenv("GITHUB_TOKEN")
	if at == "" {
		fmt.Println("Please set the GITHUB_TOKEN environment variable")
		fmt.Println("You can get one here (with repo scope): https://github.com/settings/tokens/new")
		return
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: at},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	issues, _, err := client.Issues.List(context.Background(), true, &github.IssueListOptions{
		Filter: "all",
		State:  "all",
		Labels: []string{"stale"},
	})
	if err != nil {
		panic(err)
	}

	for _, issue := range issues {
		title := issue.GetTitle()
		if len(title) == 0 {
			title = "No title."
		}
		body := issue.GetBody()
		if len(body) == 0 {
			body = "No description provided."
		}

		repo := issue.GetRepository()
		header := fmt.Sprintf("%s %s\n", repo.GetFullName()+"#"+color.FgBlue.Render(issue.GetNumber()), title)
		padding := strings.Repeat("-", len(header))
		fmt.Println(padding)
		fmt.Println(header)

		color.FgGray.Println(body)
		comments := issue.GetComments()

		rowner := issue.GetRepository().GetOwner().GetLogin()
		rname := issue.GetRepository().GetName()
		inum := issue.GetNumber()

		if comments != 0 {
			comments, _, err := client.Issues.ListComments(ctx, rowner, rname, inum, nil)
			if err != nil {
				panic(err)
			}
			fmt.Println(strings.Repeat("-", len(body)))
			for _, comment := range comments {
				color.FgGray.Printf("[%s] %s\n", comment.GetUser().GetLogin(), comment.GetBody())
			}
		}
		for {
			bio := bufio.NewReader(os.Stdin)
			fmt.Printf("What should I do? [r/o/q/?] (empty to skip): ")
			ch, _, err := bio.ReadLine()
			if err != nil {
				panic(err)
			}
			if len(ch) == 0 {
				break
			}
			switch ch[0] {
			case 'r':
				fmt.Print("Reply with what? (empty to cancel): ")
				line, _, err := bio.ReadLine()
				if err != nil {
					panic(err)
				}
				if len(line) == 0 {
					continue
				} else {
					fmt.Printf("Replying with `%s`...\n", line)
					lstr := string(line)
					comment, _, err := client.Issues.CreateComment(ctx, rowner, rname, inum, &github.IssueComment{
						Body: &lstr,
					})
					if err != nil {
						panic(err)
					}
					fmt.Printf("Reply can be viewed at %s\n", comment.GetHTMLURL())
					break
				}
			case 'o':
				url := issue.GetHTMLURL()
				fmt.Printf("Opening %s...\n", url)
				browser.OpenURL(url)
			case 'q':
				return
			case '?':
				fmt.Println(`[r]eply to the current issue
[o]pen the issue in your web browser
[q]uit the application`)
				continue
			default:
				color.Red.Println("Please enter a valid option!")
				continue
			}
			break
		}

		fmt.Printf("%s\n", padding)
	}
}
