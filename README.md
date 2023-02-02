# Drincw

Drincw (pronounced like `drink`) stands for "Drincw Really Is Not Copying WissKI".
It's a set of tools to display data from a [WissKI](http://wiss-ki.eu/) instance and help importing data.

It contains two kinds of tools:

- A __WissKI Viewer__ and __WissKI Exporter__ to display or export data found in a WissKI based on a pathbuilder and a triplestore export
- An __ODBC generator__ and __Pathbuilder Formatter__ to simplify import of data

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

5. Run the exectuables, either by placing them in your `$PATH` or telling your interpreter where they are directly.

As an alternative to steps 4 and 5, you may also run executables directly:

```bash
go run ./cmd/hangover arguments...
```

Replace `hangover` with the name of the executable you want to run.

### from a binary

We publish binaries for Mac, Linux and Windows for every release.
These can be found on the releases page on GitHub. 

## Usage

### WissKI Viewer & Exporter

#### hangover - A WissKI Viewer

The `hangover` executable implements a WissKI Viewer.
It is invoked with two parameters, the pathbuilder to a pathbuilder xml `xml` and triplestore `nquads` export.
It then starts up a server at `localhost:3000` by default.

For example:

```bash
hangover schreibkalender.xml schreibkalender.nq
```

It supports a various set of other options, which can be found using  `hangover -help`.
The most important ones are:

- `-html`, `-images`: Automatically display html and image content found within the WissKI export. By default, these are only displayed as text.
- `-public`: Set the _public URL_ this dump originates from, for example `https://wisski.example.com/`. This automatically finds all references to it within the data dump with references to the local viewer.
- `-cache`: By default all indexes of the dataset required by the viewer are constructed in main memory. This can take several gigabytes. Instead, you can specify a temporary directory to read and write temporary indexes from.
- `-export`: Index the entire dataset, then dump the export in binary into a file. Afterwards `hangover` can be invoked using only such a file (as opposed to a pathbuilder and triplestore export), skipping the indexing step. The file format may change between different builds of drincw and should be treated as a blackbox.

#### n2j - A WissKI Viewer

n2j stands for `NQuads 2 JSON` and can convert a WissKI export into json (or more general, relational) format.
Like `hangover`, it takes both a pathbuilder and export as an argument.
By default, it produces a single `.json` file on standard output.

Further options supports a various set of other options, which can be found using  `hangover -help`.

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
