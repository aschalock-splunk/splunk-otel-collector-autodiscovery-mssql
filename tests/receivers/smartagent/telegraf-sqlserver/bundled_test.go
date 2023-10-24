// Copyright Splunk, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build integration

package tests

import (
	"path"
	"runtime"
	"testing"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

func TestMssqlDockerObserver(t *testing.T) {
	testutils.SkipIfNotContainerTest(t)
	if runtime.GOOS == "darwin" {
		t.Skip("unable to share sockets between mac and d4m vm: https://github.com/docker/for-mac/issues/483#issuecomment-758836836")
	}

	server := testutils.NewContainer().WithContext(
		path.Join(".", "testdata", "server"),
	).WithExposedPorts("1433:1433").WithName("sql-server").WithNetworks(
		"mssql",
	).WillWaitForPorts("1433").WillWaitForLogs(
		"SQL Server is now ready for client connections.", "Recovery is complete.")

	client := testutils.NewContainer().WithContext(
		path.Join(".", "testdata", "client"),
	).WithName("sql-client").WithNetworks("mssql").WillWaitForLogs("name", "signalfxagent")

	containers := []testutils.Container{server, client}

	//testutils.AssertAllMetricsReceived(t, "all.yaml", "all_metrics_config.yaml", containers, nil)
	testutils.AssertAllMetricsReceived(t, "bundled.yaml", "otlp_exporter.yaml",
		containers, []testutils.CollectorBuilder{
			func(c testutils.Collector) testutils.Collector {
				cc := c.(*testutils.CollectorContainer)
				cc.Container = cc.Container.WithBinds("/var/run/docker.sock:/var/run/docker.sock:ro")
				return cc
			},
			func(collector testutils.Collector) testutils.Collector {
				return collector.WithEnv(map[string]string{
					"MSSQL_URL":                 "tcp:sql-server,1433",
					"SPLUNK_DISCOVERY_DURATION": "10s",
					// confirm that debug logging doesn't affect runtime
					"SPLUNK_DISCOVERY_LOG_LEVEL": "debug",
					"splunk.discovery.default":   "Password!",
					"HOSTNAME":                   "sql.example.com",
					"login.name":                 "signalfxagent",
				}).WithArgs(
					"--discovery",
					"--set", "splunk.discovery.receivers.mssql.config.endpoint=localhost:1433",
					"--set", `splunk.discovery.extensions.k8s_observer.enabled=false`,
					"--set", `splunk.discovery.extensions.host_observer.enabled=false`,
					"--set", `splunk.discovery.receivers.mssql.config.username=SA_ADMIN`,
					"--set", `splunk.discovery.receivers.mssql.config.login.name=signalfxagent`,
				)
			},
		},
	)
}
