# Building an intelligent log analyzer

Create an AI-powered log analysis tool that identifies patterns, anomalies, and potential issues in application logs.

## What you'll build

A CLI tool that:
- Parses log files in various formats (JSON, structured text, etc.)
- Identifies error patterns and anomalies
- Summarizes log activity and trends
- Suggests actions based on detected issues
- Generates alerts for critical problems

## Prerequisites

- Go 1.21+
- LLM API key (OpenAI, Anthropic, etc.)
- Sample log files to analyze

## Step 1: Project setup

```bash
mkdir log-analyzer
cd log-analyzer
go mod init log-analyzer
go get github.com/tmc/langchaingo
go get github.com/sirupsen/logrus  # For structured logging examples
```

## Step 2: Core log analyzer structure

Create `main.go`:

```go
package main

import (
    "bufio"
    "context"
    "encoding/json"
    "flag"
    "fmt"
    "log"
    "os"
    "regexp"
    "sort"
    "strings"
    "time"

    "github.com/tmc/langchaingo/llms"
    "github.com/tmc/langchaingo/llms/openai"
    "github.com/tmc/langchaingo/prompts"
)

type LogEntry struct {
    Timestamp time.Time `json:"timestamp"`
    Level     string    `json:"level"`
    Message   string    `json:"message"`
    Source    string    `json:"source"`
    Raw       string    `json:"raw"`
}

type LogAnalysis struct {
    TotalEntries    int                 `json:"total_entries"`
    ErrorCount      int                 `json:"error_count"`
    WarningCount    int                 `json:"warning_count"`
    TopErrors       []ErrorPattern      `json:"top_errors"`
    TimeRange       TimeRange           `json:"time_range"`
    Recommendations []string            `json:"recommendations"`
    Anomalies       []Anomaly          `json:"anomalies"`
}

type ErrorPattern struct {
    Pattern string `json:"pattern"`
    Count   int    `json:"count"`
    Example string `json:"example"`
}

type TimeRange struct {
    Start time.Time `json:"start"`
    End   time.Time `json:"end"`
}

type Anomaly struct {
    Type        string `json:"type"`
    Description string `json:"description"`
    Severity    string `json:"severity"`
    Examples    []string `json:"examples"`
}

type LogAnalyzer struct {
    llm llms.Model
}

func NewLogAnalyzer() (*LogAnalyzer, error) {
    llm, err := openai.New()
    if err != nil {
        return nil, fmt.Errorf("creating LLM: %w", err)
    }

    return &LogAnalyzer{llm: llm}, nil
}

func (la *LogAnalyzer) ParseLogFile(filename string) ([]LogEntry, error) {
    file, err := os.Open(filename)
    if err != nil {
        return nil, fmt.Errorf("opening file: %w", err)
    }
    defer file.Close()

    var entries []LogEntry
    scanner := bufio.NewScanner(file)
    
    // Common log patterns
    patterns := []*regexp.Regexp{
        // JSON logs
        regexp.MustCompile(`^\{.*\}$`),
        // Standard format: 2023-01-01 12:00:00 [ERROR] message
        regexp.MustCompile(`^(\d{4}-\d{2}-\d{2}\s+\d{2}:\d{2}:\d{2})\s+\[(\w+)\]\s+(.+)$`),
        // Nginx/Apache format
        regexp.MustCompile(`^(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}).*\[([^\]]+)\].*"([^"]*)".*(\d{3})`),
    }

    for scanner.Scan() {
        line := scanner.Text()
        if strings.TrimSpace(line) == "" {
            continue
        }

        entry := LogEntry{Raw: line}
        
        // Try JSON first
        if line[0] == '{' {
            var jsonEntry map[string]interface{}
            if err := json.Unmarshal([]byte(line), &jsonEntry); err == nil {
                entry = parseJSONLog(jsonEntry, line)
                entries = append(entries, entry)
                continue
            }
        }

        // Try structured patterns
        for _, pattern := range patterns[1:] {
            if matches := pattern.FindStringSubmatch(line); matches != nil {
                entry = parseStructuredLog(matches, line)
                break
            }
        }

        // Fallback: treat as unstructured
        if entry.Timestamp.IsZero() {
            entry = LogEntry{
                Timestamp: time.Now(), // Use current time as fallback
                Level:     inferLogLevel(line),
                Message:   line,
                Raw:       line,
            }
        }

        entries = append(entries, entry)
    }

    return entries, scanner.Err()
}

