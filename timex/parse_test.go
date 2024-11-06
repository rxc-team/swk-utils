package timex

import (
	"reflect"
	"testing"
	"time"
)

func TestToTime(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want time.Time
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToTime(tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToTimeE(t *testing.T) {

	compareT, _ := time.Parse("2006-01-02", "2022-01-02")

	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		wantD   time.Time
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "2006/01/02",
			args: args{
				s: "2022/1/2",
			},
			wantD:   compareT,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotD, err := ToTimeE(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToTimeE() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotD, tt.wantD) {
				t.Errorf("ToTimeE() = %v, want %v", gotD, tt.wantD)
			}
		})
	}
}
