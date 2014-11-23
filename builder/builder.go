package builder

import (
	"errors"
	"io"
	"io/ioutil"
	"regexp"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/pkg/archive"
	"github.com/modcloth/go-fileutils"
	"github.com/onsi/gocleanup"
	"github.com/rafecolton/go-dockerclient-quick"

	"github.com/sylphon/builder-core/filecheck"
	"github.com/sylphon/builder-core/parser"
)

var (
	// SkipPush will, when set to true, override any behavior set by a Bobfile and
	// will cause builders *NOT* to run `docker push` commands.  SkipPush is also set
	// by the `--skip-push` option when used on the command line.
	SkipPush bool

	imageWithTagRegex = regexp.MustCompile("^(.*):(.*)$")
)

/*
A Builder is the struct that actually does the work of moving files around and
executing the commands that do the docker build.
*/
type Builder struct {
	dockerClient dockerclient.DockerClient
	*logrus.Logger
	workdir         string
	isRegular       bool
	nextSubSequence *parser.SubSequence
	Stderr          io.Writer
	Stdout          io.Writer
	Builderfile     string
	contextDir      string
}

/*
SetNextSubSequence sets the next subsequence within bob to be processed. This
function is exported because it is used explicitly in tests, but in Build(), it
is intended to be used as a helper function.
*/
func (bob *Builder) SetNextSubSequence(subSeq *parser.SubSequence) {
	bob.nextSubSequence = subSeq
}

// NewBuilderOptions encapsulates all of the options necessary for creating a
// new builder
type NewBuilderOptions struct {
	Logger       *logrus.Logger
	ContextDir   string
	dockerClient dockerclient.DockerClient // default to nil for regular docker client
}

/*
NewBuilder returns an instance of a Builder struct.  The function exists in
case we want to initialize our Builders with something.
*/
func NewBuilder(opts NewBuilderOptions) (*Builder, error) {
	logger := opts.Logger
	if logger == nil {
		logger = logrus.New()
		logger.Level = logrus.PanicLevel
	}

	var client = opts.dockerClient
	if client == nil {
		var err error
		client, err = dockerclient.NewDockerClient()
		if err != nil {
			return nil, err
		}
	}

	stdout := newOutWriter(logger, "         %s")
	stderr := newOutWriter(logger, "         %s")

	if logrus.IsTerminal() {
		stdout = newOutWriter(logger, "         @{g}%s@{|}")
		stderr = newOutWriter(logger, "         @{r}%s@{|}")
	}

	return &Builder{
		dockerClient: client,
		Logger:       logger,
		isRegular:    true,
		Stdout:       stdout,
		Stderr:       stderr,
		contextDir:   opts.ContextDir,
	}, nil
}

// BuildCommandSequence performs a build from a parser-generated CommandSequence struct
func (bob *Builder) BuildCommandSequence(commandSequence *parser.CommandSequence) error {
	for _, seq := range commandSequence.Commands {
		var imageID string
		var err error

		if err := bob.cleanWorkdir(); err != nil {
			return err
		}
		bob.SetNextSubSequence(seq)
		if err := bob.setup(); err != nil {
			return err
		}

		bob.WithField("container_section", seq.Metadata.Name).
			Info("running commands for container section")

		for _, cmd := range seq.SubCommand {
			opts := &parser.DockerCmdOpts{
				DockerClient: bob.dockerClient,
				Image:        imageID,
				ImageUUID:    seq.Metadata.UUID,
				SkipPush:     SkipPush,
				Stderr:       bob.Stderr,
				Stdout:       bob.Stdout,
				Workdir:      bob.workdir,
			}
			cmd = cmd.WithOpts(opts)

			bob.WithField("command", cmd.Message()).Info("running docker command")

			if imageID, err = cmd.Run(); err != nil {
				switch err.(type) {
				case parser.NilClientError:
					continue
				default:
					return err
				}
			}
		}

		bob.attemptToDeleteTemporaryUUIDTag(seq.Metadata.UUID)
	}
	return nil
}

func (bob *Builder) attemptToDeleteTemporaryUUIDTag(uuid string) {
	if bob.dockerClient == nil {
		return
	}

	regex := ":" + uuid + "$"
	image, err := bob.dockerClient.LatestImageByRegex(regex)
	if err != nil {
		bob.WithField("err", err).Warn("error getting repo taggged with temporary tag")
	}

	for _, tag := range image.RepoTags {
		matched, err := regexp.MatchString(regex, tag)
		if err != nil {
			return
		}
		if matched {
			bob.WithFields(logrus.Fields{
				"image_id": image.ID,
				"tag":      tag,
			}).Info("deleting temporary tag")

			if err = bob.dockerClient.Client().RemoveImage(tag); err != nil {
				bob.WithField("err", err).Warn("error deleting temporary tag")
			}
			return
		}
	}
}

/*
Setup moves all of the correct files into place in the temporary directory in
order to perform the docker build.
*/
func (bob *Builder) setup() error {
	var workdir = bob.workdir
	var pathToDockerfile *filecheck.TrustedFilePath
	var err error

	if bob.nextSubSequence == nil {
		return errors.New("no command sub sequence set, cannot perform setup")
	}

	meta := bob.nextSubSequence.Metadata
	dockerfile := meta.Dockerfile
	opts := filecheck.NewTrustedFilePathOptions{File: dockerfile, Top: bob.contextDir}
	pathToDockerfile, err = filecheck.NewTrustedFilePath(opts)
	if err != nil {
		return err
	}

	if pathToDockerfile.Sanitize(); pathToDockerfile.State != filecheck.OK {
		return pathToDockerfile.Error
	}

	contextDir := pathToDockerfile.Top()
	tarStream, err := archive.TarWithOptions(contextDir, &archive.TarOptions{
		Compression: archive.Uncompressed,
		Excludes:    []string{"Dockerfile"},
	})
	if err != nil {
		return err
	}

	defer tarStream.Close()
	if err := archive.Untar(tarStream, workdir, nil); err != nil {
		return err
	}

	if err := fileutils.CpWithArgs(
		contextDir+"/"+meta.Dockerfile,
		workdir+"/Dockerfile",
		fileutils.CpArgs{PreserveModTime: true},
	); err != nil {
		return err
	}

	return nil
}

func (bob *Builder) generateWorkDir() string {
	tmp, err := ioutil.TempDir("", "bob")
	if err != nil {
		return ""
	}

	gocleanup.Register(func() {
		fileutils.RmRF(tmp)
	})

	return tmp
}

/*
cleanWorkdir effectively does a rm -rf and mkdir -p on bob's workdir.  Intended
to be used before using the workdir (i.e. before new command groups).
*/
func (bob *Builder) cleanWorkdir() error {
	workdir := bob.generateWorkDir()
	bob.workdir = workdir

	if err := fileutils.RmRF(workdir); err != nil {
		return err
	}

	return fileutils.MkdirP(workdir, 0755)
}