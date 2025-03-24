package ui

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/olekukonko/tablewriter"
)

// ValidatorMetrics represents the metrics to display in the dashboard
type ValidatorMetrics struct {
	NodeID               string
	Address              string
	Balance              string
	Registered           bool
	LastBlockProcessed   uint64
	VerificationQueueSize int
	ProcessedRequests    int
	SuccessfulSubmissions int
	FailedSubmissions    int
	Rewards              string
}

// RequestLog represents a log entry for a verification request
type RequestLog struct {
	Timestamp  time.Time
	RequestID  string
	Status     string
	TxHash     string
	Message    string
}

// ConsoleUI represents a console-based UI for the validator
type ConsoleUI struct {
	metrics        ValidatorMetrics
	logs           []RequestLog
	mutex          sync.Mutex
	maxLogs        int
	updateInterval time.Duration
	running        bool
	stopChan       chan struct{}
}

// NewConsoleUI creates a new console UI
func NewConsoleUI() *ConsoleUI {
	return &ConsoleUI{
		logs:           make([]RequestLog, 0),
		maxLogs:        100,
		updateInterval: 1 * time.Second, // More responsive updates
		stopChan:       make(chan struct{}),
	}
}

// Start starts the console UI
func (c *ConsoleUI) Start() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Only start if not already running
	if !c.running {
		c.running = true
		// Start the update routine in a goroutine
		go c.updateRoutine()
	}
}

// Stop stops the console UI
func (c *ConsoleUI) Stop() {
	if c.running {
		// Set running to false first to prevent race conditions
		c.running = false
		
		// Send stop signal in a non-blocking way
		select {
		case c.stopChan <- struct{}{}:
			// Signal sent successfully
		default:
			// Channel is full or closed, that's okay
		}
		
		// Print final message
		fmt.Println("\nConsole UI stopped")
	}
}

// UpdateMetrics updates the validator metrics
func (c *ConsoleUI) UpdateMetrics(metrics ValidatorMetrics) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.metrics = metrics
}

// AddLog adds a new log entry
func (c *ConsoleUI) AddLog(requestID, status, txHash, message string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Add new log at the beginning
	c.logs = append([]RequestLog{{
		Timestamp: time.Now(),
		RequestID: requestID,
		Status:    status,
		TxHash:    txHash,
		Message:   message,
	}}, c.logs...)

	// Trim logs if exceeding max
	if len(c.logs) > c.maxLogs {
		c.logs = c.logs[:c.maxLogs]
	}

	// Print the log immediately
	timeStr := time.Now().Format("2006/01/02 15:04:05")
	timeColor := "\033[36m" // Cyan for timestamp
	idColor := "\033[37;1m" // Bright white for ID
	statusColor := "\033[0m" // Reset
	msgColor := "\033[37m" // White for message
	txColor := "\033[33m" // Yellow for transaction hash
	resetColor := "\033[0m" // Reset
	
	switch status {
	case "success":
		statusColor = "\033[32m" // Green
	case "error":
		statusColor = "\033[31m" // Red
	case "processing":
		statusColor = "\033[33m" // Yellow
	case "pending":
		statusColor = "\033[36m" // Cyan
	case "info":
		statusColor = "\033[34m" // Blue
	}

	// Format status text consistently
	statusText := strings.ToUpper(status)

	// Print formatted log with colors
	fmt.Printf("%s%s%s %s%s%s [%s%s%s] %s%s%s\n", 
		timeColor, timeStr, resetColor, 
		idColor, requestID, resetColor, 
		statusColor, statusText, resetColor, 
		msgColor, message, resetColor)

	if txHash != "" {
		fmt.Printf("  %s→ Transaction: %s%s%s\n", msgColor, txColor, txHash, resetColor)
	}
}

// updateRoutine periodically updates the console display
func (c *ConsoleUI) updateRoutine() {
	// Use a shorter interval for more responsive updates
	ticker := time.NewTicker(c.updateInterval)
	defer ticker.Stop()

	// Initial render
	c.renderMetrics()

	for {
		select {
		case <-ticker.C:
			c.renderMetrics()
		case <-c.stopChan:
			return
		}
	}
}

