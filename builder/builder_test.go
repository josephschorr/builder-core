package builder

import (
	"os"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/rafecolton/go-dockerclient-quick"
	"github.com/sylphon/builder-core/parser"
	"github.com/sylphon/builder-core/unit-config"
)

var unitConfig = &unitconfig.UnitConfig{
	Version: 1,
	ContainerArr: []*unitconfig.ContainerSection{
		&unitconfig.ContainerSection{
			Name:       "app",
			Dockerfile: "Dockerfile",
			Registry:   "quay.io/rafecolton",
			Project:    "builder-core-test",
			Tags:       []string{"latest"},
			SkipPush:   true,
		},
	},
}

func TestBuildCommandSequence(t *testing.T) {
	var opts = parser.NewParserOptions{ContextDir: os.Getenv("PWD"), Logger: nil}
	var p = parser.NewParser(opts)
	commandSequence := p.Parse(unitConfig)

	var logger = logrus.New()
	logger.Level = logrus.DebugLevel

	builder := &Builder{
		Logger:     logger,
		contextDir: os.Getenv("GOPATH") + "/src/github.com/sylphon/builder-core/_testing",
	}

	if err := builder.BuildCommandSequence(commandSequence); err != nil {
		t.Fatal(err)
	}
}

func TestNewBuilder(t *testing.T) {
	var ops = NewBuilderOptions{
		ContextDir:   os.Getenv("PWD"),
		Logger:       &logrus.Logger{},
		dockerClient: dockerclient.FakeClient(),
	}
	_, err := NewBuilder(ops)
	if err != nil {
		t.Fatal(err)
	}
}