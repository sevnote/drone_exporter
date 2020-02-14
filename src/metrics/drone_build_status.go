package metrics

import (
	"drone_exporter/src/lib"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
)

type DroneStatus struct {
	Server             string
	PendingCurrentDesc *prometheus.Desc
	RunningCurrentDesc *prometheus.Desc
	SuccessTotalDesc   *prometheus.Desc
	KilledCurrentDesc  *prometheus.Desc
	FailureCurrentDesc *prometheus.Desc
}

// Simulate prepare the data
func (c *DroneStatus) ReallyExpensiveAssessmentOfTheSystemState() (
	pendingCurrent map[string]float64, runningCurrent map[string]float64, successTotal map[string]float64, killedCurrent map[string]float64, failureCurrent map[string]float64) {

	//Connection to MySQL
	droneDBConn, err := lib.OpenMysqlDB("192.168.25.154", 3306, "root", "Mm123456", "drone")
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	defer droneDBConn.Close()

	var pending_query = "SELECT repo_slug AS repo,count(*) AS count FROM builds INNER JOIN repos ON builds.build_repo_id=repos.repo_id WHERE builds.build_status='pending' AND builds.build_created>=UNIX_TIMESTAMP (CURRENT_TIMESTAMP-INTERVAL 5 MINUTE) GROUP BY repos.repo_slug"
	var running_query = "SELECT repo_slug AS repo,count(*) AS count FROM builds INNER JOIN repos ON builds.build_repo_id=repos.repo_id WHERE builds.build_status='running' AND builds.build_created>=UNIX_TIMESTAMP (CURRENT_TIMESTAMP-INTERVAL 5 MINUTE) GROUP BY repos.repo_slug"
	var success_query = "SELECT repo_slug AS repo,count(*) AS count FROM builds INNER JOIN repos ON builds.build_repo_id=repos.repo_id WHERE builds.build_status='success'  GROUP BY repos.repo_slug"
	var killed_query = "SELECT repo_slug AS repo,count(*) AS count FROM builds INNER JOIN repos ON builds.build_repo_id=repos.repo_id WHERE builds.build_status='killed'  AND builds.build_created>=UNIX_TIMESTAMP (CURRENT_TIMESTAMP-INTERVAL 10 MINUTE) GROUP BY repos.repo_slug"
	var failure_query = "SELECT repo_slug AS repo,count(*) AS count FROM builds INNER JOIN repos ON builds.build_repo_id=repos.repo_id WHERE builds.build_status='failure'  AND builds.build_created>=UNIX_TIMESTAMP (CURRENT_TIMESTAMP-INTERVAL 5 MINUTE) GROUP BY repos.repo_slug"
	pendingResult, err := droneDBConn.Query(pending_query)
	runningResult, err := droneDBConn.Query(running_query)
	successResult, err := droneDBConn.Query(success_query)
	killedResult, err := droneDBConn.Query(killed_query)
	failureResult, err := droneDBConn.Query(failure_query)
	if err != nil {
		return
	}

	pendingCurrent = make(map[string]float64)
	runningCurrent = make(map[string]float64)
	successTotal = make(map[string]float64)
	killedCurrent = make(map[string]float64)
	failureCurrent = make(map[string]float64)
	for pendingResult.Next() {
		var count float64
		var repo string
		err = pendingResult.Scan(&repo, &count)
		pendingCurrent[repo] = count
	}

	for runningResult.Next() {
		var count float64
		var repo string
		err = runningResult.Scan(&repo, &count)
		runningCurrent[repo] = count
	}

	for successResult.Next() {
		var count float64
		var repo string
		err = successResult.Scan(&repo, &count)
		successTotal[repo] = count
	}

	for killedResult.Next() {
		var count float64
		var repo string
		err = killedResult.Scan(&repo, &count)
		killedCurrent[repo] = count
	}

	for failureResult.Next() {
		var count float64
		var repo string
		err = failureResult.Scan(&repo, &count)
		failureCurrent[repo] = count
	}

	//End of Conver,

	return
}

// Describe simply sends the two Descs in the struct to the channel.
func (c *DroneStatus) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.PendingCurrentDesc
	ch <- c.RunningCurrentDesc
	ch <- c.KilledCurrentDesc
	ch <- c.SuccessTotalDesc
	ch <- c.FailureCurrentDesc
}

func (c *DroneStatus) Collect(ch chan<- prometheus.Metric) {
	pendingCurrent, runningCurrent, successTotal, killedCurrent, failureCurrent := c.ReallyExpensiveAssessmentOfTheSystemState()
	for repo, count := range pendingCurrent {
		ch <- prometheus.MustNewConstMetric(
			c.PendingCurrentDesc,
			prometheus.GaugeValue,
			count,
			repo,
		)
	}
	for repo, count := range runningCurrent {
		ch <- prometheus.MustNewConstMetric(
			c.RunningCurrentDesc,
			prometheus.GaugeValue,
			count,
			repo,
		)
	}
	for repo, count := range successTotal {
		ch <- prometheus.MustNewConstMetric(
			c.SuccessTotalDesc,
			prometheus.GaugeValue,
			count,
			repo,
		)
	}
	for repo, count := range killedCurrent {
		ch <- prometheus.MustNewConstMetric(
			c.KilledCurrentDesc,
			prometheus.GaugeValue,
			count,
			repo,
		)
	}
	for repo, count := range failureCurrent {
		ch <- prometheus.MustNewConstMetric(
			c.FailureCurrentDesc,
			prometheus.GaugeValue,
			count,
			repo,
		)
	}

}

func NewMetrics(server string) *DroneStatus {
	return &DroneStatus{
		Server: server,
		PendingCurrentDesc: prometheus.NewDesc(
			"drone_build_pending_current",
			"Number of Prending Current Count.",
			[]string{"repo"},
			prometheus.Labels{"server": server},
		),
		RunningCurrentDesc: prometheus.NewDesc(
			"drone_build_running_current",
			"Number of Running Current Count.",
			[]string{"repo"},
			prometheus.Labels{"server": server},
		),
		SuccessTotalDesc: prometheus.NewDesc(
			"drone_build_success_total",
			"Number of Success Total.",
			[]string{"repo"},
			prometheus.Labels{"server": server},
		),
		KilledCurrentDesc: prometheus.NewDesc(
			"drone_build_killed_current",
			"Number of Killed Current Count.",
			[]string{"repo"},
			prometheus.Labels{"server": server},
		),
		FailureCurrentDesc: prometheus.NewDesc(
			"drone_build_failure_current",
			"Number of Failure Current Count.",
			[]string{"repo"},
			prometheus.Labels{"server": server},
		),
	}
}
