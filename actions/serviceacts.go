package actions

import (
	"fmt"
	"github.com/SUSE/saptune/app"
	"github.com/SUSE/saptune/system"
	"io"
	"os"
	"strings"
)

// ignore flag for takeover
var ignoreFlag = "/run/.saptune.ignore"

// ServiceAction handles service actions like start, stop, status, enable, disable
// it controlles the systemd saptune.service
//func ServiceAction(actionName string, tuneApp *app.App) {
func ServiceAction(actionName, saptuneVersion string, tApp *app.App) {
	switch actionName {
	case "apply":
		// This action name is only used by saptune service, hence it is not advertised to end user.
		ServiceActionApply(tApp)
	case "disable":
		ServiceActionDisable()
	case "disablestop":
		ServiceActionStop(true)
	case "enable":
		ServiceActionEnable()
	case "enablestart":
		ServiceActionStart(true, tApp)
	case "restart":
		// Redirects to systemctl restart saptune.service
		// systemd uses first ExecStop, then ExecStart
		ServiceActionRestart(tApp)
	case "revert":
		// This action name is only used by saptune service, hence it is not advertised to end user.
		ServiceActionRevert(tApp)
	case "reload":
		// This action name is only used by saptune service, hence it is not advertised to end user.
		system.InfoLog("saptune is now restartig the service...")
		ServiceActionRevert(tApp)
		ServiceActionApply(tApp)
	case "start":
		ServiceActionStart(false, tApp)
	case "status":
		ServiceActionStatus(os.Stdout, tApp, saptuneVersion)
	case "stop":
		ServiceActionStop(false)
	case "takeover":
		ServiceActionTakeover(tApp)
	default:
		PrintHelpAndExit(os.Stdout, 1)
	}
}

// ServiceActionTakeover starts and enables the saptune service
// even if competing services (sapconf, tuned) are active.
// These services will be disabled and stopped
// disable and stop sapconf.service and tuned.service
func ServiceActionTakeover(tuneApp *app.App) {
	lockReleased := false
	system.InfoLog("Starting 'saptune.service', this may take some time...")

	// disable and stop 'tuned.service'
	disableAndStopTuned()

	// check, if sapconf is enabled or active
	if system.IsSapconfActive(SapconfService) {
		// disable and stop 'sapconf.service'
		lockReleased = disableAndStopSapconf(lockReleased)
	}

	if !lockReleased {
		// release Lock, to prevent deadlock with systemd service 'saptune.service'
		system.ReleaseSaptuneLock()
	}

	// enable and start 'saptune.service'
	if err := system.SystemctlEnableStart(SaptuneService); err != nil {
		system.ErrorExit("%v", err)
	}
	// saptune.service (start) then calls 'saptune service apply' to
	// tune the system
	if system.SystemctlIsRunning(SaptuneService) && system.SystemctlIsEnabled(SaptuneService) {
		system.InfoLog("Service '%s' has been enabled and started.", SaptuneService)
	} else {
		system.WarningLog("seems enabling and starting service '%s' was not successful. Please check.", SaptuneService)
	}
	if len(tuneApp.TuneForSolutions) == 0 && len(tuneApp.TuneForNotes) == 0 {
		system.InfoLog("Your system has not yet been tuned. Please visit `saptune note` and `saptune solution` to start tuning.")
	}
}

