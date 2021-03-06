package dns

import (
	"context"
	"net"
	"strings"
	"time"

	"github.com/qdm12/private-internet-access-docker/internal/constants"
)

// UseDNSInternally is to change the Go program DNS only
// It adds a 300ms delay between each DNS request as Unbound refuses connection if they are too close
// to each other from the same source port
func (c *configurator) UseDNSInternally(ip net.IP) {
	c.logger.Info("using DNS address %s internally", ip.String())
	net.DefaultResolver = &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{}
			c.internalResolverMutex.Lock()
			time.AfterFunc(300*time.Millisecond, func() {
				c.internalResolverMutex.Unlock()
			})
			return d.DialContext(ctx, "udp", net.JoinHostPort(ip.String(), "53"))
		},
	}
}

// UseDNSSystemWide changes the nameserver to use for DNS system wide
func (c *configurator) UseDNSSystemWide(ip net.IP) error {
	c.logger.Info("using DNS address %s system wide", ip.String())
	data, err := c.fileManager.ReadFile(string(constants.ResolvConf))
	if err != nil {
		return err
	}
	s := strings.TrimSuffix(string(data), "\n")
	lines := strings.Split(s, "\n")
	if len(lines) == 1 && lines[0] == "" {
		lines = nil
	}
	found := false
	for i := range lines {
		if strings.HasPrefix(lines[i], "nameserver ") {
			lines[i] = "nameserver " + ip.String()
			found = true
		}
	}
	if !found {
		lines = append(lines, "nameserver "+ip.String())
	}
	data = []byte(strings.Join(lines, "\n"))
	return c.fileManager.WriteToFile(string(constants.ResolvConf), data)
}
