package main

import (
	"testing"
	"time"
)

func TestBrakeAdmission(t *testing.T) {
	now := time.Now()
	cases := []struct {
		name, old, next, brake string
		admitted               bool
	}{
		{"held", "parked", "ready-to-drive", "on", true},
		{"released", "parked", "ready-to-drive", "off", false},
		{"missing", "parked", "ready-to-drive", "", false},
		{"invalid", "parked", "ready-to-drive", "pulled", false},
		{"startup", "", "ready-to-drive", "on", false},
		{"locked", "stand-by", "ready-to-drive", "on", false},
		{"park", "ready-to-drive", "parked", "on", false},
		{"lock", "ready-to-drive", "stand-by", "on", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s := engineState{mode: engineOn, requireLeftBrake: true}
			s.admitDrive(c.old, c.next, c.brake)
			if s.driveAdmitted != c.admitted {
				t.Fatalf("admitted=%t", s.driveAdmitted)
			}
			wantSound := c.admitted && c.next == vehicleStateReady
			if s.sounding(now, c.next) != wantSound {
				t.Fatal("unexpected sound allowance")
			}
		})
	}
}

func TestBrakeAdmissionLatchedUntilNextDrive(t *testing.T) {
	s := engineState{mode: engineOn, requireLeftBrake: true}
	s.admitDrive(vehicleStateParked, vehicleStateReady, "on")
	s.admitDrive(vehicleStateReady, vehicleStateReady, "off")
	if !s.sounding(time.Now(), vehicleStateReady) {
		t.Fatal("release or duplicate state revoked allowance")
	}
	s.style = 4
	s.volume = 60
	if !s.sounding(time.Now(), vehicleStateReady) {
		t.Fatal("style or volume revoked allowance")
	}
	s.admitDrive(vehicleStateReady, vehicleStateParked, "on")
	s.admitDrive(vehicleStateParked, vehicleStateReady, "off")
	if s.sounding(time.Now(), vehicleStateReady) {
		t.Fatal("previous interval admission was reused")
	}
	s.admitDrive(vehicleStateReady, vehicleStateReady, "on")
	if s.sounding(time.Now(), vehicleStateReady) {
		t.Fatal("brake pressed after transition granted allowance")
	}
}

func TestBrakeGateAppliesToEveryArmedMode(t *testing.T) {
	for _, mode := range []engineMode{engineOn, engineRide, engineTimed} {
		s := engineState{mode: mode, until: time.Now().Add(time.Hour), requireLeftBrake: true}
		if s.sounding(time.Now(), vehicleStateReady) {
			t.Fatalf("%s bypassed gate", mode)
		}
		s.driveAdmitted = true
		if !s.sounding(time.Now(), vehicleStateReady) {
			t.Fatalf("%s ignored admission", mode)
		}
		if s.sounding(time.Now(), vehicleStateParked) {
			t.Fatalf("%s sounded parked", mode)
		}
	}
}

func TestBrakeGateDefaultsOffAndPersists(t *testing.T) {
	c, err := loadConfigFrom(t, "{}")
	if err != nil {
		t.Fatal(err)
	}
	if *c.EngineSound.RequireLeftBrake {
		t.Fatal("gate changed default behavior")
	}
	s := engineState{mode: engineOn}
	if !s.sounding(time.Now(), vehicleStateReady) {
		t.Fatal("legacy activation changed")
	}
	engineInit(c)
	engine.mu.Lock()
	engine.s.requireLeftBrake = true
	engine.s.driveAdmitted = true
	engine.mu.Unlock()
	engineToConfig(c)
	if !*c.EngineSound.RequireLeftBrake {
		t.Fatal("gate not saved")
	}
	engineInit(c)
	fresh, _ := engineSnapshot()
	if fresh.driveAdmitted {
		t.Fatal("restart restored a stale admission")
	}
	if fresh.sounding(time.Now(), vehicleStateReady) {
		t.Fatal("restart permitted sound without observed transition")
	}
}