func parseJSONLog(data map[string]interface{}, raw string) LogEntry {
    entry := LogEntry{Raw: raw}
    
    if ts, ok := data["timestamp"].(string); ok {
        if t, err := time.Parse(time.RFC3339, ts); err == nil {
            entry.Timestamp = t
        }
    }
    
    if level, ok := data["level"].(string); ok {
        entry.Level = level
    }
    
    if msg, ok := data["message"].(string); ok {
        entry.Message = msg
    }
    
    if src, ok := data["source"].(string); ok {
        entry.Source = src
    }

    return entry
}

func parseStructuredLog(matches []string, raw string) LogEntry {
    entry := LogEntry{Raw: raw}
    
    if len(matches) >= 4 {
        if t, err := time.Parse("2006-01-02 15:04:05", matches[1]); err == nil {
            entry.Timestamp = t
        }
        entry.Level = matches[2]
        entry.Message = matches[3]
    }
    
    return entry
}

func inferLogLevel(line string) string {
    lower := strings.ToLower(line)
    switch {
    case strings.Contains(lower, "error") || strings.Contains(lower, "fatal"):
        return "ERROR"
    case strings.Contains(lower, "warn"):
        return "WARN"
    case strings.Contains(lower, "debug"):
        return "DEBUG"
    default:
        return "INFO"
    }
}

func (la *LogAnalyzer) AnalyzeLogs(entries []LogEntry) (*LogAnalysis, error) {
    if len(entries) == 0 {
        return &LogAnalysis{}, nil
    }

    // Basic statistics
    analysis := &LogAnalysis{
        TotalEntries: len(entries),
        TimeRange: TimeRange{
            Start: entries[0].Timestamp,
            End:   entries[len(entries)-1].Timestamp,
        },
    }

    // Count by level
    errorMessages := []string{}
    for _, entry := range entries {
        switch strings.ToUpper(entry.Level) {
        case "ERROR", "FATAL":
            analysis.ErrorCount++
            errorMessages = append(errorMessages, entry.Message)
        case "WARN", "WARNING":
            analysis.WarningCount++
        }
    }

    // Find error patterns
    analysis.TopErrors = findErrorPatterns(errorMessages)

    // Use AI for deeper analysis
    if err := la.performAIAnalysis(entries, analysis); err != nil {
        return nil, fmt.Errorf("AI analysis failed: %w", err)
    }

    return analysis, nil
}

func findErrorPatterns(messages []string) []ErrorPattern {
    patternCounts := make(map[string]int)
    patternExamples := make(map[string]string)
    
    for _, msg := range messages {
        // Normalize error messages by removing specific values
        pattern := normalizeErrorMessage(msg)
        patternCounts[pattern]++
        if patternExamples[pattern] == "" {
            patternExamples[pattern] = msg
        }
    }
    
    // Sort by frequency
    type kv struct {
        Pattern string
        Count   int
    }
    
    var sorted []kv
    for k, v := range patternCounts {
        sorted = append(sorted, kv{k, v})
    }
    
    sort.Slice(sorted, func(i, j int) bool {
        return sorted[i].Count > sorted[j].Count
    })
    
    var result []ErrorPattern
    for i, kv := range sorted {
        if i >= 10 { // Top 10 patterns
            break
        }
        result = append(result, ErrorPattern{
            Pattern: kv.Pattern,
            Count:   kv.Count,
            Example: patternExamples[kv.Pattern],
        })
    }
    
    return result
}

func normalizeErrorMessage(msg string) string {
    // Replace common variable patterns
    re1 := regexp.MustCompile(`\d+`)
    re2 := regexp.MustCompile(`[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}`)
    re3 := regexp.MustCompile(`\b\w+@\w+\.\w+\b`)
    
    normalized := re1.ReplaceAllString(msg, "XXX")
    normalized = re2.ReplaceAllString(normalized, "UUID")
    normalized = re3.ReplaceAllString(normalized, "EMAIL")
    
    return normalized
}

