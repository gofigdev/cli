package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"

	"cli.gofig.dev/clipb"
	"github.com/manifoldco/promptui"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name: "gofig",
		Commands: []*cli.Command{
			{
				Name: "login",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "registry",
						Aliases: []string{"r"},
						Usage:   "The registry URL to verify your login credentials and retrieve your GOPROXY URL",
						Value:   "https://gofig.dev",
						EnvVars: []string{"GOFIG_REGISTRY"},
					},
					&cli.BoolFlag{
						Name:  "token-stdin",
						Usage: "Retrieve the token credential from stdin",
					},
				},
				Action: func(c *cli.Context) error {
					if err := login(c); err != nil {
						return cli.Exit(err.Error(), 1)
					}
					return nil
				},
			},
			{
				Name: "logout",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "registry", // to remove from .netrc
						Value:   "https://gofig.dev",
						EnvVars: []string{"GOFIG_REGISTRY"},
					},
				},
				Action: func(c *cli.Context) error {
					cmd := exec.CommandContext(c.Context, "go", "env", "-w", "GOPROXY=proxy.golang.org,direct")
					cmd.Stdout = os.Stdout
					cmd.Stderr = os.Stderr
					err := cmd.Run()
					if err != nil {
						return fmt.Errorf("could not rewrite GOPROXY: %w", err)
					}
					// TODO: remove from .netrc
					return nil
				},
			},
		},
	}

	app.Run(os.Args)
}

func login(c *cli.Context) error {
	registry := c.String("registry")
	var token string
	var err error
	if c.Bool("token-stdin") {
		_, err = fmt.Scanln(&token)
		if err != nil {
			return fmt.Errorf("could not scan token from stdin: %w", err)
		}
	} else {
		prompt := promptui.Prompt{
			Label: "Token",
			Validate: func(s string) error {
				if strings.Contains(s, " ") {
					return fmt.Errorf("must not contain spaces")
				}
				return nil
			},
		}
		token, err = prompt.Run()
		if err != nil {
			return fmt.Errorf("could not get token from prompt: %w", err)
		}
	}
	gofigClient := clipb.NewAPIProtobufClient(registry, http.DefaultClient)
	resp, err := gofigClient.Proxy(c.Context, &clipb.ProxyRequest{ProxyToken: token})
	if err != nil {
		// TODO: nicer message when invalid auth
		return fmt.Errorf("error authenticating: %w", err)
	}
	proxyURL := resp.GetURL()
	url, err := url.Parse(proxyURL)
	if err != nil {
		return fmt.Errorf("error parsing %q: %w", proxyURL, err)
	}
	writeNETRC(url.Hostname(), "gofig", token, "")
	args := []string{
		"env", "-w",
		getGOPROXYValue(c.Context, proxyURL),
		getGONOSUMDBValue(c.Context, resp.GetPrivatePaths()),
		getGOPRIVATEValue(c.Context, resp.GetPrivatePaths()),
	}
	cmd := exec.CommandContext(c.Context, "go", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("could not persist GOPROXY: %w", err)
	}
	return nil
}

func getGOPROXYValue(ctx context.Context, proxyURL string) string {
	goproxy, err := getGoEnv(ctx, "GOPROXY")
	if err != nil {
		log.Fatal(err)
	}
	if goproxy == "" {
		return fmt.Sprintf("GOPROXY=%s", proxyURL)
	}
	if !strings.Contains(goproxy, proxyURL) {
		goproxy = proxyURL + "," + goproxy
	}
	return fmt.Sprintf("GOPROXY=%s", goproxy)
}

func getGONOSUMDBValue(ctx context.Context, privatePaths []string) string {
	gonosumdb, err := getGoEnv(ctx, "GONOSUMDB")
	if err != nil {
		log.Fatal(err)
	}
	for _, pp := range privatePaths {
		if strings.Contains(gonosumdb, pp) {
			continue
		}
		if gonosumdb == "" {
			gonosumdb = pp
		} else {
			gonosumdb = fmt.Sprintf("%s,%s", gonosumdb, pp)
		}
	}
	return fmt.Sprintf("GONOSUMDB=%s", gonosumdb)
}

func getGOPRIVATEValue(ctx context.Context, privatePaths []string) string {
	privateMap := map[string]struct{}{}
	for _, pp := range privatePaths {
		privateMap[strings.TrimSpace(pp)] = struct{}{}
	}
	goprivateEnv, err := getGoEnv(ctx, "GOPRIVATE")
	if err != nil {
		log.Fatal(err)
	}
	goprivate := strings.Split(goprivateEnv, ",")
	finalList := []string{}
	for _, pp := range goprivate {
		pp = strings.TrimSpace(pp)
		if pp == "" {
			continue
		}
		// An enhancement would be to check if existing value
		// "path matches" the given one and remove it.
		if _, ok := privateMap[pp]; ok {
			continue
		}
		finalList = append(finalList, pp)
	}
	return fmt.Sprintf("GOPRIVATE=%s", strings.Join(finalList, ","))
}

func getGoEnv(ctx context.Context, env string) (string, error) {
	envBts, err := exec.CommandContext(ctx, "go", "env", env).Output()
	if err != nil {
		return "", fmt.Errorf("could not get %q: %v", env, err)
	}
	return strings.TrimSpace(string(envBts)), nil
}
