package conf

// Config is the global config
var Config Conf

// Conf is used for storing data retrieved from environmental variables.
type Conf struct {
	Port      int
	LogLevel  string
	LogFormat string
	APIToken  string
	SkipPush  bool

	// for basic auth
	Username string
	Password string

	// for travis auth
	TravisToken string
	NoTravis    bool

	// for github auth
	GitHubSecret string
	NoGitHub     bool

	// docker registry credentials
	CfgUn    string
	CfgPass  string
	CfgEmail string
}
