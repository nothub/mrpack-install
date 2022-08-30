package mrpack

import (
	"os"
	"testing"
)

func Test_Overrides_Fabulously_Optimized(t *testing.T) {
	t.Parallel()
	zipPath := download(t, "https://cdn.modrinth.com/data/1KVo5zza/versions/1vRDfe1u/MR_Fabulously%20Optimized_4.2.1.mrpack")
	err := ExtractOverrides(zipPath, "fabulously_optimized", Server)
	if err != nil {
		t.Fatal(err)
	}
	err = os.Remove(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	err = os.RemoveAll("fabulously_optimized")
	if err != nil {
		t.Fatal(err)
	}
}

func Test_Overrides_Skyblocker(t *testing.T) {
	t.Parallel()
	zipPath := download(t, "https://cdn.modrinth.com/data/KmiWHzQ4/versions/1.5.0/Skyblocker-Modpack.mrpack")
	err := ExtractOverrides(zipPath, "skyblocker", Client)
	if err != nil {
		t.Fatal(err)
	}
	err = os.Remove(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	err = os.RemoveAll("skyblocker")
	if err != nil {
		t.Fatal(err)
	}
}
