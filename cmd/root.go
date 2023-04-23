package cmd

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	openai "github.com/sashabaranov/go-openai"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:  "chatgpt [message arguments...]",
		Long: "A CLI tool to interface with ChatGPT. Provide message content as arguments for a single\noneshot or enter interactive chat with no args.",
		Args: cobra.ArbitraryArgs,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				REPL()
				return
			}
			resp, err := StreamMessages(args[0])
			if err != nil {
				fmt.Printf("ChatCompletion error: %v\n", err)
				return
			}
			fmt.Println(resp.Choices[0].Message.Content)
		},
	}
)

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

// Send oneshot message.
func SendMessage(content string) (openai.ChatCompletionResponse, error) {
	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: content,
				},
			},
		},
	)
	return resp, err
}

// Send a stream of messages in chat session
func StreamMessages(content string) (openai.ChatCompletionResponse, error) {
	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))
	messages := make([]openai.ChatCompletionMessage, 0)
	content = strings.TrimSpace(content)
	content = strings.Replace(content, "\n", "", -1)
	content = strings.Replace(content, `"`, "", -1)
	// Add user chat message to the list.
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: content,
	})

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    openai.GPT3Dot5Turbo,
			Messages: messages,
		},
	)

	if err != nil {
		return resp, err
	}

	// Add chatgpt response message content to the chat log.
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleAssistant,
		Content: resp.Choices[0].Message.Content,
	})

	return resp, nil
}

// Read, execute and print ChatGPT response.
func REPL() error {
	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))
	messages := make([]openai.ChatCompletionMessage, 0)
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("ChatGPT> ")
		content, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Println("EOF received. Exiting...")
				os.Exit(0)
			}
			fmt.Fprintln(os.Stderr, err)
			continue
		}
		content = strings.TrimSpace(content)
		content = strings.Replace(content, "\n", "", -1)
		content = strings.Replace(content, `"`, "", -1)
		// Add user chat message to the list.
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: content,
		})
		if content == "exit" || content == "quit" {
			break
		}

		resp, err := client.CreateChatCompletion(
			context.Background(),
			openai.ChatCompletionRequest{
				Model:    openai.GPT3Dot5Turbo,
				Messages: messages,
			},
		)

		fmt.Println(resp.Choices[0].Message.Content)
		// Add chatgpt response message content to the chat log.
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleAssistant,
			Content: resp.Choices[0].Message.Content,
		})
	}

	return nil
}
