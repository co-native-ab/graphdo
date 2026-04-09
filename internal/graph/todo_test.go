package graph_test

import (
	"context"
	"testing"

	"github.com/co-native-ab/graphdo/internal/graph"
	"github.com/co-native-ab/graphdo/internal/testutil"
)

func TestListTodoLists(t *testing.T) {
	state := testutil.NewMockState()
	state.TodoLists = []graph.TodoList{
		{ID: "list-1", DisplayName: "Work"},
		{ID: "list-2", DisplayName: "Personal"},
	}

	srv := testutil.NewMockGraphServer(state)
	t.Cleanup(srv.Close)

	client := graph.NewClient(srv.URL, graph.StaticToken("fake-token"))
	lists, err := client.ListTodoLists(context.Background())
	if err != nil {
		t.Fatalf("ListTodoLists() error: %v", err)
	}

	if len(lists) != 2 {
		t.Fatalf("got %d lists, want 2", len(lists))
	}
	if lists[0].ID != "list-1" || lists[0].DisplayName != "Work" {
		t.Errorf("lists[0] = %+v, want {ID:list-1 DisplayName:Work}", lists[0])
	}
	if lists[1].ID != "list-2" || lists[1].DisplayName != "Personal" {
		t.Errorf("lists[1] = %+v, want {ID:list-2 DisplayName:Personal}", lists[1])
	}
}

func TestListTodos(t *testing.T) {
	state := testutil.NewMockState()
	state.Todos["list-1"] = []graph.TodoItem{
		{ID: "task-1", Title: "Buy milk", Status: "notStarted"},
		{ID: "task-2", Title: "Read book", Status: "inProgress"},
	}

	srv := testutil.NewMockGraphServer(state)
	t.Cleanup(srv.Close)

	client := graph.NewClient(srv.URL, graph.StaticToken("fake-token"))
	items, err := client.ListTodos(context.Background(), "list-1", 0, 0)
	if err != nil {
		t.Fatalf("ListTodos() error: %v", err)
	}

	if len(items) != 2 {
		t.Fatalf("got %d items, want 2", len(items))
	}
	if items[0].ID != "task-1" || items[0].Title != "Buy milk" {
		t.Errorf("items[0] = %+v, want {ID:task-1 Title:Buy milk ...}", items[0])
	}
	if items[1].ID != "task-2" || items[1].Title != "Read book" {
		t.Errorf("items[1] = %+v, want {ID:task-2 Title:Read book ...}", items[1])
	}
}

func TestCreateTodo(t *testing.T) {
	state := testutil.NewMockState()
	state.Todos["list-1"] = []graph.TodoItem{}

	srv := testutil.NewMockGraphServer(state)
	t.Cleanup(srv.Close)

	client := graph.NewClient(srv.URL, graph.StaticToken("fake-token"))
	item, err := client.CreateTodo(context.Background(), "list-1", "New task", "")
	if err != nil {
		t.Fatalf("CreateTodo() error: %v", err)
	}

	if item.Title != "New task" {
		t.Errorf("got Title %q, want %q", item.Title, "New task")
	}
	if item.Status != "notStarted" {
		t.Errorf("got Status %q, want %q", item.Status, "notStarted")
	}
	if item.ID == "" {
		t.Error("expected non-empty ID")
	}

	todos := state.GetTodos("list-1")
	if len(todos) != 1 {
		t.Fatalf("got %d todos in state, want 1", len(todos))
	}
	if todos[0].Title != "New task" {
		t.Errorf("state todo title = %q, want %q", todos[0].Title, "New task")
	}
}

func TestCreateTodoWithBody(t *testing.T) {
	state := testutil.NewMockState()
	state.Todos["list-1"] = []graph.TodoItem{}

	srv := testutil.NewMockGraphServer(state)
	t.Cleanup(srv.Close)

	client := graph.NewClient(srv.URL, graph.StaticToken("fake-token"))
	item, err := client.CreateTodo(context.Background(), "list-1", "Task with body", "Some details")
	if err != nil {
		t.Fatalf("CreateTodo() error: %v", err)
	}

	if item.Body == nil {
		t.Fatal("expected non-nil Body")
	}
	if item.Body.Content != "Some details" {
		t.Errorf("got Body.Content %q, want %q", item.Body.Content, "Some details")
	}
	if item.Body.ContentType != "text" {
		t.Errorf("got Body.ContentType %q, want %q", item.Body.ContentType, "text")
	}

	todos := state.GetTodos("list-1")
	if len(todos) != 1 {
		t.Fatalf("got %d todos in state, want 1", len(todos))
	}
	if todos[0].Body == nil || todos[0].Body.Content != "Some details" {
		t.Errorf("state todo body = %+v, want content %q", todos[0].Body, "Some details")
	}
}

