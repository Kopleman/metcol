//nolint:dupl // test-cases dupes
package metrics

import (
	"context"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/Kopleman/metcol/internal/common"
	"github.com/Kopleman/metcol/internal/common/dto"
	"github.com/Kopleman/metcol/internal/common/log"
	"github.com/Kopleman/metcol/internal/server/memstore"
	"github.com/Kopleman/metcol/internal/server/pgxstore"
	"github.com/Kopleman/metcol/internal/testutils"
	"github.com/jackc/pgx/v4"
	"github.com/pashagolub/pgxmock"
	"github.com/stretchr/testify/assert"
)

func TestMetrics_SetGauge(t *testing.T) {
	type fields struct {
		db map[string]*dto.MetricDTO
	}
	type args struct {
		name  string
		value float64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "add gauge metric 1",
			fields: fields{
				db: map[string]*dto.MetricDTO{
					"foo-gouge": {
						ID:    "foo",
						MType: "gauge",
						Delta: nil,
						Value: testutils.Pointer(0.0),
					},
				},
			},
			args: args{
				name:  "foo",
				value: 1,
			},
			wantErr: false,
		},
		{
			name: "add gauge metric 2",
			fields: fields{
				db: map[string]*dto.MetricDTO{
					"foo-gouge": {
						ID:    "foo",
						MType: "gauge",
						Delta: nil,
						Value: testutils.Pointer(1.0),
					},
				},
			},
			args: args{
				name:  "foo",
				value: 0,
			},
			wantErr: false,
		},
		{
			name: "add gauge metric 3",
			fields: fields{
				db: make(map[string]*dto.MetricDTO),
			},
			args: args{
				name:  "foo",
				value: 0,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			m := &Metrics{
				store: memstore.NewStore(tt.fields.db),
			}
			_, err := m.SetGauge(ctx, tt.args.name, tt.args.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetGauge() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			metric, ok := tt.fields.db[tt.args.name+"-"+string(common.GaugeMetricType)]
			if !ok {
				t.Errorf("metric not found in store")
				return
			}
			assert.Equalf(t, tt.args.value, *metric.Value, "SetGauge()")
		})
	}
}

func TestMetrics_SetCounter(t *testing.T) {
	type fields struct {
		db map[string]*dto.MetricDTO
	}
	type args struct {
		name  string
		value int64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "add counter metric 1",
			fields: fields{
				db: map[string]*dto.MetricDTO{
					"foo-gouge": {
						ID:    "foo",
						MType: "counter",
						Delta: testutils.Pointer(int64(0)),
						Value: nil,
					},
				},
			},
			args: args{
				name:  "foo",
				value: 1,
			},
			wantErr: false,
		},
		{
			name: "add counter metric 2",
			fields: fields{
				db: map[string]*dto.MetricDTO{
					"foo-gouge": {
						ID:    "foo",
						MType: "counter",
						Delta: testutils.Pointer(int64(1)),
						Value: nil,
					},
				},
			},
			args: args{
				name:  "foo",
				value: 1,
			},
			wantErr: false,
		},
		{
			name: "add counter metric 3",
			fields: fields{
				db: make(map[string]*dto.MetricDTO),
			},
			args: args{
				name:  "foo",
				value: 1,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			m := &Metrics{
				store: memstore.NewStore(tt.fields.db),
			}
			beforeUpdate, ok := tt.fields.db[tt.args.name+"-"+string(common.CounterMetricType)]
			if !ok {
				beforeUpdate = &dto.MetricDTO{
					Delta: testutils.Pointer(int64(0)),
					Value: nil,
					ID:    tt.args.name,
					MType: common.CounterMetricType,
				}
			}

			_, err := m.SetCounter(ctx, tt.args.name, tt.args.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetCounter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			metric, ok := tt.fields.db[tt.args.name+"-"+string(common.CounterMetricType)]
			if !ok {
				t.Errorf("metric not found in store")
				return
			}
			assert.Equalf(t, tt.args.value+*beforeUpdate.Delta, *metric.Delta, "SetCounter()")
		})
	}
}

