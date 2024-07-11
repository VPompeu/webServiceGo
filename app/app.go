package app

import (
	"database/sql"
	"fmt"
	"log"

	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/VPompeu/agenda-astrologica/models"
	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/rs/cors"
	"golang.org/x/crypto/bcrypt"
)

type App struct {
	Router  *mux.Router
	DB      *sql.DB
	Handler http.Handler
}

type Claims struct {
	Email string `json:"email"`
	ID    int    `json:"id"`
	jwt.RegisteredClaims
}

func (a *App) Initialize(user, password, dbname, dbhost, dbport string) {
	log.Print(user)
	log.Print(password)
	log.Print(dbname)
	connectionString :=
		fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=disable", user, password, dbname, dbhost, dbport)

	var err error
	a.DB, err = sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal(err)
	}
	erro := a.DB.Ping()
	if erro != nil {
		log.Print(erro)
	}
	log.Print(err)
	log.Print(a.DB)

	a.Router = mux.NewRouter()
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"}, // Substitua pelo seu domínio de origem
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           300,
		Debug:            false,
	})
	a.Handler = c.Handler(a.Router)
	a.initializeRoutes()
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func (a *App) checkDBConnection(w http.ResponseWriter, r *http.Request) {
	log.Print(a.DB)
	err := a.DB.Ping()
	log.Print("Chegayy")
	if err != nil {
		log.Print(err)
		respondWithError(w, http.StatusInternalServerError, "Database connection failed")
		return
	}
	log.Print(err)
	respondWithJSON(w, http.StatusOK, map[string]string{"status": "Database connection successful"})
}

func (a *App) getUsers(w http.ResponseWriter, r *http.Request) {
	count, _ := strconv.Atoi(r.FormValue("count"))
	start, _ := strconv.Atoi(r.FormValue("start"))

	if count > 10 || count < 1 {
		count = 10
	}
	if start < 0 {
		start = 0
	}

	users, err := models.GetUsers(a.DB, start, count)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, users)
}

func authenticate(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenHeader := r.Header.Get("Authorization")
		if tokenHeader == "" {
			respondWithError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		tokenParts := strings.Split(tokenHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			respondWithError(w, http.StatusUnauthorized, "Invalid Token")
			return
		}

		tokenStr := tokenParts[1]
		claims := &Claims{}
		jwtKey := os.Getenv("JWT_KEY")

		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtKey), nil
		})

		if err != nil {
			if err == jwt.ErrSignatureInvalid {
				respondWithError(w, http.StatusUnauthorized, "Unauthorized")
				return
			}
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		if !token.Valid {
			respondWithError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		next.ServeHTTP(w, r)
	}
}

func (a *App) getUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	u := models.User{ID: id}
	if err := u.GetUser(a.DB); err != nil {
		switch err {
		case sql.ErrNoRows:
			respondWithError(w, http.StatusNotFound, "User not found")
		default:
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	respondWithJSON(w, http.StatusOK, u)
}

func (a *App) addNote(w http.ResponseWriter, r *http.Request) {
	var n models.GlobalNote
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&n); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	// Chama a função Add para adicionar a anotação global
	err := n.Add(a.DB)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Retorna a anotação global adicionada
	respondWithJSON(w, http.StatusCreated, n)
}

func (a *App) getNotes(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	date := vars["date"]

	// Verificar se a data tem exatamente 8 caracteres
	if len(date) != 8 {
		respondWithError(w, http.StatusBadRequest, "Invalid date format")
		return
	}

	// Converter a data de DDMMYYYY para DD/MM/YYYY
	formattedDate := fmt.Sprintf("%s/%s/%s", date[:2], date[2:4], date[4:8])

	var n models.GlobalNote
	if err := n.Get(a.DB, formattedDate); err != nil {
		switch err {
		case sql.ErrNoRows:
			respondWithJSON(w, http.StatusOK, nil)
		default:
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	respondWithJSON(w, http.StatusOK, n)
}

func (a *App) addUserNoteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	var n models.UserNote
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&n); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	// Convert userID to integer and assign it to UserNote
	n.UserID, _ = strconv.Atoi(userID)

	// Chama a função Add para adicionar a nota do usuário
	err := n.AddUserNote(a.DB)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Retorna a nota do usuário adicionada
	respondWithJSON(w, http.StatusCreated, n)
}

