package game_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/go-redis/redis"
	"github.com/scribble-rs/scribble.rs/game"
	"github.com/scribble-rs/scribble.rs/game/store"
	"github.com/stretchr/testify/require"
)

func TestGame(t *testing.T) {
	redisStore, err := store.NewRedisStore(&redis.Options{
		Addr: "127.0.01:6379",
	})
	require.Nil(t, err)
	game.Store = redisStore

	game.TriggerSimpleUpdateEvent = func(eventType string, lobby *game.Lobby) {
		fmt.Println("TriggerSimpleUpdateEvent", eventType)
	}
	game.TriggerComplexUpdatePerPlayerEvent = func(eventType string, data func(*game.Player) interface{}, lobby *game.Lobby) {
		fmt.Println("TriggerComplexUpdatePerPlayerEvent", eventType)
	}
	game.TriggerComplexUpdateEvent = func(eventType string, data interface{}, lobby *game.Lobby) {
		fmt.Println("TriggerComplexUpdateEvent", eventType)
	}
	game.SendDataToOtherPlayers = func(sender *game.Player, lobby *game.Lobby, data interface{}) {
		fmt.Println("SendDataToOtherPlayers")
	}
	game.WriteAsJSON = func(player *game.Player, object interface{}) error {
		fmt.Println("WriteAsJSON")
		return nil
	}
	game.WritePublicSystemMessage = func(lobby *game.Lobby, text string) {
		fmt.Println("WritePublicSystemMessage")
	}

	bro, lobby, err := game.NewLobby("test-bro", "test-bro-session", "english", 1, game.LobbySettings{
		ClientsPerIPLimit: 5,
		CustomWords:       []string{},
		CustomWordsChance: 100,
		DrawingTime:       120,
		EnableVotekick:    true,
		MaxPlayers:        12,
		Rounds:            5,
	})

	require.Nil(t, err)
	require.NotNil(t, lobby)
	require.NotNil(t, bro)

	// no players connected
	require.Len(t, lobby.State.Players, 1)
	require.NotEmpty(t, bro.ID)
	require.NotEmpty(t, bro.GetSession())
	require.Equal(t, lobby.State.Owner, bro.ID)

	require.Contains(t, lobby.State.Players, bro.ID)
	require.Equal(t, lobby.State.Round, 0)
	require.Equal(t, lobby.State.RoundEndTime, int64(0))
	require.Equal(t, lobby.State.Drawer, "")
	require.Equal(t, lobby.State.CurrentWord, "")
	require.Empty(t, lobby.State.WordChoice)
	require.Empty(t, lobby.State.WordHints)
	require.Empty(t, lobby.State.WordHintsShown)

	// connect player
	lobby.Connect(bro)

	require.Equal(t, bro.State, game.PlayerStateGuessing)
	require.Equal(t, lobby.State.Round, 0)
	require.False(t, lobby.State.RoundEndTime > 0)
	require.Equal(t, lobby.State.Drawer, "")
	require.Equal(t, lobby.State.CurrentWord, "")
	require.Empty(t, lobby.State.WordChoice)
	require.Empty(t, lobby.State.WordHints)
	require.Empty(t, lobby.State.WordHintsShown)

	// choose word before game started
	err = lobby.HandlePacket([]byte(`{"type":"choose-word", "data":0}`), bro)
	require.NotNil(t, err)
	fmt.Println(err)

	// choose word before game started
	err = lobby.HandlePacket([]byte(`{"type":"start"}`), bro)
	require.Nil(t, err)

	time.Sleep(time.Millisecond)

	require.Equal(t, bro.State, game.PlayerStateDrawing)
	require.Equal(t, lobby.State.Round, 1)
	require.True(t, lobby.State.RoundEndTime > 0)
	require.Equal(t, lobby.State.Drawer, bro.ID)
	require.Equal(t, lobby.State.CurrentWord, "")
	require.NotEmpty(t, lobby.State.WordChoice)
	require.Empty(t, lobby.State.WordHints)
	require.Empty(t, lobby.State.WordHintsShown)

	// choose word
	err = lobby.HandlePacket([]byte(`{"type":"choose-word", "data":0}`), bro)
	require.Nil(t, err)

	require.Equal(t, bro.State, game.PlayerStateDrawing)
	require.Equal(t, lobby.State.Round, 1)
	require.True(t, lobby.State.RoundEndTime > 0)
	require.Equal(t, lobby.State.Drawer, bro.ID)
	require.NotEmpty(t, lobby.State.CurrentWord)
	require.Empty(t, lobby.State.WordChoice)
	require.NotEmpty(t, lobby.State.WordHints)
	require.NotEmpty(t, lobby.State.WordHintsShown)

	require.Len(t, lobby.State.Players, 1)

	bro2Session := lobby.JoinPlayer("test-bro2", "test-bro2-session", 0)
	require.Len(t, lobby.State.Players, 2)
	bro2 := lobby.GetPlayerBySession(bro2Session.UserSession)
	require.Equal(t, bro2.State, game.PlayerStateGuessing)
	require.Equal(t, bro2.Drawn, false)
	lobby.Connect(bro2)

	sisSession := lobby.JoinPlayer("test-sis", "test-sis-session", 0)
	require.Len(t, lobby.State.Players, 3)
	sis := lobby.GetPlayerBySession(sisSession.UserSession)
	require.Equal(t, sis.State, game.PlayerStateGuessing)
	require.Equal(t, sis.Drawn, false)
	lobby.Connect(sis)

	sis2Session := lobby.JoinPlayer("test-sis2", "test-sis2-session", 0)
	require.Len(t, lobby.State.Players, 4)
	sis2 := lobby.GetPlayerBySession(sis2Session.UserSession)
	require.Equal(t, sis2.State, game.PlayerStateGuessing)
	require.Equal(t, sis2.Drawn, false)
	lobby.Connect(sis2)

	require.Equal(t, lobby.State.Round, 1)

	time.Sleep(time.Millisecond * 10)

	require.Equal(t, lobby.State.Round, 1)

	// rounds
	for round := 1; round < 5; round++ {
		// turns
		require.Equal(t, lobby.State.Round, round)

		for turn := 1; turn < 5; turn++ {
			drawer := lobby.GetPlayerById(lobby.State.Drawer)

			// choose word before game started
			err = lobby.HandlePacket([]byte(`{"type":"choose-word", "data":0}`), drawer)
			require.Nil(t, err)

			// insert bullshit drawing data here.
			for _, p := range lobby.State.Players {
				if p == drawer {
					continue
				}

				err = lobby.HandlePacket([]byte(`{"type":"message", "data": "`+lobby.State.CurrentWord+`"}`), p)
				require.Nil(t, err)
			}

			if turn != 4 {
				require.Equal(t, lobby.State.Round, round)
			} else {
				require.Equal(t, lobby.State.Round, round+1)
			}
		}
	}
}
