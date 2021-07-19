# Groomba

[![GitHub Workflow Status](https://github.com/avbm/groomba/actions/workflows/ci.yml/badge.svg?style=flat)](https://github.com/avbm/groomba/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/badge/go%20version-%3E=1.16-61CFDD.svg?style=flat)](https://golang.org/doc/devel/release.html)
[![PkgGoDev](https://pkg.go.dev/badge/mod/github.com/avbm/groomba?style=flat)](https://pkg.go.dev/mod/github.com/avbm/groomba)

Groomba is a simple utility written in [Go](https://golang.org/) to groom your git repositories. It will rename branches older than a defined age. Unlike other tools like the [Stale Github Action](https://github.com/actions/stale), Groomba only depends on the git APIs and is agnostic of the software used to host your git repository. It will work just as well whether your repos are hosted in Github, Gitlab, Btbucket or something else.

## Installation

Download and run latest version:
```
curl -sL https://git.io/groomba | bash
```

Download and run specific version, ex v0.2.10:
```
curl -sL https://git.io/groomba | VERSION=v0.2.10 bash
```

You can add the snippet above as a step in your CI pipeline on your main branch to periodically groom your repository.

Install using `go install`
```
go install github.com/avbm/groomba/cmd/groomba@latest
```

Pre-built binaries are available only for x86_64 linux and OSX. For other architectures and OS systems please build your binaries using the [Build from souce](#build-from-source) section.

## Build from source

To build Groomba from source you need to have the golang compiler installed and configured. Clone this repository and from the root of the repository run:
```
$ cd cmd/groomba && go build .
```
This should fetch all dependencies and create a binary called `groomba` in the current directory.

## Configuration Options

To configure Groomba, you can set each configuration option in a `.groomba.toml` or `.groomba.yaml` file at the root of the repository you want to Groom. Alternately these options can also be set as environment variables. Options set as environment variables take the highest precedence.

| Name | Type | Default | Description |
|------|------|---------|-------------|
| Clobber           | bool | `false` | Toggle to enable or disable clobber mode |
| DryRun            | bool | `false` | Toggle to enable or disable dry run mode |
| MaxConcurrency    | uint8 | `4` | Set the maximum number of concurrent workers, set to 0 or 1 to disable concurrency |
| Prefix            | string | `stale/` | Identifier that will be added to the beginning of stale branch names to mark them as stale |
| StaleAgeThreshold | int | `14` | Threshold age in days for considering a branch as stale |
| StaticBranches    | []string | `["master", "main"]` | List of branches that are considered as `static` or `protected` and will be ignored |

### Clobber

`Clobber` is a bool that tells Groomba whether to run in clobber mode. In this mode, Groomba will clobber ie overwrite remote stale branches if they already exist and are not fast-forward merge-able. For example, if a repository has both branches `abc` and `stale/abc` already then with clobber mode enabled, branch `abc` will overwrite branch `stale/abc`. On the other hand if clobber mode is disabled(default), Groomba will fail to move `abc` to `stale/abc`.

Default: `false`

To set to a different value, say `true`:
```
# in .groomba.toml
clobber = true

# or in .groomba.yaml
clobber: true

# or as an environment variable
GROOMBA_CLOBBER="true"
```

Note: Any truthy value will enable: `true`, `True`, `1` or any falsy value will disable: `false`, `False`, `0`

### DryRun

`DryRun` is a bool that tells Groomba whether to run in dry run mode. In this mode, Groomba will only print out messages informing users about which branches would be moved without actually moving them.

Default: `false`

To set to a different value, say `true`:
```
# in .groomba.toml
dry_run = true

# or in .groomba.yaml
dry_run: true

# or as an environment variable
GROOMBA_DRY_RUN="true"
```

Note: Any truthy value will enable: `true`, `True`, `1` or any falsy value will disable: `false`, `False`, `0`

### MaxConcurrency

`MaxConcurrency` is a unit8 value that tells Groomba the number of worker processes to start. Each worker concurrently handles moving 1 branch.

Default: `4`

To set to a different value, say `10`:
```
# in .groomba.toml
max_concurrency = 10

# or in .groomba.yaml
max_concurrency: 10

# or as an environment variable
GROOMBA_DRY_RUN=10
```

Note: Since `MaxConcurrency` is a unit8 it can only be set to values from [0 ,255] inclusive. Setting the value to either 0 or 1 ensures only 1 worker is used ie only one branch is moved at a time.

### Prefix

`Prefix` is a string that will be added to the beginning of stale branch names to mark them as stale.

Default: `stale/`

To set to a different value, say `zzz_`, to keep stale branches at the bottom during sorted views:
```
# in .groomba.toml
prefix = zzz_

# or in .groomba.yaml
prefix: zzz_

# or as an environment variable
GROOMBA_PREFIX="zzz_"
```

### StaleAgeThreshold

`StaleAgeThreshold` is the threshold age in days for considering a branch as `stale`. It is expected to be an integer.

Default: `14`

To set to a different value, say `30`:
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

Default: `["master", "main"]`

To set to a different value, say `["latest", "staging", "production"]`:
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

## Planned Improvements

List of enhancements for Groomba in no particular order:
- A good logo: every open source tool needs a good logo ;)
- Passing command line flags and arguments: currently I am planning on adding support for arguments and flags using [Cobra](https://github.com/spf13/cobra)
- Delete (really) old branches: delete branches older than a set threshold instead of renaming them
- Add tests for failing to delete reference at remote

## Bugs and feature requests

If you notice a bug or have a feature request please feel free to file an [issue](https://github.com/avbm/groomba/issues). Merge Requests with contributions or corrections are also welcome.
