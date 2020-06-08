package sqlproxy

// Mostly copied from:
// https://github.com/GoogleCloudPlatform/cloudsql-proxy/blob/master/cmd/cloud_sql_proxy/proxy.go
// https://github.com/GoogleCloudPlatform/cloudsql-proxy/blob/master/cmd/cloud_sql_proxy/cloud_sql_proxy.go
// Not happy with it, but I cant import it due to "is a program, not an importable package"

import (
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/GoogleCloudPlatform/cloudsql-proxy/logging"
	"github.com/GoogleCloudPlatform/cloudsql-proxy/proxy/certs"
	"github.com/GoogleCloudPlatform/cloudsql-proxy/proxy/proxy"
)

const (
	minimumRefreshCfgThrottle = time.Second

	port = 3307
)

const dialersTimeout = time.Minute

type instanceConfig struct {
	Instance         string
	Network, Address string
}

// listenInstance starts listening on a new unix socket in dir to connect to the
// specified instance. New connections to this socket are sent to dst.
func listenInstance(dst chan<- proxy.Conn, cfg instanceConfig) (net.Listener, error) {
	l, err := net.Listen(cfg.Network, cfg.Address)
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			start := time.Now()
			c, err := l.Accept()
			if err != nil {
				logging.Errorf("Error in accept for %q on %v: %v", cfg, cfg.Address, err)
				if nerr, ok := err.(net.Error); ok && nerr.Temporary() {
					d := 10*time.Millisecond - time.Since(start)
					if d > 0 {
						time.Sleep(d)
					}
					continue
				}
				l.Close()
				return
			}
			logging.Verbosef("New connection for %q", cfg.Instance)

			switch clientConn := c.(type) {
			case *net.TCPConn:
				clientConn.SetKeepAlive(true)
				clientConn.SetKeepAlivePeriod(1 * time.Minute)

			}
			dst <- proxy.Conn{cfg.Instance, c}
		}
	}()

	logging.Infof("Listening on %s for %s", cfg.Address, cfg.Instance)
	return l, nil
}

func watchInstancesLoop(dir string, dst chan<- proxy.Conn, updates <-chan string, static map[string]net.Listener, cl *http.Client, cfgs []instanceConfig) {
	dynamicInstances := make(map[string]net.Listener)
	for range updates {
		list := cfgs

		stillOpen := make(map[string]net.Listener)
		for _, cfg := range list {
			instance := cfg.Instance

			// If the instance is specified in the static list don't do anything:
			// it's already open and should stay open forever.
			if _, ok := static[instance]; ok {
				continue
			}

			if l, ok := dynamicInstances[instance]; ok {
				delete(dynamicInstances, instance)
				stillOpen[instance] = l
				continue
			}

			l, err := listenInstance(dst, cfg)
			if err != nil {
				logging.Errorf("Couldn't open socket for %q: %v", instance, err)
				continue
			}
			stillOpen[instance] = l
		}

		// Any instance in dynamicInstances was not in the most recent metadata
		// update. Clean up those instances' sockets by closing them; note that
		// this does not affect any existing connections instance.
		for instance, listener := range dynamicInstances {
			logging.Infof("Closing socket for instance %v", instance)
			listener.Close()
		}

		dynamicInstances = stillOpen
	}

	for _, v := range static {
		if err := v.Close(); err != nil {
			logging.Errorf("Error closing %q: %v", v.Addr(), err)
		}
	}
	for _, v := range dynamicInstances {
		if err := v.Close(); err != nil {
			logging.Errorf("Error closing %q: %v", v.Addr(), err)
		}
	}
}

// WatchInstances handles the lifecycle of local sockets used for proxying
// local connections.  Values received from the updates channel are
// interpretted as a comma-separated list of instances.  The set of sockets in
// 'dir' is the union of 'instances' and the most recent list from 'updates'.
func WatchInstances(dir string, cfgs []instanceConfig, updates <-chan string, cl *http.Client) (<-chan proxy.Conn, error) {
	ch := make(chan proxy.Conn, 1)

	// Instances specified statically (e.g. as flags to the binary) will always
	// be available. They are ignored if also returned by the GCE metadata since
	// the socket will already be open.
	staticInstances := make(map[string]net.Listener, len(cfgs))
	for _, v := range cfgs {
		l, err := listenInstance(ch, v)
		if err != nil {
			return nil, err
		}
		staticInstances[v.Instance] = l
	}

	if updates != nil {
		go watchInstancesLoop(dir, ch, updates, staticInstances, cl, cfgs)
	}
	return ch, nil
}

// CreateProxy creates a proxy tunnel to a given instance
func CreateProxy(localPort int, instanceConnectionName CloudSQLInstance) {
	dir := "" // Not much idea what that is

	client := CreateHTTPAuthClient()

	cfgs := []instanceConfig{
		{Instance: string(instanceConnectionName), Network: "tcp", Address: net.JoinHostPort("0.0.0.0", strconv.Itoa(localPort))},
	}

	// We only need to store connections in a ConnSet if FUSE is used; otherwise
	// it is not efficient to do so.
	var connset *proxy.ConnSet

	// Initialize a source of new connections to Cloud SQL instances.
	var connSrc <-chan proxy.Conn
	updates := make(chan string)

	c, err := WatchInstances(dir, cfgs, updates, client)
	if err != nil {
		log.Fatal(err)
	}
	connSrc = c

	refreshCfgThrottle := time.Second
	logging.Infof("Ready for new connections")

	host := ""
	var maxConnections uint64 = 20
	proxyClient := &proxy.Client{
		Port:           port,
		MaxConnections: maxConnections,
		Certs: certs.NewCertSourceOpts(client, certs.RemoteOpts{
			APIBasePath:    host,
			IgnoreRegion:   true,
			UserAgent:      "goproxie",
			IPAddrTypeOpts: []string{"PUBLIC", "PRIVATE"},
		}),
		Conns:              connset,
		RefreshCfgThrottle: refreshCfgThrottle,
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT)

	var termTimeout time.Duration = 0
	go func() {
		<-signals
		logging.Infof("Received TERM signal. Waiting up to %s before terminating.", termTimeout)

		err := proxyClient.Shutdown(termTimeout)
		if err == nil {
			os.Exit(0)
		}
		logging.Errorf("Error during SIGTERM shutdown: %v", err)
		os.Exit(2)
	}()

	proxyClient.Run(connSrc)
	return
}
