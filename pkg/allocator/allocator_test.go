/*
Copyright 2020 Fairwinds

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License
*/

package allocator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_isIPV4(t *testing.T) {

	tests := []struct {
		name string
		ip   string
		want bool
	}{
		{
			name: "any string",
			ip:   "not an ip address",
			want: false,
		},
		{
			name: "ipv4",
			ip:   "192.168.0.1",
			want: true,
		},
		{
			name: "ipv6",
			ip:   "2001:db8:0:1:1:1:1:1",
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isIPV4(tt.ip))
		})
	}
}

func TestClient_setEndpointByPing(t *testing.T) {
	tests := []struct {
		name      string
		endpoints map[string]string
		want      string
		wantErr   bool
	}{
		{
			name:      "single endpoint",
			endpoints: map[string]string{"foo": ""},
			want:      "foo",
			wantErr:   true,
		},
		{
			name:      "check ping",
			endpoints: map[string]string{"example": "foo", "google": "google.com"},
			want:      "google:443",
			wantErr:   false,
		},
		{
			name:      "no valid hosts",
			endpoints: map[string]string{"example": "foo"},
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				Endpoints: tt.endpoints,
			}
			err := c.setEndpointByPing()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, c.Endpoint)
			}
		})
	}
}
