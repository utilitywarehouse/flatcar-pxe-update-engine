package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/godbus/dbus/v5"
)

const (
	// Implement the same DBus interface as update_engine
	//  - https://github.com/kinvolk/update_engine/blob/v0.4.10/src/update_engine/dbus_constants.h
	dbusName      = "com.coreos.update1"
	dbusPath      = "/com/coreos/update1"
	dbusInterface = "com.coreos.update1.Manager"
	// These interval values were taken from the original update_engine:
	//   - https://github.com/kinvolk/update_engine/blob/v0.4.10/src/update_engine/update_check_scheduler.cc#L14-L20
	intervalInitial  = 7 * time.Minute
	intervalPeriodic = 45 * time.Minute
	intervalFuzz     = 10 * time.Minute
	// This file should always exist on Flatcar
	osReleasePath = "/etc/os-release"
)

var (
	// Flag file location for kured:
	// https://kured.dev/docs/configuration/#reboot-sentinel-file--period
	kuredReleasePath = "/var/run/reboot-required"
)

// dbusConn is an interface for all the methods of *dbus.Conn that the
// updateEngine uses. This allows the connection to dbus to be mocked out in
// tests
type dbusConn interface {
	Emit(path dbus.ObjectPath, name string, values ...interface{}) error
}

// newVersionFunc returns the latest version from the given URL. Can be mocked
// out in tests.
type newVersionFunc func(versionURL string) (string, error)

type updateEngine struct {
	conn       dbusConn
	newVersion newVersionFunc
	osVersion  string
	random     *rand.Rand
	status     *status
	updateCh   chan bool
	versionURL *url.URL
}

func newUpdateEngine(versionURL string) (*updateEngine, error) {
	ov, err := osVersion()
	if err != nil {
		return nil, err
	}

	vu, err := url.Parse(versionURL)
	if err != nil {
		return nil, err
	}

	conn, err := dbus.SystemBus()
	if err != nil {
		return nil, err
	}
	reply, err := conn.RequestName(dbusName, dbus.NameFlagDoNotQueue)
	if err != nil {
		return nil, err
	}
	if reply != dbus.RequestNameReplyPrimaryOwner {
		return nil, fmt.Errorf("Name is already taken: %s", dbusName)
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	ch := make(chan bool)

	ue := &updateEngine{
		conn:       conn,
		newVersion: newVersion,
		osVersion:  ov,
		random:     r,
		status:     newStatus(),
		updateCh:   ch,
		versionURL: vu,
	}

	conn.Export(ue, dbusPath, dbusInterface)

	return ue, nil
}

// GetStatus implements com.coreos.update1.Manager.GetStatus
func (ue *updateEngine) GetStatus() (int64, float64, string, string, int64, *dbus.Error) {
	return ue.status.lastCheckedTime, ue.status.progress, ue.status.currentOperation, ue.status.newVersion, ue.status.newSize, nil
}

// ResetStatus implements com.coreos.update1.Manager.ResetStatus
func (ue *updateEngine) ResetStatus() *dbus.Error {
	ue.status = newStatus()

	log.Printf("The update status was reset")

	return nil
}

// AttemptUpdate implements com.coreos.update1.Manager.AttemptUpdate
func (ue *updateEngine) AttemptUpdate() *dbus.Error {
	log.Printf("Update check requested")

	ue.updateCh <- true

	return nil
}

func (ue *updateEngine) checkForUpdate() error {
	vu := ue.versionURL.String()
	log.Printf("Checking for new version at %s", vu)
	nv, err := ue.newVersion(vu)
	if err != nil {
		return err
	}

	ue.status.lastCheckedTime = time.Now().Unix()

	// If the latest version differs from the current OS version
	// then update the status and emit the new status to dbus
	if nv != ue.osVersion && nv != ue.status.newVersion {
		ue.status.newVersion = nv
		ue.status.currentOperation = updateStatusUpdatedNeedReboot
		if err := ue.conn.Emit(
			dbusPath,
			dbusInterface+".StatusUpdate",
			ue.status.lastCheckedTime,
			ue.status.progress,
			ue.status.currentOperation,
			ue.status.newVersion,
			ue.status.newSize,
		); err != nil {
			return err
		}

		if err := touchFile(kuredReleasePath); err != nil {
			return err
		}
		log.Printf("Updated status: %s\n", ue.status)

		return nil
	}

	log.Printf("Didn't find a new version")

	return nil
}

func (ue *updateEngine) run() {
	// Wait for a short time before performing the first status check
	id := fuzzDuration(ue.random, intervalInitial, intervalFuzz)

	ticker := time.NewTicker(id)
	defer ticker.Stop()

	log.Printf("Waiting %s before the initial update check", id)

	update := func() {
		if err := ue.checkForUpdate(); err != nil {
			log.Printf("Error checking for update: %s", err)
		}
		// Wait for a longer period between updates
		d := fuzzDuration(ue.random, intervalPeriodic, intervalFuzz)
		log.Printf("Waiting %s before next update check", d)
		ticker.Reset(d)
	}

	for {
		select {
		case <-ue.updateCh:
			update()
		case <-ticker.C:
			update()
		}
	}
}

func newVersion(versionURL string) (string, error) {
	resp, err := http.Get(versionURL)
	if err != nil {
		return "", err
	}
	defer func() {
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
	}()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return getValue("FLATCAR_VERSION", string(b))
}

func osVersion() (string, error) {
	b, err := ioutil.ReadFile(osReleasePath)
	if err != nil {
		return "", fmt.Errorf("reading file %q: %w", osReleasePath, err)
	}

	return getValue("VERSION", string(b))
}

func getValue(match, body string) (string, error) {
	sc := bufio.NewScanner(strings.NewReader(body))
	for sc.Scan() {
		spl := strings.SplitN(sc.Text(), "=", 2)

		// Just skip empty lines or lines without a value.
		if len(spl) == 1 {
			continue
		}

		if spl[0] == match {
			return spl[1], nil
		}
	}

	return "", fmt.Errorf("couldn't get value for %s", match)
}

func touchFile(fileName string) error {
	_, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		file, err := os.Create(fileName)
		defer file.Close()
		if err != nil {
			return err
		}
	} else {
		currentTime := time.Now().Local()
		err = os.Chtimes(fileName, currentTime, currentTime)
		if err != nil {
			return err
		}
	}
	return nil
}

// fuzzDuration adds a random jitter to a given duration. It's adapted from the
// equivalent method in the original update_engine:
//   - https://github.com/kinvolk/update_engine/blob/v0.4.10/src/update_engine/utils.cc#L510-L515
func fuzzDuration(r *rand.Rand, value time.Duration, fuzz time.Duration) time.Duration {
	min := int64(value.Nanoseconds() - (fuzz.Nanoseconds() / 2))
	max := int64(value.Nanoseconds() + (fuzz.Nanoseconds() / 2))

	d := r.Int63n(max-min+1) + min

	return time.Duration(d)
}
