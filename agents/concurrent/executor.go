package concurrent

import (
	"context"
	"errors"
	"fmt"
	"github.com/tmc/langchaingo/agents"
	"github.com/tmc/langchaingo/prompts"
	"strings"
	"sync"

	"github.com/tmc/langchaingo/callbacks"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/tools"
)

const _intermediateStepsOutputKey = "intermediateSteps"

// Executor is the chain responsible for running agents.
type Executor struct {
	Agent            agents.Agent
	Memory           schema.Memory
	CallbacksHandler callbacks.Handler
	ErrorHandler     *agents.ParserErrorHandler

	MaxIterations           int
	ReturnIntermediateSteps bool
}

var (
	_ chains.Chain           = &Executor{}
	_ callbacks.HandlerHaver = &Executor{}
)

type Options struct {
	Prompt                  prompts.PromptTemplate
	Memory                  schema.Memory
	CallbacksHandler        callbacks.Handler
	ErrorHandler            *agents.ParserErrorHandler
	MaxIterations           int
	ReturnIntermediateSteps bool
	OutputKey               string
	PromptPrefix            string
	FormatInstructions      string
	PromptSuffix            string

	PromptPrefixInputVariables       []string
	FormatInstructionsInputVariables []string
	PromptSuffixInputVariables       []string
	// openai
	SystemMessage string
	ExtraMessages []prompts.MessageFormatter
}

// NewExecutor creates a new agent executor with an agent and the tools the agent can use.
func NewExecutor(agent agents.Agent, options Options) *Executor {
	return &Executor{
		Agent:                   agent,
		Memory:                  options.Memory,
		MaxIterations:           options.MaxIterations,
		ReturnIntermediateSteps: options.ReturnIntermediateSteps,
		CallbacksHandler:        options.CallbacksHandler,
		ErrorHandler:            options.ErrorHandler,
	}
}

func (e *Executor) Call(ctx context.Context, inputValues map[string]any, _ ...chains.ChainCallOption) (map[string]any, error) { //nolint:lll
	inputs, err := inputsToString(inputValues)
	if err != nil {
		return nil, err
	}
	nameToTool := getNameToTool(e.Agent.GetTools())

	var nameToToolM sync.Map
	for k, tool := range nameToTool {
		nameToToolM.Store(k, tool)
	}
	steps := make([]schema.AgentStep, 0)
	for i := 0; i < e.MaxIterations; i++ {
		var finish map[string]any
		steps, finish, err = e.doIteration(ctx, steps, &nameToToolM, inputs)
		if finish != nil || err != nil {
			return finish, err
		}
	}

	if e.CallbacksHandler != nil {
		e.CallbacksHandler.HandleAgentFinish(ctx, schema.AgentFinish{
			ReturnValues: map[string]any{"output": agents.ErrNotFinished.Error()},
		})
	}
	return e.getReturn(
		&schema.AgentFinish{ReturnValues: make(map[string]any)},
		steps,
	), agents.ErrNotFinished
}

func (e *Executor) doIteration( // nolint
	ctx context.Context,
	steps []schema.AgentStep,
	nameToTool *sync.Map,
	inputs map[string]string,
) ([]schema.AgentStep, map[string]any, error) {
	actions, finish, err := e.Agent.Plan(ctx, steps, inputs)
	if errors.Is(err, agents.ErrUnableToParseOutput) && e.ErrorHandler != nil {
		formattedObservation := err.Error()
		if e.ErrorHandler.Formatter != nil {
			formattedObservation = e.ErrorHandler.Formatter(formattedObservation)
		}
		steps = append(steps, schema.AgentStep{
			Observation: formattedObservation,
		})
		return steps, nil, nil
	}
	if err != nil {
		return steps, nil, err
	}

	if len(actions) == 0 && finish == nil {
		return steps, nil, agents.ErrAgentNoReturn
	}

	if finish != nil {
		if e.CallbacksHandler != nil {
			e.CallbacksHandler.HandleAgentFinish(ctx, *finish)
		}
		return steps, e.getReturn(finish, steps), nil
	}
	type inErr struct {
		Errs   error
		Action schema.AgentAction
	}
	errItem := make(chan inErr, 1)
	StepList := make(chan schema.AgentStep, len(actions))
	for _, action := range actions {
		go func(ac schema.AgentAction, errs chan inErr, stepList chan schema.AgentStep) {
			step, err := e.doAction(ctx, nameToTool, ac)
			if err != nil {
				errs <- inErr{
					Errs:   err,
					Action: ac,
				}
			}
			stepList <- step
		}(action, errItem, StepList)

	}

	for i := 0; i < len(actions); {
		select {
		case errs, ok := <-errItem:
			if ok {
				return steps, nil, fmt.Errorf("%s,err:%e", errs.Action.Tool, errs.Errs)
			}
		case step, ok := <-StepList:
			if ok {
				steps = append(steps, step)
				i++
			}
		}
	}
	return steps, nil, nil
}

func (e *Executor) doAction(
	ctx context.Context,
	nameToTool *sync.Map,
	action schema.AgentAction,
) (schema.AgentStep, error) {
	if e.CallbacksHandler != nil {
		e.CallbacksHandler.HandleAgentAction(ctx, action)
	}
	toolAny, ok := nameToTool.Load(strings.ToUpper(action.Tool))
	if !ok {
		return schema.AgentStep{
			Action:      action,
			Observation: fmt.Sprintf("%s is not a valid tool1, try another one", action.Tool),
		}, nil
	}
	tool, ok := toolAny.(tools.Tool)
	if !ok {
		return schema.AgentStep{
			Action:      action,
			Observation: fmt.Sprintf("%s is not a valid tool2, try another one", action.Tool),
		}, nil
	}
	observation, err := tool.Call(ctx, action.ToolInput)
	if err != nil {
		return schema.AgentStep{}, err
	}

	return schema.AgentStep{
		Action:      action,
		Observation: observation,
	}, nil
}

func (e *Executor) getReturn(finish *schema.AgentFinish, steps []schema.AgentStep) map[string]any {
	if e.ReturnIntermediateSteps {
		finish.ReturnValues[_intermediateStepsOutputKey] = steps
	}

	return finish.ReturnValues
}

// GetInputKeys gets the input keys the agent of the executor expects.
// Often "input".
func (e *Executor) GetInputKeys() []string {
	return e.Agent.GetInputKeys()
}

// GetOutputKeys gets the output keys the agent of the executor returns.
func (e *Executor) GetOutputKeys() []string {
	return e.Agent.GetOutputKeys()
}

func (e *Executor) GetMemory() schema.Memory { //nolint:ireturn
	return e.Memory
}

func (e *Executor) GetCallbackHandler() callbacks.Handler { //nolint:ireturn
	return e.CallbacksHandler
}

func inputsToString(inputValues map[string]any) (map[string]string, error) {
	inputs := make(map[string]string, len(inputValues))
	for key, value := range inputValues {
		valueStr, ok := value.(string)
		if !ok {
			return nil, fmt.Errorf("%w: %s", agents.ErrExecutorInputNotString, key)
		}

		inputs[key] = valueStr
	}

	return inputs, nil
}

func getNameToTool(t []tools.Tool) map[string]tools.Tool {
	if len(t) == 0 {
		return nil
	}

	nameToTool := make(map[string]tools.Tool, len(t))
	for _, tool := range t {
		nameToTool[strings.ToUpper(tool.Name())] = tool
	}

	return nameToTool
}
