package taskname_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/coder/coder/v2/coderd/taskname"
	"github.com/coder/coder/v2/codersdk"
	"github.com/coder/coder/v2/testutil"
)

const (
	anthropicAPIKeyEnvVar    = "ANTHROPIC_API_KEY"
	anthropicAuthTokenEnvVar = "ANTHROPIC_AUTH_TOKEN"
	anthropicBaseURLEnvVar   = "ANTHROPIC_BASE_URL"
)

func TestGenerateFallback(t *testing.T) {
	t.Parallel()

	name := taskname.GenerateFallback()
	err := codersdk.NameValid(name)
	require.NoErrorf(t, err, "expected fallback to be valid workspace name, instead found %s", name)
}

func TestGetAnthropicAPIKeyFromEnv(t *testing.T) {
	// Note: Cannot use t.Parallel() with t.Setenv()

	t.Run("PreferAPIKey", func(t *testing.T) {
		// Set both environment variables
		t.Setenv(anthropicAPIKeyEnvVar, "test-api-key")
		t.Setenv(anthropicAuthTokenEnvVar, "test-auth-token")

		// Should prefer ANTHROPIC_API_KEY
		key := taskname.GetAnthropicAPIKeyFromEnv()
		require.Equal(t, "test-api-key", key)
	})

	t.Run("FallbackToAuthToken", func(t *testing.T) {
		// Only set ANTHROPIC_AUTH_TOKEN
		t.Setenv(anthropicAuthTokenEnvVar, "test-auth-token")

		// Should use ANTHROPIC_AUTH_TOKEN when API_KEY is not set
		key := taskname.GetAnthropicAPIKeyFromEnv()
		require.Equal(t, "test-auth-token", key)
	})

	t.Run("NoKeys", func(t *testing.T) {
		// Don't set any environment variables
		key := taskname.GetAnthropicAPIKeyFromEnv()
		require.Equal(t, "", key)
	})
}

func TestGetAnthropicBaseURLFromEnv(t *testing.T) {
	// Note: Cannot use t.Parallel() with t.Setenv()

	t.Run("CustomBaseURL", func(t *testing.T) {
		t.Setenv(anthropicBaseURLEnvVar, "https://custom.api.example.com")

		baseURL := taskname.GetAnthropicBaseURLFromEnv()
		require.Equal(t, "https://custom.api.example.com", baseURL)
	})

	t.Run("NoBaseURL", func(t *testing.T) {
		baseURL := taskname.GetAnthropicBaseURLFromEnv()
		require.Equal(t, "", baseURL)
	})
}

func TestGenerateTaskName(t *testing.T) {
	t.Parallel()

	t.Run("Fallback", func(t *testing.T) {
		t.Parallel()

		ctx := testutil.Context(t, testutil.WaitShort)

		name, err := taskname.Generate(ctx, "Some random prompt")
		require.ErrorIs(t, err, taskname.ErrNoAPIKey)
		require.Equal(t, "", name)
	})

	t.Run("Anthropic", func(t *testing.T) {
		t.Parallel()

		// Try ANTHROPIC_API_KEY first, then ANTHROPIC_AUTH_TOKEN
		apiKey := os.Getenv(anthropicAPIKeyEnvVar)
		if apiKey == "" {
			apiKey = os.Getenv(anthropicAuthTokenEnvVar)
		}
		if apiKey == "" {
			t.Skipf("Skipping test as neither %s nor %s is set", anthropicAPIKeyEnvVar, anthropicAuthTokenEnvVar)
		}

		ctx := testutil.Context(t, testutil.WaitShort)

		name, err := taskname.Generate(ctx, "Create a finance planning app", taskname.WithAPIKey(apiKey))
		require.NoError(t, err)
		require.NotEqual(t, "", name)

		err = codersdk.NameValid(name)
		require.NoError(t, err, "name should be valid")
	})
}
