package concurrent

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/prompts"
	"strings"
	"time"

	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/tools"
)

const (
	_finalAnswerAction = "Final Answer:"
	_defaultOutputKey  = "output"
)

type ActionItem struct {
	Action      string `json:"Action"`
	ActionInput string `json:"ActionInput"`
}

type TaskFlow struct {
	Question    string       `json:"Question"`
	Thought     string       `json:"Thought"`
	FinalAnswer string       `json:"FinalAnswer"`
	Actions     []ActionItem `json:"Actions"`
}

// ConcurrentAgent is a struct that represents an agent responsible for deciding
// what to do or give the final output if the task is finished given a set of inputs
// and previous steps taken.
//
// This agent is optimized to be used with LLMs.
type ConcurrentAgent struct {
	// Chain is the chain used to call with the values. The chain should have an
	// input called "agent_scratchpad" for the agent to put its thoughts in.
	Chain chains.Chain
	// Tools is a list of the tools the agent can use.
	Tools []tools.Tool
	// Output key is the key where the final output is placed.
	OutputKey string
	// CallbacksHandler is the handler for callbacks.
	CallbacksHandler callbacks.Handler
}

var _ agents.Agent = (*ConcurrentAgent)(nil)

func getConcurrentPrompt(tools []tools.Tool) prompts.PromptTemplate {

	return createConcurrentPrompt(
		tools,
		ConcurrentTemplateBase{_defaultMrklPrefix, []string{"today"}},
		ConcurrentTemplateBase{_defaultMrklFormatInstructions, []string{}},
		ConcurrentTemplateBase{_defaultMrklSuffix, []string{"agent_scratchpad", "input"}},
	)
}

// NewConcurrentAgent creates a new ConcurrentAgent with the given LLM model, tools,
// and options. It returns a pointer to the created agent. The opts parameter
// represents the options for the agent.
func NewConcurrentAgent(llm llms.Model, tools []tools.Tool, opts ...agents.Option) *ConcurrentAgent {

	return &ConcurrentAgent{
		Chain: chains.NewLLMChain(
			llm,
			getConcurrentPrompt(tools),
			//chains.WithCallback(options.callbacksHandler),
		),
		Tools:     tools,
		OutputKey: _defaultOutputKey,
		//CallbacksHandler: options.callbacksHandler,
	}
}

// Plan decides what action to take or returns the final result of the input.
func (a *ConcurrentAgent) Plan(
	ctx context.Context,
	intermediateSteps []schema.AgentStep,
	inputs map[string]string,
) ([]schema.AgentAction, *schema.AgentFinish, error) {
	fullInputs := make(map[string]any, len(inputs))
	for key, value := range inputs {
		fullInputs[key] = value
	}

	fullInputs["agent_scratchpad"] = constructConcurrentScratchPad(intermediateSteps)
	fullInputs["today"] = time.Now().Format("January 02, 2006")

	var stream func(ctx context.Context, chunk []byte) error

	if a.CallbacksHandler != nil {
		stream = func(ctx context.Context, chunk []byte) error {
			a.CallbacksHandler.HandleStreamingFunc(ctx, chunk)
			return nil
		}
	}

	output, err := chains.Predict(
		ctx,
		a.Chain,
		fullInputs,
		chains.WithStopWords([]string{"\nObservation:", "\n\tObservation:"}),
		chains.WithStreamingFunc(stream),
	)
	if err != nil {
		return nil, nil, err
	}
	fmt.Println("@@@@output:", output)
	return a.parseOutput(output)
}

func (a *ConcurrentAgent) GetInputKeys() []string {
	chainInputs := a.Chain.GetInputKeys()

	// Remove inputs given in plan.
	agentInput := make([]string, 0, len(chainInputs))
	for _, v := range chainInputs {
		if v == "agent_scratchpad" || v == "today" {
			continue
		}
		agentInput = append(agentInput, v)
	}

	return agentInput
}

func (a *ConcurrentAgent) GetOutputKeys() []string {
	return []string{a.OutputKey}
}

func (a *ConcurrentAgent) GetTools() []tools.Tool {
	return a.Tools
}

func constructConcurrentScratchPad(steps []schema.AgentStep) string {
	var scratchPad string
	if len(steps) > 0 {
		for _, step := range steps {
			scratchPad += "\n" + step.Action.Log
			scratchPad += "\nObservation: " + step.Observation + "\n"
		}
	}

	return scratchPad
}

func (a *ConcurrentAgent) parseOutput(output string) ([]schema.AgentAction, *schema.AgentFinish, error) {
	output = strings.Replace(output, "`", "", -1)
	output = strings.Replace(output, "json", "", 1)
	output = strings.TrimSpace(output)
	var task TaskFlow
	if err := json.Unmarshal([]byte(output), &task); err != nil {
		return nil, nil, fmt.Errorf("%s: %s", err.Error(), output)
	}
	if task.FinalAnswer != "" {
		return nil, &schema.AgentFinish{
			ReturnValues: map[string]any{
				a.OutputKey: task.FinalAnswer,
			},
			Log: output,
		}, nil
	}

	actions := make([]schema.AgentAction, 0)
	for _, action := range task.Actions {
		actions = append(actions, schema.AgentAction{
			Tool:      action.Action,
			ToolInput: action.ActionInput,
		})
	}
	if len(actions) == 0 {
		return nil, nil, fmt.Errorf("%w: %s", agents.ErrUnableToParseOutput, output)
	}

	return actions, nil, nil
}
