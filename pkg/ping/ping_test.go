package ping

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFastestTrace(t *testing.T) {
	tests := []struct {
		name       string
		traces     []Trace
		want       Trace
		wantErr    bool
		errMessage string
	}{
		{
			name: "one",
			traces: []Trace{
				{
					Host: "http://example.com",
				},
			},
			want:    Trace{Host: "http://example.com"},
			wantErr: false,
		},
		{
			name:       "empty",
			traces:     []Trace{},
			wantErr:    true,
			errMessage: "cannot handle empty slice of traces",
		},
		{
			name: "two",
			traces: []Trace{
				{
					Host:         "slower",
					ResponseTime: time.Duration(300),
				},
				{
					Host:         "faster",
					ResponseTime: time.Duration(100),
				},
			},
			want: Trace{
				Host:         "faster",
				ResponseTime: time.Duration(100),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FastestTrace(tt.traces)
			if tt.wantErr {
				assert.EqualError(t, err, tt.errMessage)
			} else {
				assert.NoError(t, err)
				assert.EqualValues(t, tt.want, got)
			}
		})
	}
}

func TestTrace_Run(t *testing.T) {
	tests := []struct {
		name    string
		trace   Trace
		wantErr bool
	}{
		{
			name: "example.com",
			trace: Trace{
				Host: "example.com",
			},
			wantErr: false,
		},
		{
			name: "invalid",
			trace: Trace{
				Host: "foo",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.trace.Run()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

		})
	}
}
