package portworx

import (
	"context"
)

func (p *Portworx) GetPXVolumes(
	ctx context.Context,
) ([]byte, error) {
	return p.buildPXAPIRequest(
		p.restClient.Post(),
		"v1/volumes/inspectwithfilters",
	).Do(ctx).Raw()
}

func (p *Portworx) DeletePXVolume(
	ctx context.Context,
	volumeId string,
) ([]byte, error) {
	return p.buildPXAPIRequest(
		p.restClient.Delete(),
		"v1/volumes/"+volumeId,
	).Do(ctx).Raw()
}
