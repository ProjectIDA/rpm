package tycon

// const (
// )

import (
	"context"
	"errors"
	"fmt"
	rlog "rpm/log"
	"strconv"
	"sync"
	"time"

	g "github.com/gosnmp/gosnmp"
)

const (
	// MinSampleInterval is hte smallest sample interval in seconds
	MinSampleInterval time.Duration = 1 * time.Second
	// MaxSampleInterval is the largest sample interval in seconds
	MaxSampleInterval time.Duration = 60 * time.Second

	relayActionOpen        int    = 0
	relayActionOpenLabel   string = "open"
	relayActionClosed      int    = 1
	relayActionClosedLabel string = "closed"
	relayActionCycle       int    = 2
	relayActionCycleLabel  string = "cycle"

	maxCycleTime int = 99999
	snmpVersion  int = 2
)

// // TPDin2Relay is the relay index type
// type TPDin2Relay int

// // Bounds for Relay Index
// const (
// 	MinRelayIndex int = 1
// 	MaxRelayIndex int = 4
// )

// TPDin2Device struct object
type TPDin2Device struct {
	host             string
	port             uint64
	ready            bool
	internalInterval time.Duration
	SNMPParams       *g.GoSNMP
	ctx              *context.Context
	mutex            sync.Mutex
	// SampleInterval   time.Duration
	CurrentScan *TPDin2Scan
}

// TPDin2Scan holds query results with timestamp
type TPDin2Scan struct {
	TS   time.Time
	Data map[string]string
}

// copy returns a pointer to a copy of the TPDin2Scan struct
func (scan *TPDin2Scan) copy() *TPDin2Scan {
	newdata := make(map[string]string)
	for key, val := range scan.Data {
		newdata[key] = val
	}

	newscan := TPDin2Scan{
		scan.TS,
		newdata,
	}

	return &newscan
}

// NewTPDin2 constructor
func NewTPDin2() *TPDin2Device {

	tp := TPDin2Device{}
	tp.ready = false
	return &tp

}

// Initialize TPDin2 object
func (tp *TPDin2Device) Initialize(host, port string) error {

	// var host, portstr string
	var portInt uint64
	var e error

	if portInt, e = strconv.ParseUint(port, 10, 16); e != nil {
		return e
	}

	tp.host = host
	tp.port = portInt
	// tp.SampleInterval = sampleInterval
	tp.ready = false
	tp.SNMPParams = nil

	rlog.DebugMsg("debug: tp.host:             %s", tp.host)
	rlog.DebugMsg("debug: tp.port:             %d", tp.port)
	rlog.DebugMsg("debug: tp.internalInterval: %s", tp.internalInterval)

	return nil
}

// Connect via SNMP to device
func (tp *TPDin2Device) Connect() error {

	if !tp.ready {

		snmpParams := &g.GoSNMP{
			Target:    tp.host,
			Port:      uint16(tp.port),
			Transport: "udp4",
			Community: "write",
			Version:   g.Version2c,
			Retries:   0,
			Timeout:   time.Duration(10) * time.Second,
		}

		if err := snmpParams.Connect(); err != nil {
			// log.Fatalf("Connect() err: %v", err)
			tp.SNMPParams = nil
			return err
		}

		tp.SNMPParams = snmpParams
		tp.ready = true
	} else {
		return errors.New("error: already connected")
	}

	return nil

}

// SetRelay to set the specified relay to the specified state
func (tp *TPDin2Device) SetRelay(relay, relayOid string, state string) error {
	return nil
}

// CycleRelay to cycle the specified relay, blocking until complete if wait==true"
func (tp *TPDin2Device) CycleRelay(relay, relayOid string, wait bool) error {

	setPDUs := []g.SnmpPDU{
		{
			Name:  relayOid,
			Type:  g.Integer,
			Value: relayActionCycle,
		},
	}
	res, err := tp.SNMPParams.Set(setPDUs)
	if err != nil {
		return err
	}

	fmt.Printf("result: %v\n", res)

	return nil
}