func (la *LogAnalyzer) performAIAnalysis(entries []LogEntry, analysis *LogAnalysis) error {
    // Prepare sample of entries for AI analysis
    sampleSize := 50
    if len(entries) < sampleSize {
        sampleSize = len(entries)
    }
    
    sample := entries[len(entries)-sampleSize:] // Last N entries
    
    template := prompts.NewPromptTemplate(`
You are an expert system administrator analyzing application logs. Based on the log data provided, identify:

1. **Anomalies**: Unusual patterns, spikes, or unexpected behaviors
2. **Recommendations**: Specific actions to improve system reliability
3. **Critical Issues**: Problems requiring immediate attention

Log Summary:
- Total Entries: {{.total_entries}}
- Errors: {{.error_count}}
- Warnings: {{.warning_count}}
- Time Range: {{.time_range}}

Top Error Patterns:
{{range .top_errors}}
- {{.Pattern}} ({{.Count}} occurrences)
{{end}}

Recent Log Sample:
{{range .sample}}
{{.timestamp}} [{{.level}}] {{.message}}
{{end}}

Respond in JSON format:
{
  "anomalies": [
    {
      "type": "error_spike|performance|security|other",
      "description": "What was detected",
      "severity": "critical|high|medium|low",
      "examples": ["example log entries"]
    }
  ],
  "recommendations": [
    "Specific actionable recommendations"
  ]
}`, []string{"total_entries", "error_count", "warning_count", "time_range", "top_errors", "sample"})

    sampleData := make([]map[string]string, len(sample))
    for i, entry := range sample {
        sampleData[i] = map[string]string{
            "timestamp": entry.Timestamp.Format(time.RFC3339),
            "level":     entry.Level,
            "message":   entry.Message,
        }
    }

    prompt, err := template.Format(map[string]any{
        "total_entries": analysis.TotalEntries,
        "error_count":   analysis.ErrorCount,
        "warning_count": analysis.WarningCount,
        "time_range":    fmt.Sprintf("%s to %s", analysis.TimeRange.Start.Format(time.RFC3339), analysis.TimeRange.End.Format(time.RFC3339)),
        "top_errors":    analysis.TopErrors,
        "sample":        sampleData,
    })
    if err != nil {
        return fmt.Errorf("formatting prompt: %w", err)
    }

    ctx := context.Background()
    response, err := la.llm.GenerateContent(ctx, []llms.MessageContent{
        llms.TextParts(llms.ChatMessageTypeHuman, prompt),
    }, llms.WithJSONMode())
    if err != nil {
        return fmt.Errorf("generating analysis: %w", err)
    }

    var aiResult struct {
        Anomalies       []Anomaly `json:"anomalies"`
        Recommendations []string  `json:"recommendations"`
    }

    if err := json.Unmarshal([]byte(response.Choices[0].Content), &aiResult); err != nil {
        return fmt.Errorf("parsing AI response: %w", err)
    }

    analysis.Anomalies = aiResult.Anomalies
    analysis.Recommendations = aiResult.Recommendations

    return nil
}

func (la *LogAnalysis) PrintReport() {
    fmt.Printf("📊 Log Analysis Report\n")
    fmt.Printf("=====================\n\n")
    
    fmt.Printf("📈 Summary:\n")
    fmt.Printf("  Total Entries: %d\n", la.TotalEntries)
    fmt.Printf("  Errors: %d\n", la.ErrorCount)
    fmt.Printf("  Warnings: %d\n", la.WarningCount)
    fmt.Printf("  Time Range: %s to %s\n\n", 
        la.TimeRange.Start.Format("2006-01-02 15:04:05"),
        la.TimeRange.End.Format("2006-01-02 15:04:05"))
    
    if len(la.TopErrors) > 0 {
        fmt.Printf("🔴 Top Error Patterns:\n")
        for i, pattern := range la.TopErrors {
            if i >= 5 { break }
            fmt.Printf("  %d. %s (%d occurrences)\n", i+1, pattern.Pattern, pattern.Count)
        }
        fmt.Println()
    }
    
    if len(la.Anomalies) > 0 {
        fmt.Printf("⚠️  Detected Anomalies:\n")
        for _, anomaly := range la.Anomalies {
            fmt.Printf("  %s - %s (%s)\n", anomaly.Type, anomaly.Description, anomaly.Severity)
        }
        fmt.Println()
    }
    
    if len(la.Recommendations) > 0 {
        fmt.Printf("💡 Recommendations:\n")
        for i, rec := range la.Recommendations {
            fmt.Printf("  %d. %s\n", i+1, rec)
        }
        fmt.Println()
    }
}

