module github.com/portworx/pds-integration-test

go 1.16

require (
	github.com/golangci/golangci-lint v1.45.0
	github.com/imdario/mergo v0.3.12 // indirect
	github.com/portworx/pds-api-go-client v0.0.0-20220224153951-f33532c56f81
	github.com/stretchr/testify v1.7.0
	golang.org/x/net v0.0.0-20220127200216-cd36cc0744dd // indirect
	k8s.io/api v0.22.2
	k8s.io/apimachinery v0.22.2
	k8s.io/client-go v0.22.2

)

replace (
	github.anim.dreamworks.com/golang/logging => github.com/portworx/golang_logging v1.0.0
	github.anim.dreamworks.com/golang/rest => github.com/portworx/golang_rest v0.0.0-20200403193337-ceb5657f6c7c
	github.anim.dreamworks.com/golang/tonic => github.com/portworx/golang_tonic v1.3.0-rc5
	github.anim.dreamworks.com/golang/utils => github.com/portworx/golang_utils v0.0.0-20181008202924-011984e32408
)