// QueryOids to get values for all device oids
func (tp *TPDin2Device) QueryOids(oids *[]string) (time.Time, map[string]string, error) {

	snmpVals, err := tp.SNMPParams.Get(*oids)
	if err != nil {
		return time.Now(), nil, err
	}

	results := make(map[string]string)
	for i, variable := range snmpVals.Variables {

		// the Value of each variable returned by Get() implements
		// interface{}. You could do a type switch...
		switch variable.Type {
		case g.OctetString:
			// fmt.Printf("string: %s\n", string(variable.Value.([]byte)))
			results[(*oids)[i]] = string(variable.Value.([]byte))
		default:
			// ... or often you're just interested in numeric values.
			// ToBigInt() will return the Value as a BigInt, for plugging
			// into your calculations.
			// fmt.Printf("number: %d\n", g.ToBigInt(variable.Value))
			results[(*oids)[i]] = g.ToBigInt(variable.Value).String()
		}
	}

	ts := time.Now()

	return ts, results, nil
}

// queryDeviceVars queries device for TPDin2 OID values
func (tp *TPDin2Device) queryDeviceVars(oids *[]string) error {

	ts, results, err := tp.QueryOids(oids)
	if err != nil {
		return err
	}
	tp.saveScan(ts, &results)

	return nil
}

// func (tp *TPDin2Device) saveScan(scan *TPDin2Scan) {
func (tp *TPDin2Device) saveScan(ts time.Time, results *map[string]string) {

	tp.mutex.Lock()
	tp.CurrentScan = &TPDin2Scan{ts, *results}
	tp.mutex.Unlock()
}

// GetScan retuns a copy of the most recent TPDin2Scan struct
func (tp *TPDin2Device) GetScan() (*TPDin2Scan, error) {

	if tp.CurrentScan == nil {
		return nil, errors.New("scan unavailable")
	}

	tp.mutex.Lock()
	scan := tp.CurrentScan.copy()
	tp.CurrentScan = nil
	tp.mutex.Unlock()

	return scan, nil
}

// PollStart start polling the connected device
func (tp *TPDin2Device) PollStart(
	ctx context.Context,
	wg *sync.WaitGroup,
	pollOids *[]string,
	sampleInterval time.Duration) error {

	tp.internalInterval = sampleInterval / 3.0

	wg.Add(1)
	defer wg.Done()

	if !tp.ready {
		rlog.WarningMsg("TP2DinDevice is not connected to host: %s", tp.host)
		return fmt.Errorf("TP2DinDevice is not connected to host: %s", tp.host)
	}

	// kick off internval polling loop
	go func(ctx context.Context, wg *sync.WaitGroup) {
		wg.Add(1)
		defer wg.Done()

		trigtime := time.Now()

		for {
			trigtime = trigtime.Add(tp.internalInterval)

			select {
			case <-time.After(time.Until(trigtime)):
				err := tp.queryDeviceVars(pollOids)
				if err != nil {
					rlog.ErrMsg(err.Error())
					continue
				}
			case <-ctx.Done():
				rlog.DebugMsg("debug: context.Done message received, shutting down internal polling loop")
				return
			}
		}

	}(ctx, wg)

	return nil
}

//// cycle relay
// func (tp *TPDin2Device) cycleRelay(relay TPDin2Relay)

func (tp2din *TPDin2Device) InitAndConnect(host, port string) error {
	err := tp2din.Initialize(host, port)
	if err != nil {
		rlog.ErrMsg("unknown error initializing structures for %s:%s, quitting", host, port)
		// log.Fatalln(err)
	}
	err = tp2din.Connect()
	if err != nil {
		rlog.CritMsg("could not connect to %s:%s, quitting", host, port)
		// log.Fatalln(fmt.Errorf("could not connect to %s:%s, quitting", host, port))
	}

	return err
}
