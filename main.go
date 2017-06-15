package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/urfave/cli"
)

type Docker struct {
	Registry string   `json:"registry"`
	Storage  string   `json:"storage_driver"`
	Token    string   `json:"token"`
	Repo     string   `json:"repo"`
	Tag      []string `json:"tag"`
	File     string   `json:"file"`
	Context  string   `json:"context"`
}

type BUILD struct {
	Number string
	Commit string
	Branch string
	Tag    string
}

var (
	buildCommit string
)

func main() {
	fmt.Printf("Drone GCR Plugin built from %s\n", buildCommit)

	app := cli.NewApp()
	app.Name = "gke plugin"
	app.Usage = "gke plugin"
	app.Action = run
	app.Version = fmt.Sprintf("1.0.0-%s", buildCommit)
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "registry",
			Usage:  "",
			EnvVar: "PLUGIN_REGISTER",
		},
		cli.StringFlag{
			Name:   "storage_driver",
			Usage:  "",
			EnvVar: "PLUGIN_STORAGE_DRIVER",
		},
		cli.StringFlag{
			Name:   "token",
			Usage:  "",
			EnvVar: "PLUGIN_TOKEN",
		},
		cli.StringFlag{
			Name:   "repo",
			Usage:  "",
			EnvVar: "PLUGIN_REPO",
		},
		cli.StringSliceFlag{
			Name:   "tag",
			Usage:  "",
			EnvVar: "PLUGIN_TAG",
		},
		cli.StringFlag{
			Name:   "file",
			Usage:  "",
			EnvVar: "PLUGIN_FILE",
		},
		cli.StringFlag{
			Name:   "context",
			Usage:  "",
			EnvVar: "PLUGIN_CONTEXT",
		},
	}

	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}

func run(c *cli.Context) error {
	vargs := Docker{}
	vargs.Registry = c.String("registry")
	vargs.Storage = c.String("storage_driver")
	vargs.Token = c.String("token")
	vargs.Repo = c.String("repo")
	vargs.Tag = c.StringSlice("tag")
	vargs.File = c.String("file")
	vargs.Context = c.String("context")

	build := BUILD{}
	build.Number = c.String("drone-build-number")
	build.Commit = c.String("drone-commit")
	build.Branch = c.String("drone-branch")
	build.Tag = c.String("drone-tag")

	// Repository name should have gcr prefix
	if len(vargs.Registry) == 0 {
		vargs.Registry = "gcr.io"
	}
	// Set the Dockerfile name
	if len(vargs.File) == 0 {
		vargs.File = "Dockerfile"
	}
	// Set the Context value
	if len(vargs.Context) == 0 {
		vargs.Context = "."
	}
	// Set the Tag value
	if len(vargs.Tag) == 0 {
		vargs.Tag = []string{"latest"}
	}
	// Concat the Registry URL and the Repository name if necessary
	if strings.Count(vargs.Repo, "/") == 1 {
		vargs.Repo = fmt.Sprintf("%s/%s", vargs.Registry, vargs.Repo)
	}
	// Trim any spaces or newlines from the token
	vargs.Token = strings.TrimSpace(vargs.Token)

	go func() {
		args := []string{"daemon"}

		if len(vargs.Storage) != 0 {
			args = append(args, "-s", vargs.Storage)
		}

		cmd := exec.Command("/usr/bin/docker", args...)
		if os.Getenv("DOCKER_LAUNCH_DEBUG") == "true" {
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
		} else {
			cmd.Stdout = ioutil.Discard
			cmd.Stderr = ioutil.Discard
		}
		trace(cmd)
		cmd.Run()
	}()

	// ping Docker until available
	for i := 0; i < 3; i++ {
		cmd := exec.Command("/usr/bin/docker", "info")
		cmd.Stdout = ioutil.Discard
		cmd.Stderr = ioutil.Discard
		err := cmd.Run()
		if err == nil {
			break
		}
		time.Sleep(time.Second * 5)
	}

	// Login to Docker
	cmd := exec.Command("/usr/bin/docker", "login", "-u", "_json_key", "-p", vargs.Token, "-e", "chunkylover53@aol.com", vargs.Registry)
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("Error while getting working directory: %s\n", err)
	}
	cmd.Dir = wd
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		fmt.Println("Login failed.")
		os.Exit(1)
	}

	// Build the container
	cmd = exec.Command("/usr/bin/docker", "build", "--pull=true", "--rm=true", "-f", vargs.File, "-t", build.Commit, vargs.Context)
	cmd.Dir = wd
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	trace(cmd)
	err = cmd.Run()
	if err != nil {
		os.Exit(1)
	}

	// Creates image tags
	for _, tag := range vargs.Tag {
		// create the full tag name
		tag_ := fmt.Sprintf("%s:%s", vargs.Repo, tag)
		if tag == "latest" {
			tag_ = vargs.Repo
		}

		// tag the build image sha
		cmd = exec.Command("/usr/bin/docker", "tag", build.Commit, tag_)
		cmd.Dir = wd
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		trace(cmd)
		err = cmd.Run()
		if err != nil {
			os.Exit(1)
		}
	}

	// Push the image and tags to the registry
	cmd = exec.Command("/usr/bin/docker", "push", vargs.Repo)
	cmd.Dir = wd
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	trace(cmd)
	err = cmd.Run()
	if err != nil {
		os.Exit(1)
	}
	return nil
}

// Trace writes each command to standard error (preceded by a ‘$ ’) before it
// is executed. Used for debugging your build.
func trace(cmd *exec.Cmd) {
	fmt.Println("$", strings.Join(cmd.Args, " "))
}