func (a *App) getUserNoteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}
	date := vars["date"]

	// Verificar se a data tem exatamente 8 caracteres
	if len(date) != 8 {
		respondWithError(w, http.StatusBadRequest, "Invalid date format")
		return
	}

	// Converter a data de DDMMYYYY para DD/MM/YYYY
	formattedDate := fmt.Sprintf("%s/%s/%s", date[:2], date[2:4], date[4:8])

	var n models.UserNote
	if err := n.GetUserNoteByDate(a.DB, userID, formattedDate); err != nil {
		if err == sql.ErrNoRows {
			respondWithJSON(w, http.StatusOK, nil)
		} else {
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	// Se encontrar a nota, retornar um array com uma única nota
	respondWithJSON(w, http.StatusOK, n)
}

func (a *App) createUser(w http.ResponseWriter, r *http.Request) {
	var u models.User
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&u); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	if u.Name == "" {
		respondWithError(w, http.StatusBadRequest, "Empty Name")
		return
	}

	if u.Email == "" {
		respondWithError(w, http.StatusBadRequest, "Empty E-mail")
		return
	}

	if len(u.Password) < 6 {
		respondWithError(w, http.StatusBadRequest, "Short Password")
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	u.Password = string(passwordHash)

	if err := u.CreateUser(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, u)
}

func (a *App) updateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var u models.User
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&u); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid resquest payload")
		return
	}
	defer r.Body.Close()
	u.ID = id

	if err := u.UpdateUser(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, u)
}

func (a *App) deleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid User ID")
		return
	}

	u := models.User{ID: id}
	if err := u.DeleteUser(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"result": "200"})
}

func (a *App) login(w http.ResponseWriter, r *http.Request) {
	var u models.User
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&u); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	authenticated, err := u.Login(a.DB)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
	} else if authenticated {
		expirationTime := time.Now().Add(24 * time.Hour)

		expirationUnix := jwt.NewNumericDate(expirationTime)
		claims := &Claims{
			Email: u.Email,
			ID:    u.ID,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: expirationUnix,
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		jwtKey := os.Getenv("JWT_KEY")
		tokenString, err := token.SignedString([]byte(jwtKey))
		if err != nil {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		respondWithJSON(w, http.StatusOK, tokenString)
	} else {
		respondWithError(w, http.StatusUnauthorized, "Falha no login")
	}
}

func (a *App) signin(w http.ResponseWriter, r *http.Request) {
	var u models.User
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&u); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	if u.Name == "" {
		respondWithError(w, http.StatusBadRequest, "Empty Name")
		return
	}

	if u.Email == "" {
		respondWithError(w, http.StatusBadRequest, "Empty E-mail")
		return
	}

	if len(u.Password) < 6 {
		respondWithError(w, http.StatusBadRequest, "Short Password")
		return
	}
	// Chama a função Register para registrar o novo usuário
	err := u.Register(a.DB)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Retorna o email do usuário registrado
	respondWithJSON(w, http.StatusCreated, map[string]string{"email": u.Email})
}

func (a *App) activateLicense(w http.ResponseWriter, r *http.Request) {
	var u models.User
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&u); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	if u.Email == "" {
		respondWithError(w, http.StatusBadRequest, "Empty E-mail")
		return
	}

	// Chama a função Register para registrar o novo usuário
	err := u.Activate(a.DB)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Retorna o email do usuário registrado
	respondWithJSON(w, http.StatusCreated, "Usuario ativado com sucesso!")
}

func (a *App) initializeRoutes() {
	a.Router.HandleFunc("/checkdb", a.checkDBConnection).Methods("GET")

	a.Router.HandleFunc("/users", authenticate(a.getUsers)).Methods("GET")
	a.Router.HandleFunc("/user/{id:[0-9]+}", authenticate(a.getUser)).Methods("GET")
	a.Router.HandleFunc("/user", authenticate(a.createUser)).Methods("POST")
	a.Router.HandleFunc("/user/{id:[0-9]+}", authenticate(a.updateUser)).Methods("PUT")
	a.Router.HandleFunc("/user/{id:[0-9]+}", authenticate(a.deleteUser)).Methods("DELETE")
	a.Router.HandleFunc("/users/{id:[0-9]+}/notes/{date:\\d{8}}", authenticate(a.getUserNoteHandler)).Methods("GET")
	a.Router.HandleFunc("/users/{id:[0-9]+}/notes", authenticate(a.addUserNoteHandler)).Methods("POST")
	a.Router.HandleFunc("/notes/{date:\\d{8}}", authenticate(a.getNotes)).Methods("GET")
	a.Router.HandleFunc("/notes", authenticate(a.addNote)).Methods("POST")
	a.Router.HandleFunc("/activate", authenticate(a.activateLicense)).Methods("POST")
	a.Router.HandleFunc("/login", a.login).Methods("POST")
	a.Router.HandleFunc("/signin", a.signin).Methods("POST")
}

func (a *App) Run(addr string) {
	log.Printf("Conectando com banco de dados!")
	erro := a.DB.Ping()
	if erro != nil {
		log.Print(erro)
	}
	log.Printf("Iniciando serviço em: %s ", addr)
	log.Fatal(http.ListenAndServe(addr, a.Handler))
	defer a.DB.Close()
}
