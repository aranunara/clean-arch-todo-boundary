package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

type todo struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return usage()
	}

	baseURL, err := detectAPIURL()
	if err != nil {
		return err
	}

	client := &http.Client{Timeout: 5 * time.Second}

	switch args[0] {
	case "url":
		fmt.Println(baseURL)
		return nil
	case "list":
		if len(args) != 1 {
			return usage()
		}
		return listTodos(client, baseURL)
	case "create":
		if len(args) < 2 {
			return usage()
		}
		return createTodo(client, baseURL, strings.Join(args[1:], " "))
	case "rename":
		if len(args) < 3 {
			return usage()
		}
		return renameTodo(client, baseURL, args[1], strings.Join(args[2:], " "))
	case "complete":
		if len(args) != 2 {
			return usage()
		}
		return completeTodo(client, baseURL, args[1])
	default:
		return usage()
	}
}

func detectAPIURL() (string, error) {
	for _, key := range []string{"TODO_API_URL", "API_URL"} {
		value := strings.TrimSpace(os.Getenv(key))
		if value != "" {
			return strings.TrimRight(value, "/"), nil
		}
	}

	output, err := exec.Command("docker", "compose", "port", "api", "8080").Output()
	if err != nil {
		return "", errors.New("API URL を検出できませんでした。先に make up-all するか、TODO_API_URL=http://127.0.0.1:PORT を指定してください")
	}

	endpoint := strings.TrimSpace(string(output))
	if endpoint == "" {
		return "", errors.New("API が起動していません。先に make up-all してください")
	}

	host, port, err := net.SplitHostPort(endpoint)
	if err != nil {
		return "", fmt.Errorf("API の公開ポートを読めませんでした: %w", err)
	}
	if host == "" || host == "0.0.0.0" || host == "::" {
		host = "127.0.0.1"
	}

	return "http://" + net.JoinHostPort(host, port), nil
}

func listTodos(client *http.Client, baseURL string) error {
	var todos []todo
	if err := doJSON(client, http.MethodGet, baseURL+"/todos", nil, &todos); err != nil {
		return err
	}
	return printJSON(todos)
}

func createTodo(client *http.Client, baseURL, title string) error {
	body := map[string]string{"title": title}
	var created todo
	if err := doJSON(client, http.MethodPost, baseURL+"/todos", body, &created); err != nil {
		return err
	}
	return printJSON(created)
}

func renameTodo(client *http.Client, baseURL, id, title string) error {
	body := map[string]string{"title": title}
	var renamed todo
	if err := doJSON(client, http.MethodPatch, baseURL+"/todos/"+id, body, &renamed); err != nil {
		return err
	}
	return printJSON(renamed)
}

func completeTodo(client *http.Client, baseURL, id string) error {
	var completed todo
	if err := doJSON(client, http.MethodPost, baseURL+"/todos/"+id+"/complete", nil, &completed); err != nil {
		return err
	}
	return printJSON(completed)
}

func doJSON(client *http.Client, method, url string, requestBody any, responseBody any) error {
	var body io.Reader
	if requestBody != nil {
		payload, err := json.Marshal(requestBody)
		if err != nil {
			return err
		}
		body = bytes.NewReader(payload)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return err
	}
	if requestBody != nil {
		req.Header.Set("content-type", "application/json")
	}

	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = res.Body.Close()
	}()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		payload, _ := io.ReadAll(res.Body)
		return fmt.Errorf("%s %s returned %s: %s", method, url, res.Status, strings.TrimSpace(string(payload)))
	}

	return json.NewDecoder(res.Body).Decode(responseBody)
}

func printJSON(value any) error {
	payload, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(payload))
	return nil
}

func usage() error {
	return errors.New(`usage:
  go run ./cmd/todoctl url
  go run ./cmd/todoctl list
  go run ./cmd/todoctl create "層の違いをメモする"
  go run ./cmd/todoctl rename 1 "Domain と Usecase の違いをメモする"
  go run ./cmd/todoctl complete 1

env:
  TODO_API_URL=http://127.0.0.1:18080`)
}