// ServiceActionStart starts the saptune service
// enable service before start, if enableService is true
func ServiceActionStart(enableService bool, tuneApp *app.App) {
	var err error
	saptuneInfo := ""
	system.InfoLog("Starting 'saptune.service', this may take some time...")
	if system.IsSapconfActive(SapconfService) {
		system.ErrorExit("ATTENTION: found an active sapconf, so refuse any action")
	}
	// release Lock, to prevent deadlock with systemd service 'saptune.service'
	system.ReleaseSaptuneLock()
	// enable and/or start 'saptune.service'
	if enableService {
		err = system.SystemctlEnableStart(SaptuneService)
		saptuneInfo = "Service 'saptune.service' has been enabled and started."
	} else {
		err = system.SystemctlStart(SaptuneService)
		saptuneInfo = "Service 'saptune.service' has been started."
	}
	if err != nil {
		system.ErrorExit("%v", err)
	}
	system.InfoLog(saptuneInfo)
	// saptune.service then calls `saptune service apply` to
	// tune the system
	if len(tuneApp.TuneForSolutions) == 0 && len(tuneApp.TuneForNotes) == 0 {
		system.InfoLog("Your system has not yet been tuned. Please visit `saptune note` and `saptune solution` to start tuning.")
	}
	if !system.SystemctlIsEnabled(SaptuneService) {
		system.InfoLog("Remember: if you wish to automatically activate the solution's tuning options after a reboot, you must enable saptune.service by running:\n    saptune service enable\n")
	}
}

// ServiceActionApply is only used by saptune service, hence it is not
// advertised to the end user. It is used to tune the system after reboot
func ServiceActionApply(tuneApp *app.App) {
	// service should fail, if sapconf.service is enabled or has exited
	// but 'active' file is available
	// /var/lib/sapconf/act_profile in sle12
	// /run/sapconf/active in sle15
	if system.IsSapconfActive(SapconfService) {
		system.ErrorExit("ATTENTION: found an active sapconf, so refuse any action")
	}
	system.InfoLog("saptune is now tuning the system...")
	if err := tuneApp.TuneAll(); err != nil {
		system.ErrorExit("%v", err)
	}
}

// ServiceActionEnable enables the saptune service
func ServiceActionEnable() {
	system.InfoLog("Enable 'saptune.service'")
	// service should fail, if sapconf.service is enabled or has exited
	// but 'active' file is available
	// /var/lib/sapconf/act_profile in sle12
	// /run/sapconf/active in sle15
	if system.IsSapconfActive(SapconfService) {
		system.ErrorExit("ATTENTION: found an active sapconf, so refuse any action")
	}
	// enable 'saptune.service'
	if err := system.SystemctlEnable(SaptuneService); err != nil {
		system.ErrorExit("%v", err)
	}
	system.InfoLog("Service 'saptune.service' has been enabled.")
	if !system.SystemctlIsRunning(SaptuneService) {
		system.InfoLog("Service 'saptune.service' is not running. Please use `saptune service start` to start the service and tune the system")
	}
}

// ServiceActionStatus checks the status of the saptune service
func ServiceActionStatus(writer io.Writer, tuneApp *app.App, saptuneVersion string) {
	infoTrigger := map[string]bool{}
	fmt.Fprintln(writer, "")
	// check for running saptune.service
	infoTrigger["saptuneStopped"], infoTrigger["remember"], infoTrigger["stenabled"] = printSaptuneStatus(writer)

	// print saptune version
	printSaptuneVers(writer, saptuneVersion)

	// Check for any enabled note/solution
	infoTrigger["notTuned"] = printNoteAndSols(writer, tuneApp)

	// staging
	printStagingStatus(writer)

	// check for running sapconf.service and print status
	infoTrigger["scenabled"] = printSapconfStatus(writer)

	// check for running tuned.service and print status
	printTunedStatus(writer)

	// check for system(d) state
	infoTrigger["chkHint"] = printSystemStatus(writer)

	printInfoBlock(writer, infoTrigger)

	// order of exit codes important for yast2 module!
	// first 'stopped', then 'notTuned', then 'ok'
	if infoTrigger["saptuneStopped"] {
		system.ErrorExit("", exitSaptuneStopped)
	}
	if infoTrigger["notTuned"] {
		system.ErrorExit("", exitNotTuned)
	}
}

