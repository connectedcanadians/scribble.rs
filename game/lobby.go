package game

import (
	"math/rand"

	uuid "github.com/satori/go.uuid"
	"github.com/vmihailenco/msgpack"
)

// Lobby represents a game session.
type Lobby struct {
	ID string

	Settings *LobbySettings

	CurrentDrawing *LobbyDrawing

	State *LobbyState

	// calculated on init
	words                 []string
	scoreEarnedByGuessers int
	alreadyUsedWords      []string
	turnDone              chan struct{}
}

func (m *Lobby) MarshalBinary() ([]byte, error) {
	return msgpack.Marshal(&m)
}

// https://github.com/go-redis/redis/issues/739
func (m *Lobby) UnmarshalBinary(data []byte) error {
	return msgpack.Unmarshal(data, &m)
}

type LobbySettings struct {
	DrawingTime       int
	Rounds            int
	MaxPlayers        int
	CustomWords       []string
	CustomWordsChance int
	ClientsPerIPLimit int
	EnableVotekick    bool
	Language          string
}

func (m *LobbySettings) MarshalBinary() ([]byte, error) {
	return msgpack.Marshal(&m)
}

func (m *LobbySettings) UnmarshalBinary(data []byte) error {
	return msgpack.Unmarshal(data, &m)
}

type LobbyState struct {
	Owner   string             // Owner references the Player that created the lobby.
	Players map[string]*Player // Players references all participants of the Lobby.
	Started bool

	Drawer       string // Drawer references the Player that is currently drawing.
	Round        int    // Round  between 0 and MaxRounds. 0 indicates that it hasn't started yet.
	RoundEndTime int64  // RoundEndTime unix timestamp

	// CurrentWord represents the word that was last selected. If no word has
	// been selected yet or the round is already over, this should be empty.
	CurrentWord    string
	WordChoice     []string    // WordChoice represents the current choice of words.
	WordHints      []*WordHint // WordHints for the current word.
	WordHintsShown []*WordHint // WordHintsShown are the same as WordHints with characters visible.

}

func (m *LobbyState) MarshalBinary() ([]byte, error) {
	return msgpack.Marshal(&m)
}

func (m *LobbyState) UnmarshalBinary(data []byte) error {
	return msgpack.Unmarshal(data, &m)
}

func (m *Packet) MarshalBinary() ([]byte, error) {
	return msgpack.Marshal(&m)
}

func (m *Packet) UnmarshalBinary(data []byte) error {
	return msgpack.Unmarshal(data, &m)
}

type LobbyDrawing struct {
	CurrentDrawing []*Packet
}

// SettingBounds defines the lower and upper bounds for the user-specified
// lobby creation input.
type SettingBounds struct {
	MinDrawingTime       int64
	MaxDrawingTime       int64
	MinRounds            int64
	MaxRounds            int64
	MinMaxPlayers        int64
	MaxMaxPlayers        int64
	MinClientsPerIPLimit int64
	MaxClientsPerIPLimit int64
}

// WordHint describes a character of the word that is to be guessed, whether
// the character should be shown and whether it should be underlined on the
// UI.
type WordHint struct {
	Character rune `json:"character"`
	Underline bool `json:"underline"`
}

// Line is the struct that a client send when drawing
type Line struct {
	FromX     float64 `json:"fromX"`
	FromY     float64 `json:"fromY"`
	ToX       float64 `json:"toX"`
	ToY       float64 `json:"toY"`
	Color     string  `json:"color"`
	LineWidth float64 `json:"lineWidth"`
	GestureID int     `json:"gestureId"`
}

// Fill represents the usage of the fill bucket.
type Fill struct {
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
	Color string  `json:"color"`
}

type Rounds struct {
	Round     int `json:"round"`
	MaxRounds int `json:"maxRounds"`
}

// NewLobby allows creating a lobby, optionally returning errors that
// occured during creation.
func NewLobby(ownerName, session, language string, avatarId int, settings LobbySettings) (*Player, *Lobby, error) {
	settings.Language = language
	lobby := &Lobby{
		ID: uuid.NewV4().String(),

		Settings: &settings,
		State: &LobbyState{
			Players: map[string]*Player{},
		},
		CurrentDrawing: &LobbyDrawing{CurrentDrawing: []*Packet{}},
		turnDone:       make(chan struct{}),
	}

	if len(settings.CustomWords) > 1 {
		rand.Shuffle(len(lobby.Settings.CustomWords), func(i, j int) {
			lobby.Settings.CustomWords[i], lobby.Settings.CustomWords[j] = lobby.Settings.CustomWords[j], lobby.Settings.CustomWords[i]
		})
	}

	lobbiesMu.Lock()
	lobbies = append(lobbies, lobby)
	lobbiesMu.Unlock()

	player := createPlayer(ownerName, session, avatarId)

	lobby.State.Players[player.ID] = player
	lobby.State.Owner = player.ID

	// Read wordlist according to the chosen language
	words, err := readWordList(settings.Language)
	if err != nil {
		//TODO Remove lobby, since we errored.
		return nil, nil, err
	}

	lobby.words = words

	Store.Save(lobby)

	return player, lobby, nil
}
