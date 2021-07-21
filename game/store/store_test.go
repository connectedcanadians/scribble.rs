package store

import (
	"reflect"
	"testing"
	"time"

	"github.com/go-redis/redis"
	"github.com/scribble-rs/scribble.rs/game"
	"github.com/stretchr/testify/require"
)

func NewTestLobby() *game.Lobby {
	return &game.Lobby{
		ID: "test-id",
		CurrentDrawing: &game.LobbyDrawing{
			CurrentDrawing: []*game.Packet{},
		},
		Settings: &game.LobbySettings{},

		State: &game.LobbyState{
			Players:        map[string]*game.Player{},
			Drawer:         "",
			Owner:          "",
			CurrentWord:    "cyborg",
			WordHints:      []*game.WordHint{&game.WordHint{}},
			WordHintsShown: []*game.WordHint{&game.WordHint{}},
			Round:          0,
			WordChoice:     []string{"cyborg", "bruce lee", "big mac"},
			RoundEndTime:   time.Now().Add(time.Second * 120).Unix(),
		},
	}
}

func requireLobbiesEqual(t *testing.T, expected, actual *game.Lobby) {
	require.NotNil(t, expected)
	require.NotNil(t, actual)
	require.NotEmpty(t, expected.ID)
	require.Equal(t, expected.ID, actual.ID)
	require.True(t, reflect.DeepEqual(expected, actual))

}

func TestSaveLobby(t *testing.T) {
	redisStore, err := NewRedisStore(&redis.Options{
		Addr: "127.0.01:6379",
	})
	require.Nil(t, err)

	stores := []*RedisStore{
		redisStore,
	}

	for _, st := range stores {
		require.NotNil(t, st)

		l := NewTestLobby()
		err := st.Save(l)
		require.Nil(t, err)

		_l, err := st.Load(l.ID)
		require.Nil(t, err)

		requireLobbiesEqual(t, l, _l)

		l.State.Players["asdf"] = &game.Player{
			ID: "asdf",
		}

		require.Nil(t, st.SaveState(l.ID, l.State))
		require.Nil(t, st.SaveSettings(l.ID, l.Settings))
		require.Nil(t, st.SaveDrawOp(l.ID, &game.Packet{Type: "a"}))
		require.Nil(t, st.SaveDrawOp(l.ID, &game.Packet{Type: "b"}))
		require.Nil(t, st.SaveDrawOp(l.ID, &game.Packet{Type: "C"}))

		_l, err = st.Load(l.ID)
		require.Nil(t, err)
		require.Len(t, _l.State.Players, 1)
		require.Len(t, _l.CurrentDrawing.CurrentDrawing, 3)

	}
}
