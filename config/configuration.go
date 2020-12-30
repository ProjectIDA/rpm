package config

import "fmt"

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
	CfgFile string
	general generalConfig
	winMain winMainConfig
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

// OIDsConfig OIDs
// type OIDsConfig struct {
// 	ProductName      string
// 	ProductVersion   string
// 	ProductBuilddate string
// 	Relays           [4]string
// 	Voltages         [4]string
// 	Currents         [4]string
// 	Temperatures     [4]string
// }

// NewRpmCfg return zero'd struct for holding RPM config file information
func NewRpmCfg() *RPMConfig {
	return &RPMConfig{"", generalConfig{"DFA", "NN", "LL"}, winMainConfig{}}
}

// Validate the rpm TOML config file
func (cfg RPMConfig) Validate() (e error) {
	return nil
}

// DumpCfg writes config to string for printing/saving
func DumpCfg(cfg *RPMConfig) string {
	cfgStr := fmt.Sprintln("General Config:")
	cfgStr += fmt.Sprintln("  Net: ", cfg.general.Net)
	cfgStr += fmt.Sprintln("  Sta: ", cfg.general.Sta)
	cfgStr += fmt.Sprintln("  Loc: ", cfg.general.Loc)

	return cfgStr
}
