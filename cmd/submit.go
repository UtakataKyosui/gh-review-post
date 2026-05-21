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
	Path       string `yaml:"path" json:"path"`
	Line       int    `yaml:"line" json:"line"`
	Body       string `yaml:"body" json:"body"`
	Side       string `yaml:"side" json:"side"`
	Suggestion string `yaml:"suggestion" json:"suggestion"`
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
	cmd.Flags().String("format", "", "Force input format: yaml or json (default: auto-detect from file extension)")
	return cmd
}

func runSubmit(cmd *cobra.Command, _ []string) error {
	filePath, _ := cmd.Flags().GetString("file")
	formatFlag, _ := cmd.Flags().GetString("format")
	prNumber, _ := cmd.Root().PersistentFlags().GetInt("pr")
	repoFlag, _ := cmd.Root().PersistentFlags().GetString("repo")
	dryRun, _ := cmd.Root().PersistentFlags().GetBool("dry-run")

	// Validate --format flag value.
	formatFlag = strings.ToLower(formatFlag)
	switch formatFlag {
	case "", "yaml", "json":
		// valid
	default:
		return cliexit.NewValidation(cliexit.ErrCodeValidation,
			fmt.Errorf("invalid --format %q; must be yaml or json", formatFlag), nil)
	}

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
	useJSON := formatFlag == "json" || (formatFlag == "" && strings.HasSuffix(strings.ToLower(filePath), ".json"))
	if useJSON {
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
		if err := enc.Encode(payload); err != nil {
			return cliexit.NewAPI(cliexit.ErrCodeAPI, fmt.Errorf("failed to encode dry-run payload: %w", err))
		}
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

	if strings.ToUpper(input.Event) == "APPROVE" {
		if err := checkSelfApprove(client, r.Host, r.Owner, r.Name, prNumber); err != nil {
			return err
		}
	}

	apiPath := fmt.Sprintf("repos/%s/%s/pulls/%d/reviews", r.Owner, r.Name, prNumber)

	body, err := json.Marshal(payload)
	if err != nil {
		return cliexit.NewAPI(cliexit.ErrCodeAPI, fmt.Errorf("failed to encode request: %w", err))
	}

	var result struct {
		ID int64 `json:"id"`
	}
	if err := client.Post(apiPath, bytes.NewReader(body), &result); err != nil {
		return cliexit.NewAPI(cliexit.ErrCodeAPI,
			fmt.Errorf("API request failed: %w", err))
	}

	reviewURL := fmt.Sprintf("https://%s/%s/%s/pull/%d#pullrequestreview-%d",
		r.Host, r.Owner, r.Name, prNumber, result.ID)

	asJSON, _ := cmd.Root().PersistentFlags().GetBool("json")
	if asJSON {
		return json.NewEncoder(cmd.OutOrStdout()).Encode(map[string]any{
			"id":  result.ID,
			"url": reviewURL,
		})
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Review submitted: %s\n", reviewURL)
	return nil
}

// commentViolation represents a single suggestion-format violation in an inline comment.
type commentViolation struct {
	CommentIndex int    `json:"comment_index"`
	Path         string `json:"path"`
	Line         int    `json:"line"`
	Reason       string `json:"reason"`
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
		if c.Line <= 0 {
			return cliexit.NewValidation(cliexit.ErrCodeValidation,
				fmt.Errorf("comment[%d]: line must be a positive integer", i), nil)
		}
		if c.Body == "" && c.Suggestion == "" {
			return cliexit.NewValidation(cliexit.ErrCodeValidation,
				fmt.Errorf("comment[%d]: body or suggestion is required", i), nil)
		}
		side := strings.ToUpper(c.Side)
		if side != "" && side != "LEFT" && side != "RIGHT" {
			return cliexit.NewValidation(cliexit.ErrCodeValidation,
				fmt.Errorf("comment[%d]: side must be LEFT or RIGHT", i), nil)
		}
	}

	bodyLines := checkBodySuggestion(input.Body)
	commentViolations := checkCommentSuggestions(input.Comments)

	if len(bodyLines) == 0 && len(commentViolations) == 0 {
		return nil
	}

	return buildValidationError(bodyLines, commentViolations)
}

// checkBodySuggestion returns 1-origin line numbers where suggestion blocks appear in the body.
// Any line starting with "```suggestion" (tag or not) is a violation in review body.
func checkBodySuggestion(body string) []int {
	var lines []int
	for i, line := range strings.Split(body, "\n") {
		if strings.HasPrefix(line, "```suggestion") {
			lines = append(lines, i+1)
		}
	}
	return lines
}

// checkCommentSuggestions validates suggestion block format across all comments.
func checkCommentSuggestions(comments []reviewComment) []commentViolation {
	var violations []commentViolation
	for i, c := range comments {
		violations = append(violations, checkSuggestionFences(i, c)...)
	}
	return violations
}

// checkSuggestionFences checks that suggestion fences in a single comment body are valid:
// no language tag, and each open fence has a matching close fence.
func checkSuggestionFences(idx int, c reviewComment) []commentViolation {
	var violations []commentViolation
	openCount := 0
	for _, line := range strings.Split(c.Body, "\n") {
		trimmed := strings.TrimRight(line, " \t\r")
		if strings.HasPrefix(trimmed, "```suggestion") {
			rest := trimmed[len("```suggestion"):]
			if strings.TrimSpace(rest) == "" {
				openCount++ // valid plain fence; tracked for missing_close check below
			} else {
				violations = append(violations, commentViolation{
					CommentIndex: idx,
					Path:         c.Path,
					Line:         c.Line,
					Reason:       "language_tag",
				})
			}
		} else if trimmed == "```" && openCount > 0 {
			openCount--
		}
	}
	if openCount > 0 {
		violations = append(violations, commentViolation{
			CommentIndex: idx,
			Path:         c.Path,
			Line:         c.Line,
			Reason:       "missing_close",
		})
	}
	return violations
}

func cvToMap(v commentViolation) map[string]any {
	return map[string]any{
		"comment_index": v.CommentIndex,
		"path":          v.Path,
		"line":          v.Line,
		"reason":        v.Reason,
	}
}

func buildValidationError(bodyLines []int, cvs []commentViolation) error {
	hasBV := len(bodyLines) > 0
	hasCV := len(cvs) > 0

	if hasBV && !hasCV {
		return cliexit.NewValidation(
			cliexit.ErrCodeBodyHasSuggestion,
			fmt.Errorf("review body contains %d suggestion block(s)", len(bodyLines)),
			map[string]any{"location": "body", "lines": bodyLines},
		)
	}

	if !hasBV && hasCV {
		vs := make([]map[string]any, len(cvs))
		for i, v := range cvs {
			vs[i] = cvToMap(v)
		}
		return cliexit.NewValidation(
			cliexit.ErrCodeSuggestionFormat,
			fmt.Errorf("suggestion format error in %d violation(s)", len(cvs)),
			map[string]any{"violations": vs},
		)
	}

	// Both body and comment violations: aggregate under VALIDATION_FAILED.
	total := 1 + len(cvs) // body counts as one aggregate entry
	violations := make([]any, 0, total)
	violations = append(violations, map[string]any{"location": "body", "lines": bodyLines})
	for _, v := range cvs {
		violations = append(violations, cvToMap(v))
	}
	return cliexit.NewValidation(
		cliexit.ErrCodeValidation,
		fmt.Errorf("validation failed: %d issue(s)", total),
		map[string]any{"violations": violations},
	)
}

// checkSelfApprove returns VALIDATION_SELF_APPROVE if the authenticated user is the PR author.
// API errors are ignored so that a transient failure doesn't block submission.
func checkSelfApprove(client *ghapi.RESTClient, host, owner, name string, prNumber int) error {
	var currentUser struct {
		Login string `json:"login"`
	}
	if err := client.Get("user", &currentUser); err != nil {
		return nil
	}

	var pr struct {
		User struct {
			Login string `json:"login"`
		} `json:"user"`
	}
	if err := client.Get(fmt.Sprintf("repos/%s/%s/pulls/%d", owner, name, prNumber), &pr); err != nil {
		return nil
	}

	if currentUser.Login == "" || pr.User.Login == "" {
		return nil
	}
	if currentUser.Login == pr.User.Login {
		return cliexit.NewValidation(
			cliexit.ErrCodeSelfApprove,
			fmt.Errorf("cannot approve your own PR: #%d was opened by %s", prNumber, pr.User.Login),
			map[string]any{"pr_number": prNumber, "author": pr.User.Login},
		)
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
		side := strings.ToUpper(c.Side)
		if side == "" {
			side = "RIGHT"
		}
		body := c.Body
		if c.Suggestion != "" {
			if body != "" {
				body += "\n```suggestion\n" + c.Suggestion + "\n```"
			} else {
				body = "```suggestion\n" + c.Suggestion + "\n```"
			}
		}
		req.Comments = append(req.Comments, reviewCommentReq{
			Path: c.Path,
			Line: c.Line,
			Side: side,
			Body: body,
		})
	}
	return req
}