// ServiceActionStop stops the saptune service
// disable service before stop, if disableService is true
func ServiceActionStop(disableService bool) {
	var err error
	saptuneInfo := ""

	system.InfoLog("Stopping 'saptune.service', this may take some time...")
	// release Lock, to prevent deadlock with systemd service 'saptune.service'
	system.ReleaseSaptuneLock()
	// disable and/or stop 'saptune.service'
	if disableService {
		err = system.SystemctlDisableStop(SaptuneService)
		saptuneInfo = "Service 'saptune.service' has been disabled and stopped."
	} else {
		err = system.SystemctlStop(SaptuneService)
		saptuneInfo = "Service 'saptune.service' has been stopped."
	}
	if err != nil {
		system.ErrorExit("%v", err)
	}
	system.InfoLog(saptuneInfo)
	// saptune.service then calls `saptune daemon revert` to
	// revert all tuned parameter
	system.InfoLog("All tuned parameters have been reverted to default.")
}

// ServiceActionRestart is only used by saptune service, hence it is not
// advertised to the end user. It is used to restart the saptune service
func ServiceActionRestart(tuneApp *app.App) {
	var err error
	system.InfoLog("Restarting 'saptune.service', this may take some time...")
	// release Lock, to prevent deadlock with systemd service 'saptune.service'
	system.ReleaseSaptuneLock()
	// restart 'saptune.service'
	if err = system.SystemctlRestart(SaptuneService); err != nil {
		system.ErrorExit("%v", err)
	}
	system.InfoLog("Service 'saptune.service' has been restarted.")
}

// ServiceActionRevert is only used by saptune service, hence it is not
// advertised to the end user. It is used to revert all the tuned parameters
// right before a system reboot
func ServiceActionRevert(tuneApp *app.App) {
	// service should fail, if sapconf.service is enabled or has exited
	// but 'active' file is available
	// /var/lib/sapconf/act_profile in sle12
	// /run/sapconf/active in sle15
	// skip these checks in case of preventing a saptune/sapconf
	// service deadlock in ServiceActionTakeover, ignoreFlag set
	if _, err := os.Stat(ignoreFlag); os.IsNotExist(err) {
		if system.IsSapconfActive(SapconfService) {
			system.ErrorExit("ATTENTION: found an active sapconf, so refuse any action")
		}
		if len(tuneApp.NoteApplyOrder) != 0 {
			system.InfoLog("saptune is now reverting all settings...")
		}
	} else {
		system.WarningLog("ignore flag set, skipping check for active sapconf service")
	}
	if err := tuneApp.RevertAll(false); err != nil {
		system.ErrorExit("%v", err)
	}
}

// ServiceActionDisable disables the saptune service
func ServiceActionDisable() {
	system.InfoLog("Disable 'saptune.service'")
	// disable 'saptune.service'
	if err := system.SystemctlDisable(SaptuneService); err != nil {
		system.ErrorExit("%v", err)
	}
	system.InfoLog("Service 'saptune.service' has been disabled.")
	if system.SystemctlIsRunning(SaptuneService) {
		system.InfoLog("Service 'saptune.service' still running. Please use `saptune service stop` to stop the service and revert the tuned parameter")
	}
}

// disableAndStopSapconf disables and stops 'sapconf.service'
func disableAndStopSapconf(lockReleased bool) bool {
	// check, if saptune is enabled or active
	if system.SystemctlIsRunning(SaptuneService) || system.SystemctlIsEnabled(SaptuneService) {
		// disable/stop saptune to prevent failed sapconf service
		// release Lock, to prevent deadlock with systemd service 'saptune.service'
		system.ReleaseSaptuneLock()
		lockReleased = true
		// set 'ignore' Flag
		ignore, err := os.Create(ignoreFlag)
		if err != nil {
			system.ErrorExit("Cannot create our 'ignore' flag - %v", err)
		}
		defer ignore.Close()
		if err := system.SystemctlDisableStop(SaptuneService); err != nil {
			os.Remove(ignoreFlag)
			system.ErrorExit("%v", err)
		}
		// remove ignore Flag
		os.Remove(ignoreFlag)
	}
	if err := system.SystemctlDisableStop(SapconfService); err != nil {
		system.ErrorExit("%v", err)
	}
	if system.SystemctlIsRunning(SapconfService) || system.SystemctlIsEnabled(SapconfService) {
		system.WarningLog("seems disabling and stopping service '%s' was not successful. Please check.", SapconfService)
	} else {
		system.InfoLog("Service '%s' disabled and stopped", SapconfService)
	}
	return lockReleased
}

