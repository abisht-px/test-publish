package test

import (
	"fmt"
	"path"
)

type api struct{
	baseURL string
}

func (a *api) endpoint(pathFragments ...string) string {
	return fmt.Sprintf("%s/%s", a.baseURL, path.Join(pathFragments...))
}
