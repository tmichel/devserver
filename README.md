# devserver

devserver is a reverse proxy and build server for web applications.

## Features

* Rebuild and restart when you hit Enter.
* Live reload on restarts and file changes
* Hot-reloading for CSS files: CSS files used via a `<link>` tag are updated in
  place without reloading the page.

## Installation

    go install github.com/tmichel/devserver@latest

Under the hood `devserver` uses [fswatch][1] for monitoring file changes.

To install fswatch on macOS run the following

    brew install fswatch

Detailed install instructions can be found in [fswatch's README][2].

## Usage

For details run

    devserver -h

### Example

    cd /path/to/your/project
    devserver \
        -build-cmd "go build -o bin/my-app"  \
        -web-root $PWD/files \
        "bin/my-app -addr {}"
    # visit http://localhost:8080

`serverCmd` defines the command for starting the server. Use `{}` as a
placeholder for the host and port. It is passed in the format of `host:port`.
`{port}` and `{host}` can also be used as placeholders.

`-build-cmd` defines the build command. Defaults to `make`. Set it to an empty
string to skip the build.

`-web-root` sets the root directory for files served by your web server. It is
used to compute absolute URLs for files. For example when `-web-root
/path/to/web-root` is set the file located at `/path/to/web-root/css/style.css`
will be reported as `/css/style.css`. This allows hot reloading CSS files.

### Example: using `go run`

Sometimes building and running are not separate. For example when using `go
run` to build and execute your program there is no separate build step. In
these situations you can specify `-build-cmd=""` to skip the build.

    cd /path/to/your/project
    devserver \
        -build-cmd "" \
        -web-root $PWD/files \
        "go run . -addr {}"
    # visit http://localhost:8080

[1]: https://github.com/emcrisostomo/fswatch
[2]: https://github.com/emcrisostomo/fswatch#installation
