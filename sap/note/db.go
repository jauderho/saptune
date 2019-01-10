package note

import (
	"github.com/SUSE/saptune/sap"
	"github.com/SUSE/saptune/system"
	"github.com/SUSE/saptune/txtparser"
	"path"
)

// 1557506 - Linux paging improvements
type LinuxPagingImprovements struct {
	SysconfigPrefix string // Used by test cases to specify alternative sysconfig location

	VMPagecacheLimitMB          uint64
	VMPagecacheLimitIgnoreDirty int
	UseAlgorithmForHANA         bool
}

func (paging LinuxPagingImprovements) Name() string {
	return "Linux paging improvements"
}
func (paging LinuxPagingImprovements) Initialise() (Note, error) {
	vmPagecach, _ := system.GetSysctlUint64(system.SysctlPagecacheLimitMB)
	vmIgnoreDirty, _ := system.GetSysctlInt(system.SysctlPagecacheLimitIgnoreDirty)
	return LinuxPagingImprovements{
		SysconfigPrefix:             paging.SysconfigPrefix,
		VMPagecacheLimitMB:          vmPagecach,
		VMPagecacheLimitIgnoreDirty: vmIgnoreDirty,
		UseAlgorithmForHANA:         true,
	}, nil
}
func (paging LinuxPagingImprovements) Optimise() (Note, error) {
	newPaging := paging
	conf, err := txtparser.ParseSysconfigFile(path.Join(newPaging.SysconfigPrefix, "/usr/share/saptune/notes/1557506"), false)
	if err != nil {
		return nil, err
	}
	inputEnable := conf.GetBool("ENABLE_PAGECACHE_LIMIT", false)
	inputOverride := conf.GetInt("OVERRIDE_PAGECACHE_LIMIT_MB", 0)

	// For HANA: new limit is 2% system memory
	newPaging.VMPagecacheLimitMB = system.GetMainMemSizeMB() * 2 / 100
	if inputOverride != 0 {
		newPaging.VMPagecacheLimitMB = uint64(inputOverride)
	}
	if !inputEnable {
		newPaging.VMPagecacheLimitMB = 0
	}
	newPaging.VMPagecacheLimitIgnoreDirty = conf.GetInt("PAGECACHE_LIMIT_IGNORE_DIRTY", 1)
	return newPaging, err
}
func (paging LinuxPagingImprovements) Apply() error {
	errs := make([]error, 0, 0)
	errs = append(errs, system.SetSysctlUint64(system.SysctlPagecacheLimitMB, paging.VMPagecacheLimitMB))
	errs = append(errs, system.SetSysctlInt(system.SysctlPagecacheLimitIgnoreDirty, paging.VMPagecacheLimitIgnoreDirty))

	err := sap.PrintErrors(errs)
	return err
}
