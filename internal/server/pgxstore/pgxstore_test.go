package pgxstore

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/Kopleman/metcol/internal/common/dto"
	"github.com/Kopleman/metcol/internal/common/log"
	"github.com/Kopleman/metcol/internal/testutils"
	"github.com/pashagolub/pgxmock/v4"
)

// TODO add more tests for other methods.

func TestMetrics_BulkCreateOrUpdate(t *testing.T) {
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
			store := NewPGXStore(&logger, mock)

			mock.ExpectBegin()

			for index, metric := range tt.args.metrics {
				indexToStr := strconv.Itoa(index)
				rows := pgxmock.
					NewRows([]string{"id", "name", "type", "value", "delta", "created_at", "updated_at", "deleted_at"}).
					AddRow(
						"00000000-0000-0000-0000-00000000000"+indexToStr,
						metric.ID,
						MetricType(metric.MType),
						metric.Value,
						metric.Delta,
						time.Now(),
						nil,
						nil,
					)
				mock.ExpectQuery(CreateOrUpdateMetric).
					WithArgs(metric.ID, MetricType(metric.MType), metric.Value, metric.Delta).
					WillReturnRows(rows)
			}

			mock.ExpectCommit()

			err = store.BulkCreateOrUpdate(ctx, tt.args.metrics)
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
