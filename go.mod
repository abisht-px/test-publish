module github.com/portworx/pds-integration-test

go 1.16

require (
	github.com/Masterminds/semver v1.5.0
	github.com/coreos/go-oidc/v3 v3.2.0
	github.com/golangci/golangci-lint v1.45.0
	github.com/jackc/pgx/v4 v4.17.2
	github.com/portworx/pds-api-go-client v0.0.0-20220810014203-e9b075efe9b0
	github.com/stretchr/testify v1.8.0
	golang.org/x/net v0.0.0-20220127200216-cd36cc0744dd // indirect
	golang.org/x/oauth2 v0.0.0-20211104180415-d3ed0bb246c8
	gopkg.in/yaml.v2 v2.4.0
	helm.sh/helm/v3 v3.7.1
	k8s.io/api v0.22.1
	k8s.io/apimachinery v0.22.1
	k8s.io/cli-runtime v0.22.1
	k8s.io/client-go v0.22.1
)

replace (
	github.anim.dreamworks.com/golang/logging => github.com/portworx/golang_logging v1.0.0
	github.anim.dreamworks.com/golang/rest => github.com/portworx/golang_rest v0.0.0-20200403193337-ceb5657f6c7c
	github.anim.dreamworks.com/golang/tonic => github.com/portworx/golang_tonic v1.3.0-rc5
	github.anim.dreamworks.com/golang/utils => github.com/portworx/golang_utils v0.0.0-20181008202924-011984e32408
	// v0.21.4 the last supporting client.authentication.k8s.io/v1alpha1 (used by aws-iam-authenticator).
	k8s.io/api => k8s.io/api v0.21.4
	k8s.io/apimachinery => k8s.io/apimachinery v0.21.4
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.21.4
	k8s.io/client-go => k8s.io/client-go v0.21.4
)
