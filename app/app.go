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

func (a *App) Initialize(user, password, dbname, dbhost string) {
	connectionString :=
		fmt.Sprintf("user=%s password=%s dbname=%s host=%s sslmode=disable", user, password, dbname, dbhost)

	var err error
	a.DB, err = sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal(err)
	}

	defer a.DB.Close()

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
	err := a.DB.Ping()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Database connection failed")
		return
	}
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

func (a *App) initializeRoutes() {
	a.Router.HandleFunc("/checkdb", a.checkDBConnection).Methods("GET")

	a.Router.HandleFunc("/users", authenticate(a.getUsers)).Methods("GET")
	a.Router.HandleFunc("/user/{id:[0-9]+}", authenticate(a.getUser)).Methods("GET")
	a.Router.HandleFunc("/user", authenticate(a.createUser)).Methods("POST")
	a.Router.HandleFunc("/user/{id:[0-9]+}", authenticate(a.updateUser)).Methods("PUT")
	a.Router.HandleFunc("/user/{id:[0-9]+}", authenticate(a.deleteUser)).Methods("DELETE")
	a.Router.HandleFunc("/login", a.login).Methods("POST")
}

func (a *App) Run(addr string) {
	log.Printf("Conectando com banco de dados!")
	log.Printf("Iniciando serviço em: %s ", addr)
	log.Fatal(http.ListenAndServe(addr, a.Handler))
}
