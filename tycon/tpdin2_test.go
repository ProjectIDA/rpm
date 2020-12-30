package tycon

import (
	"reflect"
	cfg "rpm/config"
	"testing"
	"time"
)

type CfgMock struct{}

// NewCfgMock
func NewCfgMock() *CfgMock {
	return &CfgMock{}
}
func (cfg CfgMock) Validate() (e error) {
	return nil
}

func TestInit(t *testing.T) {
	type args struct {
		host       string
		sampleRate float32
		cfg        cfg.Config
	}

	acfg := NewCfgMock()

	tests := []struct {
		name string
		args args
		want *TPDin2Device
	}{
		{
			"Construct TPDin2 Object",
			args{
				"test.com:161",
				1.0,
				acfg,
			},
			&TPDin2Device{
				// "",
				// 0,
				// nil,
				// false,
				// 0,
				// nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewTPDin2(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewTPDin2() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTPDin2Device_Init(t *testing.T) {
	type fields struct {
		host           string
		port           uint64
		c              cfg.Config
		ready          bool
		sampleInterval time.Duration
	}
	type args struct {
		hostport       string
		sampleInterval float64
		cfg            cfg.Config
	}

	acfg := NewCfgMock()

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			"TPDin2 Init: Good",
			fields{
				"test.com",
				161,
				acfg,
				false,
				1.0,
			},
			args{
				"test.com:161",
				1.0,
				acfg,
			},
			false,
		},
		{
			"TPDin2 Init: bad hostport",
			fields{
				"",
				0,
				nil,
				false,
				0,
			},
			args{
				"bad:h:ost:161",
				1.0,
				acfg,
			},
			true,
		},
		{
			"TPDin2 Init: bad hostport",
			fields{
				"",
				0,
				nil,
				false,
				0,
			},
			args{
				"badhostport",
				1.0,
				acfg,
			},
			true,
		},
		{
			"TPDin2 Init: port# > max 2^16-1",
			fields{
				"",
				0,
				nil,
				false,
				0,
			},
			args{
				"host.com:99999",
				1.0,
				acfg,
			},
			true,
		},
		{
			"TPDin2 Init: negative port",
			fields{
				"",
				0,
				nil,
				false,
				0,
			},
			args{
				"host.com:-161",
				1.0,
				acfg,
			},
			true,
		},
		{
			"TPDin2 Init: bad cfg",
			fields{
				"",
				0,
				nil,
				false,
				0,
			},
			args{
				"test.com:161",
				1.0,
				nil,
			},
			true,
		},
		{
			"TPDin2 Init: bad sampleInterval",
			fields{
				"",
				0,
				nil,
				false,
				0,
			},
			args{
				"test.com:161",
				-1.0,
				nil,
			},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tp := &TPDin2Device{}
			if err := tp.Initialize(tt.args.hostport, time.Duration(tt.args.sampleInterval)*time.Second, tt.args.cfg); (err != nil) != tt.wantErr {
				t.Errorf("TPDin2Device.Init() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tp.host != tt.fields.host {
				t.Errorf("TPDin2Device.Init()  host  = %v, want %v", tp.host, tt.fields.host)
			}
			if tp.port != tt.fields.port {
				t.Errorf("TPDin2Device.Init()  port  = %v, want %v", tp.port, tt.fields.port)
			}
			if tp.cfg != tt.fields.c {
				t.Errorf("TPDin2Device.Init()  cfg   = %v, want %v", tp.cfg, tt.fields.c)
			}
			if tp.ready != tt.fields.ready {
				t.Errorf("TPDin2Device.Init()  ready = %v, want %v", tp.ready, tt.fields.ready)
			}
			if tp.SampleInterval != time.Duration(tt.fields.sampleInterval)*time.Second {
				t.Errorf("TPDin2Device.Init()  sampr = %v, want %v", tp.SampleInterval, tt.fields.sampleInterval)
			}
		})
	}
}