func TestMetrics_GetValueAsString(t *testing.T) {
	type fields struct {
		db map[string]*dto.MetricDTO
	}
	type args struct {
		metricType common.MetricType
		name       string
	}
	tests := []struct {
		want    string
		fields  fields
		args    args
		name    string
		wantErr bool
	}{
		{
			name: "get gauge metric",
			fields: fields{
				db: map[string]*dto.MetricDTO{
					"foo-gauge": {
						ID:    "foo",
						MType: "gauge",
						Delta: nil,
						Value: testutils.Pointer(1.0),
					},
				},
			},
			args: args{
				metricType: "gauge",
				name:       "foo",
			},
			want:    "1",
			wantErr: false,
		},
		{
			name: "get counter metric",
			fields: fields{
				db: map[string]*dto.MetricDTO{
					"foo-counter": {
						ID:    "foo",
						MType: "counter",
						Delta: testutils.Pointer(int64(1)),
						Value: nil,
					},
				},
			},
			args: args{
				metricType: "counter",
				name:       "foo",
			},
			want:    "1",
			wantErr: false,
		},
		{
			name: "get counter metric",
			fields: fields{
				db: map[string]*dto.MetricDTO{
					"foo-gouge": {
						ID:    "foo",
						MType: "gauge",
						Delta: nil,
						Value: testutils.Pointer(1.0),
					},
				},
			},
			args: args{
				metricType: "counter",
				name:       "foo",
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			m := &Metrics{
				store: memstore.NewStore(tt.fields.db),
			}
			got, err := m.GetValueAsString(ctx, tt.args.metricType, tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetValueAsString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetValueAsString() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMetrics_SetMetric(t *testing.T) {
	type fields struct {
		db map[string]*dto.MetricDTO
	}
	type args struct {
		metricType common.MetricType
		name       string
		value      string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "add gauge metric",
			fields: fields{db: make(map[string]*dto.MetricDTO)},
			args: args{
				metricType: "gauge",
				name:       "foo",
				value:      "1.1",
			},
			wantErr: false,
		},
		{
			name:   "add counter metric",
			fields: fields{db: make(map[string]*dto.MetricDTO)},
			args: args{
				metricType: "counter",
				name:       "foo",
				value:      "1",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			m := &Metrics{
				store: memstore.NewStore(tt.fields.db),
			}
			err := m.SetMetric(ctx, tt.args.metricType, tt.args.name, tt.args.value)

			if tt.wantErr {
				assert.Error(t, err, "SetMetric() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			var valueToCheck string
			switch tt.args.metricType {
			case common.GaugeMetricType:
				metric, pOk := tt.fields.db[tt.args.name+"-"+string(tt.args.metricType)]
				if !pOk {
					t.Error("GaugeMetricType parse error")
					return
				}
				valueToCheck = strconv.FormatFloat(*metric.Value, 'f', -1, 64)
			case common.CounterMetricType:
				metric, pOk := tt.fields.db[tt.args.name+"-"+string(tt.args.metricType)]
				if !pOk {
					t.Error("CounterMetricType parse error")
					return
				}
				valueToCheck = strconv.FormatInt(*metric.Delta, 10)
			default:
				valueToCheck = ""
			}

			assert.Equal(t, valueToCheck, tt.args.value)
		})
	}
}

func TestMetrics_GetAllValuesAsString(t *testing.T) {
	type fields struct {
		db map[string]*dto.MetricDTO
	}
	tests := []struct {
		fields  fields
		want    map[string]string
		name    string
		wantErr bool
	}{
		{
			name: "get stored metrics",
			fields: fields{
				db: map[string]*dto.MetricDTO{
					"foo-gauge": {
						ID:    "foo",
						MType: "gauge",
						Delta: nil,
						Value: testutils.Pointer(0.1),
					},
					"bar-counter": {
						ID:    "bar",
						MType: "counter",
						Delta: testutils.Pointer(int64(2)),
						Value: nil,
					},
				},
			},
			want:    map[string]string{"foo": "0.1", "bar": "2"},
			wantErr: false,
		},
		{
			name:    "empty store",
			fields:  fields{db: map[string]*dto.MetricDTO{}},
			want:    map[string]string{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			m := &Metrics{
				store: memstore.NewStore(tt.fields.db),
			}
			got, err := m.GetAllValuesAsString(ctx)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetAllValuesAsString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMetrics_SetMetricByDto(t *testing.T) {
	type fields struct {
		db map[string]*dto.MetricDTO
	}
	type args struct {
		metricDto *dto.MetricDTO
	}
	tests := []struct {
		fields  fields
		args    args
		expect  *dto.MetricDTO
		name    string
		wantErr bool
	}{
		{
			name:   "add gouge metric",
			fields: fields{db: make(map[string]*dto.MetricDTO)},
			args: args{
				metricDto: &dto.MetricDTO{
					ID:    "foo",
					MType: "gauge",
					Value: testutils.Pointer(1.1),
				},
			},
			expect: &dto.MetricDTO{
				ID:    "foo",
				MType: "gauge",
				Value: testutils.Pointer(1.1),
			},
			wantErr: false,
		},
		{
			name:   "add counter metric",
			fields: fields{db: make(map[string]*dto.MetricDTO)},
			args: args{
				metricDto: &dto.MetricDTO{
					ID:    "foo",
					MType: "counter",
					Delta: testutils.Pointer(int64(100)),
				},
			},
			expect: &dto.MetricDTO{
				ID:    "foo",
				MType: "counter",
				Delta: testutils.Pointer(int64(100)),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			m := &Metrics{
				store: memstore.NewStore(tt.fields.db),
			}
			err := m.SetMetricByDto(ctx, tt.args.metricDto)

			if tt.wantErr {
				assert.Error(t, err, "SetMetric() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.expect, tt.args.metricDto)
		})
	}
}

func TestMetrics_GetMetricAsDTO(t *testing.T) {
	type fields struct {
		db map[string]*dto.MetricDTO
	}
	type args struct {
		metricType common.MetricType
		name       string
	}
	tests := []struct {
		fields  fields
		want    *dto.MetricDTO
		args    args
		name    string
		wantErr bool
	}{
		{
			name: "get gauge metric",
			fields: fields{db: map[string]*dto.MetricDTO{
				"foo-gauge": {
					ID:    "foo",
					MType: "gauge",
					Delta: nil,
					Value: testutils.Pointer(1.1),
				},
			}},
			args: args{
				metricType: "gauge",
				name:       "foo",
			},
			want: &dto.MetricDTO{
				ID:    "foo",
				MType: "gauge",
				Delta: nil,
				Value: testutils.Pointer(1.1),
			},
			wantErr: false,
		},
		{
			name: "get counter metric",
			fields: fields{db: map[string]*dto.MetricDTO{
				"foo-counter": {
					ID:    "foo",
					MType: "counter",
					Delta: testutils.Pointer(int64(100)),
				}},
			},
			args: args{
				metricType: "counter",
				name:       "foo",
			},
			want: &dto.MetricDTO{
				ID:    "foo",
				MType: "counter",
				Delta: testutils.Pointer(int64(100)),
			},
			wantErr: false,
		},
		{
			name: "get counter metric",
			fields: fields{db: map[string]*dto.MetricDTO{
				"foo-gauge": {
					ID:    "foo",
					MType: "counter",
					Delta: testutils.Pointer(int64(100)),
				}}},
			args: args{
				metricType: "counter",
				name:       "foo",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			m := &Metrics{
				store: memstore.NewStore(tt.fields.db),
			}
			got, err := m.GetMetricAsDTO(ctx, tt.args.metricType, tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetMetricAsDTO() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equalf(t, tt.want, got, "GetMetricAsDTO(%v, %v)", tt.args.metricType, tt.args.name)
		})
	}
}

func TestMetrics_ExportMetrics(t *testing.T) {
	type fields struct {
		db map[string]*dto.MetricDTO
	}
	tests := []struct {
		name    string
		fields  fields
		want    []*dto.MetricDTO
		wantErr bool
	}{
		{
			name: "export 1 metric",
			fields: fields{db: map[string]*dto.MetricDTO{
				"foo": {
					ID:    "foo",
					MType: "gauge",
					Delta: nil,
					Value: testutils.Pointer(1.1),
				},
			}},
			want: []*dto.MetricDTO{
				{
					ID:    "foo",
					MType: "gauge",
					Delta: nil,
					Value: testutils.Pointer(1.1),
				},
			},
			wantErr: false,
		},
		{
			name: "export 2 metrics",
			fields: fields{db: map[string]*dto.MetricDTO{
				"foo-gauge": {
					ID:    "foo",
					MType: "gauge",
					Delta: nil,
					Value: testutils.Pointer(0.1),
				},
				"bar-counter": {
					ID:    "bar",
					MType: "counter",
					Delta: testutils.Pointer(int64(2)),
					Value: nil,
				},
			}},
			want: []*dto.MetricDTO{
				{
					ID:    "foo",
					MType: "gauge",
					Delta: nil,
					Value: testutils.Pointer(0.1),
				},
				{
					ID:    "bar",
					MType: "counter",
					Delta: testutils.Pointer(int64(2)),
					Value: nil,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			m := &Metrics{
				store: memstore.NewStore(tt.fields.db),
			}
			got, err := m.ExportMetrics(ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExportMetrics() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equalf(t, tt.want, got, "ExportMetrics()")
		})
	}
}

func TestMetrics_ImportMetrics(t *testing.T) {
	type fields struct {
		db map[string]*dto.MetricDTO
	}
	type args struct {
		metricsToImport []*dto.MetricDTO
	}
	tests := []struct {
		fields  fields
		want    map[string]*dto.MetricDTO
		name    string
		args    args
		wantErr bool
	}{
		{
			name:   "import 1 metric",
			fields: fields{db: make(map[string]*dto.MetricDTO)},
			args: args{
				metricsToImport: []*dto.MetricDTO{
					{
						ID:    "foo",
						MType: "gauge",
						Delta: nil,
						Value: testutils.Pointer(1.1),
					},
				},
			},
			want: map[string]*dto.MetricDTO{
				"foo-gauge": {
					ID:    "foo",
					MType: "gauge",
					Delta: nil,
					Value: testutils.Pointer(1.1),
				},
			},
			wantErr: false,
		},
		{
			name:   "import 2 metric",
			fields: fields{db: make(map[string]*dto.MetricDTO)},
			args: args{
				metricsToImport: []*dto.MetricDTO{
					{
						ID:    "foo",
						MType: "gauge",
						Delta: nil,
						Value: testutils.Pointer(0.1),
					},
					{
						ID:    "bar",
						MType: "counter",
						Delta: testutils.Pointer(int64(2)),
						Value: nil,
					},
				},
			},
			want: map[string]*dto.MetricDTO{
				"foo-gauge": {
					ID:    "foo",
					MType: "gauge",
					Delta: nil,
					Value: testutils.Pointer(0.1),
				},
				"bar-counter": {
					ID:    "bar",
					MType: "counter",
					Delta: testutils.Pointer(int64(2)),
					Value: nil,
				},
			},
			wantErr: false,
		},
		{
			name:   "import bad metric",
			fields: fields{db: make(map[string]*dto.MetricDTO)},
			args: args{
				metricsToImport: []*dto.MetricDTO{
					{
						ID:    "foo",
						MType: "gauge",
						Delta: testutils.Pointer(int64(1)),
						Value: nil,
					},
				},
			},
			want:    make(map[string]*dto.MetricDTO),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			m := &Metrics{
				store: memstore.NewStore(tt.fields.db),
			}
			err := m.ImportMetrics(ctx, tt.args.metricsToImport)
			if (err != nil) != tt.wantErr {
				t.Errorf("ImportMetrics() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, tt.fields.db, "ImportMetrics()")
		})
	}
}

func TestMetrics_SetMetricsWithMemo(t *testing.T) {
	type fields struct {
		db map[string]*dto.MetricDTO
	}
	type args struct {
		metrics []*dto.MetricDTO
	}
	tests := []struct {
		fields  fields
		want    map[string]*dto.MetricDTO
		name    string
		args    args
		wantErr bool
	}{
		{
			name:   "set 1 metric",
			fields: fields{db: make(map[string]*dto.MetricDTO)},
			args: args{
				metrics: []*dto.MetricDTO{
					{
						ID:    "foo",
						MType: "gauge",
						Delta: nil,
						Value: testutils.Pointer(2.0),
					},
				},
			},
			want: map[string]*dto.MetricDTO{
				"foo-gauge": {
					ID:    "foo",
					MType: "gauge",
					Delta: nil,
					Value: testutils.Pointer(2.0),
				},
			},
			wantErr: false,
		},
		{
			name:   "set 2 metrics",
			fields: fields{db: make(map[string]*dto.MetricDTO)},
			args: args{
				metrics: []*dto.MetricDTO{
					{
						ID:    "foo",
						MType: "gauge",
						Delta: nil,
						Value: testutils.Pointer(1.1),
					},
					{
						ID:    "bar",
						MType: "counter",
						Delta: testutils.Pointer(int64(4)),
						Value: nil,
					},
				},
			},
			want: map[string]*dto.MetricDTO{
				"foo-gauge": {
					ID:    "foo",
					MType: "gauge",
					Delta: nil,
					Value: testutils.Pointer(1.1),
				},
				"bar-counter": {
					ID:    "bar",
					MType: "counter",
					Delta: testutils.Pointer(int64(4)),
					Value: nil,
				},
			},
			wantErr: false,
		},
		{
			name:   "set bad metric",
			fields: fields{db: make(map[string]*dto.MetricDTO)},
			args: args{
				metrics: []*dto.MetricDTO{
					{
						ID:    "foo",
						MType: "gauge",
						Delta: testutils.Pointer(int64(1)),
						Value: nil,
					},
				},
			},
			want:    make(map[string]*dto.MetricDTO),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			m := &Metrics{
				store: memstore.NewStore(tt.fields.db),
			}
			err := m.SetMetrics(ctx, tt.args.metrics)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetMetrics() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, tt.fields.db, "ImportMetrics()")
		})
	}
}

func TestMetrics_SetMetricsWithPGS(t *testing.T) {
	logger := log.MockLogger{}

	type args struct {
		metrics []*dto.MetricDTO
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "set 1 metric",
			args: args{
				metrics: []*dto.MetricDTO{
					{
						ID:    "foo",
						MType: "gauge",
						Delta: nil,
						Value: testutils.Pointer(1.1),
					},
				},
			},
			wantErr: false,
		},
		{
			name: "set 2 metrics",
			args: args{
				metrics: []*dto.MetricDTO{
					{
						ID:    "foo",
						MType: "gauge",
						Delta: nil,
						Value: testutils.Pointer(0.1),
					},
					{
						ID:    "bar",
						MType: "counter",
						Delta: testutils.Pointer(int64(2)),
						Value: nil,
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
			if err != nil {
				t.Fatal(err)
			}
			defer mock.Close()

			ctx := context.Background()
			m := &Metrics{
				store:  pgxstore.NewPGXStore(&logger, mock),
				logger: logger,
			}

			mock.ExpectBegin()

			for index, metric := range tt.args.metrics {
				mock.ExpectQuery(pgxstore.GetMetric).
					WillReturnError(pgx.ErrNoRows)

				indexToStr := strconv.Itoa(index)
				rows := pgxmock.
					NewRows([]string{"id", "name", "type", "value", "delta", "created_at", "updated_at", "deleted_at"}).
					AddRow(
						"00000000-0000-0000-0000-00000000000"+indexToStr,
						metric.ID,
						pgxstore.MetricType(metric.MType),
						metric.Value,
						metric.Delta,
						time.Now(),
						nil,
						nil,
					)
				mock.ExpectQuery(pgxstore.CreateMetric).
					WithArgs(metric.ID, pgxstore.MetricType(metric.MType), metric.Value, metric.Delta).
					WillReturnRows(rows)
			}

			mock.ExpectCommit()

			err = m.SetMetrics(ctx, tt.args.metrics)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetMetrics() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err = mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}
