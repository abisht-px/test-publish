package helminstaller

import (
	"fmt"
	"sync"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
)

// Option is a function that configures an MemoryRESTClientGetter.
type Option func(*MemoryRESTClientGetter)

// WithNamespace sets the namespace to use for the client.
func WithNamespace(namespace string) Option {
	return func(c *MemoryRESTClientGetter) {
		c.namespace = namespace
	}
}

// WithPersistent sets whether the client should persist the underlying client
// config, REST mapper, and discovery client.
func WithPersistent(persist bool) Option {
	return func(c *MemoryRESTClientGetter) {
		c.persistent = persist
	}
}

// MemoryRESTClientGetter is a resource.RESTClientGetter that uses an
// in-memory REST config, REST mapper, and discovery client.
// If configured, the client config, REST mapper, and discovery client are
// lazily initialized, and cached for subsequent calls.
type MemoryRESTClientGetter struct {
	// namespace is the namespace to use for the client.
	namespace string
	// impersonate is the username to use for the client.
	impersonate string
	// persistent indicates whether the client should persist the restMapper,
	// clientCfg, and discoveryClient. Rather than re-initializing them on
	// every call, they will be cached and reused.
	persistent bool

	cfg *rest.Config

	restMapper   meta.RESTMapper
	restMapperMu sync.Mutex

	discoveryClient discovery.CachedDiscoveryInterface
	discoveryMu     sync.Mutex

	clientCfg   clientcmd.ClientConfig
	clientCfgMu sync.Mutex
}

// setDefaults sets the default values for the MemoryRESTClientGetter.
func (c *MemoryRESTClientGetter) setDefaults() {
	if c.namespace == "" {
		c.namespace = "default"
	}
}

// NewMemoryRESTClientGetter returns a new MemoryRESTClientGetter.
func NewMemoryRESTClientGetter(cfg *rest.Config, opts ...Option) *MemoryRESTClientGetter {
	g := &MemoryRESTClientGetter{
		cfg: cfg,
	}
	for _, opts := range opts {
		opts(g)
	}
	g.setDefaults()
	return g
}

// ToRESTConfig returns the in-memory REST config.
func (c *MemoryRESTClientGetter) ToRESTConfig() (*rest.Config, error) {
	if c.cfg == nil {
		return nil, fmt.Errorf("MemoryRESTClientGetter has no REST config")
	}
	return c.cfg, nil
}

// ToDiscoveryClient returns a memory cached discovery client. Calling it
// multiple times will return the same instance.
func (c *MemoryRESTClientGetter) ToDiscoveryClient() (discovery.CachedDiscoveryInterface, error) {
	if c.persistent {
		return c.toPersistentDiscoveryClient()
	}
	return c.toDiscoveryClient()
}

func (c *MemoryRESTClientGetter) toPersistentDiscoveryClient() (discovery.CachedDiscoveryInterface, error) {
	c.discoveryMu.Lock()
	defer c.discoveryMu.Unlock()

	if c.discoveryClient == nil {
		discoveryClient, err := c.toDiscoveryClient()
		if err != nil {
			return nil, err
		}
		c.discoveryClient = discoveryClient
	}
	return c.discoveryClient, nil
}

func (c *MemoryRESTClientGetter) toDiscoveryClient() (discovery.CachedDiscoveryInterface, error) {
	config, err := c.ToRESTConfig()
	if err != nil {
		return nil, err
	}

	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, err
	}
	return memory.NewMemCacheClient(discoveryClient), nil
}

// ToRESTMapper returns a meta.RESTMapper using the discovery client. Calling
// it multiple times will return the same instance.
func (c *MemoryRESTClientGetter) ToRESTMapper() (meta.RESTMapper, error) {
	if c.persistent {
		return c.toPersistentRESTMapper()
	}
	return c.toRESTMapper()
}

func (c *MemoryRESTClientGetter) toPersistentRESTMapper() (meta.RESTMapper, error) {
	c.restMapperMu.Lock()
	defer c.restMapperMu.Unlock()

	if c.restMapper == nil {
		restMapper, err := c.toRESTMapper()
		if err != nil {
			return nil, err
		}
		c.restMapper = restMapper
	}
	return c.restMapper, nil
}

func (c *MemoryRESTClientGetter) toRESTMapper() (meta.RESTMapper, error) {
	discoveryClient, err := c.ToDiscoveryClient()
	if err != nil {
		return nil, err
	}
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(discoveryClient)
	return restmapper.NewShortcutExpander(mapper, discoveryClient), nil
}

// ToRawKubeConfigLoader returns a clientcmd.ClientConfig using
// clientcmd.DefaultClientConfig. With clientcmd.ClusterDefaults, namespace, and
// impersonate configured as overwrites.
func (c *MemoryRESTClientGetter) ToRawKubeConfigLoader() clientcmd.ClientConfig {
	if c.persistent {
		return c.toPersistentRawKubeConfigLoader()
	}
	return c.toRawKubeConfigLoader()
}

func (c *MemoryRESTClientGetter) toPersistentRawKubeConfigLoader() clientcmd.ClientConfig {
	c.clientCfgMu.Lock()
	defer c.clientCfgMu.Unlock()

	if c.clientCfg == nil {
		c.clientCfg = c.toRawKubeConfigLoader()
	}
	return c.clientCfg
}

func (c *MemoryRESTClientGetter) toRawKubeConfigLoader() clientcmd.ClientConfig {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	// use the standard defaults for this client command
	// DEPRECATED: remove and replace with something more accurate
	loadingRules.DefaultClientConfig = &clientcmd.DefaultClientConfig

	overrides := &clientcmd.ConfigOverrides{ClusterDefaults: clientcmd.ClusterDefaults}
	overrides.Context.Namespace = c.namespace
	overrides.AuthInfo.Impersonate = c.impersonate

	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, overrides)
}
