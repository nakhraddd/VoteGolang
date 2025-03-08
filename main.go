package main

import (
	"encoding/json"
	_ "fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

type DeputyVote struct {
	Username   string `json:"username"`
	DeputyName string `json:"deputy_name"`
	Vote       bool   `json:"vote"`
}

type SessionDeputyVote struct {
	Username      string `json:"username"`
	SessionDeputy string `json:"session_deputy"`
	Vote          bool   `json:"vote"`
}

type PresidentVote struct {
	Username  string `json:"username"`
	President string `json:"president"`
	Vote      bool   `json:"vote"`
}

type PetitionVote struct {
	Username string `json:"username"`
	Petition string `json:"petition"`
	Vote     bool   `json:"vote"`
}

type VoteStatusResponse struct {
	HasVoted       bool `json:"hasVoted"`
	RemainingVotes int  `json:"remainingVotes"`
}

type News struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	Title     string `json:"title"`
	Paragraph string `json:"paragraph"`
	Photo     string `json:"photo"`
}

type Petition struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type President struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	Name      string `json:"name"`
	Education string `json:"education"`
}

type SessionDeputy struct {
	ID     uint   `json:"id" gorm:"primaryKey"`
	Name   string `json:"name"`
	Region string `json:"region"`
}

type Deputy struct {
	ID     uint   `json:"id" gorm:"primaryKey"`
	Name   string `json:"name"`
	Region string `json:"region"`
}
type User struct {
	ID       uint   `json:"id" gorm:"primaryKey"`
	Username string `json:"username"`
	FullName string `json:"full_name"`
}

var router = mux.NewRouter()

func submitDeputyVote(w http.ResponseWriter, r *http.Request) {
	var vote DeputyVote
	log.Printf("Received POST request for /deputy_vote from %s", r.RemoteAddr)

	err := json.NewDecoder(r.Body).Decode(&vote)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	log.Printf("Vote for deputy '%s' by user '%s' received", vote.DeputyName, vote.Username)

	w.WriteHeader(http.StatusCreated)
	log.Printf("Response: Vote for deputy '%s' successfully created", vote.DeputyName)

	json.NewEncoder(w).Encode(vote)
}

func getDeputyVotes(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received GET request for /deputy_vote from %s", r.RemoteAddr)

	var votes []DeputyVote
	votes = append(votes, DeputyVote{Username: "user1", DeputyName: "Deputy1", Vote: true})

	w.Header().Set("Content-Type", "application/json")
	log.Printf("Response: Fetched %d votes", len(votes))

	json.NewEncoder(w).Encode(votes)
}

func getDeputyVoteStatus(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received GET request for /deputy_vote/status from %s", r.RemoteAddr)

	voteStatus := VoteStatusResponse{
		HasVoted:       false,
		RemainingVotes: 1,
	}

	w.Header().Set("Content-Type", "application/json")
	log.Printf("Response: Vote status: HasVoted=%v, RemainingVotes=%d", voteStatus.HasVoted, voteStatus.RemainingVotes)

	json.NewEncoder(w).Encode(voteStatus)
}

