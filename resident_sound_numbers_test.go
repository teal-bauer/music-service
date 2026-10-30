package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestStreamedAndResidentNumberRanges(t *testing.T) {
	for _, tc := range []struct {
		number  int
		allowed bool
	}{
		{0, false}, {1, false}, {2, true}, {239, true}, {240, false}, {255, false}, {256, false},
	} {
		if got := validStreamedSongNumber(tc.number); got != tc.allowed {
			t.Fatalf("number %d: got %v", tc.number, got)
		}
	}
	if got := nextSongNumber([]Song{{Number: 8}, {Number: 240}, {Number: 255}}); got != 9 {
		t.Fatalf("resident numbers affected streamed allocation: %d", got)
	}
	if got := nextSongNumber([]Song{{Number: 239}, {Number: 240}}); got != 0 {
		t.Fatalf("expected explicit exhaustion, got %d", got)
	}
}

func TestFreshResidentDefaultUsesControllerNumber(t *testing.T) {
	dir := t.TempDir()
	if _, err := syncDefaultSongs(dir, embeddedSongs); err != nil {
		t.Fatal(err)
	}
	catalog := readCatalog(t, dir)
	index := catalogIndexByDefault(catalog, "tng-warp2")
	if index < 0 || catalog[index].Number != 240 || catalog[index].Source != "builtin-samples" {
		t.Fatalf("incorrect resident sample: %#v", catalog)
	}
	if got := nextSongNumber(catalog); got != 10 {
		t.Fatalf("wrong next streamed number: %d", got)
	}
}

func TestLegacyResidentSevenAndItsIdentityStayUnchanged(t *testing.T) {
	dir := t.TempDir()
	if _, err := syncDefaultSongs(dir, embeddedSongs); err != nil {
		t.Fatal(err)
	}
	catalog := readCatalog(t, dir)
	index := catalogIndexByDefault(catalog, "tng-warp2")
	catalog[index].Number = 7
	catalog[index].Name = "TNG Warp 2"
	// Model an older installation where number 7 belongs only to the sample.
	for i := range catalog {
		if i != index && catalog[i].Number == 7 {
			catalog[i].Number = 10
		}
	}
	if err := writeSongsFile(filepath.Join(dir, "songs.json"), catalog); err != nil {
		t.Fatal(err)
	}
	if _, err := syncDefaultSongs(dir, embeddedSongs); err != nil {
		t.Fatal(err)
	}
	result := readCatalog(t, dir)
	index = catalogIndexByDefault(result, "tng-warp2")
	if index < 0 || result[index].Number != 7 || result[index].Name != "TNG Warp 2" {
		t.Fatalf("legacy controller id changed without installation: %#v", result)
	}
}

func TestResidentNumberCollisionIsNotOverwritten(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "songs"), 0700); err != nil {
		t.Fatal(err)
	}
	original := []Song{{Number: 240, Name: "user", Tracks: []SongTrack{{File: "mine.bin"}}}}
	if err := writeSongsFile(filepath.Join(dir, "songs.json"), original); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "songs", "mine.bin"), []byte("my track"), 0600); err != nil {
		t.Fatal(err)
	}
	_, err := syncDefaultSongs(dir, embeddedSongs)
	if err == nil {
		t.Fatal("occupied resident slot must not be overwritten")
	}
	data, err := os.ReadFile(filepath.Join(dir, "songs.json"))
	if err != nil {
		t.Fatal(err)
	}
	var result []Song
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 || result[0].Number != 240 || result[0].Name != "user" {
		t.Fatal("catalogue changed on collision")
	}
}
