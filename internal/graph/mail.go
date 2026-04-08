package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// User represents a Microsoft Graph user profile.
type User struct {
	ID                string `json:"id"`
	DisplayName       string `json:"displayName"`
	Mail              string `json:"mail"`
	UserPrincipalName string `json:"userPrincipalName"`
}

type sendMailRequest struct {
	Message sendMailMessage `json:"message"`
}

type sendMailMessage struct {
	Subject      string              `json:"subject"`
	Body         sendMailBody        `json:"body"`
	ToRecipients []sendMailRecipient `json:"toRecipients"`
}

type sendMailBody struct {
	ContentType string `json:"contentType"`
	Content     string `json:"content"`
}

type sendMailRecipient struct {
	EmailAddress sendMailAddress `json:"emailAddress"`
}

type sendMailAddress struct {
	Address string `json:"address"`
}

// GetMe returns the authenticated user's profile.
func (c *Client) GetMe(ctx context.Context) (*User, error) {
	resp, err := c.do(ctx, http.MethodGet, "/me", nil)
	if err != nil {
		return nil, fmt.Errorf("get me: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("get me: %w", err)
	}

	return &user, nil
}

// SendMail sends an email to the specified recipient via Microsoft Graph.
func (c *Client) SendMail(ctx context.Context, to, subject, body string, html bool) error {
	contentType := "Text"
	if html {
		contentType = "HTML"
	}

	payload := sendMailRequest{
		Message: sendMailMessage{
			Subject: subject,
			Body: sendMailBody{
				ContentType: contentType,
				Content:     body,
			},
			ToRecipients: []sendMailRecipient{
				{EmailAddress: sendMailAddress{Address: to}},
			},
		},
	}

	resp, err := c.do(ctx, http.MethodPost, "/me/sendMail", payload)
	if err != nil {
		return fmt.Errorf("send mail: %w", err)
	}
	_ = resp.Body.Close()

	return nil
}
