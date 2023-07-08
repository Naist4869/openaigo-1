package chatgpt

import (
	"context"
	"fmt"

	"github.com/otiai10/openaigo"
	"github.com/otiai10/openaigo/functioncall"
)

type Client struct {
	openaigo.Client `json:"-"`

	// Model: ID of the model to use.
	// Currently, only gpt-3.5-turbo and gpt-3.5-turbo-0301 are supported.
	Model string `json:"model"`

	// Temperature: What sampling temperature to use, between 0 and 2.
	// Higher values like 0.8 will make the output more random, while lower values like 0.2 will make it more focused and deterministic.
	// We generally recommend altering this or top_p but not both.
	// Defaults to 1.
	Temperature float32 `json:"temperature,omitempty"`

	// TopP: An alternative to sampling with temperature, called nucleus sampling,
	// where the model considers the results of the tokens with top_p probability mass.
	// So 0.1 means only the tokens comprising the top 10% probability mass are considered.
	// We generally recommend altering this or temperature but not both.
	// Defaults to 1.
	TopP float32 `json:"top_p,omitempty"`

	// N: How many chat completion choices to generate for each input message.
	// Defaults to 1.
	N int `json:"n,omitempty"`

	// TODO:
	// Stream: If set, partial message deltas will be sent, like in ChatGPT.
	// Tokens will be sent as data-only server-sent events as they become available,
	// with the stream terminated by a data: [DONE] message.
	// Stream bool `json:"stream,omitempty"`

	// TODO:
	// StreamCallback is a callback funciton to handle stream response.
	// If provided, this library automatically set `Stream` `true`.
	// This field is added by github.com/otiai10/openaigo only to handle Stream.
	// Thus, it is omitted when the client excute HTTP request.
	// StreamCallback func(res ChatCompletionResponse, done bool, err error) `json:"-"`

	// Stop: Up to 4 sequences where the API will stop generating further tokens.
	// Defaults to null.
	Stop []string `json:"stop,omitempty"`

	// MaxTokens: The maximum number of tokens allowed for the generated answer.
	// By default, the number of tokens the model can return will be (4096 - prompt tokens).
	MaxTokens int `json:"max_tokens,omitempty"`

	// PresencePenalty: Number between -2.0 and 2.0.
	// Positive values penalize new tokens based on whether they appear in the text so far,
	// increasing the model's likelihood to talk about new topics.
	// See more information about frequency and presence penalties.
	// https://platform.openai.com/docs/api-reference/parameter-details
	PresencePenalty float32 `json:"presence_penalty,omitempty"`

	// FrequencyPenalty: Number between -2.0 and 2.0.
	// Positive values penalize new tokens based on their existing frequency in the text so far,
	// decreasing the model's likelihood to repeat the same line verbatim.
	// See more information about frequency and presence penalties.
	// https://platform.openai.com/docs/api-reference/parameter-details
	FrequencyPenalty float32 `json:"frequency_penalty,omitempty"`

	// LogitBias: Modify the likelihood of specified tokens appearing in the completion.
	// Accepts a json object that maps tokens (specified by their token ID in the tokenizer)
	// to an associated bias value from -100 to 100.
	// Mathematically, the bias is added to the logits generated by the model prior to sampling.
	// The exact effect will vary per model, but values between -1 and 1 should decrease or increase likelihood of selection;
	// values like -100 or 100 should result in a ban or exclusive selection of the relevant token.
	LogitBias map[string]int `json:"logit_bias,omitempty"`

	// User: A unique identifier representing your end-user, which can help OpenAI to monitor and detect abuse. Learn more.
	// https://platform.openai.com/docs/guides/safety-best-practices/end-user-ids
	User string `json:"user,omitempty"`

	// Functions: A list of functions which GPT is allowed to request to call.
	// Functions []Function `json:"functions,omitempty"`
	Functions functioncall.Funcs `json:"functions,omitempty"`

	// FunctionCall: You ain't need it. Default is "auto".
	FunctionCall string `json:"function_call,omitempty"`
}

type Message openaigo.Message

func New(apikey, model string) *Client {
	return &Client{
		Client: openaigo.Client{
			APIKey: apikey,
		},
		Model: model,
	}
}

func (c *Client) Chat(ctx context.Context, conv []Message) ([]Message, error) {
	// Create messages from conv
	messages := make([]openaigo.Message, len(conv))
	for i, m := range conv {
		messages[i] = openaigo.Message(m)
	}
	// Create request
	req := openaigo.ChatRequest{
		Model:     c.Model,
		Messages:  messages,
		Functions: functioncall.Funcs(c.Functions),
		// TODO: more options from from *Client
	}
	// Call API
	res, err := c.Client.Chat(ctx, req)
	if err != nil {
		return conv, err
	}
	conv = append(conv, Message(res.Choices[0].Message))

	if res.Choices[0].Message.FunctionCall != nil {
		call := res.Choices[0].Message.FunctionCall
		conv = append(conv, Func(call.Name(), c.Functions.Call(call)))
		return c.Chat(ctx, conv)
	}

	return conv, nil
}

func User(message string) Message {
	return Message{
		Role:    "user",
		Content: message,
	}
}

func Func(name string, data interface{}) Message {
	return Message{
		Role:    "function",
		Name:    name,
		Content: fmt.Sprintf("%+v\n", data),
	}
}

func System(message string) Message {
	return Message{
		Role:    "system",
		Content: message,
	}
}