// renderMetrics renders the metrics table to the console
func (c *ConsoleUI) renderMetrics() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Clear the screen using ANSI escape code
	fmt.Print("\033[H\033[2J")

	// Print a separator to clearly mark the new output
	headerColor := "\033[1;36m" // Bold Cyan
	resetColor := "\033[0m"
	fmt.Println(headerColor + strings.Repeat("═", 80) + resetColor)
	
	// Center the title with color
	title := "VALIDATOR NODE STATUS"
	padding := (80 - len(title)) / 2
	fmt.Println(headerColor + strings.Repeat(" ", padding) + title + strings.Repeat(" ", 80 - padding - len(title)) + resetColor)
	
	fmt.Println(headerColor + strings.Repeat("═", 80) + resetColor)
	fmt.Println()

	// Create metrics table
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"METRIC", "VALUE"})
	table.SetBorder(true)
	table.SetHeaderColor(
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiCyanColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiCyanColor},
	)
	table.SetColumnColor(
		tablewriter.Colors{tablewriter.FgHiWhiteColor},
		tablewriter.Colors{tablewriter.FgHiYellowColor},
	)

	// Add metrics data
	registeredStatus := "\033[31mNo\033[0m" // Red color for No
	if c.metrics.Registered {
		registeredStatus = "\033[32mYes\033[0m" // Green color for Yes
	}

	table.Append([]string{"Node ID", c.metrics.NodeID})
	table.Append([]string{"Address", c.metrics.Address})
	table.Append([]string{"Balance", c.metrics.Balance + " ETH"})
	table.Append([]string{"Registered", registeredStatus})
	table.Append([]string{"Last Block", fmt.Sprintf("%d", c.metrics.LastBlockProcessed)})
	table.Append([]string{"Queue Size", fmt.Sprintf("%d", c.metrics.VerificationQueueSize)})
	table.Append([]string{"Processed Requests", fmt.Sprintf("%d", c.metrics.ProcessedRequests)})
	table.Append([]string{"Successful Submissions", fmt.Sprintf("%d", c.metrics.SuccessfulSubmissions)})
	table.Append([]string{"Failed Submissions", fmt.Sprintf("%d", c.metrics.FailedSubmissions)})
	table.Append([]string{"Rewards", c.metrics.Rewards + " ETH"})

	// Render the table
	table.Render()

	// Print recent logs header
	fmt.Println()
	// Use the same header color variables from above
	fmt.Println(headerColor + "RECENT VERIFICATION REQUESTS" + resetColor)
	fmt.Println(headerColor + strings.Repeat("─", 80) + resetColor)

	// Create logs table
	logsTable := tablewriter.NewWriter(os.Stdout)
	logsTable.SetHeader([]string{"TIME", "REQUEST ID", "STATUS", "TRANSACTION", "MESSAGE"})
	logsTable.SetBorder(true)
	// Set column widths to accommodate full transaction hash
	logsTable.SetColWidth(80)
	logsTable.SetColMinWidth(3, 66) // Transaction column
	logsTable.SetHeaderColor(
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiCyanColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiCyanColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiCyanColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiCyanColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiCyanColor},
	)
	logsTable.SetColumnColor(
		tablewriter.Colors{tablewriter.FgWhiteColor},
		tablewriter.Colors{tablewriter.FgHiWhiteColor},
		tablewriter.Colors{tablewriter.FgHiWhiteColor},
		tablewriter.Colors{tablewriter.FgHiWhiteColor},
		tablewriter.Colors{tablewriter.FgHiWhiteColor},
	)

	// Add log entries (most recent first)
	for i, log := range c.logs {
		if i >= 10 { // Show only the last 10 logs
			break
		}

		timeStr := log.Timestamp.Format("15:04:05")
		
		// Format status with color
		statusStr := log.Status
		switch log.Status {
		case "success":
			statusStr = "\033[32mSUCCESS\033[0m" // Green
		case "error":
			statusStr = "\033[31mERROR\033[0m" // Red
		case "processing":
			statusStr = "\033[33mPROCESSING\033[0m" // Yellow
		case "pending":
			statusStr = "\033[36mPENDING\033[0m" // Cyan
		case "info":
			statusStr = "\033[34mINFO\033[0m" // Blue
		}

		// Format transaction hash with color
		txHash := log.TxHash
		if len(txHash) > 0 {
			// Add 0x prefix if missing
			if !strings.HasPrefix(txHash, "0x") {
				txHash = "0x" + txHash
			}
			// Add color
			txHash = "\033[33m" + txHash + "\033[0m" // Yellow for transaction hash
		}

		logsTable.Append([]string{
			timeStr,
			log.RequestID,
			statusStr,
			txHash,
			log.Message,
		})
	}

	// Render the logs table
	logsTable.Render()

	// Print footer
	fmt.Println(headerColor + strings.Repeat("─", 80) + resetColor)
	fmt.Println("Last updated: " + time.Now().Format("2006/01/02 15:04:05"))
	fmt.Println()
}

// RenderOnce renders the metrics once without starting the update routine
func (c *ConsoleUI) RenderOnce() {
	c.renderMetrics()
}
