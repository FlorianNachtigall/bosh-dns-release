package server

import (
	"errors"
	"net"
	"math/rand"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	"github.com/cloudfoundry/bosh-utils/logger"
	"github.com/miekg/dns"
)

type Dialer func(string, string) (net.Conn, error)

//go:generate counterfeiter . Upcheck

type Upcheck interface {
	IsUp() error
}

type DNSAnswerValidatingUpcheck struct {
	target        string
	upCheckDomain string
	network       string
	logger        logger.Logger
}

func NewDNSAnswerValidatingUpcheck(target string, upcheckDomain string, network string, logger logger.Logger) Upcheck {
	return DNSAnswerValidatingUpcheck{
		target:        target,
		upCheckDomain: upcheckDomain,
		network:       network,
		logger:        logger,
	}
}

func (uc DNSAnswerValidatingUpcheck) IsUp() error {
	var err error
	uc.target, err = determineHost(uc.target)
	if err != nil {
		return uc.wrapError(err)
	}

	dnsClient := dns.Client{Net: uc.network}
	request := dns.Msg{
		Question: []dns.Question{
			{Name: uc.upCheckDomain, Qtype: dns.TypeA},
		},
	}
	request.Id = uint16(rand.Uint32())

	uc.logger.Debug("upcheck", "Sending upcheck %d to %s over %s", request.Id, uc.target, uc.network)
	msg, _, err := dnsClient.Exchange(&request, uc.target)

	if err != nil {
		return uc.wrapError(err)
	}
	if msg.Rcode != dns.RcodeSuccess {
		return uc.wrapError(errors.New("DNS resolve failed"))
	}

	if len(msg.Answer) == 0 {
		return uc.wrapError(errors.New("DNS upcheck found no answers"))
	}

	aRecord, ok := msg.Answer[0].(*dns.A)
	if !ok {
		return uc.wrapError(errors.New("upcheck must return A record"))
	}

	if !aRecord.A.Equal(net.ParseIP("127.0.0.1")) {
		return uc.wrapError(errors.New("DNS upcheck does not return the correct answer"))
	}

	return nil
}

func determineHost(target string) (string, error) {
	host, port, err := net.SplitHostPort(target)
	if err != nil {
		return "", err
	}

	if host == "0.0.0.0" {
		return net.JoinHostPort("127.0.0.1", port), nil
	}

	return target, nil
}

func (h DNSAnswerValidatingUpcheck) wrapError(err error) error {
	return bosherr.WrapErrorf(err, "on %s", h.network)
}