func main() {
    var (
        file   = flag.String("file", "", "Log file to analyze")
        output = flag.String("output", "", "Output file for JSON report")
        watch  = flag.Bool("watch", false, "Watch file for changes")
    )
    flag.Parse()

    if *file == "" {
        fmt.Println("Usage: log-analyzer -file=application.log")
        os.Exit(1)
    }

    analyzer, err := NewLogAnalyzer()
    if err != nil {
        log.Fatal(err)
    }

    if *watch {
        // Watch mode - simplified version
        fmt.Printf("👀 Watching %s for changes...\n", *file)
        for {
            if err := analyzeFile(analyzer, *file, *output); err != nil {
                log.Printf("Analysis error: %v", err)
            }
            time.Sleep(30 * time.Second)
        }
    } else {
        if err := analyzeFile(analyzer, *file, *output); err != nil {
            log.Fatal(err)
        }
    }
}

func analyzeFile(analyzer *LogAnalyzer, filename, outputFile string) error {
    fmt.Printf("🔍 Analyzing %s...\n", filename)
    
    entries, err := analyzer.ParseLogFile(filename)
    if err != nil {
        return fmt.Errorf("parsing log file: %w", err)
    }

    analysis, err := analyzer.AnalyzeLogs(entries)
    if err != nil {
        return fmt.Errorf("analyzing logs: %w", err)
    }

    analysis.PrintReport()

    if outputFile != "" {
        data, err := json.MarshalIndent(analysis, "", "  ")
        if err != nil {
            return fmt.Errorf("marshaling report: %w", err)
        }
        
        if err := os.WriteFile(outputFile, data, 0644); err != nil {
            return fmt.Errorf("writing report: %w", err)
        }
        fmt.Printf("📄 Report saved to %s\n", outputFile)
    }

    return nil
}
```

## Step 3: Create sample log file

Create `sample.log` for testing:

```
2024-01-15 10:30:01 [INFO] Application started successfully
2024-01-15 10:30:02 [INFO] Database connection established
2024-01-15 10:30:15 [ERROR] Failed to process user request: invalid email format user@
2024-01-15 10:30:16 [WARN] High memory usage detected: 85%
2024-01-15 10:30:17 [ERROR] Database timeout after 30s
2024-01-15 10:30:18 [ERROR] Failed to process user request: invalid email format admin@
2024-01-15 10:30:19 [INFO] Request processed successfully
2024-01-15 10:30:25 [ERROR] Database timeout after 30s
2024-01-15 10:30:30 [FATAL] Out of memory error - application terminating
2024-01-15 10:30:31 [INFO] Application shutdown initiated
```

## Step 4: Run the analyzer

```bash
export OPENAI_API_KEY="your-openai-api-key-here"
go run main.go -file=sample.log -output=report.json
```

## Step 5: Enhanced real-time monitoring

Create `monitor.go` for continuous monitoring:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/fsnotify/fsnotify"
    "github.com/tmc/langchaingo/llms"
    "github.com/tmc/langchaingo/chains"
)

type LogMonitor struct {
    analyzer    *LogAnalyzer
    watcher     *fsnotify.Watcher
    alertChain  chains.Chain
    thresholds  MonitoringThresholds
}

type MonitoringThresholds struct {
    ErrorsPerMinute   int
    CriticalKeywords  []string
    ResponseTimeLimit time.Duration
}

func NewLogMonitor(analyzer *LogAnalyzer) (*LogMonitor, error) {
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        return nil, err
    }

    // Create alert chain for notifications
    alertChain := chains.NewLLMChain(analyzer.llm, prompts.NewPromptTemplate(`
Generate a concise alert message for this log analysis:

{{.analysis}}

