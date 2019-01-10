package note

import (
	"fmt"
	"github.com/SUSE/saptune/system"
	"github.com/SUSE/saptune/txtparser"
	"os"
	"path"
	"strconv"
	"testing"
)

func TestGetServiceName(t *testing.T) {
	val := GetServiceName("UuiddSocket")
	if val != "uuidd.socket" {
		t.Fatal(val)
	}
	val = GetServiceName("Sysstat")
	if val != "sysstat" {
		t.Fatal(val)
	}
	val = GetServiceName("UnkownService")
	if val != "" {
		t.Fatal(val)
	}
}

func TestOptSysctlVal(t *testing.T) {
	op := txtparser.Operator("=")
	val := OptSysctlVal(op, "TestParam", "120", "100")
	if val != "100" {
		t.Fatal(val)
	}
	val = OptSysctlVal(op, "TestParam", "120 300 200", "100 330 180")
	if val != "100	330	180" {
		t.Fatal(val)
	}
	val = OptSysctlVal(op, "TestParam", "120 300", "100 330 180")
	if val != "" {
		t.Fatal(val)
	}
}

//GetBlkVal
func TestOptBlkVal(t *testing.T) {
	val := OptBlkVal("IO_SCHEDULER", "sda@cfq", "noop")
	if val != "sda@noop" {
		t.Fatal(val)
	}
	val = OptBlkVal("IO_SCHEDULER", "sda@cfq", "NoOP")
	if val != "sda@noop" {
		t.Fatal(val)
	}
	val = OptBlkVal("IO_SCHEDULER", "sda@cfq sdb@cfq sdc@none", "noop")
	if val != "sda@noop sdb@noop sdc@noop" {
		t.Fatal(val)
	}
	val = OptBlkVal("NRREQ", "sda@128", "512")
	if val != "sda@512" {
		t.Fatal(val)
	}
	val = OptBlkVal("NRREQ", "sda@128", "0")
	if val != "sda@1024" {
		t.Fatal(val)
	}
	val = OptBlkVal("NRREQ", "sda@128 sdb@512 sdc@1024", "512")
	if val != "sda@512 sdb@512 sdc@512" {
		t.Fatal(val)
	}
}
//SetBlkVal

//GetLimitsVal
func TestOptLimitsVal(t *testing.T) {
	val := OptLimitsVal("LIMIT_HARD", "@dom1:0 @dom2:0 @dom3:0", "32800", "nofile", "@dom1 @dom2 @dom3")
	if val != "@dom1:32800 @dom2:32800 @dom3:32800 " {
		t.Fatal(val)
	}
	val = OptLimitsVal("LIMIT_HARD", "@dom1: @dom2: @dom3:", "32800", "nofile", "@dom1 @dom2 @dom3")
	if val != "@dom1:32800 @dom2:32800 @dom3:32800 " {
		t.Fatal(val)
	}
	val = OptLimitsVal("LIMIT_SOFT", "@dom1:unlimited @dom2:infinity @dom3:-1", "32800", "nofile", "@dom1 @dom2 @dom3")
	if val != "@dom1:unlimited @dom2:infinity @dom3:-1 " {
		t.Fatal(val)
	}
	val = OptLimitsVal("LIMIT_HARD", "@dom1:0 @dom2:0 @dom3:0", "32800", "memlock", "@dom1 @dom2 @dom3")
	if val != "@dom1:32800 @dom2:32800 @dom3:32800 " {
		t.Fatal(val)
	}
	calcLimit := system.GetMainMemSizeMB()*1024 - (system.GetMainMemSizeMB() * 1024 * 10 / 100)
	val = OptLimitsVal("LIMIT_SOFT", "@dom1:0 @dom2:0 @dom3:0", "0", "memlock", "@dom1 @dom2 @dom3")
	if val != fmt.Sprintf("@dom1:%d @dom2:%d @dom3:%d ", calcLimit, calcLimit, calcLimit) {
		t.Fatal(val, calcLimit)
	}
	val = OptLimitsVal("LIMIT_SOFT", "@dom1:67108864 @dom2:67108864 @dom3:67108864", "0", "memlock", "@dom1 @dom2 @dom3")
	if val != "@dom1:67108864 @dom2:67108864 @dom3:67108864 " {
		t.Fatal(val, calcLimit)
	}

	val = OptLimitsVal("LIMIT_ITEM", "", "nofile", "nofile", "@dom1 @dom2 @dom3")
	if val != "@dom1:nofile @dom2:nofile @dom3:nofile " {
		t.Fatal(val)
	}
	val = OptLimitsVal("LIMIT_ITEM", "0", "memlock", "memlock", "dom5")
	if val != "dom5:memlock " {
		t.Fatal(val)
	}

	val = OptLimitsVal("LIMIT_DOMAIN", "", "@dom1 @dom2 @dom3", "nofile", "@dom1 @dom2 @dom3")
	if val != "@dom1 @dom2 @dom3 " {
		t.Fatal(val)
	}
}
//SetLimitsVal

