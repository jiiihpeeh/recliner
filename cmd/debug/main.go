package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/jiiihpeeh/recliner/app"
	"github.com/jiiihpeeh/recliner/c"
	"github.com/jiiihpeeh/recliner/events"
	"github.com/jiiihpeeh/recliner/hooks"
	"github.com/jiiihpeeh/recliner/vdom"
)

var (
	globalCategories = make(map[string][]string)
	mu               sync.Mutex
)

func parseMessage(line string) (string, string) {
	// Parse format: [timestamp] [category] message
	re := regexp.MustCompile(`^\[(\d{2}:\d{2}:\d{2})\] \[([^\]]+)\] (.+)$`)
	matches := re.FindStringSubmatch(line)
	if len(matches) == 4 {
		return matches[2], matches[3] // category, message
	}
	// Fallback for uncategorized messages
	return "GENERAL", line
}

func addMessage(line string) {
	category, message := parseMessage(line)
	mu.Lock()
	globalCategories[category] = append(globalCategories[category], message)

	// Keep only last 1000 messages per category
	if len(globalCategories[category]) > 1000 {
		globalCategories[category] = globalCategories[category][1:]
	}
	mu.Unlock()
}

func getCategoryNames() []string {
	mu.Lock()
	var names []string
	for name := range globalCategories {
		names = append(names, name)
	}
	mu.Unlock()
	sort.Strings(names)
	return names
}

func getCategories() map[string][]string {
	mu.Lock()
	// Copy the map
	categories := make(map[string][]string)
	for k, v := range globalCategories {
		categories[k] = make([]string, len(v))
		copy(categories[k], v)
	}
	mu.Unlock()
	return categories
}

func startConnection(socketPath string) {
	go func() {
		for {
			fmt.Println("Connecting to ReCLIner debug server...")

			conn, err := net.Dial("unix", socketPath)
			if err != nil {
				fmt.Printf("Failed to connect to debug server: %v\n", err)
				fmt.Println("Make sure ReCLIner is running with debug mode enabled.")
				fmt.Println("Retrying in 1/4 second...")
				time.Sleep(250 * time.Millisecond)
				continue
			}
			defer conn.Close()

			fmt.Println("Connected! Displaying debug output:")
			fmt.Println("===================================")

			scanner := bufio.NewScanner(conn)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if line != "" {
					addMessage(line)
				}
			}

			if err := scanner.Err(); err != nil {
				fmt.Printf("Connection lost: %v\n", err)
				fmt.Println("Attempting to reconnect...")
				time.Sleep(1 * time.Second)
				continue
			}

			// If we reach here, the connection was closed gracefully
			fmt.Println("Debug server closed the connection.")
			break
		}
	}()
}

func createApp(props any) vdom.Node {
	hc := hooks.GetContext()

	tick, setTick := hooks.UseState[int](hc, 0)
	currentTab, setCurrentTab := hooks.UseState[int](hc, 0)

	// Update categories from global every tick
	categories := getCategories()
	categoryNames := getCategoryNames()

	// Ensure current tab is valid
	if currentTab >= len(categoryNames) && len(categoryNames) > 0 {
		setCurrentTab(len(categoryNames) - 1)
		currentTab = len(categoryNames) - 1
	}

	// Timer to update UI
	hc.UseEffect(func() func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		go func() {
			for range ticker.C {
				setTick(tick + 1)
			}
		}()
		return func() {
			ticker.Stop()
		}
	}, []any{})

	// Keyboard navigation
	hc.UseInput(func(e events.KeyPressEvent) {
		switch e.Key {
		case "left":
			if currentTab > 0 {
				setCurrentTab(currentTab - 1)
			}
		case "right":
			if currentTab < len(categoryNames)-1 {
				setCurrentTab(currentTab + 1)
			}
		case "q":
			os.Exit(0)
		}
	}, []any{currentTab, len(categoryNames)})

	if len(categoryNames) == 0 {
		return c.Text(c.TextProps{Content: "No debug messages received yet..."})
	}

	// Create tabs
	var tabItems []c.TabItem
	for _, name := range categoryNames {
		messages := categories[name]
		var children []vdom.Node
		for _, msg := range messages {
			children = append(children, c.Text(c.TextProps{Content: msg}))
		}
		content := c.Box(c.BoxProps{
			BorderStyle: c.BorderStyleNone,
			Padding:     0,
			Style: c.StyleProps{
				Display:       c.DisplayFlex,
				FlexDirection: c.FlexDirectionColumn,
			},
		}, children...)
		tabItems = append(tabItems, c.TabItem{
			Title:   name,
			Content: content,
		})
	}

	onChange := func(tabID string) {
		for i, name := range categoryNames {
			if name == tabID {
				setCurrentTab(i)
				break
			}
		}
	}

	return c.Tabs(c.TabsProps{
		Items:     tabItems,
		ActiveTab: categoryNames[currentTab],
		OnChange:  onChange,
	})
}

func main() {
	socketPath := "/tmp/recliner_debug.sock"

	// Start connection in background
	startConnection(socketPath)

	// Create ReCLIner app
	appInstance := app.NewWithOptions(func(props any) vdom.Node {
		return createApp(props)
	}, app.AppOptions{})

	if err := appInstance.Run(); err != nil {
		fmt.Printf("Error running app: %v\n", err)
	}
}
