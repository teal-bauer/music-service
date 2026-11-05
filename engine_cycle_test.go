package main

import (
	"encoding/json"
	"testing"
)

func TestDriveOptionsDefaultOff(t *testing.T) {
	c, err := loadConfigFrom(t, "{}")
	if err != nil {
		t.Fatal(err)
	}
	if *c.EngineSound.CycleOnDrive || *c.EngineSound.NotifyOnDrive {
		t.Fatal("drive features changed defaults")
	}
}

func TestDriveNotificationContainsStyleWithoutVolume(t *testing.T) {
	for style, label := range []string{"V8", "Moped", "Sport", "V-Twin", "Electric"} {
		var payload map[string]any
		if err := json.Unmarshal([]byte(engineDriveNotification(engineState{mode: engineOn, style: style, volume: 65})), &payload); err != nil {
			t.Fatal(err)
		}
		if payload["title"] != "Engine sound: "+label {
			t.Fatal(payload)
		}
		if len(payload) != 5 || payload["ttl_ms"] != float64(8000) {
			t.Fatal(payload)
		}
		if _, ok := payload["volume"]; ok {
			t.Fatal("volume included")
		}
	}
	var payload map[string]any
	json.Unmarshal([]byte(engineDriveNotification(engineState{mode: engineOff})), &payload)
	if payload["title"] != "Engine sound off" {
		t.Fatal(payload)
	}
}

func TestChangedDriveNotificationIsDiscardedBeforeConnecting(t *testing.T) {
	previous := cfg
	cfg = nil
	defer func() { cfg = previous }()
	wanted := engineState{mode: engineOn, style: 1, notifyOnDrive: true}
	for _, current := range []engineState{
		{mode: engineOn, style: 1, notifyOnDrive: false},
		{mode: engineOn, style: 2, notifyOnDrive: true},
		{mode: engineOff, style: 1, notifyOnDrive: true},
	} {
		engine.mu.Lock()
		engine.vehicleState = vehicleStateReady
		engine.driveGeneration = 3
		engine.s = current
		engine.mu.Unlock()
		engineNotifyDrive(wanted, 3)
	}
	engine.mu.Lock()
	engine.vehicleState = ""
	engine.mu.Unlock()
}

func TestOldDriveNotificationIsDiscardedBeforeConnecting(t *testing.T) {
	previous := cfg
	cfg = nil
	defer func() { cfg = previous }()
	engine.mu.Lock()
	engine.vehicleState = vehicleStateReady
	engine.driveGeneration = 2
	engine.mu.Unlock()
	engineNotifyDrive(engineState{mode: engineOn}, 1)
	engine.mu.Lock()
	engine.vehicleState = vehicleStateParked
	engine.mu.Unlock()
	engineNotifyDrive(engineState{mode: engineOn}, 2)
	engine.mu.Lock()
	engine.vehicleState = ""
	engine.mu.Unlock()
}
