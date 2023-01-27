package core

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

type agentConfigLocked struct {
	bucket    string
	tlsConfig *tls.Config
	username  string
	password  string

	latestConfig *routeConfig
}

type Agent struct {
	lock   sync.Mutex
	config agentConfigLocked

	poller      ConfigPoller
	configMgr   *RouteConfigManager
	connMgr     KvClientManager
	collections CollectionResolver
	retries     RetryManager
	vbs         *vbucketRouter

	crud *CrudComponent
}

func CreateAgent(ctx context.Context, opts AgentOptions) (*Agent, error) {
	var srcHTTPAddrs []string
	for _, hostPort := range opts.HTTPAddrs {
		if opts.TLSConfig == nil {
			ep := fmt.Sprintf("http://%s", hostPort)
			srcHTTPAddrs = append(srcHTTPAddrs, ep)
		} else {
			ep := fmt.Sprintf("https://%s", hostPort)
			srcHTTPAddrs = append(srcHTTPAddrs, ep)
		}
	}

	agent := &Agent{
		config: agentConfigLocked{
			bucket:    opts.BucketName,
			tlsConfig: opts.TLSConfig,
			username:  opts.Username,
			password:  opts.Password,
		},

		poller: newhttpConfigPoller(srcHTTPAddrs, httpPollerProperties{
			ConfHTTPRetryDelay:   10 * time.Second,
			ConfHTTPRedialPeriod: 10 * time.Second,
			ConfHTTPMaxWait:      5 * time.Second,
			HttpClient:           http.DefaultClient,
			BucketName:           opts.BucketName,
			Username:             opts.Username,
			Password:             opts.Password,
		}),
		configMgr: newConfigManager(),
		retries:   NewRetryManagerFastFail(),
	}

	clients := make(map[string]*KvClientConfig)
	for _, addr := range opts.MemdAddrs {
		clients[addr] = &KvClientConfig{
			Address:        addr,
			TlsConfig:      agent.config.tlsConfig,
			SelectedBucket: agent.config.bucket,
			Username:       agent.config.username,
			Password:       agent.config.password,
		}
	}
	connMgr, err := NewKvClientManager(&KvClientManagerConfig{
		NumPoolConnections: 1,
		Clients:            clients,
	}, nil)
	if err != nil {
		return nil, err
	}
	agent.connMgr = connMgr

	collections, err := NewCollectionResolverCached(&CollectionResolverCachedOptions{
		Resolver: &CollectionResolverMemd{
			connMgr: agent.connMgr,
		},
		ResolveTimeout: 10 * time.Second,
	})
	if err != nil {
		return nil, err
	}
	agent.collections = collections

	agent.vbs = newVbucketRouter()

	agent.configMgr.RegisterCallback(func(rc *routeConfig) {
		agent.lock.Lock()
		agent.config.latestConfig = rc
		agent.updateStateLocked()
		agent.lock.Unlock()
	})

	err = agent.startConfigWatcher(ctx)
	if err != nil {
		return nil, err
	}

	agent.crud = &CrudComponent{
		collections: agent.collections,
		retries:     agent.retries,
		// errorResolver: new,
		connManager: agent.connMgr,
		vbs:         agent.vbs,
	}

	return agent, nil
}

func (agent *Agent) updateStateLocked() {
	log.Printf("updating config: %+v %+v", agent.config, *agent.config.latestConfig)
	routeCfg := agent.config.latestConfig

	var mgmtList []string
	var serverList []string
	if agent.config.tlsConfig == nil {
		serverList = make([]string, len(routeCfg.kvServerList.NonSSLEndpoints))
		copy(serverList, routeCfg.kvServerList.NonSSLEndpoints)
		mgmtList = make([]string, len(routeCfg.mgmtEpList.NonSSLEndpoints))
		copy(mgmtList, routeCfg.mgmtEpList.NonSSLEndpoints)
	} else {
		serverList = make([]string, len(routeCfg.kvServerList.SSLEndpoints))
		copy(serverList, routeCfg.kvServerList.SSLEndpoints)
		mgmtList = make([]string, len(routeCfg.mgmtEpList.SSLEndpoints))
		copy(mgmtList, routeCfg.mgmtEpList.SSLEndpoints)
	}

	// TODO(brett19): Need to make this ADD the new endpoints first, then update the
	// vbucket map, and then reconfigure again to drop the old endpoints.  Otherwise
	// vbucket mapping and connection dispatch will race and loop.

	clients := make(map[string]*KvClientConfig)
	for _, addr := range serverList {
		clients[addr] = &KvClientConfig{
			Address:        addr,
			TlsConfig:      agent.config.tlsConfig,
			SelectedBucket: agent.config.bucket,
			Username:       agent.config.username,
			Password:       agent.config.password,
		}
	}
	agent.connMgr.Reconfigure(&KvClientManagerConfig{
		NumPoolConnections: 1,
		Clients:            clients,
	})

	agent.vbs.UpdateRoutingInfo(&vbucketRoutingInfo{
		vbmap:      routeCfg.vbMap,
		serverList: serverList,
	})

	agent.poller.UpdateEndpoints(mgmtList)
}

func (agent *Agent) startConfigWatcher(ctx context.Context) error {
	configCh, err := agent.poller.Watch(ctx)
	if err != nil {
		return err
	}

	var firstConfig *TerseConfigJsonWithSource
	select {
	case config := <-configCh:
		firstConfig = config
	case <-ctx.Done():
		return ctx.Err()
	}

	agent.configMgr.ApplyConfig(firstConfig.SourceHostname, firstConfig.Config)

	go func() {
		for config := range configCh {
			agent.configMgr.ApplyConfig(config.SourceHostname, config.Config)
		}
	}()

	return nil
}
