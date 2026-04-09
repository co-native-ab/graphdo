package graph_test

import (
	"context"
	"testing"

	"github.com/co-native-ab/graphdo/internal/graph"
	"github.com/co-native-ab/graphdo/internal/testutil"
)

func TestGetMe(t *testing.T) {
	state := testutil.NewMockState()
	state.User = graph.User{
		ID:                "user-1",
		DisplayName:       "Test User",
		Mail:              "test@example.com",
		UserPrincipalName: "test@example.com",
	}

	srv := testutil.NewMockGraphServer(state)
	t.Cleanup(srv.Close)

	client := graph.NewClient(srv.URL, graph.StaticToken("fake-token"))
	user, err := client.GetMe(context.Background())
	if err != nil {
		t.Fatalf("GetMe() error: %v", err)
	}

	if user.ID != "user-1" {
		t.Errorf("got ID %q, want %q", user.ID, "user-1")
	}
	if user.DisplayName != "Test User" {
		t.Errorf("got DisplayName %q, want %q", user.DisplayName, "Test User")
	}
	if user.Mail != "test@example.com" {
		t.Errorf("got Mail %q, want %q", user.Mail, "test@example.com")
	}
	if user.UserPrincipalName != "test@example.com" {
		t.Errorf("got UserPrincipalName %q, want %q", user.UserPrincipalName, "test@example.com")
	}
}

func TestSendMail(t *testing.T) {
	state := testutil.NewMockState()
	srv := testutil.NewMockGraphServer(state)
	t.Cleanup(srv.Close)

	client := graph.NewClient(srv.URL, graph.StaticToken("fake-token"))
	err := client.SendMail(context.Background(), "bob@example.com", "Hello", "Hi Bob", false)
	if err != nil {
		t.Fatalf("SendMail() error: %v", err)
	}

	mails := state.GetSentMails()
	if len(mails) != 1 {
		t.Fatalf("got %d sent mails, want 1", len(mails))
	}
	if mails[0].To != "bob@example.com" {
		t.Errorf("got To %q, want %q", mails[0].To, "bob@example.com")
	}
	if mails[0].Subject != "Hello" {
		t.Errorf("got Subject %q, want %q", mails[0].Subject, "Hello")
	}
	if mails[0].Body != "Hi Bob" {
		t.Errorf("got Body %q, want %q", mails[0].Body, "Hi Bob")
	}
	if mails[0].ContentType != "Text" {
		t.Errorf("got ContentType %q, want %q", mails[0].ContentType, "Text")
	}
}

func TestSendMailHTML(t *testing.T) {
	state := testutil.NewMockState()
	srv := testutil.NewMockGraphServer(state)
	t.Cleanup(srv.Close)

	client := graph.NewClient(srv.URL, graph.StaticToken("fake-token"))
	err := client.SendMail(context.Background(), "alice@example.com", "News", "<h1>Hello</h1>", true)
	if err != nil {
		t.Fatalf("SendMail() error: %v", err)
	}

	mails := state.GetSentMails()
	if len(mails) != 1 {
		t.Fatalf("got %d sent mails, want 1", len(mails))
	}
	if mails[0].ContentType != "HTML" {
		t.Errorf("got ContentType %q, want %q", mails[0].ContentType, "HTML")
	}
	if mails[0].Body != "<h1>Hello</h1>" {
		t.Errorf("got Body %q, want %q", mails[0].Body, "<h1>Hello</h1>")
	}
}
