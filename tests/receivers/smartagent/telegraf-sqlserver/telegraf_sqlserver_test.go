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
	"testing"

	"github.com/signalfx/splunk-otel-collector/tests/testutils"
)

var server = testutils.NewContainer().WithContext(
	path.Join(".", "testdata", "server"),
).WithExposedPorts("1433:1433").WithName("sql-server").WithNetworks("mssql").WillWaitForPorts("1433").WillWaitForLogs("SQL Server is now ready for client connections.", "Recovery is complete.")

var client = testutils.NewContainer().WithContext(
	path.Join(".", "testdata", "client"),
).WithName("sql-client").WithNetworks("mssql").WillWaitForLogs("name", "signalfxagent")

var mssql_containers = []testutils.Container{server, client}

func TestTelegrafSQLServerReceiverProvidesAllMetrics(t *testing.T) {

	//testutils.AssertAllMetricsReceived(t, "all.yaml", "all_metrics_config.yaml", mssql_containers, nil)
	testutils.AssertAllMetricsReceived(t, "all.yaml", "all_metrics_config.yaml",
		mssql_containers, []testutils.CollectorBuilder{
			func(collector testutils.Collector) testutils.Collector {
				return collector.WithEnv(map[string]string{"MSSQLDB_URL": "mssql://otel:Password!@localhost:1443"})
			},
		},
	)
}
