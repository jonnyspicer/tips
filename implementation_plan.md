# Tips CLI Tool - Go Implementation

## Overview
Create a simple command-line tool in Go that displays helpful tips on various topics and refreshes them periodically. This is an MVP focused on core functionality.

## Core Features

### 1. JSON Data Store
- Use a local `.json` file stored at `~/.tips.json`
- Structure should include:
  - Tip ID (unique identifier)
  - Topic/category
  - Tip content
  - Timestamp when added
- Implement functions to read from and write to this file
- Handle file creation if it doesn't exist

### 2. LLM Integration
- Use Google Gemini 2.0 Flash with structured output
- Function to call LLM with a specific topic and request multiple tips
- Default to 20 tips per API call (user configurable via `--count` flag)
- Use structured output to ensure consistent JSON response format
- Parse LLM response and save new tips to JSON file
- Include error handling for API failures

### 3. Display Tips
- Show a randomly selected tip in the terminal
- Support command-line arguments:
  - `--topic` or `-t`: Filter tips by topic(s)
  - `--refresh` or `-r`: Minutes between tip changes (default: 60)
- Clear terminal and display new tip when refresh interval expires
- Format tip display nicely in terminal

### 4. Mark Tips as Known
- Allow user to press a key (like 'k' or 'd') to mark current tip as "known"
- Remove marked tip from JSON file immediately
- Display a new random tip after removal
- Show confirmation message when tip is removed

## Command Structure
```
tips [OPTIONS] [COMMAND]

Commands:
  show     Display tips (default command)
  generate Generate new tips for a topic
  
Options:
  -t, --topic    Filter by topic (can specify multiple)
  -r, --refresh  Refresh interval in minutes (default: 60)
  -c, --count    Number of tips to generate per API call (default: 20)
  
Examples:
  tips                           # Show random tip, refresh every 60 min
  tips -t programming            # Show programming tips only
  tips -r 30                     # Refresh every 30 minutes
  tips generate -t "cooking"     # Generate 20 new cooking tips
  tips generate -t "go" -c 10    # Generate 10 new Go programming tips
```

## Technical Requirements

### Dependencies
- Use Google AI Go SDK: `google.golang.org/genai`
- Standard Go libraries:
  - `encoding/json` for JSON handling
  - `context` for API calls
  - `flag` or `cobra` for CLI argument parsing
  - `time` for refresh intervals
  - `math/rand` for random selection
  - `os` for file path handling

### Data Structure Example
```json
{
  "tips": [
    {
      "id": "uuid-here",
      "topic": "programming",
      "content": "Use meaningful variable names to make your code self-documenting.",
      "created_at": "2025-06-05T10:00:00Z"
    }
  ]
}
```

### Key Functions Needed
1. `loadTips()` - Read tips from JSON file at `~/.tips.json`
2. `saveTips()` - Write tips to JSON file
3. `generateTips(topic string, count int)` - Call Gemini API for new tips using structured output
4. `getRandomTip(topics []string)` - Get random tip, optionally filtered
5. `removeTip(id string)` - Remove tip from data store
6. `displayTip()` - Format and show tip in terminal
7. `handleUserInput()` - Listen for 'k' key to mark as known

### Gemini API Integration
Use the provided example pattern with structured output:
- Model: `gemini-2.0-flash`
- Response format: JSON array of tip objects
- Schema should define tip structure with required fields
- Environment variable: `GOOGLE_API_KEY`
- Handle API errors and rate limiting gracefully

## Implementation Notes
- Keep the MVP simple - focus on core functionality first
- Use `GOOGLE_API_KEY` environment variable for Gemini API access
- Handle edge cases: empty JSON file, no tips for topic, API errors
- Make the terminal display clean and readable
- Consider using goroutines for non-blocking refresh timer and user input
- Default JSON file location: `~/.tips.json` (expand home directory programmatically)

## Future Enhancements (Not for MVP)
- Multiple data store backends
- Tip categories and difficulty levels
- Statistics tracking
- Import/export functionality
- Web interface