func submitSessionDeputyVote(w http.ResponseWriter, r *http.Request) {
	var vote SessionDeputyVote
	log.Printf("Received POST request for from %s", r.RemoteAddr)

	err := json.NewDecoder(r.Body).Decode(&vote)
	if err != nil {
		log.Printf("Error decoding request body: %v", err)

		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(vote)
}

func getSessionDeputyVotes(w http.ResponseWriter, r *http.Request) {
	var votes []SessionDeputyVote
	log.Printf("Received GET request for from %s", r.RemoteAddr)
	votes = append(votes, SessionDeputyVote{Username: "user1", SessionDeputy: "SessionDeputy1", Vote: true})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(votes)
}

func getSessionDeputyVoteStatus(w http.ResponseWriter, r *http.Request) {

	voteStatus := VoteStatusResponse{
		HasVoted:       false,
		RemainingVotes: 1,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(voteStatus)
}

func submitPresidentVote(w http.ResponseWriter, r *http.Request) {
	var vote PresidentVote
	log.Printf("Received POST request for /president_vote from %s", r.RemoteAddr)

	err := json.NewDecoder(r.Body).Decode(&vote)
	if err != nil {
		log.Printf("Error decoding request body: %v", err)

		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("Vote for president '%s' by user '%s' received", vote.President, vote.Username)

	w.WriteHeader(http.StatusCreated)
	log.Printf("Response: Vote for president '%s' successfully created", vote.President)

	json.NewEncoder(w).Encode(vote)
}

func getPresidentVotes(w http.ResponseWriter, r *http.Request) {
	var votes []PresidentVote
	log.Printf("Received GET request for /president_vote/ from %s", r.RemoteAddr)
	votes = append(votes, PresidentVote{Username: "user1", President: "President1", Vote: true})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(votes)
}

func getPresidentVoteStatus(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received GET request for /president_vote/status from %s", r.RemoteAddr)

	voteStatus := VoteStatusResponse{
		HasVoted:       false,
		RemainingVotes: 1,
	}

	w.Header().Set("Content-Type", "application/json")
	log.Printf("Response: Vote status: HasVoted=%v, RemainingVotes=%d", voteStatus.HasVoted, voteStatus.RemainingVotes)

	json.NewEncoder(w).Encode(voteStatus)
}

func submitPetitionVote(w http.ResponseWriter, r *http.Request) {
	var vote PetitionVote
	err := json.NewDecoder(r.Body).Decode(&vote)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(vote)
}

func getPetitionVotes(w http.ResponseWriter, r *http.Request) {
	var votes []PetitionVote
	votes = append(votes, PetitionVote{Username: "user1", Petition: "Petition1", Vote: true})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(votes)
}

func getPetitionVoteStatus(w http.ResponseWriter, r *http.Request) {

	voteStatus := VoteStatusResponse{
		HasVoted:       false,
		RemainingVotes: 1,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(voteStatus)
}

func getNews(w http.ResponseWriter, r *http.Request) {
	var news []News
	log.Printf("Received GET request for /vote/general_news from %s", r.RemoteAddr)
	news = append(news, News{ID: 1, Title: "Election News", Paragraph: "The election process has started.", Photo: "url_to_photo"})

	w.Header().Set("Content-Type", "application/json")
	log.Printf("Response: Fetched %d news items", len(news))

	json.NewEncoder(w).Encode(news)
}

func getPetition(w http.ResponseWriter, r *http.Request) {
	var petitions []Petition
	log.Printf("Received GET request for /vote/petition from %s", r.RemoteAddr)
	petitions = append(petitions, Petition{ID: 1, Title: "Petition1", Description: "Petition description"})

	w.Header().Set("Content-Type", "application/json")
	log.Printf("Response: Fetched %d petitions", len(petitions))

	json.NewEncoder(w).Encode(petitions)
}

func getPresident(w http.ResponseWriter, r *http.Request) {
	var presidents []President
	log.Printf("Received GET request for /vote/president from %s", r.RemoteAddr)
	presidents = append(presidents, President{ID: 1, Name: "President1", Education: "PhD"})

	w.Header().Set("Content-Type", "application/json")
	log.Printf("Response: Fetched %d presidents", len(presidents))

	json.NewEncoder(w).Encode(presidents)
}

func getSessionDeputy(w http.ResponseWriter, r *http.Request) {
	var sessionDeputies []SessionDeputy
	sessionDeputies = append(sessionDeputies, SessionDeputy{ID: 1, Name: "SessionDeputy1", Region: "Region1"})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sessionDeputies)
}

func getDeputy(w http.ResponseWriter, r *http.Request) {
	var deputies []Deputy
	deputies = append(deputies, Deputy{ID: 1, Name: "SessionDeputy1", Region: "Region1"})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(deputies)
}

func getAllUsers(w http.ResponseWriter, r *http.Request) {
	var users []User
	users = append(users, User{ID: 1, Username: "user1", FullName: "User One"})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func createUser(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func initServer() {
	router.HandleFunc("/deputy_vote", submitDeputyVote).Methods("POST")
	router.HandleFunc("/deputy_vote", getDeputyVotes).Methods("GET")
	router.HandleFunc("/deputy_vote/status", getDeputyVoteStatus).Methods("GET")

	router.HandleFunc("/session_deputy_vote", submitSessionDeputyVote).Methods("POST")
	router.HandleFunc("/session_deputy_vote", getSessionDeputyVotes).Methods("GET")
	router.HandleFunc("/session_deputy_vote/status", getSessionDeputyVoteStatus).Methods("GET")

	router.HandleFunc("/president_vote", submitPresidentVote).Methods("POST")
	router.HandleFunc("/president_vote", getPresidentVotes).Methods("GET")
	router.HandleFunc("/president_vote/status", getPresidentVoteStatus).Methods("GET")

	router.HandleFunc("/petition_vote", submitPetitionVote).Methods("POST")
	router.HandleFunc("/petition_vote", getPetitionVotes).Methods("GET")
	router.HandleFunc("/petition_vote/status", getPetitionVoteStatus).Methods("GET")

	router.HandleFunc("/vote/general_news", getNews).Methods("GET")
	router.HandleFunc("/vote/petition", getPetition).Methods("GET")
	router.HandleFunc("/vote/president", getPresident).Methods("GET")
	router.HandleFunc("/vote/session_deputy", getSessionDeputy).Methods("GET")
	router.HandleFunc("/vote/deputy", getDeputy).Methods("GET")

	router.HandleFunc("/votes_count/users_data", getAllUsers).Methods("GET")
	router.HandleFunc("/votes_count/users_data", createUser).Methods("POST")

	log.Fatal(http.ListenAndServe(":8080", router))
}

func main() {
	initServer()
}
