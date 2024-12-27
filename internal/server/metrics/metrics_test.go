package metrics

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/Kopleman/metcol/internal/common"
	"github.com/Kopleman/metcol/internal/common/dto"
	"github.com/Kopleman/metcol/internal/server/memstore"
	"github.com/Kopleman/metcol/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func TestMetrics_SetGauge(t *testing.T) {
	type fields struct {
		db map[string]any
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
				db: map[string]any{"foo-gouge": 0},
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
				db: map[string]any{"foo-gouge": 1},
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
				db: make(map[string]any),
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
			m := &Metrics{
				store: memstore.NewStore(tt.fields.db),
			}
			_, err := m.SetGauge(tt.args.name, tt.args.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetGauge() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(tt.fields.db[tt.args.name+"-"+string(common.GougeMetricType)], tt.args.value) {
				t.Errorf("SetGauge() got = %v, want %v", tt.fields.db[tt.args.name], tt.args.value)
			}
		})
	}
}

func TestMetrics_SetCounter(t *testing.T) {
	type fields struct {
		db map[string]any
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
				db: map[string]any{"foo-counter": int64(0)},
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
				db: map[string]any{"foo-counter": int64(1)},
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
				db: make(map[string]any),
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
			m := &Metrics{
				store: memstore.NewStore(tt.fields.db),
			}
			beforeUpdate, ok := tt.fields.db[tt.args.name+"-"+string(common.CounterMetricType)]
			if !ok {
				beforeUpdate = int64(0)
			}
			parsed, pOk := beforeUpdate.(int64)
			if !pOk {
				t.Error("beforeUpdate parse error")
				return
			}

			_, err := m.SetCounter(tt.args.name, tt.args.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetCounter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(tt.fields.db[tt.args.name+"-"+string(common.CounterMetricType)], tt.args.value+parsed) {
				t.Errorf("SetCounter() got = %v, want %v", tt.fields.db[tt.args.name], tt.args.value)
			}
		})
	}
}

