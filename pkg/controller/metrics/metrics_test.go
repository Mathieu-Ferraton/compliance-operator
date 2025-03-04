/*
Copyright 2021 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package metrics

import (
	"errors"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/require"

	"github.com/openshift/compliance-operator/pkg/apis/compliance/v1alpha1"
	"github.com/openshift/compliance-operator/pkg/controller/metrics/metricsfakes"
)

var errTest = errors.New("")

func TestRegisterMetrics(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		prepare   func(*metricsfakes.FakeImpl)
		shouldErr bool
	}{
		{ // success
			prepare: func(*metricsfakes.FakeImpl) {},
		},
		{ // error Register fails
			prepare: func(mock *metricsfakes.FakeImpl) {
				mock.RegisterReturns(errTest)
			},
			shouldErr: true,
		},
	} {
		mock := &metricsfakes.FakeImpl{}
		tc.prepare(mock)

		sut := New()
		sut.impl = mock

		err := sut.Register()

		if tc.shouldErr {
			require.NotNil(t, err)
		} else {
			require.Nil(t, err)
		}
	}
}

func TestFileIntegrityMetrics(t *testing.T) {
	t.Parallel()

	getMetricValue := func(col prometheus.Collector) int {
		c := make(chan prometheus.Metric, 1)
		col.Collect(c)
		m := dto.Metric{}
		err := (<-c).Write(&m)
		require.Nil(t, err)
		return int(*m.Counter.Value)
	}

	for _, tc := range []struct {
		when func(m *Metrics)
		then func(m *Metrics)
	}{
		{ // single active
			when: func(m *Metrics) {
				m.IncComplianceScanStatus("foo", v1alpha1.ComplianceScanStatus{
					Result: "bar",
					Phase:  "baz",
				})
			},
			then: func(m *Metrics) {
				ctr, err := m.metrics.metricComplianceScanStatus.GetMetricWith(prometheus.Labels{metricLabelScanName: "foo",
					metricLabelScanResult: "bar",
					metricLabelScanPhase:  "baz",
				})
				require.Nil(t, err)
				require.Equal(t, 1, getMetricValue(ctr))
			},
		},
	} {
		mock := &metricsfakes.FakeImpl{}
		sut := New()
		sut.impl = mock

		tc.when(sut)
		tc.then(sut)
	}
}
