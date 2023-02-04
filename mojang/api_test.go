package mojang

import (
	"testing"
)

func TestGetPlayerUuid(t *testing.T) {
	player, err := GetPlayer("lit_furnace")
	if err != nil {
		t.Fatal(err.Error())
	}
	if player.Name != "lit_furnace" || player.Uuid != "8be60c03-25c5-4e57-ab5d-0081e8736cf8" {
		t.Fatal("Wrong player data!")
	}
}
