package sync

import (
	"fmt"
	"time"

	"github.com/SuzumiyaAoba/entry/internal/config"
	"github.com/go-resty/resty/v2"
	"gopkg.in/yaml.v3"
)

var GitHubAPIURL = "https://api.github.com"

type GistFile struct {
	Content string `json:"content"`
}

type Gist struct {
	Files       map[string]GistFile `json:"files"`
	Description string              `json:"description"`
	Public      bool                `json:"public"`
}

type Client struct {
	Token  string
	client *resty.Client
}

func NewClient(token string) *Client {
	client := resty.New().
		SetBaseURL(GitHubAPIURL).
		SetTimeout(10 * time.Second).
		SetHeader("Accept", "application/vnd.github.v3+json")

	if token != "" {
		client.SetHeader("Authorization", "token "+token)
	}

	return &Client{
		Token:  token,
		client: client,
	}
}

func (c *Client) GetGist(gistID string) (*config.Config, error) {
	var gist Gist
	resp, err := c.client.R().
		SetResult(&gist).
		Get("/gists/" + gistID)

	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, fmt.Errorf("failed to get gist: %s", resp.Status())
	}

	file, ok := gist.Files["config.yml"]
	if !ok {
		return nil, fmt.Errorf("config.yml not found in gist")
	}

	var cfg config.Config
	if err := yaml.Unmarshal([]byte(file.Content), &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config from gist: %w", err)
	}

	return &cfg, nil
}

func (c *Client) UpdateGist(gistID string, cfg *config.Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	gist := Gist{
		Files: map[string]GistFile{
			"config.yml": {Content: string(data)},
		},
		Description: "Entry Configuration (Updated " + time.Now().Format(time.RFC3339) + ")",
	}

	resp, err := c.client.R().
		SetBody(gist).
		Patch("/gists/" + gistID)

	if err != nil {
		return err
	}

	if resp.IsError() {
		return fmt.Errorf("failed to update gist: %s - %s", resp.Status(), resp.String())
	}

	return nil
}

func (c *Client) CreateGist(cfg *config.Config, public bool) (string, error) {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return "", err
	}

	gist := Gist{
		Files: map[string]GistFile{
			"config.yml": {Content: string(data)},
		},
		Description: "Entry Configuration",
		Public:      public,
	}

	var respData struct {
		ID string `json:"id"`
	}

	resp, err := c.client.R().
		SetBody(gist).
		SetResult(&respData).
		Post("/gists")

	if err != nil {
		return "", err
	}

	if resp.IsError() {
		return "", fmt.Errorf("failed to create gist: %s - %s", resp.Status(), resp.String())
	}

	return respData.ID, nil
}
