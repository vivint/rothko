module "github.com/vivint/rothko"

require (
	"github.com/BurntSushi/toml" v0.3.0
	"github.com/fsnotify/fsnotify" v1.4.7
	"github.com/gogo/protobuf" v1.0.0
	"github.com/kisielk/gotool" v1.0.0
	"github.com/neelance/astrewrite" v0.0.0-20160511093645-99348263ae86
	"github.com/neelance/sourcemap" v0.0.0-20151028013722-8c68805598ab
	"github.com/pkg/errors" v0.8.0
	"github.com/robertkrimen/godocdown" v0.0.0-20130622164427-0bfa04905481
	"github.com/spf13/cobra" v0.0.1
	"github.com/spf13/pflag" v1.0.0
	"github.com/stretchr/testify" v1.2.1
	"github.com/urfave/cli" v1.20.0
	"github.com/zeebo/errs" v0.1.0
	"github.com/zeebo/float16" v0.1.0
	"github.com/zeebo/live" v0.0.0-20180301045707-148ee9a0fa56
	"github.com/zeebo/tdigest" v0.1.0
	"go.uber.org/atomic" v1.3.1
	"go.uber.org/multierr" v1.1.0
	"go.uber.org/zap" v1.7.1
	"golang.org/x/crypto" v0.0.0-20180226093711-beaf6a35706e
	"golang.org/x/image" v0.0.0-20171214225156-12117c17ca67
	"golang.org/x/sys" v0.0.0-20180224232135-f6cff0780e54
	"golang.org/x/tools" v0.0.0-20180226184358-10db2d12cfa8
)

require (
	// we intentionally keep this at a git ref version until
	// golang.org/issue/23954 is fixed
	"github.com/stretchr/objx" v0.0.0-20180129182003-8a3f715
)
