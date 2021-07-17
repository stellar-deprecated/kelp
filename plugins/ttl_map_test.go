package plugins

import (
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestNewTTLMap(t *testing.T) {
	type args struct {
		ttl int64
	}
	tests := []struct {
		name string
		args args
		want *Map
	}{
		{
			name: "Time set correct",
			args: args{
				ttl: 10,
			},
			want: &Map{
				Mutex: &sync.Mutex{},
				ttl:   10,
				data:  map[string]data{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewTTLMap(tt.args.ttl); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMap_Destroy(t *testing.T) {
	type fields struct {
		data      map[string]data
		ttl       int64
		destroyed bool
		Mutex     *sync.Mutex
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "Ttl Map Destroyed",
			fields: fields{
				data:      map[string]data{},
				ttl:       10,
				destroyed: false,
				Mutex:     &sync.Mutex{},
			},
			wantErr: false,
		},
		{
			name: "Ttl Map Already destoryed",
			fields: fields{
				data:      map[string]data{},
				ttl:       10,
				destroyed: true,
				Mutex:     &sync.Mutex{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			heap := &Map{
				data:      tt.fields.data,
				ttl:       tt.fields.ttl,
				destroyed: tt.fields.destroyed,
				Mutex:     tt.fields.Mutex,
			}
			if err := heap.Destroy(); (err != nil) != tt.wantErr {
				t.Errorf("Map.Destroy() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMap_Set(t *testing.T) {
	type fields struct {
		data      map[string]data
		ttl       int64
		destroyed bool
		Mutex     *sync.Mutex
	}
	type args struct {
		key   string
		value string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "Set value in ttl map",
			args: args{
				key:   "test",
				value: "test",
			},
			fields: fields{
				data:      map[string]data{},
				ttl:       60,
				destroyed: false,
				Mutex:     &sync.Mutex{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			heap := &Map{
				data:      tt.fields.data,
				ttl:       tt.fields.ttl,
				destroyed: tt.fields.destroyed,
				Mutex:     tt.fields.Mutex,
			}
			heap.Set(tt.args.key, tt.args.value)

			if _, ok := heap.data[tt.args.key]; !ok {
				t.Errorf("Map.Set() Value for key:`%s` not set!", tt.args.key)
			}

			if heap.data[tt.args.key].value != tt.args.value {
				t.Errorf("Map.Set() Value for key:`%s` not the same as args.key!", tt.args.key)
			}

			if heap.data[tt.args.key].timestamp <= time.Now().Unix() {
				t.Errorf("Map.Set() timestamp of heap.data[%s] is less than current time!", tt.args.key)
			}
		})
	}
}

func TestMap_cleanup(t *testing.T) {
	type fields struct {
		data      map[string]data
		ttl       int64
		destroyed bool
		Mutex     *sync.Mutex
	}

	values := fields{
		data: map[string]data{
			"test": {
				value:     "test",
				timestamp: time.Now().Unix() - 10,
			},
		},
		ttl:       1,
		destroyed: false,
	}

	t.Run("Cleanup ttl map", func(t *testing.T) {
		heap := &Map{
			data:      values.data,
			ttl:       values.ttl,
			destroyed: values.destroyed,
			Mutex:     &sync.Mutex{},
		}

		isStoped := false

		go func() {
			time.Sleep(2 * time.Second)

			if _, ok := heap.data["test"]; ok {
				t.Errorf("Map.cleanup(), items are not removed after expiration")
			}

			heap.Destroy()

			time.Sleep(2 * time.Second)

			if !isStoped {
				t.Errorf("Map.cleanup(), cleanup is not stoped after heap is destroyed")
			}
		}()

		heap.cleanup()

		isStoped = true
	})
}

func TestMap_Get(t *testing.T) {
	type fields struct {
		data      map[string]data
		ttl       int64
		destroyed bool
		Mutex     *sync.Mutex
	}
	type args struct {
		key string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
		want1  bool
		want2  time.Time
	}{
		{
			name: "Get value ttl map",
			fields: fields{
				data: map[string]data{
					"test": {
						value:     "test",
						timestamp: time.Now().Unix() + 60,
					},
				},
				ttl:       60,
				destroyed: false,
				Mutex:     &sync.Mutex{},
			},
			want:  "test",
			want1: true,
			args: args{
				key: "test",
			},
		},
		{
			name: "Get value ttl map, removed when expired",
			fields: fields{
				data: map[string]data{
					"test": {
						value:     "test",
						timestamp: time.Now().Unix() - 1,
					},
				},
				ttl:       60,
				destroyed: false,
				Mutex:     &sync.Mutex{},
			},
			want:  "",
			want1: false,
			args: args{
				key: "test",
			},
		},

		{
			name: "Get value ttl map, empty when is not value",
			fields: fields{
				data:      map[string]data{},
				ttl:       60,
				destroyed: false,
				Mutex:     &sync.Mutex{},
			},
			want:  "",
			want1: false,
			args: args{
				key: "test",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			heap := &Map{
				data:      tt.fields.data,
				ttl:       tt.fields.ttl,
				destroyed: tt.fields.destroyed,
				Mutex:     tt.fields.Mutex,
			}
			got, got1, _ := heap.Get(tt.args.key)
			if got != tt.want {
				t.Errorf("Map.Get() value = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("Map.Get() isValue = %v, want %v", got1, tt.want1)
			}

		})
	}
}

func TestMap_Del(t *testing.T) {
	type fields struct {
		data      map[string]data
		ttl       int64
		destroyed bool
		Mutex     *sync.Mutex
	}
	type args struct {
		key string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "Del ttl map",
			fields: fields{
				data: map[string]data{
					"test": {
						value:     "test",
						timestamp: time.Now().Unix() + 60,
					},
				},
				ttl:       60,
				destroyed: false,
				Mutex:     &sync.Mutex{},
			},
			args: args{
				key: "test",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			heap := &Map{
				data:      tt.fields.data,
				ttl:       tt.fields.ttl,
				destroyed: tt.fields.destroyed,
				Mutex:     tt.fields.Mutex,
			}
			heap.Del(tt.args.key)

			if _, ok := heap.data[tt.args.key]; ok {
				t.Errorf("Map.Del() , value is not deleted")
			}
		})
	}
}
