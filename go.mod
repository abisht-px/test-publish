module github.com/portworx/pds-integration-test

go 1.16

require (
	github.anim.dreamworks.com/DreamCloud/stella-api v0.0.0-00010101000000-000000000000
	github.com/google/uuid v1.2.0
	github.com/stretchr/testify v1.7.0
)

replace (
	github.anim.dreamworks.com/DreamCloud/stella-api => github.com/portworx/DreamCloud_stella-api v0.0.0-20211001065015-f32adb49906c
	github.anim.dreamworks.com/golang/logging => github.com/portworx/golang_logging v1.0.0
	github.anim.dreamworks.com/golang/rest => github.com/portworx/golang_rest v0.0.0-20200403193337-ceb5657f6c7c
	github.anim.dreamworks.com/golang/tonic => github.com/portworx/golang_tonic v1.3.0-rc5
	github.anim.dreamworks.com/golang/utils => github.com/portworx/golang_utils v0.0.0-20181008202924-011984e32408
)
