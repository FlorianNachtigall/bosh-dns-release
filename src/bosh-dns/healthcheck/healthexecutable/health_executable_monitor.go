package healthexecutable

import (
	"strings"
	"time"

	"sync"

	"code.cloudfoundry.org/clock"
	"github.com/cloudfoundry/bosh-utils/logger"
	"github.com/cloudfoundry/bosh-utils/system"
)

type HealthExecutableMonitor struct {
	healthExecutablePaths []string
	cmdRunner             system.CmdRunner
	clock                 clock.Clock
	interval              time.Duration
	shutdown              chan struct{}
	status                bool
	mutex                 *sync.Mutex
	logger                logger.Logger
}

func NewHealthExecutableMonitor(
	healthExecutablePaths []string,
	cmdRunner system.CmdRunner,
	clock clock.Clock,
	interval time.Duration,
	shutdown chan struct{},
	logger logger.Logger,
) *HealthExecutableMonitor {
	monitor := &HealthExecutableMonitor{
		healthExecutablePaths: healthExecutablePaths,
		cmdRunner:             cmdRunner,
		clock:                 clock,
		interval:              interval,
		shutdown:              shutdown,
		mutex:                 &sync.Mutex{},
		logger:                logger,
	}

	monitor.runChecks()
	go monitor.run()

	return monitor
}

func (m *HealthExecutableMonitor) Status() bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.status
}

func (m *HealthExecutableMonitor) run() {
	timer := m.clock.NewTimer(m.interval)
	m.logger.Debug("HealthExecutableMonitor", "starting monitor for [%s] with interval %v", strings.Join(m.healthExecutablePaths, ", "), m.interval)

	for {
		select {
		case <-m.shutdown:
			m.logger.Debug("HealthExecutableMonitor", "stopping")
			timer.Stop()
			return
		case <-timer.C():
			m.runChecks()
			timer.Reset(m.interval)
		}
	}
}

func (m *HealthExecutableMonitor) runChecks() {
	var allSucceeded = true

	for _, executable := range m.healthExecutablePaths {
		_, _, exitStatus, err := m.runExecutable(executable)
		if err != nil {
			allSucceeded = false
			m.logger.Error("HealthExecutableMonitor", "Error occurred executing '%s': %v", executable, err)
		} else if exitStatus != 0 {
			allSucceeded = false
		}
	}

	m.mutex.Lock()
	m.status = allSucceeded
	m.mutex.Unlock()
}
