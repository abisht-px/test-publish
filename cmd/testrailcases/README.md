# Sync Testcase Data 

1. Create a testdata file by running the below tool :
```bash
make doc | jq -r '.'
```

2. Run the go command directly from the root of the project : 
```bash
go run ./cmd/testrailcases/post-testdata.go
```

## Environment Configuration

Add below variables in the existing env file by refering to secret manager 'testrail-api-secret' from the console :

| Key                        | Default value                              | Description                                                                                                       |
|----------------------------|--------------------------------------------|-------------------------------------------------------------------------------------------------------------------|
| TESTRAIL_USER          |                                            | User to authenticate api requests to testrail.                                                             |
| TESTRAIL_API_KEY  |                                            | Api key to authenticate api requests to testrail                                    |
| TESTDATA_PATH  |                                            | Path to the testdata file generated from doc tool

```dotenv
# Testrail related variables.
TESTRAIL_USER=''
TESTRAIL_API_KEY=''
TESTDATA_PATH=''
```

