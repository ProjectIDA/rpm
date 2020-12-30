package tycon

import (
	"reflect"
	"testing"
)

func Test_getRelayOids(t *testing.T) {
	tests := []struct {
		name string
		want []string
	}{
		{
			"Check Relay OID Values", []string{
				"enterprises.45621.2.2.1.0",
				"enterprises.45621.2.2.2.0",
				"enterprises.45621.2.2.3.0",
				"enterprises.45621.2.2.4.0",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tp := &TPDin2Device{}
			if got := *tp.getRelayOids(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getRelayOids() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getVoltageOids(t *testing.T) {
	tests := []struct {
		name string
		want []string
	}{
		{
			"Check Voltage OID Values", []string{
				"enterprises.45621.2.2.5.0",
				"enterprises.45621.2.2.6.0",
				"enterprises.45621.2.2.7.0",
				"enterprises.45621.2.2.8.0",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tp := &TPDin2Device{}
			if got := *tp.getVoltageOids(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getVoltageOids() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getCurrentOids(t *testing.T) {
	tests := []struct {
		name string
		want []string
	}{
		{
			"Check Current OID Values", []string{
				"enterprises.45621.2.2.9.0",
				"enterprises.45621.2.2.10.0",
				"enterprises.45621.2.2.11.0",
				"enterprises.45621.2.2.12.0",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tp := &TPDin2Device{}
			if got := *tp.getCurrentOids(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getCurrentOids() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getTemperatureOids(t *testing.T) {
	tests := []struct {
		name string
		want []string
	}{
		{
			"Check Temperature OID Values", []string{
				"enterprises.45621.2.2.13.0",
				"enterprises.45621.2.2.14.0",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tp := &TPDin2Device{}
			if got := *tp.getTemperatureOids(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getTemperatureOids() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_dynamicOidsCnt(t *testing.T) {
	tests := []struct {
		name string
		want int
	}{
		{
			"length",
			14,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// there should be 14 dynamic oids
			tp := &TPDin2Device{}
			if got := *tp.dynamicOids(); len(got) != tt.want {
				t.Errorf("dynamicOids() length = %v, want %v", len(got), tt.want)
			}
		})
	}
}

func Test_staticOidsCnt(t *testing.T) {
	tests := []struct {
		name string
		want int
	}{
		{
			"length",
			3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tp := &TPDin2Device{}
			if got := *tp.staticOids(); len(got) != tt.want {
				t.Errorf("staticOids() length = %v, want %v", len(got), tt.want)
			}
		})
	}
}
