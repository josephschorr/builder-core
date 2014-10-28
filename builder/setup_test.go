package builder_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/rafecolton/docker-builder/builder"

	"io/ioutil"
	"os"
	"os/exec"
	"sort"

	"github.com/modcloth/go-fileutils"
	"github.com/rafecolton/docker-builder/parser"
)

var _ = Describe("Setup", func() {
	var (
		branch          string
		rev             string
		short           string
		top             string
		subject         *Builder
		baseSubSequence = &parser.SubSequence{
			Metadata: &parser.SubSequenceMetadata{
				Name:       "base",
				Dockerfile: "Dockerfile.base",
			},
			SubCommand: []parser.DockerCmd{
				&parser.BuildCmd{},
				&parser.TagCmd{Tag: "quay.io/modcloth/style-gallery:base"},
			},
		}
		appSubSequence = &parser.SubSequence{
			Metadata: &parser.SubSequenceMetadata{
				Name:       "app",
				Dockerfile: "Dockerfile",
			},
			SubCommand: []parser.DockerCmd{
				&parser.BuildCmd{},
				&parser.TagCmd{Repo: "quay.io/modcloth/style-gallery", Tag: branch},
				&parser.TagCmd{Repo: "quay.io/modcloth/style-gallery", Tag: rev},
				&parser.TagCmd{Repo: "quay.io/modcloth/style-gallery", Tag: short},
				&parser.PushCmd{Image: "quay.io/modcloth/style-gallery", Tag: branch},
				&parser.PushCmd{Image: "quay.io/modcloth/style-gallery", Tag: rev},
				&parser.PushCmd{Image: "quay.io/modcloth/style-gallery", Tag: short},
			},
		}
	)

	BeforeEach(func() {
		subject, _ = NewBuilder(nil, false)
		top = os.Getenv("PWD")
		git, _ := fileutils.Which("git")
		// branch
		branchCmd := &exec.Cmd{
			Path: git,
			Dir:  top,
			Args: []string{git, "rev-parse", "-q", "--abbrev-ref", "HEAD"},
		}

		branchBytes, _ := branchCmd.Output()
		branch = string(branchBytes)[:len(branchBytes)-1]

		// rev
		revCmd := &exec.Cmd{
			Path: git,
			Dir:  top,
			Args: []string{git, "rev-parse", "-q", "HEAD"},
		}
		revBytes, _ := revCmd.Output()
		rev = string(revBytes)[:len(revBytes)-1]

		// short
		shortCmd := &exec.Cmd{
			Path: git,
			Dir:  top,
			Args: []string{git, "describe", "--always", "--dirty", "--tags"},
		}
		shortBytes, _ := shortCmd.Output()
		short = string(shortBytes)[:len(shortBytes)-1]
	})

	Context("with the base container sequence", func() {
		It("places the correct files in the workdir", func() {
			subject.SetNextSubSequence(baseSubSequence)
			subject.CleanWorkdir()
			subject.Setup()

			expectedFiles := []string{
				"Dockerfile",
				"Gemfile",
				"Gemfile.lock",
				"foo",
				"README.txt",
				"other_file.txt",
				"spec",
			}

			files, _ := ioutil.ReadDir(subject.Workdir())
			fileNames := make([]string, len(files), len(files))

			for i, v := range files {
				fileNames[i] = v.Name()
			}

			sort.Strings(fileNames)
			sort.Strings(expectedFiles)
			Expect(fileNames).To(Equal(expectedFiles))
		})
	})

	Context("with the app container sequence", func() {
		It("places the correct files in the workdir", func() {
			subject.SetNextSubSequence(appSubSequence)
			subject.CleanWorkdir()
			subject.Setup()

			expectedFiles := []string{
				"Dockerfile",
				"Dockerfile.base",
				"Gemfile",
				"Gemfile.lock",
				"foo",
				"README.txt",
				"other_file.txt",
				"spec",
			}

			files, _ := ioutil.ReadDir(subject.Workdir())
			fileNames := make([]string, len(files), len(files))
			for i, v := range files {
				fileNames[i] = v.Name()
			}

			sort.Strings(fileNames)
			sort.Strings(expectedFiles)

			Expect(fileNames).To(Equal(expectedFiles))
		})
	})
})
