package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// TodoCmd is the argument type for the todo subcommand.
type TodoCmd struct {
	List     *TodoListCmd     `arg:"subcommand:list" help:"list todos"`
	Show     *TodoShowCmd     `arg:"subcommand:show" help:"show a single todo"`
	Create   *TodoCreateCmd   `arg:"subcommand:create" help:"create a todo"`
	Update   *TodoUpdateCmd   `arg:"subcommand:update" help:"update a todo"`
	Complete *TodoCompleteCmd `arg:"subcommand:complete" help:"mark a todo as completed"`
	Delete   *TodoDeleteCmd   `arg:"subcommand:delete" help:"delete a todo"`
}

// TodoListCmd is the argument type for the todo list subcommand.
type TodoListCmd struct {
	Top  int `arg:"--top" help:"maximum number of items to return" default:"20"`
	Skip int `arg:"--skip" help:"number of items to skip (for pagination)" default:"0"`
}

// TodoShowCmd is the argument type for the todo show subcommand.
type TodoShowCmd struct {
	ID string `arg:"--id,required" help:"task ID"`
}

// TodoUpdateCmd is the argument type for the todo update subcommand.
type TodoUpdateCmd struct {
	ID    string `arg:"--id,required" help:"task ID"`
	Title string `arg:"--title" help:"new task title"`
	Body  string `arg:"--body" help:"new task body"`
}

// TodoCreateCmd is the argument type for the todo create subcommand.
type TodoCreateCmd struct {
	Title string `arg:"--title,required" help:"task title"`
	Body  string `arg:"--body" help:"task body"`
}

// TodoCompleteCmd is the argument type for the todo complete subcommand.
type TodoCompleteCmd struct {
	ID string `arg:"--id,required" help:"task ID"`
}

// TodoDeleteCmd is the argument type for the todo delete subcommand.
type TodoDeleteCmd struct {
	ID string `arg:"--id,required" help:"task ID"`
}

// --- MCP tool input types ---

type mcpTodoListInput struct {
	Top  int `json:"top,omitempty" jsonschema:"maximum number of items to return (default 20)"`
	Skip int `json:"skip,omitempty" jsonschema:"number of items to skip for pagination (default 0)"`
}

type mcpTodoIDInput struct {
	ID string `json:"id" jsonschema:"the task ID"`
}

type mcpTodoCreateInput struct {
	Title string `json:"title" jsonschema:"the task title"`
	Body  string `json:"body,omitempty" jsonschema:"the task body (optional)"`
}

type mcpTodoUpdateInput struct {
	ID    string `json:"id" jsonschema:"the task ID"`
	Title string `json:"title,omitempty" jsonschema:"new task title (leave empty to keep unchanged)"`
	Body  string `json:"body,omitempty" jsonschema:"new task body (leave empty to keep unchanged)"`
}

// --- CLI handlers ---

func runTodo(ctx context.Context, cmd *TodoCmd, deps *Dependencies) error {
	switch {
	case cmd.List != nil:
		return runTodoList(ctx, cmd.List, deps)
	case cmd.Show != nil:
		return runTodoShow(ctx, cmd.Show, deps)
	case cmd.Create != nil:
		return runTodoCreate(ctx, cmd.Create, deps)
	case cmd.Update != nil:
		return runTodoUpdate(ctx, cmd.Update, deps)
	case cmd.Complete != nil:
		return runTodoComplete(ctx, cmd.Complete, deps)
	case cmd.Delete != nil:
		return runTodoDelete(ctx, cmd.Delete, deps)
	default:
		return fmt.Errorf("missing subcommand — run 'graphdo todo --help' for usage")
	}
}

