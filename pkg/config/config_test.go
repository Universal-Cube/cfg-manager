package config

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	m := New()
	assert.NotNil(t, m, "New should return a non-nil manager")
	assert.NotNil(t, m.data, "data map should be initialized")
	assert.Truef(t, m.caseSensitive, "caseSensitive should be true by default")

	m = New(WithCaseSensitive(false))
	assert.Falsef(t, m.caseSensitive, "caseSensitive should be false")
}

func TestManager_Load(t *testing.T) {
	jsonContent := `{
		"app": {
			"name": "TestApp",
			"version": "1.0.0"
		},
		"server": {
			"port": 8080
		}
	}`

	m := New()
	err := m.Load(strings.NewReader(jsonContent), FormatJSON)
	require.NoError(t, err, "Load should not return an error for valid JSON")

	name, err := m.GetString("app.name")
	require.NoError(t, err, "GetString should not return an error for valid key")
	assert.Equal(t, "TestApp", name, "GetString should return the correct value")

	port, err := m.GetInt("server.port")
	require.NoError(t, err)
	assert.Equal(t, 8080, port)

	yamlContent := `app:
  name: YAMLApp
  version: 2.0.0
server:
  port: 9090
  host: localhost`

	m = New()
	err = m.Load(strings.NewReader(yamlContent), FormatYAML)
	require.NoError(t, err, "Load should not return an error for valid YAML")

	name, err = m.GetString("app.name")
	require.NoError(t, err, "GetString should not return an error for valid key")
	assert.Equal(t, "YAMLApp", name, "GetString should return the correct value")

	port, err = m.GetInt("server.port")
	require.NoError(t, err, "GetInt should not return an error for valid key")
	assert.Equal(t, 9090, port, "GetInt should return the correct value")

	host, err := m.GetString("server.host")
	require.NoError(t, err, "GetString should not return an error for valid key")
	assert.Equal(t, "localhost", host, "GetString should return the correct value")

	err = m.Load(strings.NewReader("invalid content"), Format("invalid"))
	assert.Error(t, err, "Load should return an error for invalid format")
}