func TestGetVmVal(t *testing.T) {
	val := GetVmVal("THP")
	if val != "always" && val != "madvise" && val != "never" {
		t.Fatalf("wrong value '%+v' for THP.\n", val)
	}
	val = GetVmVal("KSM")
	if val != "1" && val != "0" {
		t.Fatalf("wrong value '%+v' for KSM.\n", val)
	}
}

func TestOptVmVal(t *testing.T) {
	val := OptVmVal("THP", "always")
	if val != "always" {
		t.Fatal(val)
	}
	val = OptVmVal("THP", "unknown")
	if val != "never" {
		t.Fatal(val)
	}
	val = OptVmVal("KSM", "1")
	if val != "1" {
		t.Fatal(val)
	}
	val = OptVmVal("KSM", "2")
	if val != "0" {
		t.Fatal(val)
	}
	val = OptVmVal("UNKOWN_PARAMETER", "unknown")
	if val != "unknown" {
		t.Fatal(val)
	}
}
//SetVmVal

func TestGetCpuVal(t *testing.T) {
	val := GetCpuVal("force_latency")
	if val != "all:none" {
		t.Logf("force_latency supported: '%s'\n", val)
	}
	val = GetCpuVal("energy_perf_bias")
	if val != "all:none" {
		t.Logf("energy_perf_bias supported: '%s'\n", val)
	}
	val = GetCpuVal("governor")
	if val != "all:none" && val != "" {
		t.Logf("governor supported: '%s'\n", val)
	}
}

func TestOptCpuVal(t *testing.T) {
	val := OptCpuVal("force_latency", "1000", "70")
	if val != "70" {
		t.Fatal(val)
	}

	val = OptCpuVal("energy_perf_bias", "all:15", "performance")
	if val != "all:0" {
		t.Fatal(val)
	}
	val = OptCpuVal("energy_perf_bias", "cpu0:15 cpu1:6 cpu2:0", "performance")
	if val != "cpu0:0 cpu1:0 cpu2:0" {
		t.Fatal(val)
	}
/* future feature
	val = OptCpuVal("energy_perf_bias", "cpu0:6 cpu1:6 cpu2:6", "cpu0:performance cpu1:normal cpu2:powersave")
	if val != "cpu0:0 cpu1:6 cpu2:15" {
		t.Fatal(val)
	}
	val = OptCpuVal("energy_perf_bias", "all:6", "cpu0:performance cpu1:normal cpu2:powersave")
	if val != "cpu0:performance cpu1:normal cpu2:powersave" {
		t.Fatal(val)
	}
*/

	val = OptCpuVal("governor", "all:powersave", "performance")
	if val != "all:performance" {
		t.Fatal(val)
	}
	val = OptCpuVal("governor", "cpu0:powersave cpu1:performance cpu2:powersave", "performance")
	if val != "cpu0:performance cpu1:performance cpu2:performance" {
		t.Fatal(val)
	}
/* future feature
	val = OptCpuVal("governor", "cpu0:powersave cpu1:performance cpu2:powersave", "cpu0:performance cpu1:powersave cpu2:performance")
	if val != "cpu0:performance cpu1:powersave cpu2:performance" {
		t.Fatal(val)
	}
	val = OptCpuVal("energy_perf_bias", "all:powersave", "cpu0:performance cpu1:powersave cpu2:performance")
	if val != "cpu0:performance cpu1:powersave cpu2:performance" {
		t.Fatal(val)
	}
*/
}
//SetCpuVal