func runTodoList(ctx context.Context, cmd *TodoListCmd, deps *Dependencies) error {
	cfg, err := loadConfig(deps.ConfigDir)
	if err != nil {
		return err
	}

	items, err := deps.GraphClient.ListTodos(ctx, cfg.TodoListID, cmd.Top, cmd.Skip)
	if err != nil {
		return fmt.Errorf("listing todos: %w", err)
	}

	enc := json.NewEncoder(deps.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(items); err != nil {
		return fmt.Errorf("encoding todos: %w", err)
	}

	return nil
}

func runTodoShow(ctx context.Context, cmd *TodoShowCmd, deps *Dependencies) error {
	cfg, err := loadConfig(deps.ConfigDir)
	if err != nil {
		return err
	}

	slog.Debug("getting todo", "id", cmd.ID, "list", cfg.TodoListID)

	item, err := deps.GraphClient.GetTodo(ctx, cfg.TodoListID, cmd.ID)
	if err != nil {
		return fmt.Errorf("getting todo: %w", err)
	}

	enc := json.NewEncoder(deps.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(item); err != nil {
		return fmt.Errorf("encoding todo: %w", err)
	}

	return nil
}

func runTodoUpdate(ctx context.Context, cmd *TodoUpdateCmd, deps *Dependencies) error {
	if cmd.Title == "" && cmd.Body == "" {
		return fmt.Errorf("at least one of --title or --body must be provided")
	}

	cfg, err := loadConfig(deps.ConfigDir)
	if err != nil {
		return err
	}

	slog.Debug("updating todo", "id", cmd.ID, "list", cfg.TodoListID)

	item, err := deps.GraphClient.UpdateTodo(ctx, cfg.TodoListID, cmd.ID, cmd.Title, cmd.Body)
	if err != nil {
		return fmt.Errorf("updating todo: %w", err)
	}

	enc := json.NewEncoder(deps.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(item); err != nil {
		return fmt.Errorf("encoding todo: %w", err)
	}

	_, _ = fmt.Fprintln(deps.Stderr, "✓ Task updated")
	return nil
}

func runTodoCreate(ctx context.Context, cmd *TodoCreateCmd, deps *Dependencies) error {
	cfg, err := loadConfig(deps.ConfigDir)
	if err != nil {
		return err
	}

	slog.Debug("creating todo", "title", cmd.Title, "list", cfg.TodoListID)

	item, err := deps.GraphClient.CreateTodo(ctx, cfg.TodoListID, cmd.Title, cmd.Body)
	if err != nil {
		return fmt.Errorf("creating todo: %w", err)
	}

	enc := json.NewEncoder(deps.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(item); err != nil {
		return fmt.Errorf("encoding todo: %w", err)
	}

	return nil
}

func runTodoComplete(ctx context.Context, cmd *TodoCompleteCmd, deps *Dependencies) error {
	cfg, err := loadConfig(deps.ConfigDir)
	if err != nil {
		return err
	}

	slog.Debug("completing todo", "id", cmd.ID, "list", cfg.TodoListID)

	if err := deps.GraphClient.CompleteTodo(ctx, cfg.TodoListID, cmd.ID); err != nil {
		return fmt.Errorf("completing todo: %w", err)
	}

	_, _ = fmt.Fprintln(deps.Stderr, "✓ Task marked as completed")
	return nil
}

func runTodoDelete(ctx context.Context, cmd *TodoDeleteCmd, deps *Dependencies) error {
	cfg, err := loadConfig(deps.ConfigDir)
	if err != nil {
		return err
	}

	slog.Debug("deleting todo", "id", cmd.ID, "list", cfg.TodoListID)

	if err := deps.GraphClient.DeleteTodo(ctx, cfg.TodoListID, cmd.ID); err != nil {
		return fmt.Errorf("deleting todo: %w", err)
	}

	_, _ = fmt.Fprintln(deps.Stderr, "✓ Task deleted")
	return nil
}

// --- MCP tool registration ---

func registerTodoMCPTools(s *mcp.Server, deps *Dependencies) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "todo_list",
		Description: "List todos from your Microsoft To Do list",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input mcpTodoListInput) (*mcp.CallToolResult, any, error) {
		cfg, err := loadConfig(deps.ConfigDir)
		if err != nil {
			return mcpErrResult(err), nil, nil
		}
		top := input.Top
		if top <= 0 {
			top = 20
		}
		items, err := deps.GraphClient.ListTodos(ctx, cfg.TodoListID, top, input.Skip)
		if err != nil {
			return mcpErrResult(fmt.Errorf("listing todos: %w", err)), nil, nil
		}
		return mcpJSONResult(items), nil, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "todo_show",
		Description: "Get a single todo by ID from your Microsoft To Do list",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input mcpTodoIDInput) (*mcp.CallToolResult, any, error) {
		cfg, err := loadConfig(deps.ConfigDir)
		if err != nil {
			return mcpErrResult(err), nil, nil
		}
		item, err := deps.GraphClient.GetTodo(ctx, cfg.TodoListID, input.ID)
		if err != nil {
			return mcpErrResult(fmt.Errorf("getting todo: %w", err)), nil, nil
		}
		return mcpJSONResult(item), nil, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "todo_create",
		Description: "Create a new todo in your Microsoft To Do list",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input mcpTodoCreateInput) (*mcp.CallToolResult, any, error) {
		cfg, err := loadConfig(deps.ConfigDir)
		if err != nil {
			return mcpErrResult(err), nil, nil
		}
		item, err := deps.GraphClient.CreateTodo(ctx, cfg.TodoListID, input.Title, input.Body)
		if err != nil {
			return mcpErrResult(fmt.Errorf("creating todo: %w", err)), nil, nil
		}
		return mcpJSONResult(item), nil, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "todo_update",
		Description: "Update a todo's title or body in your Microsoft To Do list",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input mcpTodoUpdateInput) (*mcp.CallToolResult, any, error) {
		if input.Title == "" && input.Body == "" {
			return mcpErrResult(fmt.Errorf("at least one of title or body must be provided")), nil, nil
		}
		cfg, err := loadConfig(deps.ConfigDir)
		if err != nil {
			return mcpErrResult(err), nil, nil
		}
		item, err := deps.GraphClient.UpdateTodo(ctx, cfg.TodoListID, input.ID, input.Title, input.Body)
		if err != nil {
			return mcpErrResult(fmt.Errorf("updating todo: %w", err)), nil, nil
		}
		return mcpJSONResult(item), nil, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "todo_complete",
		Description: "Mark a todo as completed in your Microsoft To Do list",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input mcpTodoIDInput) (*mcp.CallToolResult, any, error) {
		cfg, err := loadConfig(deps.ConfigDir)
		if err != nil {
			return mcpErrResult(err), nil, nil
		}
		if err := deps.GraphClient.CompleteTodo(ctx, cfg.TodoListID, input.ID); err != nil {
			return mcpErrResult(fmt.Errorf("completing todo: %w", err)), nil, nil
		}
		return mcpTextResult("✓ Task marked as completed"), nil, nil
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "todo_delete",
		Description: "Delete a todo from your Microsoft To Do list",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input mcpTodoIDInput) (*mcp.CallToolResult, any, error) {
		cfg, err := loadConfig(deps.ConfigDir)
		if err != nil {
			return mcpErrResult(err), nil, nil
		}
		if err := deps.GraphClient.DeleteTodo(ctx, cfg.TodoListID, input.ID); err != nil {
			return mcpErrResult(fmt.Errorf("deleting todo: %w", err)), nil, nil
		}
		return mcpTextResult("✓ Task deleted"), nil, nil
	})
}
