# go-testify-report

A simple tool to generate jUnit results from go test logs.

This tool is a wrapper over github.com/jstemmer/go-junit-report/, with following additional support -

1. Generate report from multiple files in a directory
2. Generate report from a URL
3. Generate report from a zipped file

## Usage

```shell
make build

./bin/go-testify-report --help

# Generate report from log files in a directory
./bin/go-testify-report -d <dir-path>

# Generate report from a URL
./bin/go-testify-report -f <url>

# Generate report from log files in a zipped directory
./bin/go-testify-report -gzip <zipped-file>

# Generate report from log files in a zipped directory by downloading from a URL
./bin/go-testify-report -gzip <zipped-file-url>
```

## Docker Usage

### Help Command

```shell
docker run --rm portworx/pds-integration-test-tools:dev-a09233fc-dirty /bin/go-testify-report --help
```

Output -

```shell
Usage of /bin/go-testify-report:
  -d (string) Directory path containing log files
  -f (string) Log file name
  -gzip (string) Zipped File path
  -o (string) File path to write report to
  -sysLogs (boolean) Include sys logs
```

### Attaching a volume

```shell
docker run --rm -v $(pwd):/reports go-testify-report:latest /bin/go-testify-report -d=<dir-path>
```

### Save report

```shell
docker run --rm -v $(pwd):/reports go-testify-report:latest /bin/go-testify-report -f=<url> -o=./report.xml
```
