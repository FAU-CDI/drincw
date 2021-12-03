# Drincw

Drincw (pronounced like `drink`) stands for "Drink Really Is Not Copying WissKI".
It's a quick tool to help with importing data into [WissKI](http://wiss-ki.eu/) using odbc.

*Work in progress*

*Documentation may be oudated and incomplete*

## Installation

### from source

1. Install [Go](https://go.dev/), Version 1.17 or newer
2. Clone this repository somewhere.
3. Fetch dependencies using standard go tools:

```bash
go get ./cmd/makeodb ./cmd/pbfmt
```

4. Build the executables from the `cmd` subdirectory:

```bash
go build ./cmd/makeodb -o makeodbc
go build ./cmd/pbfmt -o pbfmt
```

5. Run the exectuables, either by placing them in your `$PATH` or telling your interpreter where they are directly.

As an alternative to steps 4 and 5, you may also run executables directly:

```bash
go run ./cmd/makeodbc arguments...
go run ./cmd/pbfmt arguments...
```

## Usage


Each executable takes a pathbuilder as an argument.
This can be given either as a (relative or absolute) path or a http(s) URL.

## pbfmt - Formatting a pathbuilder

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

### makeodbc - Generating an odbc

The `makeodbc` executable can be used to generate an `odbc` import script from a pathbuilder.

#### Basic usage

```bash
# generate a basic odbc from a pathbuilder
makeodbc path/to/pathbuilder.xml
```

#### Using selectors

Selectors can be used to limit the fields to be imported and determine the sql statements to generate.
`selector` files are json with comments. 

```bash
# generate a sample selector file for the pathbuilder
makeodbc -dump-selectors path/to/pathbuilder.xml

# generate an odbc using only bundles and fields available in the given selectors
makeodbc -load-selectors path/to/selectors.json path/to/pathbuilder.xml
```

#### Previewing SQL

To generate the sql a particular import would run:

```bash
# print the sql that importing the given bundle would run
makeodbc -sql my_bundle_name path/to/pathbuilder.xml

# print the sql that importing the given bundle and odbc would run
makeodbc -sql my_bundle_name -load-selectors path/to/selectors.json  path/to/pathbuilder.xml
```

## License

Licensed under the terms of [AGPL 3.0](https://github.com/FAU-CDI/drincw/blob/main/LICENSE) for everyone.
Aditionally licensed under the terms of the standard GPL license, version 3, for internal usage at FAU-CDI only. 