// disableAndStopTuned disables and stops 'tuned.service'
func disableAndStopTuned() {
	if system.IsServiceAvailable(TunedService) {
		if err := system.SystemctlDisableStop(TunedService); err != nil {
			system.ErrorExit("%v", err)
		}
		if system.SystemctlIsRunning(TunedService) || system.SystemctlIsEnabled(TunedService) {
			system.WarningLog("seems disabling and stopping service '%s' was not successful. Please check.", TunedService)
		} else {
			system.InfoLog("Service '%s' disabled and stopped", TunedService)
		}
	}
}

// printSapconfStatus prints status of sapconf.service
func printSapconfStatus(writer io.Writer) bool {
	scenabled := false
	fmt.Fprintf(writer, "sapconf.service:        ")
	if system.IsServiceAvailable(SapconfService) {
		if system.SystemctlIsEnabled(SapconfService) {
			fmt.Fprintf(writer, "enabled/")
			scenabled = true
		} else {
			fmt.Fprintf(writer, "disabled/")
		}
		if system.SystemctlIsRunning(SapconfService) {
			fmt.Fprintf(writer, "active\n")
		} else {
			fmt.Fprintf(writer, "stopped\n")
		}
	} else {
		fmt.Fprintf(writer, "not available\n")
	}
	return scenabled
}

// printTunedStatus prints status of tuned.service
func printTunedStatus(writer io.Writer) {
	fmt.Fprintf(writer, "tuned.service:          ")
	if system.IsServiceAvailable(TunedService) {
		if system.SystemctlIsEnabled(TunedService) {
			fmt.Fprintf(writer, "enabled/")
		} else {
			fmt.Fprintf(writer, "disabled/")
		}
		if system.SystemctlIsRunning(TunedService) {
			fmt.Fprintf(writer, "running (profile: '%s')\n", system.GetTunedAdmProfile())
		} else {
			fmt.Fprintf(writer, "stopped\n")
		}
	} else {
		fmt.Fprintf(writer, "not available\n")
	}
}

// printNoteAndSols prints all enabled/active notes and solutions
func printNoteAndSols(writer io.Writer, tuneApp *app.App) bool {
	notTuned := true
	fmt.Fprintf(writer, "configured solution:    ")
	if len(tuneApp.TuneForSolutions) > 0 {
		fmt.Fprintf(writer, "%s", tuneApp.TuneForSolutions[0])
		notTuned = false
	}
	fmt.Fprintf(writer, "\n")
	fmt.Fprintf(writer, "configured Notes:       ")
	if len(tuneApp.TuneForNotes) > 0 {
		for _, noteID := range tuneApp.TuneForNotes {
			fmt.Fprintf(writer, " "+noteID)
		}
		notTuned = false
	}
	fmt.Fprintf(writer, "\n")
	fmt.Fprintf(writer, "order of enabled notes: ")
	if len(tuneApp.NoteApplyOrder) != 0 {
		fmt.Fprintf(writer, "%s", strings.Join(tuneApp.NoteApplyOrder, " "))
	}
	fmt.Fprintf(writer, "\n")
	fmt.Fprintf(writer, "applied Notes:          ")
	appliedNotes, _ := tuneApp.State.List()
	if len(appliedNotes) != 0 {
		fmt.Fprintf(writer, "%s", strings.Join(appliedNotes, " "))
	}
	fmt.Fprintf(writer, "\n")
	return notTuned
}

// printSaptuneVers prints saptune version
func printSaptuneVers(writer io.Writer, saptuneVersion string) {
	// print saptune rpm version and date
	// because of the need of 'reproducible' builds, we can not use a
	// build date in the 'official' saptune binary, so 'RPMDate' will
	// report 'undef'
	fmt.Fprintf(writer, "saptune package:        '%s'", RPMVersion)
	if RPMDate != "undef" {
		fmt.Fprintf(writer, " (%s)", RPMDate)
	}
	fmt.Fprintf(writer, "\n")
	fmt.Fprintf(writer, "configured version:     '%s'\n", saptuneVersion)

}

