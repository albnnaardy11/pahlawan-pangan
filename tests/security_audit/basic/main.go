package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// PahlawanSecurityAudit simulates real-world attack vectors on the API.
// Based on OWASP Top 10 and common microservice vulnerabilities.
func main() {
	fmt.Println("🛡️  [SECURITY AUDIT] Starting Penetration Test Simulation...")
	baseURL := "http://localhost:8080/api/v1"

	var wg sync.WaitGroup

	// 1. RATE LIMITING ATTACK (DoS Simulation)
	// Attempt to flood the API with requests to trigger the Rate Limiter/Load Shedder.
	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Println("\n🔥 [ATTACK] Launching Rate Limit Flood (DoS Simulation)...")
		successCount := 0
		blockedCount := 0
		client := &http.Client{Timeout: 2 * time.Second}

		for i := 0; i < 100; i++ {
			resp, err := client.Get(baseURL + "/marketplace?lat=-6.2&lon=106.8")
			if err != nil {
				// Likely connection reset or timeout (Load Shedder working)
				blockedCount++
				continue
			}
			defer resp.Body.Close()
			if resp.StatusCode == 429 {
				blockedCount++ // 429 Too Many Requests
			} else if resp.StatusCode == 503 {
				blockedCount++ // 503 Service Unavailable (Load Shedder)
			} else {
				successCount++
			}
			time.Sleep(10 * time.Millisecond) // Fast fire
		}
		fmt.Printf("   -> Results: %d Requests Passed, %d Requests BLOCKED/SHED. (Target: >0 Blocked)\n", successCount, blockedCount)
	}()

	// 2. SQL INJECTION (Validation & Sanitization Test)
	// Attempt to bypass auth or extract data via SQLi patterns in query params.
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(500 * time.Millisecond) // Wait for Flood to stabilize
		fmt.Println("\n💉 [ATTACK] Injecting SQL Payload in Parameters...")
		
		// Payload: ' OR '1'='1
		sqliURL := baseURL + "/marketplace?lat=-6.2&lon=106.8' OR '1'='1"
		resp, err := http.Get(sqliURL)
		if err != nil {
			fmt.Printf("   -> [ERROR] Connection failed: %v\n", err)
			return
		}
		defer resp.Body.Close()
		
		body, _ := io.ReadAll(resp.Body)
		if strings.Contains(string(body), "postgres") || strings.Contains(string(body), "syntax error") {
			fmt.Println("   -> ❌ VULNERABLE! Database error leaked in response.")
		} else if resp.StatusCode == 400 || resp.StatusCode == 422 || resp.StatusCode == 500 {
			fmt.Println("   -> ✅ SAFE. Request rejected or handled gracefully.")
		} else {
			fmt.Printf("   -> ⚠️  Check Manual. Status: %d\n", resp.StatusCode)
		}
	}()

	// 3. CHAOS MONKEY (Resilience Test)
	// Trigger the Chaos Middleware to simulate random failures.
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(1 * time.Second)
		fmt.Println("\n🐵 [ATTACK] Triggering Chaos Monkey (Simulated Network Failure)...")
		
		req, _ := http.NewRequest("GET", baseURL+"/marketplace?lat=-6.2&lon=106.8", nil)
		req.Header.Set("X-Chaos-Simulate", "true")
		req.Header.Set("X-Chaos-Failure-Rate", "1.0") // 100% Failure Rate
		
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("   -> [INFO] Connection dropped (Expected behavior).")
			return
		}
		defer resp.Body.Close()
		
		if resp.StatusCode == 500 {
			fmt.Println("   -> ✅ RESILIENCE VERIFIED. Ops/SRE alerts should be firing now.")
		} else {
			fmt.Printf("   -> ❌ Chaos Failed. Status: %d\n", resp.StatusCode)
		}
	}()
	
	// 4. XSS / SCRIPT INJECTION (Input Sanitization)
	// Attempt to store a script in a text field (e.g., surplus food_type)
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(1500 * time.Millisecond)
		fmt.Println("\n🦠 [ATTACK] Attempting Stored XSS Injection...")
		
		payload := map[string]interface{}{
			"id":             "xss-test-1",
			"provider_id":    "hacker-001",
			"food_type":      "<script>alert('pwned')</script>", // Malicious Script
			"quantity_kgs":   10.5,
			"original_price": 50000,
			"discount_price": 10000,
			"status":         "available",
			"expiry_time":    time.Now().Add(24 * time.Hour),
			"lat":            -6.2,
			"lon":            106.8,
		}
		jsonPayload, _ := json.Marshal(payload)
		
		resp, err := http.Post(baseURL+"/surplus", "application/json", bytes.NewBuffer(jsonPayload))
		if err != nil {
			fmt.Printf("   -> [ERROR] Connection failed: %v\n", err)
			return
		}
		defer resp.Body.Close()
		
		if resp.StatusCode == 422 || resp.StatusCode == 400 {
			fmt.Println("   -> ✅ SAFE. Input validation blocked the payload.")
		} else if resp.StatusCode == 201 {
			fmt.Println("   -> ⚠️  WARNING. Payload accepted. Check Output Encoding in Frontend!")
		} else {
			fmt.Printf("   -> ℹ️  Response Status: %d\n", resp.StatusCode)
		}
	}()

	wg.Wait()
	fmt.Println("\n🏁 [AUDIT] Penetration Test Complete.")
}
