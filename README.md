# Drincw

Drincw (pronounced like `drink`) stands for "Drink Really Is Not Copying WissKI".
It's a quick tool to help with importing data into [WissKI](http://wiss-ki.eu/) using odbc.

* Work in progress *

## Usage

```bash
# generate a blank odbc for pathbuilder
go run ./cmd/makeodbc path/to/pathbuilder.xml

# format a pathbuilder as xml
go run ./cmd/pbfmt -pretty path/to/pathbuilder.xml

# format a pathbuilder as ascii text
go run ./cmd/pbfmt -ascii path/to/pathbuilder.xml
```