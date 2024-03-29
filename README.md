# Drincw

Drincw (pronounced like `drink`) stands for "Drincw Really Is Not Copying WissKI".
It's a set of tools to display data from a [WissKI](http://wiss-ki.eu/) instance and help importing data.

It also contains __ODBC generator__ and __Pathbuilder Formatter__ to simplify import of data.
The WissKI viewer and related tools have been moved the [hangover](https://github.com/FAU-CDI/hangover) repository.

#### Why the name?
Cause this project really isn't copying WissKI or intended to fully replace it.
It is an auxiliary tool on top. 

*Work in progress*

*Documentation may be oudated and incomplete*

## Installation

### from source

1. Install [Go](https://go.dev/), Version 1.18 or newer
2. Install [Yarn](https://yarnpkg.com/) (to build some frontends)
3. Clone this repository somewhere.
4. Fetch dependencies:

```bash
make deps
```

5. Use the `Makefile` to build dependencies into the `dist` directory:

```bash
make all
```

5. Run the executables, either by placing them in your `$PATH` or telling your interpreter where they are directly.

As an alternative to steps 4 and 5, you may also run executables directly:

```bash
go run ./cmd/hangover arguments...
```

Replace `hangover` with the name of the executable you want to run.

### from a binary

We publish binaries for Mac, Linux and Windows for every release.
These can be found on the releases page on GitHub. 

## Usage

### WissKI Import

Each executable takes a pathbuilder as an argument.
This can be given either as a (relative or absolute) path or a http(s) URL.

#### pbfmt - Formatting a pathbuilder

The `pbfmt` executable takes two logical parameters, the pathbuilder to format and the mode to format it in)

Pathbuilders can be formatted in three ways:

- XML (default)
- Prettyfied XML (`-pretty`)
- ASCII text (`-ascii`)


Examples:

```bash
# Format the pathbuilder stored in pathbuilder.xml as ascii
pbfmt -ascii pathbuilder.xml

# Format the pathbuilder from the provided url as pretty xml
pbfmt -pretty https://mywisski.example.com/sites/default/files/wisski_pathbuilder/export/default_00000000T000000

# Format the pathbuilder as xml
pbfmt pathbuilder.xml
```

#### makeodbc - Generating an odbc

The `makeodbc` executable can be used to generate an `odbc` import script from a pathbuilder.

##### Basic usage

```bash
# generate a basic odbc from a pathbuilder
makeodbc path/to/pathbuilder.xml
```

##### Using selectors

Selectors can be used to limit the fields to be imported and determine the sql statements to generate.
`selector` files are json with comments. 

```bash
# generate a sample selector file for the pathbuilder
makeodbc -dump-selectors path/to/pathbuilder.xml

# generate an odbc using only bundles and fields available in the given selectors
makeodbc -load-selectors path/to/selectors.json path/to/pathbuilder.xml
```

##### Previewing SQL

To generate the sql a particular import would run:

```bash
# print the sql that importing the given bundle would run
makeodbc -sql my_bundle_name path/to/pathbuilder.xml

# print the sql that importing the given bundle and odbc would run
makeodbc -sql my_bundle_name -load-selectors path/to/selectors.json  path/to/pathbuilder.xml
```

```bash
# alternatively do this directly on the generated odbc
dummysql /path/to/odbc.xml tablename
```

#### addict - gui for makeodbc

An experimental gui for makeodbc.

#### odbcd - web interface for generating odbc via the browser

```bash
# build 
cd cmd/odbcd && yarn install && yarn dist
# run the executable
go run . -listen localhost:8080
```

For debugging (with live reloading)
```bash
# start the development server for the frontend
cd cmd/odbcd && yarn install && yarn dev

# start the debugging server
go run -tags debug ./cmd/odbcd
```

An example instance is running at https://odbc.kwarc.info/. 

#### ps2 - generate sparql queries for a field

Generate a simple sparql query to view values of a single field.

```bash
ps2 path/to/pathbuilder.xml name-of-some-path
```

#### pbdot - generate a dot graph from a pathbuilder

This is a (highly experimental) program that renders a bundle into a dot file for graphviz.
For example to generate an svg image representing a specific bundle (and child bundles) use:

```bash
echo '{"ecrm":"http://erlangen-crm.org/170309/"}' | pbdot -prefixes - /path/to/pathbuilder.xml bundlename | dot -T svg > output.svg
```

## Deployment


[![Publish Docker Image](https://github.com/FAU-CDI/drincw/actions/workflows/docker.yml/badge.svg)](https://github.com/FAU-CDI/drincw/actions/workflows/docker.yml)

Only the Data Importer is available as a Docker Image on [GitHub Packages](https://github.com/FAU-CDI/drincw/pkgs/container/odbcd) at the moment.
Automatically built on every commit.

```bash
 docker run -ti -p 8080:8080 ghcr.io/fau-cdi/odbcd
```

## Development

During development standard go tools are used.
Commands can be found in `./cmd/`.
Packages are documented and tested where applicable. 

Some files are generated, in particular the legal notices and frontend assets.
This requires some external tools written in go.
The frontend assets furthermore require node packages to be installed using [yarn](https://yarnpkg.com/).

A Makefile exists to simply the setup on a fresh system.
To install all (go) dependencies required for a build, run `make deps`.
To regenerate all assets, run `make generate`.
To build the `dist` directory, run `make all`.

go executables remain buildable without installing external dependencies.

## License

Licensed under the terms of [AGPL 3.0](https://github.com/FAU-CDI/drincw/blob/main/LICENSE) for everyone.
Aditionally licensed under the terms of the standard GPL license, version 3, for internal usage at FAU-CDI only. 
