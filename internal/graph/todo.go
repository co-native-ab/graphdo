package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// TodoList represents a Microsoft To Do task list.
type TodoList struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
}

// TodoItem represents a single task in a Microsoft To Do list.
type TodoItem struct {
	ID     string    `json:"id"`
	Title  string    `json:"title"`
	Status string    `json:"status"`
	Body   *ItemBody `json:"body,omitempty"`
}

// ItemBody holds the content and format of a todo item body.
type ItemBody struct {
	Content     string `json:"content"`
	ContentType string `json:"contentType"`
}

type graphListResponse[T any] struct {
	Value []T `json:"value"`
}

// ListTodoLists returns all of the user's Microsoft To Do task lists.
func (c *Client) ListTodoLists(ctx context.Context) ([]TodoList, error) {
	resp, err := c.do(ctx, http.MethodGet, "/me/todo/lists", nil)
	if err != nil {
		return nil, fmt.Errorf("list todo lists: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var result graphListResponse[TodoList]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("list todo lists: %w", err)
	}

	return result.Value, nil
}

// ListTodos returns tasks in the specified todo list with optional pagination.
// Pass top=0 and skip=0 to retrieve all tasks without pagination.
func (c *Client) ListTodos(ctx context.Context, listID string, top, skip int) ([]TodoItem, error) {
	if listID == "" {
		return nil, fmt.Errorf("list todos: list ID must not be empty")
	}

	path := fmt.Sprintf("/me/todo/lists/%s/tasks", listID)
	params := url.Values{}
	if top > 0 {
		params.Set("$top", strconv.Itoa(top))
	}
	if skip > 0 {
		params.Set("$skip", strconv.Itoa(skip))
	}
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	resp, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("list todos: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var result graphListResponse[TodoItem]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("list todos: %w", err)
	}

	return result.Value, nil
}

// GetTodo returns a single task from the specified todo list.
func (c *Client) GetTodo(ctx context.Context, listID, taskID string) (*TodoItem, error) {
	if listID == "" {
		return nil, fmt.Errorf("get todo: list ID must not be empty")
	}
	if taskID == "" {
		return nil, fmt.Errorf("get todo: task ID must not be empty")
	}

	path := fmt.Sprintf("/me/todo/lists/%s/tasks/%s", listID, taskID)
	resp, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("get todo: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var item TodoItem
	if err := json.NewDecoder(resp.Body).Decode(&item); err != nil {
		return nil, fmt.Errorf("get todo: %w", err)
	}

	return &item, nil
}

// UpdateTodo updates the title and/or body of a task in the specified todo list.
func (c *Client) UpdateTodo(ctx context.Context, listID, taskID, title, body string) (*TodoItem, error) {
	if listID == "" {
		return nil, fmt.Errorf("update todo: list ID must not be empty")
	}
	if taskID == "" {
		return nil, fmt.Errorf("update todo: task ID must not be empty")
	}

	type updateRequest struct {
		Title string    `json:"title,omitempty"`
		Body  *ItemBody `json:"body,omitempty"`
	}

	payload := updateRequest{}
	if title != "" {
		payload.Title = title
	}
	if body != "" {
		payload.Body = &ItemBody{
			Content:     body,
			ContentType: "text",
		}
	}

	path := fmt.Sprintf("/me/todo/lists/%s/tasks/%s", listID, taskID)
	resp, err := c.do(ctx, http.MethodPatch, path, payload)
	if err != nil {
		return nil, fmt.Errorf("update todo: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var item TodoItem
	if err := json.NewDecoder(resp.Body).Decode(&item); err != nil {
		return nil, fmt.Errorf("update todo: %w", err)
	}

	return &item, nil
}

// CreateTodo creates a new task in the specified todo list.
func (c *Client) CreateTodo(ctx context.Context, listID, title, body string) (*TodoItem, error) {
	if listID == "" {
		return nil, fmt.Errorf("create todo: list ID must not be empty")
	}
	if title == "" {
		return nil, fmt.Errorf("create todo: title must not be empty")
	}

	type createRequest struct {
		Title string    `json:"title"`
		Body  *ItemBody `json:"body,omitempty"`
	}

	payload := createRequest{Title: title}
	if body != "" {
		payload.Body = &ItemBody{
			Content:     body,
			ContentType: "text",
		}
	}

	path := fmt.Sprintf("/me/todo/lists/%s/tasks", listID)
	resp, err := c.do(ctx, http.MethodPost, path, payload)
	if err != nil {
		return nil, fmt.Errorf("create todo: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var item TodoItem
	if err := json.NewDecoder(resp.Body).Decode(&item); err != nil {
		return nil, fmt.Errorf("create todo: %w", err)
	}

	return &item, nil
}

// CompleteTodo marks a task as completed in the specified todo list.
func (c *Client) CompleteTodo(ctx context.Context, listID, taskID string) error {
	if listID == "" {
		return fmt.Errorf("complete todo: list ID must not be empty")
	}
	if taskID == "" {
		return fmt.Errorf("complete todo: task ID must not be empty")
	}

	payload := struct {
		Status string `json:"status"`
	}{Status: "completed"}

	path := fmt.Sprintf("/me/todo/lists/%s/tasks/%s", listID, taskID)
	resp, err := c.do(ctx, http.MethodPatch, path, payload)
	if err != nil {
		return fmt.Errorf("complete todo: %w", err)
	}
	_ = resp.Body.Close()

	return nil
}

// DeleteTodo removes a task from the specified todo list.
func (c *Client) DeleteTodo(ctx context.Context, listID, taskID string) error {
	if listID == "" {
		return fmt.Errorf("delete todo: list ID must not be empty")
	}
	if taskID == "" {
		return fmt.Errorf("delete todo: task ID must not be empty")
	}

	path := fmt.Sprintf("/me/todo/lists/%s/tasks/%s", listID, taskID)
	resp, err := c.do(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return fmt.Errorf("delete todo: %w", err)
	}
	_ = resp.Body.Close()

	return nil
}
