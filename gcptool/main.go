package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	functions "cloud.google.com/go/functions/apiv1"
	"cloud.google.com/go/functions/apiv1/functionspb"
	"google.golang.org/api/iterator"
)

type Config struct {
	ProjectID string
	Region    string
}

type Function struct {
	Name       string    `json:"name"`
	Status     string    `json:"status"`
	Runtime    string    `json:"runtime"`
	Trigger    string    `json:"trigger"`
	UpdateTime time.Time `json:"updateTime"`
	Version    string    `json:"version"`
}

func main() {
	var (
		help      = flag.Bool("h", false, "Show help")
		projectID = flag.String("p", "", "GCP Project ID")
		region    = flag.String("r", "us-central1", "GCP Region")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] COMMAND [ARGS]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Commands:\n")
		fmt.Fprintf(os.Stderr, "  deploy <function-name>     Deploy cloud function\n")
		fmt.Fprintf(os.Stderr, "  describe <function-name>   Describe cloud function details\n")
		fmt.Fprintf(os.Stderr, "  list                       List all cloud functions\n")
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nEnvironment Variables:\n")
		fmt.Fprintf(os.Stderr, "  GOOGLE_APPLICATION_CREDENTIALS  Path to service account key\n")
		fmt.Fprintf(os.Stderr, "  GCP_PROJECT_ID                  Default project ID\n")
	}

	flag.Parse()

	if *help || len(flag.Args()) == 0 {
		flag.Usage()
		os.Exit(0)
	}

	// Get project ID from flag or environment
	if *projectID == "" {
		*projectID = os.Getenv("GCP_PROJECT_ID")
		if *projectID == "" {
			log.Fatal("Project ID required. Use -p flag or set GCP_PROJECT_ID environment variable")
		}
	}

	config := &Config{
		ProjectID: *projectID,
		Region:    *region,
	}

	command := flag.Args()[0]
	args := flag.Args()[1:]

	switch command {
	case "deploy":
		if len(args) == 0 {
			log.Fatal("Function name required for deploy command")
		}
		handleDeploy(config, args[0], args[1:])
	case "describe":
		if len(args) == 0 {
			log.Fatal("Function name required for describe command")
		}
		handleDescribe(config, args[0])
	case "list":
		handleList(config)
	default:
		fmt.Printf("Unknown command: %s\n", command)
		flag.Usage()
		os.Exit(1)
	}
}

