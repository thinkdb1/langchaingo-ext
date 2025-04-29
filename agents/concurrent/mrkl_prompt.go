package concurrent

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/tmc/langchaingo/prompts"
	"github.com/tmc/langchaingo/tools"
)

const (
	_defaultMrklPrefix = `Today is {{.today}}.
Answer the following questions to the best of your ability. You can use the following tools	:
{{.tool_descriptions}}
`

	_defaultMrklFormatInstructions = `
	Generate a JSON-formatted data structure based on the information provided below.Output must strictly adhere to the standard JSON structure without any additional characters or strings.
Content requirements:
	•	If a task has actions that can be executed in parallel, then the “actions” field of that task can contain multiple actions.
	•	If the actions of a task cannot be executed in parallel, then the “actions” field of that task must contain only one action.
Output example:
{
	"Question": "the input question you must answer",
	"Thought": "You should always be thinking about what needs to be done and what can be executed concurrently",
	"FinalAnswer": "the final answer to the original input question,it must be a empty string if not end",
	"Actions": [{
		"Action": "the actions to take, should be in the [ {{.tool_names}} ]",
		"ActionInput": "the input to the action"
	}, {
		"Action": "the actions to take, should be in the [ {{.tool_names}} ]",
		"ActionInput": "the input to the action"
	}]
}
`

	_defaultMrklSuffix = `Begin!

Question: {{.input}}
{{.agent_scratchpad}}`
)

type ConcurrentTemplateBase struct {
	Template       string
	InputVariables []string
}

func createConcurrentPrompt(tools []tools.Tool, prefix, instructions, suffix ConcurrentTemplateBase) prompts.PromptTemplate {
	template := strings.Join([]string{prefix.Template, instructions.Template, suffix.Template}, "\n\n")
	inputVariables := make([]string, 0, len(prefix.InputVariables)+
		len(instructions.InputVariables)+
		len(suffix.InputVariables))
	inputVariables = append(inputVariables, prefix.InputVariables...)
	inputVariables = append(inputVariables, instructions.InputVariables...)
	inputVariables = append(inputVariables, suffix.InputVariables...)

	if err := checkConcurrentTemplate(template); err != nil {
		log.Println(err.Error())
	}

	return prompts.PromptTemplate{
		Template:       template,
		TemplateFormat: prompts.TemplateFormatGoTemplate,
		InputVariables: inputVariables,
		PartialVariables: map[string]any{
			"tool_names":        toolNames(tools),
			"tool_descriptions": toolDescriptions(tools),
		},
	}
}

// checkMrklPrompt check Prompt for PartialVariables.
func checkConcurrentTemplate(template string) error {
	re := regexp.MustCompile(`\{\{\.(.*?)\}\}`)
	matches := re.FindAllStringSubmatch(template, -1)
	matchesMap := make(map[string]struct{})
	for _, match := range matches {
		matchesMap[match[1]] = struct{}{}
	}
	mustMatches := []string{"tool_names", "tool_descriptions"}
	for _, v := range mustMatches {
		if _, ok := matchesMap[v]; !ok {
			return errors.New(v + " is not contained in option template")
		}
	}
	return nil
}

func toolNames(tools []tools.Tool) string {
	var tn strings.Builder
	for i, tool := range tools {
		if i > 0 {
			tn.WriteString(", ")
		}
		tn.WriteString(tool.Name())
	}

	return tn.String()
}

func toolDescriptions(tools []tools.Tool) string {
	var ts strings.Builder
	for _, tool := range tools {
		ts.WriteString(fmt.Sprintf("- %s: %s\n", tool.Name(), tool.Description()))
	}

	return ts.String()
}