func TestGetMemVal(t *testing.T) {
	val := GetMemVal("VSZ_TMPFS_PERCENT")
	if val == "-1" {
		t.Log("/dev/shm not found")
	}
	val = GetMemVal("ShmFileSystemSizeMB")
	if val == "-1" {
		t.Log("/dev/shm not found")
	}
	val = GetMemVal("UNKOWN_PARAMETER")
	if val != "" {
		t.Fatal(val)
	}
}

func TestOptMemVal(t *testing.T) {
	val := OptMemVal("VSZ_TMPFS_PERCENT", "47", "80", "0", "80")
	if val != "80" {
		t.Fatal(val)
	}
	val = OptMemVal("VSZ_TMPFS_PERCENT", "-1", "75", "0", "75")
	if val != "75" {
		t.Fatal(val)
	}

	size75 := uint64(system.GetTotalMemSizeMB())*75/100
	size80 := uint64(system.GetTotalMemSizeMB())*80/100

	val = OptMemVal("ShmFileSystemSizeMB", "16043", "0", "0", "80")
	if val != strconv.FormatUint(size80, 10) {
		t.Fatal(val)
	}
	val = OptMemVal("ShmFileSystemSizeMB", "-1", "0", "0", "80")
	if val != "-1" {
		t.Fatal(val)
	}

	val = OptMemVal("ShmFileSystemSizeMB", "16043", "0", "0", "0")
	if val != strconv.FormatUint(size75, 10) {
		t.Fatal(val)
	}
	val = OptMemVal("ShmFileSystemSizeMB", "-1", "0", "0", "0")
	if val != "-1" {
		t.Fatal(val)
	}

	val = OptMemVal("ShmFileSystemSizeMB", "16043", "25605", "25605", "80")
	if val != "25605" {
		t.Fatal(val)
	}
	val = OptMemVal("ShmFileSystemSizeMB", "-1", "25605", "25605", "80")
	if val != "-1" {
		t.Fatal(val)
	}

	val = OptMemVal("ShmFileSystemSizeMB", "16043", "25605", "25605", "0")
	if val != "25605" {
		t.Fatal(val)
	}
	val = OptMemVal("ShmFileSystemSizeMB", "-1", "25605", "25605", "0")
	if val != "-1" {
		t.Fatal(val)
	}

	val = OptMemVal("UNKOWN_PARAMETER", "16043", "0", "0", "0")
	if val != "" {
		t.Fatal(val)
	}
	val = OptMemVal("UNKOWN_PARAMETER", "-1", "0", "0", "0")
	if val != "" {
		t.Fatal(val)
	}
}
//SetMemVal

func TestGetRpmVal(t *testing.T) {
	val := GetRpmVal("rpm:glibc")
	if val == "" {
		t.Log("rpm 'glibc' not found")
	}
}

func TestOptRpmVal(t *testing.T) {
	val := OptRpmVal("rpm:glibc", "NO_OPT")
	if val != "NO_OPT" {
		t.Fatal(val)
	}
}

func TestSetRpmVal(t *testing.T) {
	val := SetRpmVal("NO_OPT")
	if val != nil {
		t.Fatal(val)
	}
}

func TestGetGrubVal(t *testing.T) {
	val := GetGrubVal("grub:processor.max_cstate")
	if val == "NA" {
		t.Log("'processor.max_cstate' not found in kernel cmdline")
	}
	val = GetGrubVal("grub:UNKNOWN")
	if val != "NA" {
		t.Fatal(val)
	}
}

