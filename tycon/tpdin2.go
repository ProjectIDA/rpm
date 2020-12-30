package tycon

// const (
// )

import (
	"context"
	"errors"
	"fmt"
	"net"
	"rpm/config"
	"strconv"
	"sync"
	"time"

	g "github.com/gosnmp/gosnmp"
)

const (
	relayActionOpen        int    = 0
	relayActionOpenLabel   string = "open"
	relayActionClosed      int    = 1
	relayActionClosedLabel string = "closed"
	relayActionCycle       int    = 2
	relayActionCycleLabel  string = "cycle"

	maxCycleTime int = 99999
	snmpVersion  int = 2
)

// TPDin2Device struct object
type TPDin2Device struct {
	host             string
	port             uint64
	cfg              config.Config
	ready            bool
	internalInterval time.Duration
	snmpParams       *g.GoSNMP
	ctx              *context.Context
	mutex            sync.Mutex
	SampleInterval   time.Duration
	CurrentScan      *TPDin2Scan
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
func (tp *TPDin2Device) Initialize(hostport string, sampleInterval time.Duration, cfg config.Config) error {

	var host, portstr string
	var portInt uint64
	var e error

	host, portstr, e = net.SplitHostPort(hostport)
	if e != nil {
		return e
	}

	if portInt, e = strconv.ParseUint(portstr, 10, 16); e != nil {
		return e
	}

	// check port is numeric: 1 < port < 2^16-1

	if (sampleInterval < time.Duration(1.0*time.Second)) || (sampleInterval > time.Duration(3600*time.Second)) {
		return errors.New("invalid sampleInterval, must be between 1.0 and 3600.0 seconds")
	}

	if cfg == nil {
		return errors.New("cfg must not be nil")
	}
	// if err := cfg.Validate(); err != nil {
	// 	return err
	// }

	tp.host = host
	tp.port = portInt
	tp.SampleInterval = sampleInterval
	tp.internalInterval = tp.SampleInterval / 2
	tp.cfg = cfg
	tp.ready = false
	tp.snmpParams = nil

	fmt.Println("debug: tp.host = ", tp.host)
	fmt.Println("debug: tp.port = ", tp.port)
	fmt.Println("debug: tp.SampleInterval = ", tp.SampleInterval)
	fmt.Println("debug: tp.internalInterval = ", tp.internalInterval)

	return nil
}

// Connect via SNMP to device
func (tp *TPDin2Device) Connect() error {

	// g.Default.Target = tp.host
	// err := g.Default.Connect()
	// if err != nil {
	// 	log.Fatalf("Connect() error: %v", err)
	// }
	// // defer g.Default.Conn.Close()
	// tp.ready = true

	if !tp.ready {

		timeoutms := int64(tp.SampleInterval * 1000.0)

		snmpParams := &g.GoSNMP{
			Target:    tp.host,
			Port:      uint16(tp.port),
			Transport: "udp4",
			Community: "readwrite",
			Version:   g.Version2c,
			Retries:   3,
			Timeout:   time.Duration(timeoutms) * time.Millisecond,
			// Logger:    log.New(os.Stdout, "", 0),
			// Transport: "udp4",
		}
		if err := snmpParams.Connect(); err != nil {
			// log.Fatalf("Connect() err: %v", err)
			tp.snmpParams = nil
			return err
		}

		tp.snmpParams = snmpParams
		tp.ready = true
	} else {
		return errors.New("error: already connected")
	}

	return nil

}

// // ExportScan will send scan results to stdout formated properly for txtoida fmt=2
// func (tp *TPDin2Device) ExportScan() error {

// 	scan := tp.getScan()

// }

// formatScan returns scan results as a string accoding to txtoida fmt=2 format
func (scan *TPDin2Scan) formatScan() {

}

// queryDeviceVars queries device for TPDin2 OID values
func (tp *TPDin2Device) queryDeviceVars() error {

	oidList := *tp.dynamicOids()

	// TODO: Remove and get real OIDS from TPDIN2 config struct
	oidList = []string{"1.3.6.1.2.1.1.3.0"}

	// fmt.Println("debug: querying device")
	snmpVals, err := tp.snmpParams.Get(oidList)
	ts := time.Now()
	if err != nil {
		fmt.Println("error: querying device")
		return err
	}

	results := make(map[string]string)
	for i, variable := range snmpVals.Variables {
		// fmt.Printf("%d: oid: %s ", i, variable.Name)

		// the Value of each variable returned by Get() implements
		// interface{}. You could do a type switch...
		switch variable.Type {
		case g.OctetString:
			// fmt.Printf("string: %s\n", string(variable.Value.([]byte)))
			results[(oidList)[i]] = string(variable.Value.([]byte))
		default:
			// ... or often you're just interested in numeric values.
			// ToBigInt() will return the Value as a BigInt, for plugging
			// into your calculations.
			// fmt.Printf("number: %d\n", g.ToBigInt(variable.Value))
			results[(oidList)[i]] = g.ToBigInt(variable.Value).String()
		}
	}

	tp.saveScan(ts, &results)

	return err
}

// func (tp *TPDin2Device) saveScan(scan *TPDin2Scan) {
func (tp *TPDin2Device) saveScan(ts time.Time, results *map[string]string) {

	tp.mutex.Lock()
	tp.CurrentScan = &TPDin2Scan{ts, *results}
	tp.mutex.Unlock()
}

// GetScan retuns a copy of the most recent TPDin2Scan struct
func (tp *TPDin2Device) GetScan() *TPDin2Scan {

	if tp.CurrentScan == nil {
		return nil
	}

	tp.mutex.Lock()
	scan := tp.CurrentScan.copy()
	tp.mutex.Unlock()

	return scan
}

// PollStart start polling the connected device
func (tp *TPDin2Device) PollStart(ctx context.Context, wg *sync.WaitGroup) error {
	wg.Add(1)
	defer wg.Done()

	if !tp.ready {
		return fmt.Errorf("TP2DinDevice is not connected to host: %s", tp.host)
	}

	fmt.Println("spawning internal polling loop...")
	// kick off internval polling loop
	go func(ctx context.Context, wg *sync.WaitGroup) {
		wg.Add(1)
		defer wg.Done()

		trigtime := time.Now()

		for {
			trigtime = trigtime.Add(tp.internalInterval)
			// fmt.Println("debug: now=", time.Now().String(), "  trigtime=", trigtime.String())

			select {
			case <-time.After(time.Until(trigtime)):
				// fmt.Println("debug: device being queried...")
				err := tp.queryDeviceVars()
				if err != nil {
					fmt.Println(err)
					continue
				}
				// fmt.Println("debug: device queried without error")
			case <-ctx.Done():
				fmt.Println("debug: context.Done message received, shutting down internal polling loop")
				return
			}
		}

	}(ctx, wg)

	return nil
}
