package cloudprint

import (
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

var apiClient *APIClient

func init() {
	godotenv.Load()
	apiClient = newTestAPIClient()
}

func newTestAPIClient() *APIClient {
	return NewAPIClient(
		os.Getenv("CLIENT_ID"),
		os.Getenv("CLIENT_SECRET"),
		NewMemoryStore(),
	)
}

func TestAPIClient_Authorize(t *testing.T) {
	token := os.Getenv("ACCESS_TOKEN")
	if token != "" {
		apiClient.tokenStore.Set(apiClient.clientID, token)
		return
	}

	err := apiClient.Authorize()
	assert.NoError(t, err)

	token, err = apiClient.tokenStore.Get(apiClient.clientID)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestAPIClient_DeletePrinter(t *testing.T) {
	machineCode := os.Getenv("MACHINE_CODE")

	_, err := apiClient.DeletePrinter(machineCode)
	assert.NoError(t, err)
}

func TestAPIClient_AddPrinter(t *testing.T) {
	machineCode := os.Getenv("MACHINE_CODE")
	msign := os.Getenv("MSIGN")

	_, err := apiClient.AddPrinter(machineCode, msign)
	assert.NoError(t, err)
}

func TestAPIClient_Print(t *testing.T) {
	machineCode := os.Getenv("MACHINE_CODE")
	content := "Across the Great Wall we can reach every corner in the world."

	_, err := apiClient.Print(machineCode, content)
	assert.NoError(t, err)
}

func TestAPIClient_GetPrinterStatus(t *testing.T) {
	machineCode := os.Getenv("MACHINE_CODE")

	_, err := apiClient.GetPrinterStatus(machineCode)
	assert.NoError(t, err)
}
