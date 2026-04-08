// Package testutil provides test helpers including a mock Microsoft Graph API server.
package testutil

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"

	"github.com/co-native-ab/graphdo/internal/graph"
)

// SentMail records the details of a mail sent through the mock server.
type SentMail struct {
	To          string
	Subject     string
	Body        string
	ContentType string
}

// MockState holds the in-memory state for the mock Graph API server.
type MockState struct {
	mu        sync.Mutex
	User      graph.User
	TodoLists []graph.TodoList
	Todos     map[string][]graph.TodoItem
	SentMails []SentMail
	nextID    int
}

// NewMockState creates a new empty MockState.
func NewMockState() *MockState {
	return &MockState{
		Todos: make(map[string][]graph.TodoItem),
	}
}

// GetSentMails returns a snapshot of all mails sent through the mock server.
func (s *MockState) GetSentMails() []SentMail {
	s.mu.Lock()
	defer s.mu.Unlock()
	mails := make([]SentMail, len(s.SentMails))
	copy(mails, s.SentMails)
	return mails
}

// GetTodos returns a snapshot of the todos for the given list ID.
func (s *MockState) GetTodos(listID string) []graph.TodoItem {
	s.mu.Lock()
	defer s.mu.Unlock()
	items := s.Todos[listID]
	result := make([]graph.TodoItem, len(items))
	copy(result, items)
	return result
}

func (s *MockState) genID() string {
	s.nextID++
	return fmt.Sprintf("mock-%d", s.nextID)
}

