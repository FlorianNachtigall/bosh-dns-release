package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"testing"
	"time"
)

func TestDNS(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "dns")
}

var (
	pathToServer string
)

var _ = SynchronizedBeforeSuite(func() []byte {
	path, err := gexec.Build("bosh-dns/dns")
	Expect(err).NotTo(HaveOccurred())
	SetDefaultEventuallyTimeout(2 * time.Second)
	SetDefaultEventuallyPollingInterval(500 * time.Millisecond)
	return []byte(path)
}, func(data []byte) {
	pathToServer = string(data)
})

var _ = SynchronizedAfterSuite(func() {
}, func() {
	gexec.CleanupBuildArtifacts()
})
