package game

import (
	"fmt"

	"github.com/kr/pretty"
)

// GetLobby returns a Lobby that has a matching ID or no Lobby if none could
// be found.
func GetLobby(id string) *Lobby {
	for _, l := range lobbies {
		if l.ID == id {
			return l
		}
	}

	return nil
}

func GetLoadLobby(id string) (*Lobby, error) {
	lobby := GetLobby(id)
	if lobby != nil {
		return lobby, nil
	}

	lobby, err := Store.Load(id)
	if err != nil {
		return nil, err
	}

	pretty.Println(lobby)

	lobby.turnDone = make(chan struct{})

	// Read wordlist according to the chosen language
	words, err := readWordList(lobby.Settings.Language)
	if err != nil {
		//TODO Remove lobby, since we errored.
		return nil, err
	}

	lobby.words = words
	lobby.alreadyUsedWords = []string{}

	lobbiesMu.Lock()
	lobbies = append(lobbies, lobby)
	lobbiesMu.Unlock()

	for _, p := range lobby.State.Players {
		fmt.Println("Loaded Player", p.Name, p.ID)
	}

	return lobby, nil
}

func (l *Lobby) HasConnectedPlayers() bool {
	for _, p := range l.State.Players {
		if p.Connected {
			return true
		}
	}

	return false
}
func (l *Lobby) nextDrawer() *Player {
	for _, p := range l.State.Players {
		if !p.Drawn && p.Connected {
			return p
		}
	}
	return nil
}

// GetPlayerBySession searches for a player, identifying them by usersession.
func (l *Lobby) GetPlayerBySession(userSession string) *Player {
	for _, player := range l.State.Players {
		if player.UserSession == userSession {
			return player
		}
	}

	return nil
}

// GetPlayer searches for a player, identifying them by usersession.
func (l *Lobby) GetPlayerById(id string) *Player {
	for _id, player := range l.State.Players {
		if id == _id {
			return player
		}
	}

	return nil
}

func (l *Lobby) GetAvailableWordHints(player *Player) []*WordHint {
	//The draw simple gets every character as a word-hint. We basically abuse
	//the hints for displaying the word, instead of having yet another GUI
	//element that wastes space.
	if player.State == PlayerStateDrawing || player.State == PlayerStateStandby {
		return l.State.WordHintsShown
	} else {
		return l.State.WordHints
	}
}

func (l *Lobby) canDraw(player *Player) bool {
	return l.State.Drawer == player.ID && l.State.CurrentWord != ""
}

func (l *Lobby) isAnyoneStillGuessing() bool {
	for _, p := range l.State.Players {
		if p.State == PlayerStateGuessing && p.Connected {
			return true
		}
	}

	return false
}

func (l *Lobby) IsFull() bool {
	numNeeded := l.Settings.MaxPlayers

	for _, p := range l.State.Players {
		if p.Connected {
			numNeeded--
			if numNeeded == 0 {
				return true
			}
		}
	}

	return false
}
