package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"net/http"
	"strconv"
	"strings"

	"github.com/scribble-rs/scribble.rs/game"
)

func getLobbyHandler(r *http.Request) (*game.Lobby, error) {
	lobbyID := r.URL.Query().Get("lobby_id")
	if lobbyID == "" {

		lobbyCookie, err := r.Cookie("X-LobbyId")
		if err != nil || lobbyCookie.Value == "" {
			return nil, errors.New("the requested lobby doesn't exist")
		}
		lobbyID = lobbyCookie.Value
	}

	lobby, err := game.GetLoadLobby(lobbyID)

	if err != nil {
		fmt.Println(err)
		return nil, errors.New("the requested lobby doesn't exist")
	}

	return lobby, nil
}

func userSession(r *http.Request) string {
	sessionCookie, noCookieError := r.Cookie("X-UserSession")
	if noCookieError == nil && sessionCookie.Value != "" {
		return sessionCookie.Value
	}

	session, ok := r.Header["Usersession"]
	if ok {
		return session[0]
	}

	return ""
}

func getAvatarId(r *http.Request) (int, bool) {
	if r.Form.Get("avatarId") != "" {
		avatar, err := strconv.Atoi(r.Form.Get("avatarId"))
		if err != nil {
			fmt.Println("avatar id was bad int")
			return 0, false
		}
		return avatar, true
	}

	avatarCookie, noCookieError := r.Cookie("X-Avatar")
	if noCookieError == nil {
		avatar, err := strconv.Atoi(avatarCookie.Value)
		if err != nil {
			fmt.Println("avatar id was bad int")
			return 0, false
		}
		return avatar, true
	}

	return 0, false
}
func getPlayer(lobby *game.Lobby, r *http.Request) *game.Player {
	return lobby.GetPlayerBySession(userSession(r))
}

func getPlayernameHandler(r *http.Request) string {
	name := r.Form.Get("name")
	if len(name) > 0 {
		return trimDownTo(name, 30)
	}

	usernameCookie, noCookieError := r.Cookie("X-Username")
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

	return "" // game.GeneratePlayerName()
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
	err = json.NewEncoder(w).Encode(lobby.State.Players)
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
	err = json.NewEncoder(w).Encode(game.Rounds{Round: lobby.State.Round, MaxRounds: lobby.Settings.Rounds})
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

func enoughIPs(r *http.Request, l *game.Lobby) bool {
	matches := 0

	for _, p := range l.State.Players {
		socket := p.GetWebsocket()
		if socket != nil && remoteAddressToSimpleIP(socket.RemoteAddr().String()) == remoteAddressToSimpleIP(r.RemoteAddr) {
			matches++
		}
	}

	if matches >= l.Settings.ClientsPerIPLimit {
		return false
	}
	return true
}

const (
	DrawingBoardBaseWidth  = 1600
	DrawingBoardBaseHeight = 900
)

// LobbyData is the data necessary for initially displaying all data of
// the lobbies webpage.
type LobbyData struct {
	AgoraToken             string `json:"agoraToken"`
	AgoraUID               uint32 `json:"agoraUid"`
	AgoraAppID             string `json:"agoraAppID"`
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

	// redirect to home page if no username yet
	playerName := getPlayernameHandler(r)
	playerAvatar, ok := getAvatarId(r)
	if playerName == "" || !ok {
		http.SetCookie(w, &http.Cookie{
			Name:     "X-LobbyId",
			Value:    lobby.ID,
			Path:     "/",
			SameSite: http.SameSiteLaxMode,
		})
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// clear redirect otherwise
	http.SetCookie(w, &http.Cookie{
		Name:     "X-LobbyId",
		Value:    "",
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})

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

	if lobby.IsFull() {
		userFacingError(w, "Sorry, but the lobby is full.")
		return
	}

	if !enoughIPs(r, lobby) {
		userFacingError(w, "Sorry, but you have exceeded the maximum number of clients per IP.")
		return
	}

	uid, token, err := getAgoraToken(lobby.ID)
	if err != nil {
		fmt.Println("could not get agora token", err)
	}
	pageData := &LobbyData{
		AgoraUID:               uid,
		AgoraToken:             token,
		AgoraAppID:             AgoraAppID,
		LobbyID:                lobby.ID,
		DrawingBoardBaseWidth:  DrawingBoardBaseWidth,
		DrawingBoardBaseHeight: DrawingBoardBaseHeight,
	}

	lobbyPlayer := getPlayer(lobby, r)
	if lobbyPlayer == nil {
		lobbyPlayer = lobby.JoinPlayer(playerName, "", playerAvatar)
	}

	// Use the players generated usersession and pass it as a cookie.
	http.SetCookie(w, &http.Cookie{
		Name:     "X-UserSession",
		Value:    lobbyPlayer.GetSession(),
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})
	// http.SetCookie(w, &http.Cookie{
	// 	Name:     "X-Username",
	// 	Value:    player.Name,
	// 	Path:     "/",
	// 	SameSite: http.SameSiteLaxMode,
	// })

	templateError := lobbyPage.ExecuteTemplate(w, "lobby.html", pageData)
	if templateError != nil {
		panic(templateError)
	}
}
