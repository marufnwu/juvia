package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"juvia/internal/apitest/runner"
)

func main() {
	baseURL := flag.String("url", "http://192.168.0.211:18473", "Panel URL")
	username := flag.String("username", "", "Admin username")
	password := flag.String("password", "", "Admin password")
	skipSetup := flag.Bool("skip-setup", false, "Skip setup even if panel is not initialized")
	destructive := flag.Bool("destructive", false, "Allow destructive operations (SSL issue/renew, service restart, updates apply/rollback)")
	flag.Parse()

	if *username == "" {
		*username = os.Getenv("JUVIA_TEST_USERNAME")
	}
	if *password == "" {
		*password = os.Getenv("JUVIA_TEST_PASSWORD")
	}

	if *username == "" || *password == "" {
		fmt.Println("Error: username and password are required")
		fmt.Println("Provide via --username/--password flags or JUVIA_TEST_USERNAME/JUVIA_TEST_PASSWORD env vars")
		flag.Usage()
		os.Exit(1)
	}

	fmt.Printf("Testing panel at: %s\n", *baseURL)
	fmt.Printf("Username: %s\n", *username)
	fmt.Printf("Destructive mode: %v\n", *destructive)

	client := runner.NewAPIClient(*baseURL, *username, *password)

	health, err := client.HealthCheck()
	if err != nil {
		fmt.Printf("❌ Health check failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ Health check: %s\n", health.Body)

	setupRequired, err := client.CheckSetupStatus()
	if err != nil {
		fmt.Printf("⚠️  Could not check setup status: %v\n", err)
	} else if setupRequired && !*skipSetup {
		fmt.Println("Panel requires setup. Running first-run setup...")
		if err := client.FirstRunSetup(*username, *password, *username+"@example.com", "test-server"); err != nil {
			fmt.Printf("⚠️  Setup failed: %v\n", err)
		} else {
			fmt.Println("✅ Setup completed")
		}
	}

	fmt.Println("Logging in...")
	if err := client.Login(); err != nil {
		fmt.Printf("❌ Login failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ Logged in successfully")

	fmt.Println("\nRunning API endpoint tests (stateful)...")
	results, err := client.Run(*destructive)
	if err != nil {
		fmt.Printf("❌ Test run failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\nCleaning up created resources...")
	cleanupResults := client.Cleanup()
	cleanupFailed := 0
	for _, cr := range cleanupResults {
		if !cr.Success {
			cleanupFailed++
		}
	}
	fmt.Printf("✅ Cleanup done: %d ops, %d failed\n", len(cleanupResults), cleanupFailed)

	report := runner.GenerateReport(*baseURL, results, cleanupResults)

	fmt.Printf("\n=== Test Results ===\n")
	fmt.Printf("Total endpoints: %d\n", report.Summary.TotalEndpoints)
	fmt.Printf("Called: %d\n", report.Summary.Called)
	fmt.Printf("Successful: %d\n", report.Summary.Successful)
	fmt.Printf("Failed: %d\n", report.Summary.Failed)
	fmt.Printf("Skipped: %d\n", report.Summary.Skipped)
	fmt.Printf("Agent Dependent: %d\n", report.Summary.AgentDependent)
	fmt.Printf("Avg response time: %d ms\n", report.Summary.AvgResponseMS)
	fmt.Printf("Success rate: %.1f%%\n", report.Summary.Coverage)

	fmt.Printf("\n=== By Category ===\n")
	for _, cs := range report.ByCategory {
		fmt.Printf("  %s: %d/%d successful\n", cs.Category, cs.Successful, cs.Total)
	}

	jsonFile := "api-test-report.json"
	if err := report.SaveJSON(jsonFile); err != nil {
		log.Printf("Failed to save JSON report: %v", err)
	} else {
		fmt.Printf("✅ JSON report saved: %s\n", jsonFile)
	}

	mdFile := "api-test-report.md"
	if err := report.SaveMarkdown(mdFile); err != nil {
		log.Printf("Failed to save Markdown report: %v", err)
	} else {
		fmt.Printf("✅ Markdown report saved: %s\n", mdFile)
	}

	fmt.Printf("\n=== Failed Endpoints (%d) ===\n", report.Summary.Failed)
	for _, result := range results {
		if !result.Success && result.ErrorMessage != "skipped (non-destructive mode)" {
			bodyPreview := result.Body
			if len(bodyPreview) > 100 {
				bodyPreview = bodyPreview[:100] + "..."
			}
			bodyPreview = strings.ReplaceAll(bodyPreview, "\n", " ")
			fmt.Printf("❌ %s %s → %d: %s\n", result.Method, result.Path, result.Status, bodyPreview)
		}
	}

	if report.Summary.Failed > 0 {
		os.Exit(1)
	}
}