// NewMockGraphServer creates an httptest.Server that simulates the Microsoft Graph API.
func NewMockGraphServer(state *MockState) *httptest.Server {
	mux := http.NewServeMux()

	checkAuth := func(w http.ResponseWriter, r *http.Request) bool {
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			http.Error(w, `{"error":{"code":"Unauthorized","message":"missing token"}}`, http.StatusUnauthorized)
			return false
		}
		return true
	}

	writeJSON := func(w http.ResponseWriter, status int, v any) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		if err := json.NewEncoder(w).Encode(v); err != nil {
			http.Error(w, `{"error":{"code":"InternalError","message":"encode failed"}}`, http.StatusInternalServerError)
		}
	}

	mux.HandleFunc("GET /me", func(w http.ResponseWriter, r *http.Request) {
		if !checkAuth(w, r) {
			return
		}
		state.mu.Lock()
		user := state.User
		state.mu.Unlock()
		writeJSON(w, http.StatusOK, user)
	})

	mux.HandleFunc("POST /me/sendMail", func(w http.ResponseWriter, r *http.Request) {
		if !checkAuth(w, r) {
			return
		}

		var req struct {
			Message struct {
				Subject string `json:"subject"`
				Body    struct {
					ContentType string `json:"contentType"`
					Content     string `json:"content"`
				} `json:"body"`
				ToRecipients []struct {
					EmailAddress struct {
						Address string `json:"address"`
					} `json:"emailAddress"`
				} `json:"toRecipients"`
			} `json:"message"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":{"code":"BadRequest","message":"invalid body"}}`, http.StatusBadRequest)
			return
		}

		var to string
		if len(req.Message.ToRecipients) > 0 {
			to = req.Message.ToRecipients[0].EmailAddress.Address
		}

		state.mu.Lock()
		state.SentMails = append(state.SentMails, SentMail{
			To:          to,
			Subject:     req.Message.Subject,
			Body:        req.Message.Body.Content,
			ContentType: req.Message.Body.ContentType,
		})
		state.mu.Unlock()

		w.WriteHeader(http.StatusAccepted)
	})

	mux.HandleFunc("GET /me/todo/lists", func(w http.ResponseWriter, r *http.Request) {
		if !checkAuth(w, r) {
			return
		}
		state.mu.Lock()
		lists := state.TodoLists
		state.mu.Unlock()

		if lists == nil {
			lists = []graph.TodoList{}
		}
		writeJSON(w, http.StatusOK, map[string]any{"value": lists})
	})

	mux.HandleFunc("GET /me/todo/lists/{listID}/tasks/{taskID}", func(w http.ResponseWriter, r *http.Request) {
		if !checkAuth(w, r) {
			return
		}
		listID := r.PathValue("listID")
		taskID := r.PathValue("taskID")

		state.mu.Lock()
		defer state.mu.Unlock()

		items := state.Todos[listID]
		for _, item := range items {
			if item.ID == taskID {
				writeJSON(w, http.StatusOK, item)
				return
			}
		}

		http.Error(w, `{"error":{"code":"NotFound","message":"task not found"}}`, http.StatusNotFound)
	})

	mux.HandleFunc("GET /me/todo/lists/{listID}/tasks", func(w http.ResponseWriter, r *http.Request) {
		if !checkAuth(w, r) {
			return
		}
		listID := r.PathValue("listID")

		state.mu.Lock()
		items, ok := state.Todos[listID]
		state.mu.Unlock()

		if !ok {
			http.Error(w, `{"error":{"code":"NotFound","message":"list not found"}}`, http.StatusNotFound)
			return
		}

		if items == nil {
			items = []graph.TodoItem{}
		}

		// Apply $skip and $top pagination
		skip := 0
		if s := r.URL.Query().Get("$skip"); s != "" {
			if v, err := strconv.Atoi(s); err == nil && v > 0 {
				skip = v
			}
		}
		if skip > len(items) {
			skip = len(items)
		}
		items = items[skip:]

		if t := r.URL.Query().Get("$top"); t != "" {
			if v, err := strconv.Atoi(t); err == nil && v > 0 && v < len(items) {
				items = items[:v]
			}
		}

		writeJSON(w, http.StatusOK, map[string]any{"value": items})
	})

	mux.HandleFunc("POST /me/todo/lists/{listID}/tasks", func(w http.ResponseWriter, r *http.Request) {
		if !checkAuth(w, r) {
			return
		}
		listID := r.PathValue("listID")

		var req struct {
			Title string          `json:"title"`
			Body  *graph.ItemBody `json:"body,omitempty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":{"code":"BadRequest","message":"invalid body"}}`, http.StatusBadRequest)
			return
		}

		state.mu.Lock()
		item := graph.TodoItem{
			ID:     state.genID(),
			Title:  req.Title,
			Status: "notStarted",
			Body:   req.Body,
		}
		state.Todos[listID] = append(state.Todos[listID], item)
		state.mu.Unlock()

		writeJSON(w, http.StatusCreated, item)
	})

	mux.HandleFunc("PATCH /me/todo/lists/{listID}/tasks/{taskID}", func(w http.ResponseWriter, r *http.Request) {
		if !checkAuth(w, r) {
			return
		}
		listID := r.PathValue("listID")
		taskID := r.PathValue("taskID")

		var req struct {
			Status string          `json:"status"`
			Title  string          `json:"title"`
			Body   *graph.ItemBody `json:"body"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":{"code":"BadRequest","message":"invalid body"}}`, http.StatusBadRequest)
			return
		}

		state.mu.Lock()
		defer state.mu.Unlock()

		items := state.Todos[listID]
		for i, item := range items {
			if item.ID != taskID {
				continue
			}
			if req.Status != "" {
				items[i].Status = req.Status
			}
			if req.Title != "" {
				items[i].Title = req.Title
			}
			if req.Body != nil {
				items[i].Body = req.Body
			}
			writeJSON(w, http.StatusOK, items[i])
			return
		}

		http.Error(w, `{"error":{"code":"NotFound","message":"task not found"}}`, http.StatusNotFound)
	})

	mux.HandleFunc("DELETE /me/todo/lists/{listID}/tasks/{taskID}", func(w http.ResponseWriter, r *http.Request) {
		if !checkAuth(w, r) {
			return
		}
		listID := r.PathValue("listID")
		taskID := r.PathValue("taskID")

		state.mu.Lock()
		defer state.mu.Unlock()

		items := state.Todos[listID]
		for i, item := range items {
			if item.ID == taskID {
				state.Todos[listID] = append(items[:i], items[i+1:]...)
				w.WriteHeader(http.StatusNoContent)
				return
			}
		}

		http.Error(w, `{"error":{"code":"NotFound","message":"task not found"}}`, http.StatusNotFound)
	})

	return httptest.NewServer(mux)
}
