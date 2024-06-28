package integration

import (
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"
)

const baseAddress = "http://balancer:8090"

var client = http.Client{
	Timeout: 3 * time.Second,
}

func TestBalancer(t *testing.T) {
    if _, exists := os.LookupEnv("INTEGRATION_TEST"); !exists {
        t.Skip("Integration test is not enabled")
    }

    // Кількість запитів для тестування
    numRequests := 100

    // Мапа для підрахунку кількості запитів до кожного сервера
    serverCounts := make(map[string]int)

    for i := 0; i < numRequests; i++ {
        // Генеруємо унікальний шлях для кожного запиту
        path := fmt.Sprintf("/api/v1/some-data/%d", i)
        
        resp, err := client.Get(fmt.Sprintf("%s%s", baseAddress, path))
        if err != nil {
            t.Errorf("Request failed: %v", err)
            continue
        }
        defer resp.Body.Close()

        if resp.StatusCode != http.StatusOK {
            t.Errorf("Unexpected status code: %d", resp.StatusCode)
        }

        server := resp.Header.Get("lb-from")
        if server == "" {
            t.Errorf("Missing lb-from header")
        } else {
            serverCounts[server]++
        }

        t.Logf("Request %d: response from [%s]", i, server)
    }

    // Перевіряємо розподіл запитів між серверами
    t.Logf("Server distribution: %v", serverCounts)
    
    if len(serverCounts) != 3 {
        t.Errorf("Expected requests to be distributed among 3 servers, but got %d", len(serverCounts))
    }

    // Перевіряємо, чи розподіл більш-менш рівномірний
    expectedCount := numRequests / 3
    for server, count := range serverCounts {
        if count < expectedCount/2 || count > expectedCount*2 {
            t.Errorf("Uneven distribution for server %s: %d requests (expected around %d)", server, count, expectedCount)
        }
    }
}
