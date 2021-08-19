package constants

const (
	Version = "1.0.2095.36"
)

var (
	Development = false
	ConfPath    = "/etc/pritunl-endpoint.json"
	VarDir      = "/var/lib/pritunl-endpoint"
)

//timestamp := time.Now().UTC().Add(-2160 * time.Hour)
//cpuUsage = 30.0
//memUsage = 50.0
//
//for i := 0; i < 1296000; i++ {
//timestamp = timestamp.Add(1 * time.Minute)
//cpuUsage = utils.RandFloatData(cpuUsage, 10, 20, 80, 90, 1)
//memUsage = utils.RandFloatData(memUsage, 10, 20, 80, 90, 1)
//
//doc := &System{
//Timestamp: timestamp,
//CpuUsage:  cpuUsage,
//MemTotal:  memTotal,
//MemUsage:  memUsage,
//SwapTotal: swapTotal,
//SwapUsage: swapUsage,
//}
//
//stream.Append(doc)
//}
//
//println("***************************************************")
//println("done")
//println("***************************************************")
//
//time.Sleep(1000000 * time.Second)