func TestMetrics_GetValueAsString(t *testing.T) {
	type fields struct {
		db map[string]any
	}
	type args struct {
		metricType common.MetricType
		name       string
	}
	tests := []struct {
		want    any
		fields  fields
		args    args
		name    string
		wantErr bool
	}{
		{
			name:   "get gauge metric",
			fields: fields{db: map[string]any{"foo-gauge": float64(1)}},
			args: args{
				metricType: "gauge",
				name:       "foo",
			},
			want:    "1",
			wantErr: false,
		},
		{
			name:   "get counter metric",
			fields: fields{db: map[string]any{"foo-counter": int64(1)}},
			args: args{
				metricType: "counter",
				name:       "foo",
			},
			want:    "1",
			wantErr: false,
		},
		{
			name:   "get counter metric",
			fields: fields{db: map[string]any{"foo-gauge": 1}},
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
			m := &Metrics{
				store: memstore.NewStore(tt.fields.db),
			}
			got, err := m.GetValueAsString(tt.args.metricType, tt.args.name)
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
		db map[string]any
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
			name:   "add gouge metric",
			fields: fields{db: make(map[string]any)},
			args: args{
				metricType: "gauge",
				name:       "foo",
				value:      "1.1",
			},
			wantErr: false,
		},
		{
			name:   "add counter metric",
			fields: fields{db: make(map[string]any)},
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
			m := &Metrics{
				store: memstore.NewStore(tt.fields.db),
			}
			err := m.SetMetric(tt.args.metricType, tt.args.name, tt.args.value)

			if tt.wantErr {
				assert.Error(t, err, "SetMetric() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			var valueToCheck string
			switch tt.args.metricType {
			case common.GougeMetricType:
				parsed, pOk := tt.fields.db[tt.args.name+"-"+string(tt.args.metricType)].(float64)
				if !pOk {
					t.Error("GougeMetricType parse error")
					return
				}
				valueToCheck = strconv.FormatFloat(parsed, 'f', -1, 64)
			case common.CounterMetricType:
				parsed, pOk := tt.fields.db[tt.args.name+"-"+string(tt.args.metricType)].(int64)
				if !pOk {
					t.Error("CounterMetricType parse error")
					return
				}
				valueToCheck = strconv.FormatInt(parsed, 10)
			default:
				valueToCheck = ""
			}

			assert.Equal(t, valueToCheck, tt.args.value)
		})
	}
}

func TestMetrics_GetAllValuesAsString(t *testing.T) {
	type fields struct {
		db map[string]any
	}
	tests := []struct {
		fields  fields
		want    map[string]string
		name    string
		wantErr bool
	}{
		{
			name:    "get stored metrics",
			fields:  fields{db: map[string]any{"foo-gauge": float64(0.1), "bar-counter": int64(2)}},
			want:    map[string]string{"foo": "0.1", "bar": "2"},
			wantErr: false,
		},
		{
			name:    "empty memstore",
			fields:  fields{db: map[string]any{}},
			want:    map[string]string{},
			wantErr: false,
		},
		{
			name:    "empty memstore",
			fields:  fields{db: map[string]any{"foo-gauge": 1}},
			want:    map[string]string{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Metrics{
				store: memstore.NewStore(tt.fields.db),
			}
			got, err := m.GetAllValuesAsString()

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
		db map[string]any
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
			fields: fields{db: make(map[string]any)},
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
			fields: fields{db: make(map[string]any)},
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
			m := &Metrics{
				store: memstore.NewStore(tt.fields.db),
			}
			err := m.SetMetricByDto(tt.args.metricDto)

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
		db map[string]any
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
			name:   "get gauge metric",
			fields: fields{db: map[string]any{"foo-gauge": 1.1}},
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
			name:   "get counter metric",
			fields: fields{db: map[string]any{"foo-counter": int64(100)}},
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
			name:   "get counter metric",
			fields: fields{db: map[string]any{"foo-gauge": 1}},
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
			m := &Metrics{
				store: memstore.NewStore(tt.fields.db),
			}
			got, err := m.GetMetricAsDTO(tt.args.metricType, tt.args.name)
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
		db map[string]any
	}
	tests := []struct {
		name    string
		fields  fields
		want    []*dto.MetricDTO
		wantErr bool
	}{
		{
			name:   "export 1 metric",
			fields: fields{db: map[string]any{"foo-gauge": 1.1}},
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
			name:   "export 2 metrics",
			fields: fields{db: map[string]any{"foo-gauge": float64(0.1), "bar-counter": int64(2)}},
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
		{
			name:    "export bad metric",
			fields:  fields{db: map[string]any{"foo-gauge": "bad staff"}},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Metrics{
				store: memstore.NewStore(tt.fields.db),
			}
			got, err := m.ExportMetrics()
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
		db map[string]any
	}
	type args struct {
		metricsToImport []*dto.MetricDTO
	}
	tests := []struct {
		fields  fields
		want    map[string]any
		name    string
		args    args
		wantErr bool
	}{
		{
			name:   "import 1 metric",
			fields: fields{db: make(map[string]any)},
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
			want:    map[string]any{"foo-gauge": 1.1},
			wantErr: false,
		},
		{
			name:   "import 2 metric",
			fields: fields{db: make(map[string]any)},
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
			want:    map[string]any{"foo-gauge": float64(0.1), "bar-counter": int64(2)},
			wantErr: false,
		},
		{
			name:   "import bad metric",
			fields: fields{db: make(map[string]any)},
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
			want:    map[string]any{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Metrics{
				store: memstore.NewStore(tt.fields.db),
			}
			err := m.ImportMetrics(tt.args.metricsToImport)
			if (err != nil) != tt.wantErr {
				t.Errorf("ImportMetrics() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equalf(t, tt.want, tt.fields.db, "ImportMetrics()")
		})
	}
}
