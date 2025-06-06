# API Integration Testing

This document explains how to run integration tests that make real API calls to verify the tips CLI works with actual LLM providers.

## ⚠️ Important Notes

- **Costs**: These tests make real API calls and will incur small costs (~$0.10 per full test run)
- **Rate Limits**: Tests include built-in rate limiting and retry logic
- **API Keys**: You need valid API keys for at least one provider

## Quick Start

1. **Set up API keys** (choose one or more):
   ```bash
   export OPENAI_API_KEY="your-openai-key"
   export ANTHROPIC_API_KEY="your-anthropic-key" 
   export GOOGLE_API_KEY="your-google-key"
   ```

2. **Run the tests**:
   ```bash
   ./scripts/run-api-tests.sh
   ```

3. **Or run manually**:
   ```bash
   go test -tags="integration,api" -v ./...
   ```

## Configuration

### Environment Variables

Copy `.env.example` to `.env` and configure:

```bash
# Required: At least one API key
OPENAI_API_KEY=your_openai_api_key_here
ANTHROPIC_API_KEY=your_anthropic_api_key_here  
GOOGLE_API_KEY=your_google_api_key_here

# Optional: Test limits
TEST_MAX_COST_USD=0.10
TEST_MAX_RETRIES=3
TEST_RATE_LIMIT_SECONDS=1
```

### Cost Control

Tests are designed with cost control in mind:

- **Maximum cost per run**: $0.10 USD (configurable)
- **Minimal requests**: Only 2 tips generated per provider
- **Short prompts**: Uses simple test topics like "testing"
- **Estimated costs**:
  - OpenAI (GPT-3.5): ~$0.004 per test
  - Anthropic (Claude): ~$0.016 per test  
  - Google (Gemini): ~$0.002 per test

## Test Types

### Basic Integration Tests
```bash
# Test all available providers
go test -tags="integration,api" -run TestRealAPIIntegration -v

# Test specific provider only
ANTHROPIC_API_KEY="" GOOGLE_API_KEY="" go test -tags="integration,api" -run TestRealAPIIntegration -v
```

### Error Handling Tests
```bash
go test -tags="integration,api" -run TestAPIErrorHandling -v
```

## CI/CD Integration

### GitHub Actions

API tests run automatically:

1. **On schedule**: Weekly (to catch API changes)
2. **On demand**: Include `[test-api]` in your commit message
3. **Manual trigger**: Use GitHub Actions UI

### Setting up API Keys in GitHub

1. Go to your repository Settings > Secrets and variables > Actions
2. Add these secrets (optional - tests will skip if not present):
   - `OPENAI_API_KEY`
   - `ANTHROPIC_API_KEY`  
   - `GOOGLE_API_KEY`

## Safety Features

### Rate Limiting
- **1 second delay** between requests
- **Exponential backoff** on retries
- **Jitter** to avoid thundering herd

### Retry Logic
- **3 retry attempts** for failed requests
- **Automatic retry** for rate limits, timeouts, server errors
- **No retry** for authentication errors, invalid requests

### Cost Protection
- **Pre-flight cost estimation** before each request
- **Configurable spending limits** per test run
- **Token count monitoring**

### Request Timeouts
- **30 second timeout** per individual request
- **10 minute timeout** for entire test suite

## Troubleshooting

### Common Issues

**"No API keys found"**
```bash
# Make sure at least one key is set
echo $OPENAI_API_KEY
export OPENAI_API_KEY="your-key-here"
```

**"Rate limit exceeded"**
```bash
# Tests automatically retry with backoff
# If persistent, increase rate limit delay:
export TEST_RATE_LIMIT_SECONDS=5
```

**"Cost limit exceeded"**
```bash
# Increase the cost limit if needed:
export TEST_MAX_COST_USD=0.25
```

**Tests timing out**
```bash
# Run with longer timeout:
go test -tags="integration,api" -timeout=15m -v ./...
```

### Debug Mode

Enable verbose logging:
```bash
# Set log level for debugging
export LOG_LEVEL=debug
./scripts/run-api-tests.sh
```

## Manual Testing

For manual verification without automated tests:

```bash
# Test OpenAI provider
export OPENAI_API_KEY="your-key"
./tips generate -p openai -t "testing" -c 2

# Test Anthropic provider  
export ANTHROPIC_API_KEY="your-key"
./tips generate -p anthropic -t "testing" -c 2

# Test Google provider
export GOOGLE_API_KEY="your-key" 
./tips generate -p google -t "testing" -c 2
```

## Best Practices

1. **Use test API keys** if available (some providers offer test/sandbox keys)
2. **Monitor costs** - check your provider dashboards regularly
3. **Run locally first** before pushing to CI
4. **Limit test frequency** - don't run on every commit
5. **Use minimal test data** - small prompts and low token counts

## Contributing

When adding new API tests:

1. **Follow cost limits** - keep requests minimal
2. **Add retry logic** for new error types  
3. **Update cost estimates** when providers change pricing
4. **Test locally** before submitting PR
5. **Document new test cases** in this README