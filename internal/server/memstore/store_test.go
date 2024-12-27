package memstore

import (
	"context"
	"reflect"
	"testing"

	"github.com/Kopleman/metcol/internal/common/dto"
	"github.com/Kopleman/metcol/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func TestStore_Create(t *testing.T) {
	type fields struct {
		db map[string]*dto.MetricDTO
	}
	type args struct {
		value *dto.MetricDTO
	}
	tests := []struct {
		args    args
		fields  fields
		name    string
		wantErr bool
	}{
		{
			name:   "add record to db",
			fields: fields{db: make(map[string]*dto.MetricDTO)},
			args: args{
				value: &dto.MetricDTO{
					ID:    "foo",
					MType: "gauge",
					Delta: nil,
					Value: testutils.Pointer(0.0),
				},
			},
			wantErr: false,
		},
		{
			name: "should throw error on duplicates",
			fields: fields{
				db: map[string]*dto.MetricDTO{
					"foo-gauge": {
						ID:    "foo",
						MType: "gauge",
						Delta: nil,
						Value: testutils.Pointer(0.0),
					},
				},
			},
			args: args{
				value: &dto.MetricDTO{
					ID:    "foo",
					MType: "gauge",
					Delta: nil,
					Value: testutils.Pointer(0.0),
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			s := NewStore(tt.fields.db)
			if err := s.Create(ctx, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStore_Read(t *testing.T) {
	type fields struct {
		db map[string]*dto.MetricDTO
	}
	type args struct {
		key string
	}
	tests := []struct {
		want    *dto.MetricDTO
		fields  fields
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "should read value from memstore",
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
			args: args{key: "foo-gouge"},
			want: &dto.MetricDTO{
				ID:    "foo",
				MType: "gauge",
				Delta: nil,
				Value: testutils.Pointer(0.0),
			},
			wantErr: false,
		},
		{
			name: "should throw error if value not presented",
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
			args:    args{key: "another-foo"},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			s := &Store{
				db: tt.fields.db,
			}
			got, err := s.Read(ctx, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("Read() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Read() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStore_Update(t *testing.T) {
	type fields struct {
		db map[string]*dto.MetricDTO
	}
	type args struct {
		value *dto.MetricDTO
	}
	tests := []struct {
		args    args
		fields  fields
		name    string
		wantErr bool
	}{
		{
			name: "should update record",
			fields: fields{
				db: map[string]*dto.MetricDTO{
					"foo-gauge": {
						ID:    "foo",
						MType: "gauge",
						Delta: nil,
						Value: testutils.Pointer(0.0),
					},
				},
			},
			args: args{
				value: &dto.MetricDTO{
					ID:    "foo",
					MType: "gauge",
					Delta: nil,
					Value: testutils.Pointer(1.0),
				},
			},
			wantErr: false,
		},
		{
			name: "should throw error if value not presented",
			fields: fields{
				db: map[string]*dto.MetricDTO{
					"foo-gauge": {
						ID:    "foo",
						MType: "gauge",
						Delta: nil,
						Value: testutils.Pointer(0.0),
					},
				},
			},
			args: args{
				value: &dto.MetricDTO{
					ID:    "bar",
					MType: "gauge",
					Delta: nil,
					Value: testutils.Pointer(1.0),
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			s := &Store{
				db: tt.fields.db,
			}
			err := s.Update(ctx, tt.args.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Update() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			metric, ok := tt.fields.db[tt.args.value.ID+"-"+string(tt.args.value.MType)]
			if !ok {
				t.Errorf("metric not found in store")
				return
			}
			assert.Equalf(t, tt.args.value, metric, "SetGauge()")
		})
	}
}

func TestStore_Delete(t *testing.T) {
	type fields struct {
		db map[string]*dto.MetricDTO
	}
	type args struct {
		key string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "should delete record",
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
			args:    args{key: "foo-gouge"},
			wantErr: false,
		},
		{
			name: "should throw error if value not presented",
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
			args:    args{key: "another-foo"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			s := &Store{
				db: tt.fields.db,
			}
			err := s.Delete(ctx, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.fields.db[tt.args.key] != nil {
				t.Errorf("Delete() got = %v, want nil", tt.fields.db[tt.args.key])
			}
		})
	}
}
