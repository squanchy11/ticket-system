package sessionHandler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

type UserAccounts struct {
	Users []User `json:"users"`
}

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Salt     string `json:"salt"`
}

var users UserAccounts

func HandleError(err error) {
	if err != nil {
		panic(err)
	}
}

// Gib den relativen Pfad zum Ressourcen-Verzeichnis zur Laufzeit zurück.
func GetAssetsDir() string {
	path, err := os.Getwd()
	HandleError(err)

	// Fallunterscheidung für Aufruf über main.go oder *_test.go aus Unterverzeichnissen.
	if filepath.Base(path) == "ticketBackend" {
		return "./assets/"
	} else {
		return "../assets/"
	}
}

// Lies users.json und importiere alle Benutzerdaten nach &users.
func loadUserData() {
	userData, err := ioutil.ReadFile(GetAssetsDir() + "users.json")
	HandleError(err)

	err = json.Unmarshal(userData, &users)
	HandleError(err)
}

// Speichern einer Sitzung in Form von 2 Cookies: UserID und UserName.
func setSession(id int, username string, response http.ResponseWriter) {
	cookieUserID := &http.Cookie{
		Name:  "sessionUserID",
		Value: strconv.Itoa(id),
		Path:  "/",
	}
	cookieUserName := &http.Cookie{
		Name:  "sessionUserName",
		Value: username,
		Path:  "/",
	}

	http.SetCookie(response, cookieUserID)
	http.SetCookie(response, cookieUserName)
}

// Beim Beenden einer Sitzung werden alle Cookies gelöscht (mit nil überschrieben).
func clearSession(response http.ResponseWriter) {
	cookieUserID := &http.Cookie{
		Name:   "sessionUserID",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	cookieUserName := &http.Cookie{
		Name:   "sessionUserName",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}

	http.SetCookie(response, cookieUserID)
	http.SetCookie(response, cookieUserName)
}

// Lies den Namen des aktiven Nutzers aus dem Session-Cookie.
func GetSessionUserName(request *http.Request) string {
	if cookie, err := request.Cookie("sessionUserName"); err == nil {
		return cookie.Value
	} else {
		HandleError(err)
		return ""
	}
}

// Lies die ID des aktiven Nutzers aus dem Session-Cookie.
func GetSessionUserID(request *http.Request) int {
	if cookie, err := request.Cookie("sessionUserID"); err == nil {
		i, _ := strconv.Atoi(cookie.Value)
		return i
	} else {
		fmt.Println(err)
		return 0
	}
}

// A-3.2:
// Der Zugang für die Bearbeiter soll durch Benutzernamen und Passwort geschützt sein.
//
// Überprüfe anhand der Session-Cookies, ob ein Benutzer eingeloggt ist.
// Benutzer eingeloggt = true; Benutzer ist nicht eingeloggt = false.
func IsUserLoggedIn(request *http.Request) bool {
	return GetSessionUserID(request) != 0 && GetSessionUserName(request) != ""
}

// A-3.2:
// Der Zugang für die Bearbeiter soll durch Benutzernamen und Passwort geschützt sein.
//
// Authentifizierung des Benutzers.
// Die Eingaben des Nutzers werden mit gespeicherten Credentials abgeglichen.
func LoginHandler(response http.ResponseWriter, request *http.Request) {
	inputUsername := request.FormValue("username")
	inputPassword := request.FormValue("password")
	redirectTarget := "/"

	if inputUsername != "" && inputPassword != "" {
		// Lade die aktuellen Daten für registrierte Nutzer.
		loadUserData()

		// Überprüfe, ob der Benutzer registriert ist und prüfe dann das Passwort.
		for i := 0; i < len(users.Users); i++ {
			if users.Users[i].Username == inputUsername {

				// Berechne den Hashwert des Input-Passworts mit dem Salt-Wert, der zum verifizerten Benutzer gehört.
				// Die Authentifizerung war erfolgreich, wenn dieser Hashwert mit dem Gespeicherten übereinstimmt.
				if GetHash(inputPassword, users.Users[i].Salt) == users.Users[i].Password {
					setSession(users.Users[i].ID, users.Users[i].Username, response)
					redirectTarget = "/internal"
				}
			}
		}
	}

	http.Redirect(response, request, redirectTarget, 302)
}

// A-3.2:
// Der Zugang für die Bearbeiter soll durch Benutzernamen und Passwort geschützt sein.
//
// Beenden einer Sitzung und ausloggen des Benutzers durch löschen der Session-Cookies.
func LogoutHandler(response http.ResponseWriter, request *http.Request) {
	clearSession(response)
	http.Redirect(response, request, "/", 302)
}

/*
// Registrieren und speichern eines neuen Benutzers in users.json.
// Dabei wird der Hashwert des Passworts mit einem persönlichen Salt-Wert verschleiert.
// Der Salt-Wert wird für spätere Abgleiche beider Hashwerte benötigt und folglich gespeichert.
func RegistrationHandler(response http.ResponseWriter, request *http.Request) {
	hash, salt := HashString(request.FormValue("password"))
	loadUserData()

	users.Users = append(users.Users, User{
		ID:       len(users.Users),
		Username: request.FormValue("username"),
		Password: hash,
		Salt:     salt,
	})

	usersJson, _ := json.Marshal(users)
	err := ioutil.WriteFile(getPathForUserData(), usersJson, 0644)
	HandleError(err)

	http.Redirect(response, request, "/", 302)
}
*/
