package commands

import (
	"sync"

	stks "github.com/daidokoro/qaz/stacks"

	"github.com/daidokoro/qaz/logger"
	"github.com/daidokoro/qaz/repo"
)

var (
	config  Config
	stacks  map[string]*stks.Stack
	region  string
	project string
	wg      sync.WaitGroup
	gitrepo repo.Repo
	log     = logger.Logger{
		DebugMode: &run.debug,
		Colors:    &run.colors,
	}
)

// config environment variable
const (
	configENV     = "QAZ_CONFIG"
	defaultconfig = "config.yml"

	// OutputRegex for printing yaml/json output
	OutputRegex = `(?m)^[ ]*([^\r\n:]+?)\s*:`
)

// run.var used as a central point for command data from flags
var run = struct {
	cfgSource  string
	tplSource  string
	profile    string
	tplSources []string
	stacks     map[string]string
	all        bool
	version    bool
	request    string
	debug      bool
	funcEvent  string
	lambdAsync bool
	changeName string
	stackName  string
	rollback   bool
	colors     bool
	cfgRaw     string
	gituser    string
	gitpass    string
	gitrsa     string
}{}
