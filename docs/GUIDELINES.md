# Development Guidelines

Note - These guidelines are not strictly enforced but are encouraged to maintain the test suites better and make our
lives easier.

`framework` contains all the necessary common code that is required to bootstrap a new test suite.

## Use of flags

It is always better to define test parameters (or inputs) as reusable flags. Flags are intuitive and their usage can
be defined in code itself. Flags can be marked as required so the user knows required parameters for tests to run.
Eg -

```go
flag.StringVar(
    &PDSHelmChartVersion,
    "pdsHelmChartVersion",
    "0",
    "PDS Helm Chart Version. 
\n - if value is 0 helm installation is skipped
\n - if value is empty, chart version is fetched from CP",
)

flag.StringVar(
    &TargetClusterKubeconfig,
    "targetClusterKubeconfig",
    "",
    "Path to target cluster's kubeconfig. For running tests within the cluster set this as empty",
)
```

A user can list the flags using `{binary} --help`

## Initialize what is needed

A very important aspect of making test suite lightweight is to initialize objects only that will be required by that
particular test suite.

For Eg -

```go
func (c *ControlPlane) MustInitializeTestData(
    ctx context.Context, t tests.T, 
    accountName, tenantName, projectName, namePrefix string,
) {
    c.mustHavePDStestAccount(ctx, t, accountName)
    c.mustHavePDStestTenant(ctx, t, tenantName)
    c.mustHavePDStestProject(ctx, t, projectName)
    c.mustLoadImageVersions(ctx, t)
    c.mustCreateStorageOptions(ctx, t, namePrefix)
    c.mustCreateApplicationTemplates(ctx, t, namePrefix)
}

cp.MustInitializeTestData(ctx, t, "", "", "", "")
```

This method initializes all the objects that belongs to a control plane. But not necessarily each test requires all of
these objects. We can let the client decide about the initializations by making use of optional arguments as following

```go
type InitializeOption func(context.Context, tests.T, *ControlPlane)

func WithAccountName(accountName string) InitializeOption {
    return func(ctx context.Context, t tests.T, c *ControlPlane) {
        c.mustHavePDStestAccount(ctx, t, accountName)
    }
}

func WithTenantName(tenantName string) InitializeOption {
    return func(ctx context.Context, t tests.T, c *ControlPlane) {
        c.mustHavePDStestTenant(ctx, t, tenantName)
    }
}

func WithProjectName(projectName string) InitializeOption {
    return func(ctx context.Context, t tests.T, c *ControlPlane) {
        c.mustHavePDStestProject(ctx, t, projectName)
    }
}

func WithLoadImageVersions() InitializeOption {
    return func(ctx context.Context, t tests.T, c *ControlPlane) {
        c.mustLoadImageVersions(ctx, t)
    }
}

func WithCreateTemplatesAndStorageOptions(namePrefix string) InitializeOption {
    return func(ctx context.Context, t tests.T, c *ControlPlane) {
        c.mustCreateStorageOptions(ctx, t, namePrefix)
        c.mustCreateApplicationTemplates(ctx, t, namePrefix)
    }
}

func (c *ControlPlane) MustInitializeTestDataWithOptions(ctx context.Context, t tests.T, opts ...InitializeOption) {
    for _, o := range opts {
        o(ctx, t, c)
    }
}

cp.MustInitializeTestDataWithOptions(
    ctx, t,
    WithAccountName(),
    WithTenantName(),
    WithProjectName(), 
    )
```

## Decorate Test Suites

Test suite decoration comes handy when we are debugging a failure by breaking a long-running test into subtests and
attaching relevant logging.

[Source Code](./examples/decorated_test.go)

Eg -

```go
go test -v ./examples/
```

## Add delays and retries

Wrap API calls with retries wherever needed. Make use of testify's
[Eventually](https://pkg.go.dev/github.com/stretchr/testify/assert#Eventuallyf) assertion.

Eg -

```go
func (c *ControlPlane) MustDeleteBackup(ctx context.Context, t tests.T, backupID string, localOnly bool) {
    resp, err := c.PDS.BackupsApi.ApiBackupsIdDelete(ctx, backupID).LocalOnly(localOnly).Execute()
    api.RequireNoError(t, resp, err)
}

func (c *ControlPlane) DeleteBackup(
    ctx context.Context, 
    t tests.T, 
    backupID string, 
    localOnly bool) (*http.Response, error) {
    return c.PDS.BackupsApi.ApiBackupsIdDelete(ctx, backupID).LocalOnly(localOnly).Execute()
}

// Without Eventually block -
controlPlane.MustDeleteBackup(ctx, s.T(), backup.GetId(), false)    

// With Eventually block
var err error
s.Eventually(func() bool {
    resp, err = controlPlane.DeleteBackup(ctx, s.T(), backup.GetId(), false)
    if err != nil {return false}
    return true
    },
    framework.DefaultTimeout,
    framework.DefaultPollPeriod,
    err,
)
```