// printStagingStatus prints the status of the staging area
func printStagingStatus(writer io.Writer) {
	fmt.Fprintf(writer, "staging:                ")
	stagingSwitch := getStagingFromConf()
	if stagingSwitch {
		fmt.Fprintf(writer, "enabled\n")
	} else {
		fmt.Fprintf(writer, "disabled\n")
	}
	_, files := system.ListDir(StagingSheets, "")
	fmt.Fprintf(writer, "staging area:           %s\n", strings.Join(files, " "))
	fmt.Fprintln(writer, "")
}

// printSaptuneStatus checks for running saptune.service and print status
func printSaptuneStatus(writer io.Writer) (bool, bool, bool) {
	remember := false
	saptuneStopped := false
	stenabled := false
	fmt.Fprintf(writer, "saptune.service:        ")
	if !system.SystemctlIsEnabled(SaptuneService) {
		fmt.Fprintf(writer, "disabled/")
		remember = true
	} else {
		fmt.Fprintf(writer, "enabled/")
		stenabled = true
	}
	if system.SystemctlIsRunning(SaptuneService) {
		fmt.Fprintf(writer, "active\n")
	} else {
		fmt.Fprintf(writer, "stopped\n")
		saptuneStopped = true
	}
	return saptuneStopped, remember, stenabled
}

// printSystemStatus prints the state of the systemd
func printSystemStatus(writer io.Writer) bool {
	chkHint := false
	state, err := system.GetSystemState()
	fmt.Fprintf(writer, "system state:           %s\n", state)
	if err != nil {
		chkHint = true
	}
	return chkHint
}

// printInfoBlock prints additional info for the status
func printInfoBlock(writer io.Writer, infoTrigger map[string]bool) {
	fmt.Fprintln(writer, "")
	if infoTrigger["remember"] {
		fmt.Fprintf(writer, "Remember: if you wish to automatically activate the note's and solution's tuning options after a reboot, you must enable ")
		if infoTrigger["saptuneStopped"] {
			fmt.Fprintf(writer, "and start saptune.service by running:\n 'saptune service enablestart'.\n")
		} else {
			fmt.Fprintf(writer, "saptune.service by running:\n 'saptune service enable'.\n")
		}
	}
	if infoTrigger["notTuned"] {
		fmt.Fprintf(writer, "Your system has not yet been tuned. Please visit `saptune note` and `saptune solution` to start tuning.\n")
	}
	if infoTrigger["stenabled"] && infoTrigger["scenabled"] {
		fmt.Fprintf(writer, "WARNING! saptune.service and sapconf.service are BOTH enabled!\nOnly one tool may tune the system.\n")
	}
	if infoTrigger["chkHint"] {
		fmt.Fprintf(writer, "The system state is NOT ok.\n")
	}
	if (infoTrigger["stenabled"] && infoTrigger["scenabled"]) || infoTrigger["chkHint"] {
		fmt.Fprintf(writer, "Please call '/bin/saptune_check' to get guidance to resolve the issues!\n")
	}

	fmt.Fprintln(writer, "")
}

// DaemonAction handles daemon actions like start, stop, status asm.
// still available for compatibility reasons
func DaemonAction(writer io.Writer, actionName, saptuneVersion string, tuneApp *app.App) {
	serviceAction := actionName
	if actionName == "start" {
		serviceAction = "takeover"
	}
	if actionName == "stop" {
		serviceAction = "disablestop"
	}
	system.WarningLog("ATTENTION: the argument 'daemon' is deprecated!. saptune will forward the request to 'saptune service %s'.\nFor the future please use 'saptune service %s'.", serviceAction, serviceAction)
	switch actionName {
	case "start":
		ServiceActionTakeover(tuneApp)
	case "status":
		ServiceActionStatus(writer, tuneApp, saptuneVersion)
	case "stop":
		// disablestop
		ServiceActionStop(true)
	default:
		PrintHelpAndExit(os.Stdout, 1)
	}
}