func TestCompleteTodo(t *testing.T) {
	state := testutil.NewMockState()
	state.Todos["list-1"] = []graph.TodoItem{
		{ID: "task-1", Title: "Buy milk", Status: "notStarted"},
	}

	srv := testutil.NewMockGraphServer(state)
	t.Cleanup(srv.Close)

	client := graph.NewClient(srv.URL, graph.StaticToken("fake-token"))
	err := client.CompleteTodo(context.Background(), "list-1", "task-1")
	if err != nil {
		t.Fatalf("CompleteTodo() error: %v", err)
	}

	todos := state.GetTodos("list-1")
	if len(todos) != 1 {
		t.Fatalf("got %d todos in state, want 1", len(todos))
	}
	if todos[0].Status != "completed" {
		t.Errorf("got Status %q, want %q", todos[0].Status, "completed")
	}
}

func TestListTodosPagination(t *testing.T) {
	state := testutil.NewMockState()
	state.Todos["list-1"] = []graph.TodoItem{
		{ID: "t1", Title: "Task 1", Status: "notStarted"},
		{ID: "t2", Title: "Task 2", Status: "notStarted"},
		{ID: "t3", Title: "Task 3", Status: "notStarted"},
		{ID: "t4", Title: "Task 4", Status: "notStarted"},
		{ID: "t5", Title: "Task 5", Status: "notStarted"},
	}
	srv := testutil.NewMockGraphServer(state)
	defer srv.Close()
	client := graph.NewClient(srv.URL, graph.StaticToken("test-token"))
	ctx := context.Background()

	// Page 1: top=2, skip=0
	items, err := client.ListTodos(ctx, "list-1", 2, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].ID != "t1" || items[1].ID != "t2" {
		t.Errorf("got wrong items: %v", items)
	}

	// Page 2: top=2, skip=2
	items, err = client.ListTodos(ctx, "list-1", 2, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].ID != "t3" || items[1].ID != "t4" {
		t.Errorf("got wrong items: %v", items)
	}

	// Page 3: top=2, skip=4 — only 1 item left
	items, err = client.ListTodos(ctx, "list-1", 2, 4)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}

	// No pagination — get all
	items, err = client.ListTodos(ctx, "list-1", 0, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 5 {
		t.Fatalf("expected 5 items, got %d", len(items))
	}
}

func TestGetTodo(t *testing.T) {
	state := testutil.NewMockState()
	state.Todos["list-1"] = []graph.TodoItem{
		{ID: "task-1", Title: "Buy milk", Status: "notStarted", Body: &graph.ItemBody{Content: "whole milk", ContentType: "text"}},
	}
	srv := testutil.NewMockGraphServer(state)
	defer srv.Close()
	client := graph.NewClient(srv.URL, graph.StaticToken("test-token"))

	item, err := client.GetTodo(context.Background(), "list-1", "task-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.Title != "Buy milk" {
		t.Errorf("got title %q, want %q", item.Title, "Buy milk")
	}
	if item.Body == nil || item.Body.Content != "whole milk" {
		t.Errorf("got unexpected body: %+v", item.Body)
	}
}

func TestGetTodoNotFound(t *testing.T) {
	state := testutil.NewMockState()
	state.Todos["list-1"] = []graph.TodoItem{}
	srv := testutil.NewMockGraphServer(state)
	defer srv.Close()
	client := graph.NewClient(srv.URL, graph.StaticToken("test-token"))

	_, err := client.GetTodo(context.Background(), "list-1", "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent task")
	}
}

func TestUpdateTodo(t *testing.T) {
	state := testutil.NewMockState()
	state.Todos["list-1"] = []graph.TodoItem{
		{ID: "task-1", Title: "Buy milk", Status: "notStarted"},
	}
	srv := testutil.NewMockGraphServer(state)
	defer srv.Close()
	client := graph.NewClient(srv.URL, graph.StaticToken("test-token"))

	item, err := client.UpdateTodo(context.Background(), "list-1", "task-1", "Buy oat milk", "from the store")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if item.Title != "Buy oat milk" {
		t.Errorf("got title %q, want %q", item.Title, "Buy oat milk")
	}
	if item.Body == nil || item.Body.Content != "from the store" {
		t.Errorf("got unexpected body: %+v", item.Body)
	}

	// Verify state was updated
	items := state.GetTodos("list-1")
	if items[0].Title != "Buy oat milk" {
		t.Errorf("state not updated: got title %q", items[0].Title)
	}
}

func TestDeleteTodo(t *testing.T) {
	state := testutil.NewMockState()
	state.Todos["list-1"] = []graph.TodoItem{
		{ID: "task-1", Title: "Buy milk", Status: "notStarted"},
		{ID: "task-2", Title: "Read book", Status: "notStarted"},
	}

	srv := testutil.NewMockGraphServer(state)
	t.Cleanup(srv.Close)

	client := graph.NewClient(srv.URL, graph.StaticToken("fake-token"))
	err := client.DeleteTodo(context.Background(), "list-1", "task-1")
	if err != nil {
		t.Fatalf("DeleteTodo() error: %v", err)
	}

	todos := state.GetTodos("list-1")
	if len(todos) != 1 {
		t.Fatalf("got %d todos in state, want 1", len(todos))
	}
	if todos[0].ID != "task-2" {
		t.Errorf("remaining todo ID = %q, want %q", todos[0].ID, "task-2")
	}
}
