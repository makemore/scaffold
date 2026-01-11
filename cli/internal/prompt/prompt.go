// Package prompt handles interactive user prompts
package prompt

import (
	"fmt"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/makemore/scaffold/internal/config"
)

// PromptForVariables prompts the user for each variable defined in the manifest
func PromptForVariables(cfg *config.Manifest, existingVars map[string]string) (map[string]string, error) {
	result := make(map[string]string)

	// Copy existing vars
	for k, v := range existingVars {
		result[k] = v
	}

	for _, v := range cfg.Variables {
		// Skip if already provided
		if _, exists := result[v.Name]; exists {
			continue
		}

		value, err := promptForVariable(v)
		if err != nil {
			return nil, err
		}
		result[v.Name] = value
	}

	return result, nil
}

func promptForVariable(v config.Variable) (string, error) {
	// Build the prompt message
	message := v.Name
	if v.Description != "" {
		message = v.Description
	}

	switch v.Type {
	case "select", "choice":
		return promptSelect(message, v.Choices, v.Default)
	case "confirm", "boolean":
		return promptConfirm(message, v.Default == "true")
	default:
		return promptInput(message, v.Default)
	}
}

func promptInput(message, defaultValue string) (string, error) {
	var result string
	prompt := &survey.Input{
		Message: message,
		Default: defaultValue,
	}
	if err := survey.AskOne(prompt, &result); err != nil {
		return "", err
	}
	return result, nil
}

func promptSelect(message string, options []string, defaultValue string) (string, error) {
	if len(options) == 0 {
		return promptInput(message, defaultValue)
	}

	var result string
	prompt := &survey.Select{
		Message: message,
		Options: options,
		Default: defaultValue,
	}
	if err := survey.AskOne(prompt, &result); err != nil {
		return "", err
	}
	return result, nil
}

func promptConfirm(message string, defaultValue bool) (string, error) {
	var result bool
	prompt := &survey.Confirm{
		Message: message,
		Default: defaultValue,
	}
	if err := survey.AskOne(prompt, &result); err != nil {
		return "", err
	}
	if result {
		return "true", nil
	}
	return "false", nil
}

// PromptForTemplate prompts the user to select a template if none provided
func PromptForTemplate(templates map[string]string) (string, error) {
	if len(templates) == 0 {
		return "", fmt.Errorf("no templates available")
	}

	options := make([]string, 0, len(templates))
	for name, desc := range templates {
		options = append(options, fmt.Sprintf("%s - %s", name, desc))
	}

	var result string
	prompt := &survey.Select{
		Message: "Select a template:",
		Options: options,
	}
	if err := survey.AskOne(prompt, &result); err != nil {
		return "", err
	}

	// Extract template name from selection
	parts := strings.SplitN(result, " - ", 2)
	return parts[0], nil
}

// PromptForProjectName prompts for project name if not provided
func PromptForProjectName() (string, error) {
	var result string
	prompt := &survey.Input{
		Message: "Project name:",
	}
	if err := survey.AskOne(prompt, &result, survey.WithValidator(survey.Required)); err != nil {
		return "", err
	}
	return result, nil
}

