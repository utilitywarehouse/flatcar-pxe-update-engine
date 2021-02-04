package main

import (
	"flag"
	"log"
	"time"
)

const (
	defaultVersionURL = "https://stable.release.flatcar-linux.net/amd64-usr/current/version.txt"
	// These interval values were taken from the original update_engine:
	//   - https://github.com/kinvolk/update_engine/blob/v0.4.10/src/update_engine/update_check_scheduler.cc#L14-L20
	defaultIntervalInitial  = 7 * time.Minute
	defaultIntervalPeriodic = 45 * time.Minute
	// The default in update_engine was 10 minutes but PXE hosts generally
	// take longer than VMs to boot, so if you're relying on the fuzz to
	// naturally distribute reboots then 20 minutes should avoid more
	// collisions.
	defaultIntervalFuzz = 20 * time.Minute
)

var (
	flagVersionURL       = flag.String("version-url", defaultVersionURL, "Remote location of a version.txt file")
	flagIntervalInitial  = flag.Duration("interval-initial", defaultIntervalInitial, "The period to wait before performing the initial update check")
	flagIntervalPeriodic = flag.Duration("interval-periodic", defaultIntervalPeriodic, "The period to wait between update checks")
	flagIntervalFuzz     = flag.Duration("interval-fuzz", defaultIntervalFuzz, "The period to fuzz the intervals by")
)

func main() {
	flag.Parse()

	ue, err := newUpdateEngine(updateEngineConfig{
		intervalInitial:  *flagIntervalInitial,
		intervalPeriodic: *flagIntervalPeriodic,
		intervalFuzz:     *flagIntervalFuzz,
		versionURL:       *flagVersionURL,
	})
	if err != nil {
		log.Fatalln(err)
	}

	ue.run()
}