func handleDeploy(config *Config, functionName string, options []string) {
	fmt.Printf("🚀 Deploying function '%s' to project '%s'...\n", functionName, config.ProjectID)

	// Parse deployment options
	environment := "prod"
	version := ""
	clean := false
	source := "./function"

	for i, opt := range options {
		switch opt {
		case "-e":
			if i+1 < len(options) {
				environment = options[i+1]
			}
		case "-v":
			if i+1 < len(options) {
				version = options[i+1]
			}
		case "-c":
			clean = true
		case "-s":
			if i+1 < len(options) {
				source = options[i+1]
			}
		}
	}

	// Clean and rebuild if requested
	if clean {
		fmt.Println("🧹 Cleaning and rebuilding...")
		cleanCmd := exec.Command("go", "clean", "-cache")
		cleanCmd.Dir = source
		if err := cleanCmd.Run(); err != nil {
			log.Printf("Warning: Clean failed: %v", err)
		}
	}

	// Build deployment command
	deployCmd := []string{
		"gcloud", "functions", "deploy", functionName,
		"--gen2",
		"--runtime=go121",
		"--region=" + config.Region,
		"--source=" + source,
		"--entry-point=HandleRequest",
		"--trigger=https",
		"--allow-unauthenticated",
		"--project=" + config.ProjectID,
	}

	// Add environment label
	if environment != "" {
		deployCmd = append(deployCmd, "--update-labels=environment="+environment)
	}

	// Add version if specified
	if version != "" {
		deployCmd = append(deployCmd, "--update-labels=version="+version)
	}

	// Execute deployment
	cmd := exec.Command(deployCmd[0], deployCmd[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Fatalf("❌ Deployment failed: %v", err)
	}

	fmt.Printf("✅ Function '%s' deployed successfully!\n", functionName)

	// Get and display function URL
	describeCmd := exec.Command("gcloud", "functions", "describe", functionName,
		"--region="+config.Region,
		"--project="+config.ProjectID,
		"--format=value(serviceConfig.uri)")

	if output, err := describeCmd.Output(); err == nil {
		url := strings.TrimSpace(string(output))
		if url != "" {
			fmt.Printf("🌐 Function URL: %s\n", url)
		}
	}
}

func handleDescribe(config *Config, functionName string) {
	ctx := context.Background()

	client, err := functions.NewCloudFunctionsClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// Get function details
	functionPath := fmt.Sprintf("projects/%s/locations/%s/functions/%s",
		config.ProjectID, config.Region, functionName)

	req := &functionspb.GetFunctionRequest{
		Name: functionPath,
	}

	function, err := client.GetFunction(ctx, req)
	if err != nil {
		log.Fatalf("Failed to get function: %v", err)
	}

	// Create function info
	funcInfo := Function{
		Name:       function.Name,
		Status:     function.Status.String(),
		Runtime:    function.Runtime,
		UpdateTime: function.UpdateTime.AsTime(),
	}

	// Determine trigger type
	if function.GetHttpsTrigger() != nil {
		funcInfo.Trigger = "HTTPS"
	} else if function.GetEventTrigger() != nil {
		funcInfo.Trigger = "Event"
	} else {
		funcInfo.Trigger = "Unknown"
	}

	// Get version from labels if available
	if labels := function.GetLabels(); labels != nil {
		if v, exists := labels["version"]; exists {
			funcInfo.Version = v
		}
	}

	// Display function details
	fmt.Printf("📋 Function Details:\n")
	fmt.Printf("Name: %s\n", strings.Split(funcInfo.Name, "/")[5])
	fmt.Printf("Status: %s\n", funcInfo.Status)
	fmt.Printf("Runtime: %s\n", funcInfo.Runtime)
	fmt.Printf("Trigger: %s\n", funcInfo.Trigger)
	fmt.Printf("Last Updated: %s\n", funcInfo.UpdateTime.Format(time.RFC3339))

	if funcInfo.Version != "" {
		fmt.Printf("Version: %s\n", funcInfo.Version)
	}

	// Display URL if HTTPS trigger
	if function.GetHttpsTrigger() != nil {
		fmt.Printf("URL: %s\n", function.GetHttpsTrigger().Url)
	}

	// Display source info
	if source := function.GetSourceArchiveUrl(); source != "" {
		fmt.Printf("Source: %s\n", source)
	}
}

func handleList(config *Config) {
	ctx := context.Background()

	client, err := functions.NewCloudFunctionsClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// List functions
	parent := fmt.Sprintf("projects/%s/locations/%s", config.ProjectID, config.Region)
	req := &functionspb.ListFunctionsRequest{
		Parent: parent,
	}

	it := client.ListFunctions(ctx, req)
	functions := []Function{}

	for {
		function, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("Failed to list functions: %v", err)
		}

		funcInfo := Function{
			Name:       strings.Split(function.Name, "/")[5], // Extract just the function name
			Status:     function.Status.String(),
			Runtime:    function.Runtime,
			UpdateTime: function.UpdateTime.AsTime(),
		}

		// Determine trigger type
		if function.GetHttpsTrigger() != nil {
			funcInfo.Trigger = "HTTPS"
		} else if function.GetEventTrigger() != nil {
			funcInfo.Trigger = "Event"
		} else {
			funcInfo.Trigger = "Unknown"
		}

		// Get version from labels
		if labels := function.GetLabels(); labels != nil {
			if v, exists := labels["version"]; exists {
				funcInfo.Version = v
			}
		}

		functions = append(functions, funcInfo)
	}

	if len(functions) == 0 {
		fmt.Printf("No functions found in project '%s' region '%s'\n", config.ProjectID, config.Region)
		return
	}

	fmt.Printf("📦 Cloud Functions in project '%s':\n\n", config.ProjectID)
	fmt.Printf("%-20s %-10s %-10s %-8s %-10s %s\n", "NAME", "STATUS", "RUNTIME", "TRIGGER", "VERSION", "LAST UPDATED")
	fmt.Printf("%s\n", strings.Repeat("-", 90))

	for _, f := range functions {
		version := f.Version
		if version == "" {
			version = "n/a"
		}
		fmt.Printf("%-20s %-10s %-10s %-8s %-10s %s\n",
			f.Name, f.Status, f.Runtime, f.Trigger, version, f.UpdateTime.Format("2006-01-02 15:04"))
	}
}
