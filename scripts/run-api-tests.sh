#!/bin/bash
set -e

# Script to run API integration tests with proper setup and safety measures

echo "🧪 API Integration Test Runner"
echo "=============================="

# Check if API keys are available
has_openai_key=false
has_anthropic_key=false
has_google_key=false

if [ ! -z "$OPENAI_API_KEY" ]; then
    has_openai_key=true
    echo "✅ OpenAI API key found"
fi

if [ ! -z "$ANTHROPIC_API_KEY" ]; then
    has_anthropic_key=true
    echo "✅ Anthropic API key found"
fi

if [ ! -z "$GOOGLE_API_KEY" ]; then
    has_google_key=true
    echo "✅ Google API key found"
fi

if [ "$has_openai_key" = false ] && [ "$has_anthropic_key" = false ] && [ "$has_google_key" = false ]; then
    echo "❌ No API keys found. Please set at least one of:"
    echo "   - OPENAI_API_KEY"
    echo "   - ANTHROPIC_API_KEY" 
    echo "   - GOOGLE_API_KEY"
    echo ""
    echo "Example:"
    echo "   export OPENAI_API_KEY=\"your-key-here\""
    echo "   ./scripts/run-api-tests.sh"
    exit 1
fi

# Confirm with user before running (to avoid accidental costs)
echo ""
echo "⚠️  WARNING: These tests will make real API calls and may incur costs!"
echo "   - Tests are designed to be minimal (2 tips per provider)"
echo "   - Rate limiting is implemented to avoid hitting limits"
echo "   - Retry logic handles temporary failures"

if [ "$CI" != "true" ]; then
    echo ""
    read -p "Do you want to continue? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Cancelled by user"
        exit 1
    fi
fi

echo ""
echo "🏃 Running API integration tests..."

# Set test timeout (10 minutes max)
export TEST_TIMEOUT="10m"

# Run the tests with specific tags
go test -tags="integration,api" -timeout="$TEST_TIMEOUT" -v ./... 2>&1 | tee api_test_results.log

# Check test results
if [ ${PIPESTATUS[0]} -eq 0 ]; then
    echo ""
    echo "✅ All API integration tests passed!"
    echo "📄 Test results saved to: api_test_results.log"
else
    echo ""
    echo "❌ Some API integration tests failed!"
    echo "📄 Check api_test_results.log for details"
    exit 1
fi