func TestOptGrubVal(t *testing.T) {
	val := OptGrubVal("grub:processor.max_cstate", "NO_OPT")
	if val != "NO_OPT" {
		t.Fatal(val)
	}
}

func TestSetGrubVal(t *testing.T) {
	val := SetGrubVal("NO_OPT")
	if val != nil {
		t.Fatal(val)
	}
}

func TestGetUuiddVal(t *testing.T) {
	val := GetUuiddVal()
	if val != "start" && val != "stop" {
		t.Fatal(val)
	}
}

func TestOptUuiddVal(t *testing.T) {
	val := OptUuiddVal("start")
	if val != "start" {
		t.Fatal(val)
	}
	val = OptUuiddVal("stop")
	if val != "start" {
		t.Fatal(val)
	}
	val = OptUuiddVal("unknown")
	if val != "start" {
		t.Fatal(val)
	}
}
//SetUuiddVal

func TestGetServiceVal(t *testing.T) {
	val := GetServiceVal("UnkownService")
	if val != "" {
		t.Fatal(val)
	}
	val = GetServiceVal("UuiddSocket")
	if val != "start" && val != "stop" && val != "" {
		t.Fatal(val)
	}
}

func TestOptServiceVal(t *testing.T) {
	val := OptServiceVal("UnkownService", "start")
	if val != "" {
		t.Fatal(val)
	}
	val = OptServiceVal("UuiddSocket", "start")
	if val != "start" {
		t.Fatal(val)
	}
	val = OptServiceVal("UuiddSocket", "stop")
	if val != "start" {
		t.Fatal(val)
	}
	val = OptServiceVal("UuiddSocket", "unknown")
	if val != "start" {
		t.Fatal(val)
	}
	val = OptServiceVal("Sysstat", "start")
	if val != "start" {
		t.Fatal(val)
	}
	val = OptServiceVal("Sysstat", "stop")
	if val != "stop" {
		t.Fatal(val)
	}
	val = OptServiceVal("Sysstat", "unknown")
	if val != "start" {
		t.Fatal(val)
	}
}

func TestSetServiceVal(t *testing.T) {
	val := SetServiceVal("UnkownService", "start")
	if val != nil {
		t.Fatal(val)
	}
}

func TestGetLoginVal(t *testing.T) {
	val, err := GetLoginVal("Unkown")
	if val != "" || err != nil {
		t.Fatal(val)
	}

	val, err = GetLoginVal("UserTasksMax")
	if _, errno := os.Stat("/etc/systemd/logind.conf.d/sap.conf"); errno != nil {
		if !os.IsNotExist(errno) {
			if val != "" || err == nil {
				t.Fatal(val)
			}
		} else {
			if val != "" || err != nil {
				t.Fatal(val)
			}
		}
	}
}

func TestOptLoginVal(t *testing.T) {
	val := OptLoginVal("unkown")
	if val != "unkown" {
		t.Fatal(val)
	}
	val = OptLoginVal("infinity")
	if val != "infinity" {
		t.Fatal(val)
	}
	val = OptLoginVal("")
	if val != "" {
		t.Fatal(val)
	}
}
// SetLoginVal

func TestGetPagecacheVal(t *testing.T) {
	prepare := LinuxPagingImprovements{SysconfigPrefix: OSPackageInGOPATH}
	val := GetPagecacheVal("ENABLE_PAGECACHE_LIMIT", &prepare)
	if val != "yes" && val != "no" {
		t.Fatal(val)
	}
	if prepare.VMPagecacheLimitMB == 0 && val != "no" {
		t.Fatal(val)
	}
	if prepare.VMPagecacheLimitMB > 0 && val != "yes" {
		t.Fatal(val)
	}

	prepare = LinuxPagingImprovements{SysconfigPrefix: OSPackageInGOPATH}
	val = GetPagecacheVal("PAGECACHE_LIMIT_IGNORE_DIRTY", &prepare)
	if val != strconv.Itoa(prepare.VMPagecacheLimitIgnoreDirty) {
		t.Fatal(val)
	}

	prepare = LinuxPagingImprovements{SysconfigPrefix: OSPackageInGOPATH}
	val = GetPagecacheVal("OVERRIDE_PAGECACHE_LIMIT_MB", &prepare)
	if prepare.VMPagecacheLimitMB == 0 && val != "" {
		t.Fatal(val)
	}
	if prepare.VMPagecacheLimitMB > 0 && val != strconv.FormatUint(prepare.VMPagecacheLimitMB, 10) {
		t.Fatal(val)
	}

	prepare = LinuxPagingImprovements{SysconfigPrefix: OSPackageInGOPATH}
	val = GetPagecacheVal("UNKOWN", &prepare)
	if val != "" {
		t.Fatal(val)
	}
}

