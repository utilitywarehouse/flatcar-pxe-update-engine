package main

import (
	"fmt"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/stretchr/testify/assert"
)

type dbusConnRecorder struct {
	Path   dbus.ObjectPath
	Name   string
	Values []interface{}
}

// Emit implements dbusConn and records the arguments passed to Emit
func (d *dbusConnRecorder) Emit(path dbus.ObjectPath, name string, values ...interface{}) error {
	d.Path = path
	d.Name = name
	d.Values = values

	return nil
}

func mockNewVersionFuncFactory(version string) newVersionFunc {
	return func(string) (string, error) {
		return version, nil
	}
}

func TestUpdateEngineAttemptUpdate(t *testing.T) {
	ue := &updateEngine{updateCh: make(chan bool)}

	var err error

	go func() {
		err = ue.AttemptUpdate()
	}()

	select {
	case ret := <-ue.updateCh:
		assert.True(t, ret)
	case <-time.After(3 * time.Second):
		t.Fatal("expected to receive update request on channel")
	}

	assert.Nil(t, err)
}

func TestUpdateEngineCheckForUpdate(t *testing.T) {
	versionURL, _ := url.Parse(defaultVersionURL)

	s := newStatus()

	d := &dbusConnRecorder{}

	// Test path for the kured release loaction
	kuredReleasePath = fmt.Sprintf("%s/reboot-required", t.TempDir())

	ue := updateEngine{
		conn:       d,
		status:     s,
		osVersion:  "2605.11.0",
		newVersion: mockNewVersionFuncFactory("2605.11.0"),
		versionURL: versionURL,
	}

	// The os version and new version match so there shouldn't be any update
	// statuses emitted
	if err := ue.checkForUpdate(); err != nil {
		t.Fatal("expected nil error")
	}
	assert.Equal(t, &dbusConnRecorder{}, ue.conn)
	assert.Greater(t, ue.status.lastCheckedTime, int64(0))

	// Now that the version is different the status should be updated and a
	// signal should have been emitted with the expected values
	ue.newVersion = mockNewVersionFuncFactory("2605.12.0")
	err := ue.checkForUpdate()

	assert.Nil(t, err)
	assert.Equal(t, d.Path, dbus.ObjectPath(dbusPath))
	assert.Equal(t, d.Name, dbusInterface+".StatusUpdate")
	assert.Equal(t, d.Values[2], updateStatusUpdatedNeedReboot)
	assert.Equal(t, d.Values[3], "2605.12.0")
	assert.Equal(t, ue.status.currentOperation, updateStatusUpdatedNeedReboot)
	assert.Equal(t, ue.status.newVersion, "2605.12.0")
	// Check that we have the flag file for kured
	if _, err := os.Stat(kuredReleasePath); os.IsNotExist(err) {
		t.Fatal("Missing release flag file for kured")
	}
}

func TestUpdateEngineResetStatus(t *testing.T) {
	ue := &updateEngine{
		status: &status{
			lastCheckedTime:  12345,
			progress:         12.6,
			currentOperation: updateStatusUpdatedNeedReboot,
			newVersion:       "2605.12.0",
			newSize:          10,
		},
	}
	err := ue.ResetStatus()

	assert.Nil(t, err)
	assert.Equal(t, ue.status, newStatus())
}

func TestUpdateEngineGetStatus(t *testing.T) {
	s := &status{
		lastCheckedTime:  12345,
		progress:         12.6,
		currentOperation: updateStatusUpdatedNeedReboot,
		newVersion:       "2605.12.0",
		newSize:          10,
	}

	ue := &updateEngine{
		status: s,
	}

	lastCheckedTime, progress, currentOperation, newVersion, newSize, err := ue.GetStatus()

	assert.Nil(t, err)
	assert.Equal(t, s.lastCheckedTime, lastCheckedTime)
	assert.Equal(t, s.progress, progress)
	assert.Equal(t, s.currentOperation, currentOperation)
	assert.Equal(t, s.newVersion, newVersion)
	assert.Equal(t, s.newSize, newSize)
}
