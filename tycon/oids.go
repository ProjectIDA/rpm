package tycon

const (
	// static oids
	oidProductNameOID      string = "enterprises.45621.2.1.1.0"
	oidProductVersionOID   string = "enterprises.45621.2.1.2.0"
	oidProductBuildDateOID string = "enterprises.45621.2.1.3.0"

	// dynamic oids
	oidRelay1       string = "enterprises.45621.2.2.1.0"
	oidRelay2       string = "enterprises.45621.2.2.2.0"
	oidRelay3       string = "enterprises.45621.2.2.3.0"
	oidRelay4       string = "enterprises.45621.2.2.4.0"
	oidVoltage1     string = "enterprises.45621.2.2.5.0"
	oidVoltage2     string = "enterprises.45621.2.2.6.0"
	oidVoltage3     string = "enterprises.45621.2.2.7.0"
	oidVoltage4     string = "enterprises.45621.2.2.8.0"
	oidCurrent1     string = "enterprises.45621.2.2.9.0"
	oidCurrent2     string = "enterprises.45621.2.2.10.0"
	oidCurrent3     string = "enterprises.45621.2.2.11.0"
	oidCurrent4     string = "enterprises.45621.2.2.12.0"
	oidTemperature1 string = "enterprises.45621.2.2.13.0"
	oidTemperature2 string = "enterprises.45621.2.2.14.0"

	testOid1 string = "1.3.6.1.2.1.1.3.0"
	testOid2 string = "1.3.6.1.2.1.1.4.0"
)

var (
	relayOids   = []string{oidRelay1, oidRelay2, oidRelay3, oidRelay4}
	voltageOids = []string{oidVoltage1, oidVoltage2, oidVoltage3, oidVoltage4}
	currentOids = []string{oidCurrent1, oidCurrent2, oidCurrent3, oidCurrent4}
	tempOids    = []string{oidTemperature1, oidTemperature2}
	testOids    = []string{testOid1, testOid2}

	deviceOids []string
	dataOids   []string
)

func init() {
	dataOids = []string{}
	dataOids = append(dataOids, relayOids...)
	dataOids = append(dataOids, voltageOids...)
	dataOids = append(dataOids, currentOids...)
	dataOids = append(dataOids, tempOids...)

	dataOids = testOids

	deviceOids = []string{
		oidProductNameOID,
		oidProductBuildDateOID,
		oidProductVersionOID,
	}

}

func staticOids() *[]string {
	return &[]string{
		oidProductNameOID,
		oidProductBuildDateOID,
		oidProductVersionOID}
}

func dynamicOids() *[]string {
	dynoids := []string{}
	dynoids = append(dynoids, *getRelayOids()...)
	dynoids = append(dynoids, *getVoltageOids()...)
	dynoids = append(dynoids, *getCurrentOids()...)
	dynoids = append(dynoids, *getTemperatureOids()...)
	return &dynoids
}

func getRelayOids() *[]string {
	return &[]string{
		oidRelay1,
		oidRelay2,
		oidRelay3,
		oidRelay4,
	}
}

func getVoltageOids() *[]string {
	return &[]string{
		oidVoltage1,
		oidVoltage2,
		oidVoltage3,
		oidVoltage4,
	}
}

func getCurrentOids() *[]string {
	return &[]string{
		oidCurrent1,
		oidCurrent2,
		oidCurrent3,
		oidCurrent4,
	}
}

func getTemperatureOids() *[]string {
	return &[]string{
		oidTemperature1,
		oidTemperature2,
	}
}

func getTestOids() *[]string {
	return &[]string{
		testOid1,
		testOid2,
	}
}
