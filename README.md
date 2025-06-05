# Tips CLI

A command-line tool for generating, storing, and displaying helpful tips on various topics using AI language models.

## Features

- **AI-Powered Generation**: Generate tips using OpenAI, Anthropic, or Google AI models
- **Interactive Display**: Browse tips in an interactive terminal interface
- **Topic Filtering**: Organize and filter tips by topic
- **Local Storage**: Tips are stored locally in JSON format
- **Auto-refresh**: Automatically cycle through tips at configurable intervals
- **Mark as Known**: Remove tips you've learned to focus on new content

## Prerequisites

- Go 1.19 or later
- API key for one of the supported LLM providers

## Installation

### From Source

```bash
git clone https://github.com/yourusername/tips
cd tips
go build -o tips
sudo mv tips /usr/local/bin/
```

### Using Homebrew (Coming Soon)

```bash
brew install tips
```

## Configuration

Set your API key as an environment variable:

```bash
# For OpenAI (default)
export OPENAI_API_KEY="your-openai-api-key"

# For Anthropic
export ANTHROPIC_API_KEY="your-anthropic-api-key"

# For Google AI
export GOOGLE_API_KEY="your-google-api-key"
```

Optionally configure the model to use:

```bash
# Default: openai/gpt-4o
export TIPS_MODEL="anthropic/claude-3-sonnet-20240229"
export TIPS_MODEL="google/gemini-pro"
```

## Usage

### Generate Tips
Generate new tips for specific topics:

```bash
# Generate 20 programming tips (default count)
./tips generate -t programming

# Generate 10 cooking tips
./tips generate -t cooking -c 10

# Generate tips for multiple topics
./tips generate -t "go programming" -t "web development" -c 5
```

### Display Tips
Show tips with automatic refresh:

```bash
# Show random tips, refresh every 60 minutes (default)
./tips

# Show only programming tips
./tips -t programming

# Show tips and refresh every 30 minutes
./tips -r 30

# Show tips from multiple topics
./tips -t programming -t cooking
```

### Clear Tips

Remove all stored tips from local storage:

```bash
# Delete all tips
./tips clear
```

This permanently deletes the `~/.tips.json` file and all stored tips.

### Interactive Controls
While viewing tips:
- Press `n` to immediately show the next tip
- Press `k` to mark the current tip as "known" (removes it permanently)
- Press `q` to quit
- Tips automatically refresh based on the interval you set

## Command Reference

```
tips [OPTIONS] [COMMAND]

Commands:
  show     Display tips (default command)
  generate Generate new tips for a topic
  clear    Delete all stored tips
  
Options:
  -t, --topic    Filter by topic (can specify multiple)
  -r, --refresh  Refresh interval in minutes (default: 60)
  -c, --count    Number of tips to generate per API call (default: 20)
```

## Examples

```bash
# Basic usage
./tips                                    # Show random tip, refresh every 60 min
./tips -t programming                     # Show programming tips only
./tips -r 30                              # Refresh every 30 minutes
./tips generate -t "cooking"              # Generate 20 new cooking tips
./tips generate -t "go" -c 10             # Generate 10 new Go programming tips
./tips clear                              # Delete all stored tips

# Advanced usage
./tips -t programming -t cooking -r 15    # Show programming and cooking tips, refresh every 15 min
./tips generate -t "machine learning" -t "data science" -c 15  # Generate 15 tips for each topic
```

## Data Storage

Tips are stored locally in `~/.tips.json` with the following structure:

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

## Model Configuration

The tool supports multiple AI providers through the `TIPS_MODEL` environment variable:

### Supported Providers and Models

**OpenAI (default)**
- `openai/gpt-4o` (default)
- `openai/gpt-4o-mini`
- `openai/gpt-3.5-turbo`

**Anthropic**
- `anthropic/claude-3-5-sonnet-20241022`
- `anthropic/claude-3-5-haiku-20241022`
- `anthropic/claude-3-opus-20240229`

**Google**
- `google/gemini-2.5-flash`
- `google/gemini-1.5-pro`

### Environment Variables Required

Make sure you have the appropriate API key set:
- OpenAI: `OPENAI_API_KEY`
- Anthropic: `ANTHROPIC_API_KEY`  
- Google: `GOOGLE_API_KEY`

The tool will automatically detect which provider you're using based on the `TIPS_MODEL` format and check for the corresponding API key.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.
