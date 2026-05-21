package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	ghapi "github.com/cli/go-gh/v2/pkg/api"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/UtakataKyosui/gh-review-post/internal/auth"
	"github.com/UtakataKyosui/gh-review-post/internal/cliexit"
	"github.com/UtakataKyosui/gh-review-post/internal/repo"
)

// reviewInput is the YAML/JSON schema for --file.
type reviewInput struct {
	Event    string         `yaml:"event" json:"event"`
	Body     string         `yaml:"body" json:"body"`
	Comments []reviewComment `yaml:"comments" json:"comments"`
}

type reviewComment struct {
	Path string `yaml:"path" json:"path"`
	Line int    `yaml:"line" json:"line"`
	Body string `yaml:"body" json:"body"`
}

// reviewRequest is the GitHub REST API payload.
type reviewRequest struct {
	Body     string           `json:"body,omitempty"`
	Event    string           `json:"event"`
	Comments []reviewCommentReq `json:"comments,omitempty"`
}

type reviewCommentReq struct {
	Path string `json:"path"`
	Line int    `json:"line"`
	Side string `json:"side"`
	Body string `json:"body"`
}

func newSubmitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit",
		Short: "Submit a PR review from a YAML/JSON file",
		Args:  cobra.NoArgs,
		RunE:  runSubmit,
	}
	cmd.Flags().StringP("file", "f", "", "Path to review YAML/JSON file (required)")
	_ = cmd.MarkFlagRequired("file")
	return cmd
}

func runSubmit(cmd *cobra.Command, _ []string) error {
	filePath, _ := cmd.Flags().GetString("file")
	prNumber, _ := cmd.Root().PersistentFlags().GetInt("pr")
	repoFlag, _ := cmd.Root().PersistentFlags().GetString("repo")
	dryRun, _ := cmd.Root().PersistentFlags().GetBool("dry-run")

	// Resolve repository.
	r, err := repo.Resolve(repoFlag)
	if err != nil {
		return err
	}

	// Read and parse input file.
	raw, err := os.ReadFile(filePath)
	if err != nil {
		return cliexit.NewValidation(cliexit.ErrCodeValidation,
			fmt.Errorf("cannot read file %q: %w", filePath, err), nil)
	}

	var input reviewInput
	lower := strings.ToLower(filePath)
	if strings.HasSuffix(lower, ".json") {
		if err := json.Unmarshal(raw, &input); err != nil {
			return cliexit.NewValidation(cliexit.ErrCodeValidation,
				fmt.Errorf("invalid JSON: %w", err), nil)
		}
	} else {
		if err := yaml.Unmarshal(raw, &input); err != nil {
			return cliexit.NewValidation(cliexit.ErrCodeValidation,
				fmt.Errorf("invalid YAML: %w", err), nil)
		}
	}

	if err := validateInput(&input); err != nil {
		return err
	}

	payload := buildPayload(&input)

	if dryRun {
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		_ = enc.Encode(payload)
		return nil
	}

	// Authenticate.
	token, err := auth.Token(r.Host)
	if err != nil {
		return err
	}

	client, err := ghapi.NewRESTClient(ghapi.ClientOptions{
		Host:      r.Host,
		AuthToken: token,
	})
	if err != nil {
		return cliexit.NewAPI(cliexit.ErrCodeAPI,
			fmt.Errorf("failed to create API client: %w", err))
	}

	apiPath := fmt.Sprintf("repos/%s/%s/pulls/%d/reviews", r.Owner, r.Name, prNumber)

	body, err := json.Marshal(payload)
	if err != nil {
		return cliexit.NewAPI(cliexit.ErrCodeAPI, fmt.Errorf("failed to encode request: %w", err))
	}

	var result map[string]any
	if err := client.Post(apiPath, bytes.NewReader(body), &result); err != nil {
		return cliexit.NewAPI(cliexit.ErrCodeAPI,
			fmt.Errorf("API request failed: %w", err))
	}

	reviewID, _ := result["id"].(float64)
	fmt.Fprintf(cmd.OutOrStdout(), "Review submitted: https://github.com/%s/%s/pull/%d#pullrequestreview-%d\n",
		r.Owner, r.Name, prNumber, int(reviewID))
	return nil
}

func validateInput(input *reviewInput) error {
	event := strings.ToUpper(input.Event)
	switch event {
	case "APPROVE", "REQUEST_CHANGES", "COMMENT", "":
	default:
		return cliexit.NewValidation(cliexit.ErrCodeValidation,
			fmt.Errorf("invalid event %q; must be APPROVE, REQUEST_CHANGES, or COMMENT", input.Event), nil)
	}
	for i, c := range input.Comments {
		if c.Path == "" {
			return cliexit.NewValidation(cliexit.ErrCodeValidation,
				fmt.Errorf("comment[%d]: path is required", i), nil)
		}
		if c.Body == "" {
			return cliexit.NewValidation(cliexit.ErrCodeValidation,
				fmt.Errorf("comment[%d]: body is required", i), nil)
		}
	}
	return nil
}

func buildPayload(input *reviewInput) reviewRequest {
	event := strings.ToUpper(input.Event)
	if event == "" {
		event = "COMMENT"
	}
	req := reviewRequest{
		Body:  input.Body,
		Event: event,
	}
	for _, c := range input.Comments {
		req.Comments = append(req.Comments, reviewCommentReq{
			Path: c.Path,
			Line: c.Line,
			Side: "RIGHT",
			Body: c.Body,
		})
	}
	return req
}
