package portworx

import (
	"context"
	"errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

var ErrNoPXServiceFound = errors.New("no PX service found")

type Portworx struct {
	namespace  string
	restClient rest.Interface
}

func New(restClient rest.Interface, namespace string) Portworx {
	return Portworx{
		restClient: restClient,
		namespace:  namespace,
	}
}

func (p *Portworx) buildPXAPIRequest(baseRequest *rest.Request, pathSuffix string) *rest.Request {
	return baseRequest.Namespace(p.namespace).
		Resource("services").
		Name("portworx-api:9021").
		SubResource("proxy").
		Suffix(pathSuffix)
}

func FindPXNamespace(ctx context.Context, corev1 corev1.CoreV1Interface) (string, error) {
	services, err := corev1.Services("").
		List(ctx, metav1.ListOptions{
			FieldSelector: fields.SelectorFromSet(map[string]string{
				"metadata.name": "portworx-api",
			}).String(),
		})
	if err != nil {
		return "", err
	}
	if len(services.Items) == 0 {
		return "", ErrNoPXServiceFound
	}
	return services.Items[0].Namespace, nil
}
