package store

import (
	"reflect"
	"testing"
)

func TestStore_Create(t *testing.T) {
	type fields struct {
		db map[string]any
	}
	type args struct {
		value any
		key   string
	}
	tests := []struct {
		args    args
		fields  fields
		name    string
		wantErr bool
	}{
		{
			name:   "add record to db",
			fields: fields{db: make(map[string]any)},
			args: args{
				key:   "foo",
				value: "bar",
			},
			wantErr: false,
		},
		{
			name:   "should throw error on duplicates",
			fields: fields{db: map[string]any{"foo": "bar"}},
			args: args{
				key:   "foo",
				value: "bar",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewStore(tt.fields.db)
			if err := s.Create(tt.args.key, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStore_Read(t *testing.T) {
	type fields struct {
		db map[string]any
	}
	type args struct {
		key string
	}
	tests := []struct {
		want    any
		fields  fields
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "should read value from store",
			fields:  fields{db: map[string]any{"foo": "bar"}},
			args:    args{key: "foo"},
			want:    "bar",
			wantErr: false,
		},
		{
			name:    "should throw error if value not presented",
			fields:  fields{db: map[string]any{"foo": "bar"}},
			args:    args{key: "another-foo"},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Store{
				db: tt.fields.db,
			}
			got, err := s.Read(tt.args.key)
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
		db map[string]any
	}
	type args struct {
		value any
		key   string
	}
	tests := []struct {
		args    args
		fields  fields
		name    string
		wantErr bool
	}{
		{
			name:    "should update record",
			fields:  fields{db: map[string]any{"foo": "bar"}},
			args:    args{key: "foo", value: "baz"},
			wantErr: false,
		},
		{
			name:    "should throw error if value not presented",
			fields:  fields{db: map[string]any{"foo": "bar"}},
			args:    args{key: "another-foo"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Store{
				db: tt.fields.db,
			}
			err := s.Update(tt.args.key, tt.args.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Update() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(tt.fields.db[tt.args.key], tt.args.value) {
				t.Errorf("Update() got = %v, want %v", tt.fields.db[tt.args.key], tt.args.value)
			}
		})
	}
}

func TestStore_Delete(t *testing.T) {
	type fields struct {
		db map[string]any
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
			name:    "should delete record",
			fields:  fields{db: map[string]any{"foo": "bar"}},
			args:    args{key: "foo"},
			wantErr: false,
		},
		{
			name:    "should throw error if value not presented",
			fields:  fields{db: map[string]any{"foo": "bar"}},
			args:    args{key: "another-foo"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Store{
				db: tt.fields.db,
			}
			err := s.Delete(tt.args.key)
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
