package communication

import (
	"encoding/json"
	"errors"
	"html"
	"net/http"
	"strings"

	"github.com/scribble-rs/scribble.rs/game"
)

func getLobbyHandler(r *http.Request) (*game.Lobby, error) {
	lobbyID := r.URL.Query().Get("lobby_id")
	if lobbyID == "" {
		return nil, errors.New("the requested lobby doesn't exist")
	}

	lobby := game.GetLobby(lobbyID)

	if lobby == nil {
		return nil, errors.New("the requested lobby doesn't exist")
	}

	return lobby, nil
}

func userSession(r *http.Request) string {
	sessionCookie, noCookieError := r.Cookie("usersession")
	if noCookieError == nil && sessionCookie.Value != "" {
		return sessionCookie.Value
	}

	session, ok := r.Header["Usersession"]
	if ok {
		return session[0]
	}

	return ""
}

func getPlayer(lobby *game.Lobby, r *http.Request) *game.Player {
	return lobby.GetPlayer(userSession(r))
}

func getPlayernameHandler(r *http.Request) string {
	usernameCookie, noCookieError := r.Cookie("username")
	if noCookieError == nil {
		username := html.EscapeString(strings.TrimSpace(usernameCookie.Value))
		if username != "" {
			return trimDownTo(username, 30)
		}
	}

	parseError := r.ParseForm()
	if parseError == nil {
		username := r.Form.Get("username")
		if username != "" {
			return trimDownTo(username, 30)
		}
	}

	return game.GeneratePlayerName()
}

func trimDownTo(text string, size int) string {
	if len(text) <= size {
		return text
	}

	return text[:size]
}

// getPlayersHandler returns divs for all players in the lobby to the calling client.
func getPlayersHandler(w http.ResponseWriter, r *http.Request) {
	lobby, err := getLobbyHandler(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if getPlayer(lobby, r) == nil {
		http.Error(w, "you aren't part of this lobby", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(lobby.Players)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

//getRoundsHandler returns the html structure and data for the current round info.
func getRoundsHandler(w http.ResponseWriter, r *http.Request) {
	lobby, err := getLobbyHandler(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if getPlayer(lobby, r) == nil {
		http.Error(w, "you aren't part of this lobby", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(game.Rounds{Round: lobby.Round, MaxRounds: lobby.MaxRounds})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// getWordHintHandler returns the html structure and data for the current word hint.
func getWordHintHandler(w http.ResponseWriter, r *http.Request) {
	lobby, err := getLobbyHandler(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	player := getPlayer(lobby, r)
	if player == nil {
		http.Error(w, "you aren't part of this lobby", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(lobby.GetAvailableWordHints(player))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

const (
	DrawingBoardBaseWidth  = 1600
	DrawingBoardBaseHeight = 900
)

// LobbyData is the data necessary for initially displaying all data of
// the lobbies webpage.
type LobbyData struct {
	LobbyID                string `json:"lobbyId"`
	DrawingBoardBaseWidth  int    `json:"drawingBoardBaseWidth"`
	DrawingBoardBaseHeight int    `json:"drawingBoardBaseHeight"`
}

// ssrEnterLobbyHandler opens a lobby, either opening it directly or asking for a lobby.
func ssrEnterLobbyHandler(w http.ResponseWriter, r *http.Request) {
	lobby, err := getLobbyHandler(r)
	if err != nil {
		userFacingError(w, err.Error())
		return
	}

	// TODO Improve this. Return metadata or so instead.
	userAgent := strings.ToLower(r.UserAgent())
	if !(strings.Contains(userAgent, "gecko") || strings.Contains(userAgent, "chrom") || strings.Contains(userAgent, "opera") || strings.Contains(userAgent, "safari")) {
		userFacingError(w, "Sorry, no robots allowed.")
		return
	}

	//FIXME Temporary
	if strings.Contains(userAgent, "iphone") || strings.Contains(userAgent, "android") {
		userFacingError(w, "Sorry, mobile is currently not supported.")
		return
	}

	player := getPlayer(lobby, r)

	pageData := &LobbyData{
		LobbyID:                lobby.ID,
		DrawingBoardBaseWidth:  DrawingBoardBaseWidth,
		DrawingBoardBaseHeight: DrawingBoardBaseHeight,
	}

	var templateError error

	if player == nil {
		if len(lobby.Players) >= lobby.MaxPlayers {
			userFacingError(w, "Sorry, but the lobby is full.")
			return
		}

		matches := 0
		for _, otherPlayer := range lobby.Players {
			socket := otherPlayer.GetWebsocket()
			if socket != nil && remoteAddressToSimpleIP(socket.RemoteAddr().String()) == remoteAddressToSimpleIP(r.RemoteAddr) {
				matches++
			}
		}

		if matches >= lobby.ClientsPerIPLimit {
			userFacingError(w, "Sorry, but you have exceeded the maximum number of clients per IP.")
			return
		}

		var playerName = getPlayernameHandler(r)
		userSession := lobby.JoinPlayer(playerName)

		// Use the players generated usersession and pass it as a cookie.
		http.SetCookie(w, &http.Cookie{
			Name:     "usersession",
			Value:    userSession,
			Path:     "/",
			SameSite: http.SameSiteStrictMode,
		})
	}

	templateError = lobbyPage.ExecuteTemplate(w, "lobby.html", pageData)
	if templateError != nil {
		panic(templateError)
	}
}
