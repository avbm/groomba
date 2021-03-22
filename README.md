# Groomba

[![GitHub Workflow Status](https://github.com/avbm/groomba/actions/workflows/ci.yml/badge.svg?style=flat)](https://github.com/avbm/groomba/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/badge/go%20version-%3E=1.16-61CFDD.svg?style=flat)](https://golang.org/doc/devel/release.html)
[![PkgGoDev](https://pkg.go.dev/badge/mod/github.com/avbm/groomba?style=flat)](https://pkg.go.dev/mod/github.com/avbm/groomba)

Groomba is a simple utility written in [Go](https://golang.org/) to groom your git repositories. It will rename branches older than a defined age. Unlike other tools like the [Stale Github Action](https://github.com/actions/stale), Groomba only depends on the git APIs and is agnostic of the software used to host your git repository. It will work just as well whether your repos are hosted in Github, Gitlab, Btbucket or something else.

## Installation

Download and install latest version (assuming linux):
```
LATEST_VERSION=$(curl https://github.com/avbm/groomba/releases/latest | sed 's/^.*href="//' | cut -d'"' -f1 | rev | cut -d/ -f1 | rev)
curl "https://github.com/avbm/groomba/releases/download/${LATEST_VERSION}/groomba_linux_amd64" -o groomba
chmod +x groomba
./groomba
```

You can add the snippet above as a step in your CI pipeline on your main branch to periodically groom your repository.

Replace `linux` with `darwin` for Mac OSX builds.
Pre-built binaries are available only for x86_64 linux and OSX. For other architectures and OS systems please build your binaries using the [Build from souce](#build-from-source) section.

## Build from source

To build Groomba from source you need to have the golang compiler installed and configured. Clone this repository and from the root of the repository run:
```
$ go build .
```
This should fetch all dependencies and create a binary called `groomba` in the current directory.

## Configuration Options

To configure Groomba, you can set each configuration option in a `.groomba.toml` or `.groomba.yaml` file at the root of the repository you want to Groom. Alternately these options can also be set as environment variables. Options set as environment variables take the highest precedence.

Currently only 2 configuration options are supported:

### StaleAgeThreshold

`StaleAgeThreshold` is the threshold age in days for considering a branch as `stale`. It is expected to be an integer.

Default value is `14`

To set to a different value say 30:
```
# in .groomba.toml
stale_age_threshold = 30

# or in .groomba.yaml
stale_age_threshold: 30

# or as an environment variable
GROOMBA_STALE_AGE_THRESHOLD=3
```

### StaticBranches

`StaticBranches` is a list of branches that Groomba considers as `static` or `protected` and will ignore.

Default is `["master", "main"]`

To set to a different value say `["latest", "staging", "production"]`:
```
# in .groomba.toml
static_branches = ["latest", "staging", "production"]

# or in .groomba.yaml
stale_age_threshold: ["latest", "staging", "production"]

# or this also works in yaml
stale_age_threshold:
  - latest
  - staging
  - production

# or as an environment variable
GROOMBA_STATIC_BRANCHES="latest,staging,production"
```

### Noop

`Noop` is boolean that tells Groomba whether to run in noop mode. In noop mode, Groomba will only print out messages informing users about which branches would be moved without actually moving them.

Default is `false`

To set to a different value say `true`:
```
# in .groomba.toml
noop = true

# or in .groomba.yaml
noop: true

# or as an environment variable
GROOMBA_NOOP="true"
```

Note: Any truthy value will enable: `true`, `True`, `1` or any falsy value will disable: `false`, `False`, `0`

## Planned Improvements

List of enhancements for Groomba in no particular order:
- A good logo: every open source tool needs a good logo ;)
- Passing command line flags and arguments: currently I am planning on adding support for arguments and flags using [Cobra](https://github.com/spf13/cobra)
- Noop mode: add a noop mode where groomba just outputs changes it would have made without making them
- Custom rename prefix: currently Groomba only supports renaming to a `stale/` prefix. Ideally this should be a configurable parameter
- Delete (really) old branches: delete branches older than a set threshold instead of renaming them
- Better install scripts and documentation

## Bugs and feature requests

If you notice a bug or have a feature request please feel free to file an [issue](https://github.com/avbm/groomba/issues). Merge Requests with contributions or corrections are also welcome.