func TestOptPagecacheVal(t *testing.T) {
	initPrepare, _ := LinuxPagingImprovements{SysconfigPrefix: OSPackageInGOPATH, VMPagecacheLimitMB: 0, VMPagecacheLimitIgnoreDirty: 0, UseAlgorithmForHANA: true}.Initialise()
	prepare := initPrepare.(LinuxPagingImprovements)

	val := OptPagecacheVal("UNKNOWN", "unknown", &prepare)
	if val != "unknown" {
		t.Fatal(val)
	}
	val = OptPagecacheVal("ENABLE_PAGECACHE_LIMIT", "yes", &prepare)
	if val != "yes" {
		t.Fatal(val)
	}
	val = OptPagecacheVal("ENABLE_PAGECACHE_LIMIT", "no", &prepare)
	if val != "no" {
		t.Fatal(val)
	}
	val = OptPagecacheVal("ENABLE_PAGECACHE_LIMIT", "unknown", &prepare)
	if val != "no" {
		t.Fatal(val)
	}
	val = OptPagecacheVal("PAGECACHE_LIMIT_IGNORE_DIRTY", "2", &prepare)
	if val != "2" {
		t.Fatal(val)
	}
	if val != strconv.Itoa(prepare.VMPagecacheLimitIgnoreDirty) {
		t.Fatal(val, prepare.VMPagecacheLimitIgnoreDirty)
	}
	val = OptPagecacheVal("PAGECACHE_LIMIT_IGNORE_DIRTY", "1", &prepare)
	if val != "1" {
		t.Fatal(val)
	}
	if val != strconv.Itoa(prepare.VMPagecacheLimitIgnoreDirty) {
		t.Fatal(val, prepare.VMPagecacheLimitIgnoreDirty)
	}
	val = OptPagecacheVal("PAGECACHE_LIMIT_IGNORE_DIRTY", "0", &prepare)
	if val != "0" {
		t.Fatal(val)
	}
	if val != strconv.Itoa(prepare.VMPagecacheLimitIgnoreDirty) {
		t.Fatal(val, prepare.VMPagecacheLimitIgnoreDirty)
	}
	val = OptPagecacheVal("PAGECACHE_LIMIT_IGNORE_DIRTY", "unknown", &prepare)
	if val != "1" {
		t.Fatal(val)
	}
	if val != strconv.Itoa(prepare.VMPagecacheLimitIgnoreDirty) {
		t.Fatal(val, prepare.VMPagecacheLimitIgnoreDirty)
	}

	PCTestConf := path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/pcTest1")
	initPrepare, _ = LinuxPagingImprovements{SysconfigPrefix: PCTestConf, VMPagecacheLimitMB: 0, VMPagecacheLimitIgnoreDirty: 0, UseAlgorithmForHANA: true}.Initialise()
	prepare = initPrepare.(LinuxPagingImprovements)
	val = OptPagecacheVal("OVERRIDE_PAGECACHE_LIMIT_MB", "unknown", &prepare)
	if val != "" || prepare.VMPagecacheLimitMB > 0 {
		t.Fatal(val, prepare.VMPagecacheLimitMB)
	}

	calc := system.GetMainMemSizeMB() * 2 / 100
	PCTestConf = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/pcTest2")
	initPrepare, _ = LinuxPagingImprovements{SysconfigPrefix: PCTestConf, VMPagecacheLimitMB: 0, VMPagecacheLimitIgnoreDirty: 0, UseAlgorithmForHANA: true}.Initialise()
	prepare = initPrepare.(LinuxPagingImprovements)
	val = OptPagecacheVal("OVERRIDE_PAGECACHE_LIMIT_MB", "unknown", &prepare)
	if val == "" || val == "0" {
		t.Fatal(val)
	}
	if val != strconv.FormatUint(prepare.VMPagecacheLimitMB, 10) {
		t.Fatal(val, prepare.VMPagecacheLimitMB)
	}
	if val != strconv.FormatUint(calc, 10) {
		t.Fatal(val, calc)
	}

	PCTestConf = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/pcTest3")
	initPrepare, _ = LinuxPagingImprovements{SysconfigPrefix: PCTestConf, VMPagecacheLimitMB: 0, VMPagecacheLimitIgnoreDirty: 0, UseAlgorithmForHANA: true}.Initialise()
	prepare = initPrepare.(LinuxPagingImprovements)
	val = OptPagecacheVal("OVERRIDE_PAGECACHE_LIMIT_MB", "unknown", &prepare)
	if val != "" || prepare.VMPagecacheLimitMB > 0 {
		t.Fatal(val, prepare.VMPagecacheLimitMB)
	}

	PCTestConf = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/pcTest4")
	initPrepare, _ = LinuxPagingImprovements{SysconfigPrefix: PCTestConf, VMPagecacheLimitMB: 0, VMPagecacheLimitIgnoreDirty: 0, UseAlgorithmForHANA: true}.Initialise()
	prepare = initPrepare.(LinuxPagingImprovements)
	val = OptPagecacheVal("OVERRIDE_PAGECACHE_LIMIT_MB", "unknown", &prepare)
	if val == "" || val == "0" {
		t.Fatal(val)
	}
	if val != strconv.FormatUint(prepare.VMPagecacheLimitMB, 10) {
		t.Fatal(val, prepare.VMPagecacheLimitMB)
	}
	if val != strconv.FormatUint(calc, 10) {
		t.Fatal(val, calc)
	}

	PCTestConf = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/pcTest5")
	initPrepare, _ = LinuxPagingImprovements{SysconfigPrefix: PCTestConf, VMPagecacheLimitMB: 0, VMPagecacheLimitIgnoreDirty: 0, UseAlgorithmForHANA: true}.Initialise()
	prepare = initPrepare.(LinuxPagingImprovements)
	val = OptPagecacheVal("OVERRIDE_PAGECACHE_LIMIT_MB", "unknown", &prepare)
	if val != "" || prepare.VMPagecacheLimitMB > 0 {
		t.Fatal(val, prepare.VMPagecacheLimitMB)
	}

	PCTestConf = path.Join(os.Getenv("GOPATH"), "/src/github.com/SUSE/saptune/testdata/pcTest6")
	initPrepare, _ = LinuxPagingImprovements{SysconfigPrefix: PCTestConf, VMPagecacheLimitMB: 0, VMPagecacheLimitIgnoreDirty: 0, UseAlgorithmForHANA: true}.Initialise()
	prepare = initPrepare.(LinuxPagingImprovements)
	val = OptPagecacheVal("OVERRIDE_PAGECACHE_LIMIT_MB", "unknown", &prepare)
	if val == "" || val == "0" {
		t.Fatal(val)
	}
	if val != strconv.FormatUint(prepare.VMPagecacheLimitMB, 10) {
		t.Fatal(val, prepare.VMPagecacheLimitMB)
	}
	if val != "641" {
		t.Fatal(val)
	}

}

func TestSetPagecacheVal(t *testing.T) {
	prepare := LinuxPagingImprovements{SysconfigPrefix: OSPackageInGOPATH, VMPagecacheLimitMB: 0, VMPagecacheLimitIgnoreDirty: 0, UseAlgorithmForHANA: true}
	val := SetPagecacheVal("UNKNOWN", &prepare)
	if val != nil {
		t.Fatal(val)
	}
}