Format as: [SEVERITY] Brief description - Action needed
Keep under 140 characters.`, []string{"analysis"}))

    return &LogMonitor{
        analyzer:   analyzer,
        watcher:    watcher,
        alertChain: alertChain,
        thresholds: MonitoringThresholds{
            ErrorsPerMinute:   10,
            CriticalKeywords:  []string{"fatal", "out of memory", "database down"},
            ResponseTimeLimit: 5 * time.Second,
        },
    }, nil
}

func (lm *LogMonitor) Start(filename string) error {
    err := lm.watcher.Add(filename)
    if err != nil {
        return err
    }

    fmt.Printf("🚨 Monitoring %s for critical issues...\n", filename)

    for {
        select {
        case event, ok := <-lm.watcher.Events:
            if !ok {
                return nil
            }
            if event.Op&fsnotify.Write == fsnotify.Write {
                go lm.checkForAlerts(filename)
            }
        case err, ok := <-lm.watcher.Errors:
            if !ok {
                return nil
            }
            log.Printf("Watcher error: %v", err)
        }
    }
}

func (lm *LogMonitor) checkForAlerts(filename string) {
    // Read last N lines and check for critical issues
    entries, err := lm.analyzer.ParseLogFile(filename)
    if err != nil {
        log.Printf("Error parsing file: %v", err)
        return
    }

    // Check recent entries (last minute)
    recent := lm.getRecentEntries(entries, time.Minute)
    if lm.shouldAlert(recent) {
        analysis, err := lm.analyzer.AnalyzeLogs(recent)
        if err != nil {
            log.Printf("Error analyzing logs: %v", err)
            return
        }

        alert, err := chains.Run(context.Background(), lm.alertChain, 
            fmt.Sprintf("Analysis: %+v", analysis))
        if err != nil {
            log.Printf("Error generating alert: %v", err)
            return
        }

        fmt.Printf("🚨 ALERT: %s\n", alert)
        // Here you would send to Slack, email, etc.
    }
}

func (lm *LogMonitor) getRecentEntries(entries []LogEntry, duration time.Duration) []LogEntry {
    cutoff := time.Now().Add(-duration)
    var recent []LogEntry
    
    for i := len(entries) - 1; i >= 0; i-- {
        if entries[i].Timestamp.Before(cutoff) {
            break
        }
        recent = append([]LogEntry{entries[i]}, recent...)
    }
    
    return recent
}

func (lm *LogMonitor) shouldAlert(entries []LogEntry) bool {
    errorCount := 0
    for _, entry := range entries {
        if entry.Level == "ERROR" || entry.Level == "FATAL" {
            errorCount++
        }
        
        // Check for critical keywords
        for _, keyword := range lm.thresholds.CriticalKeywords {
            if strings.Contains(strings.ToLower(entry.Message), keyword) {
                return true
            }
        }
    }
    
    return errorCount >= lm.thresholds.ErrorsPerMinute
}
```

## Step 6: Integration with observability tools

Create `integrations.go`:

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
)

type SlackAlert struct {
    Text string `json:"text"`
}

func (lm *LogMonitor) sendSlackAlert(message string, webhookURL string) error {
    alert := SlackAlert{Text: fmt.Sprintf("Log Alert: %s", message)}
    
    jsonData, err := json.Marshal(alert)
    if err != nil {
        return err
    }
    
    resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    return nil
}

// Prometheus metrics
type MetricsCollector struct {
    errorCount   int
    warningCount int
}

func (mc *MetricsCollector) UpdateFromAnalysis(analysis *LogAnalysis) {
    mc.errorCount += analysis.ErrorCount
    mc.warningCount += analysis.WarningCount
}

// Export to Prometheus format
func (mc *MetricsCollector) PrometheusMetrics() string {
    return fmt.Sprintf(`
# HELP log_errors_total Total number of error log entries
# TYPE log_errors_total counter
log_errors_total %d

# HELP log_warnings_total Total number of warning log entries  
# TYPE log_warnings_total counter
log_warnings_total %d
`, mc.errorCount, mc.warningCount)
}
```

## Use cases

This log analyzer can be used for:

1. **Production Monitoring**: Detect issues before they become critical
2. **Incident Response**: Quickly understand what went wrong
3. **Performance Analysis**: Identify slow queries and bottlenecks
4. **Security Monitoring**: Detect suspicious patterns
5. **Capacity Planning**: Understand usage patterns and growth

## Advanced features

- **Machine Learning**: Train models on historical log patterns
- **Correlation Analysis**: Link errors across multiple services
- **Predictive Alerts**: Warn before issues occur
- **Custom Dashboards**: Visual representations of log data
- **Automated Remediation**: Trigger fixes for known issues

This tutorial demonstrates how LangChainGo can power sophisticated operational tools that provide real business value!