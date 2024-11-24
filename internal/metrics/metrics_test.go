package metrics

import (
	"github.com/Kopleman/metcol/internal/store"
	"github.com/stretchr/testify/assert"
	"reflect"
	"strconv"
	"testing"
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
				store: store.NewStore(tt.fields.db),
			}
			err := m.SetGauge(tt.args.name, tt.args.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetGauge() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(tt.fields.db[tt.args.name+"-"+string(GougeMetricType)], tt.args.value) {
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
				store: store.NewStore(tt.fields.db),
			}
			beforeUpdate, ok := tt.fields.db[tt.args.name+"-"+string(CounterMetricType)]
			if !ok {
				beforeUpdate = int64(0)
			}
			parsed := beforeUpdate.(int64)

			err := m.SetCounter(tt.args.name, tt.args.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetCounter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(tt.fields.db[tt.args.name+"-"+string(CounterMetricType)], tt.args.value+parsed) {
				t.Errorf("SetCounter() got = %v, want %v", tt.fields.db[tt.args.name], tt.args.value)
			}
		})
	}
}

func TestMetrics_Get(t *testing.T) {
	type fields struct {
		db map[string]any
	}
	type args struct {
		metricType MetricType
		name       string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    any
		wantErr bool
	}{
		{
			name:   "get gauge metric",
			fields: fields{db: map[string]any{"foo-gauge": 1}},
			args: args{
				metricType: "gauge",
				name:       "foo",
			},
			want:    1,
			wantErr: false,
		},
		{
			name:   "get counter metric",
			fields: fields{db: map[string]any{"foo-counter": 1}},
			args: args{
				metricType: "counter",
				name:       "foo",
			},
			want:    1,
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
				store: store.NewStore(tt.fields.db),
			}
			got, err := m.Get(tt.args.metricType, tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Get() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMetrics_SetMetric(t *testing.T) {
	type fields struct {
		db map[string]any
	}
	type args struct {
		metricType MetricType
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
				store: store.NewStore(tt.fields.db),
			}
			err := m.SetMetric(tt.args.metricType, tt.args.name, tt.args.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetMetric() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			var valueToCheck string
			switch tt.args.metricType {
			case GougeMetricType:
				valueToCheck = strconv.FormatFloat(tt.fields.db[tt.args.name+"-"+string(tt.args.metricType)].(float64), 'f', -1, 64)
			case CounterMetricType:
				valueToCheck = strconv.FormatInt(tt.fields.db[tt.args.name+"-"+string(tt.args.metricType)].(int64), 10)
			default:
				valueToCheck = ""
			}

			assert.Equal(t, valueToCheck, tt.args.value)
		})
	}
}
