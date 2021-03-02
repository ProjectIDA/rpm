package config

import (
	"fmt"
	"io"
)

// Config interface for RPM
type Config interface {
	Validate() error
}

// NewConfig constructor
func NewConfig() *RPMConfig {
	return &RPMConfig{}
}

// RPMConfig hold the RPM configuration structure
type RPMConfig struct {
	General generalConfig
	WinMain winMainConfig
	Oids    TyconOids
	CfgFile string
}

// GeneralConfig top lebel config settings
type generalConfig struct {
	Sta string
	Net string
	Loc string
}

// WinMainConfig display labels for realtime monitoring
type winMainConfig struct {
	LBL220vac   string
	LBL110vac   string
	LBLCpu1     string
	LBLCpu2     string
	LBLLoadvolt string
	LBLLoadamp  string
	LBLBatvolt  string
	LBLBatamp   string
	LBLPwrsup   string
	LBLBatcha   string
	LBLTyctmp   string
	LBLBattmp   string
	LBLRackamp  string
	LBLVaultamp string
	LBLAuxamp   string
}

// TyconOids wraps the info for different categrories of Oids
type TyconOids struct {
	Static   []OidInfo
	Tests    []OidInfo
	Relays   []OidInfo
	Voltages []OidInfo
	Currents []OidInfo
	Temps    []OidInfo
}

// OidInfo holds detailed info for each Oid endpoint
type OidInfo struct {
	Oid      string
	Chancode string
	Label    string
	Function string
	// Cycletime int
}

// DataOids provides list of list of Data Oids
func (toids *TyconOids) DataOids() *[][]OidInfo {
	return &[][]OidInfo{
		toids.Static,
		toids.Relays,
		toids.Voltages,
		toids.Currents,
		toids.Temps,
	}
}

// Validate the rpm TOML config file
func (cfg RPMConfig) Validate() (e error) {
	return nil
}

// DumpCfg writes config to string for printing/saving
func (cfg *RPMConfig) DumpCfg(writer io.Writer) {

	fmt.Fprintf(writer, "%v\n", *&cfg.General)
	fmt.Fprintf(writer, "%v\n", *&cfg.WinMain)
	listlist := [][]OidInfo{*&cfg.Oids.Static, *&cfg.Oids.Relays, *&cfg.Oids.Voltages, *&cfg.Oids.Currents, *&cfg.Oids.Temps}
	for _, list := range listlist {
		for _, detail := range list {
			fmt.Fprintf(writer, "%v\n", detail)
		}
	}
	return
}

// DataOidsInfo is a convenience func to generate an ordered list of OIDS that have real data for polling/querying
func (cfg *RPMConfig) DataOidsInfo() ([]string, []OidInfo) {

	cnt := len(cfg.Oids.Voltages) + len(cfg.Oids.Currents) + len(cfg.Oids.Relays) + len(cfg.Oids.Temps)

	oidInfo := make([]OidInfo, 0, cnt)
	oidInfo = append(oidInfo, cfg.Oids.Currents...)
	oidInfo = append(oidInfo, cfg.Oids.Voltages...)
	oidInfo = append(oidInfo, cfg.Oids.Relays...)
	oidInfo = append(oidInfo, cfg.Oids.Temps...)

	oids := make([]string, 0, cnt)
	for _, oidinfo := range oidInfo {
		oids = append(oids, oidinfo.Oid)
	}

	return oids, oidInfo

}

// StaticOidsInfo is a convenience func to generate an ordered list of OIDS that have device static values
func (cfg *RPMConfig) StaticOidsInfo() ([]string, []OidInfo) {

	cnt := len(cfg.Oids.Static)

	oidInfo := make([]OidInfo, 0, cnt)
	oidInfo = append(oidInfo, cfg.Oids.Static...)

	oids := make([]string, 0, cnt)
	for _, oidinfo := range oidInfo {
		oids = append(oids, oidinfo.Oid)
	}

	return oids, oidInfo

}

// RelayOidsInfo is a convenience func to generate an ordered list of Relay OIDS
func (cfg *RPMConfig) RelayOidsInfo() ([]string, []OidInfo) {

	cnt := len(cfg.Oids.Relays)

	oidInfo := make([]OidInfo, 0, cnt)
	oidInfo = append(oidInfo, cfg.Oids.Relays...)

	oids := make([]string, 0, cnt)
	for _, oidinfo := range oidInfo {
		oids = append(oids, oidinfo.Oid)
	}

	return oids, oidInfo

}
