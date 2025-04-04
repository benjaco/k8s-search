package main

import (
	"context"
	"flag"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"log"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	// Uncomment the following line if you need to use in-cluster config
	// "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	// Define command-line flags
	findPatterenPtr := flag.String("string", "", "String to search for in the deployed config and secret")
	caseSensitivePtr := flag.Bool("casesensitive", false, "Enable case sensitive search")
	verbosePtr := flag.Bool("verbose", false, "Enable verbose logging")
	flag.Parse()

	// Set up logging
	if *verbosePtr {
		log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	} else {
		log.SetFlags(0)
	}
	log.SetOutput(os.Stdout)
	if *findPatterenPtr == "" {
		log.Fatal("Please specify --string")
		return
	}
	findPatteren := *findPatterenPtr
	globalFound := false

	log.Printf("Searching for >>%s<< (case sensitive: %t)", findPatteren, *caseSensitivePtr)

	clientset, err := getKubernetesClient()
	if err != nil {
		log.Fatalf("Failed to create Kubernetes client: %v", err)
	}

	secrets := clientset.CoreV1().Secrets("")
	secretList, err := secrets.List(context.Background(), metav1.ListOptions{})
	if err != nil {
		log.Fatalf("Failed to list secrets: %v", err)
	}
	for _, item := range secretList.Items {
		if matches(item.Name, findPatteren, *caseSensitivePtr) {
			log.Printf("Found secret with name: %s in namespace %s", item.Name, item.Namespace)
		}

		// log.Printf("Searchig secret in namespace: %s %s", item.Namespace, item.Name)
		for name, stringdata := range item.StringData {
			// log.Printf("Seaching secret in namespace: %s; Secret: %s - %s : %s", item.Namespace, item.Name, name, stringdata)

			if matches(name, findPatteren, *caseSensitivePtr) || matches(stringdata, findPatteren, *caseSensitivePtr) {
				log.Printf("Found secret in namespace: %s; Secret: %s - %s : %s", item.Namespace, item.Name, name, stringdata)
				globalFound = true
			}
		}
		for name, stringdata := range item.Data {
			str, valid := safeConvert(stringdata)
			if valid {
				if matches(name, findPatteren, *caseSensitivePtr) || matches(str, findPatteren, *caseSensitivePtr) {
					log.Printf("Found secret in namespace: %s; Secret: %s - %s : %s", item.Namespace, item.Name, name, str)
					globalFound = true
				}
			} else {
				if *verbosePtr {
					log.Printf("Value could not be decoded: %s; Secret: %s - %s", item.Namespace, item.Name, name)
				}
			}

		}
	}

	configs := clientset.CoreV1().ConfigMaps("")
	configMapList, err := configs.List(context.Background(), metav1.ListOptions{})
	if err != nil {
		log.Fatalf("Failed to list config maps: %v", err)
	}
	for _, item := range configMapList.Items {
		for name, stringdata := range item.Data {
			if matches(name, findPatteren, *caseSensitivePtr) || matches(stringdata, findPatteren, *caseSensitivePtr) {
				log.Printf("Found config in namespace: %s; %s : %s", item.Namespace, name, stringdata)
				globalFound = true
			}
		}
	}

	// Set exit code based on whether any differences were found
	if globalFound {
		fmt.Println("Found some results.")
		os.Exit(1) // Indicates failure due to differences
	} else {
		fmt.Println("Nothing found.")
		os.Exit(0) // Indicates success
	}
}

// getKubernetesClient initializes and returns a Kubernetes clientset
func getKubernetesClient() (*kubernetes.Clientset, error) {
	// Use the current context in kubeconfig
	kubeconfigPath := filepath.Join(homeDir(), ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("error building kubeconfig: %w", err)
	}

	// Create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("error creating Kubernetes client: %w", err)
	}

	return clientset, nil
}

// homeDir returns the home directory for the current user
func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // Windows
}

const maxDataSize = 10000

// safeConvert attempts to convert a byte slice to a string.
// It returns the string and true if conversion is successful; otherwise, an empty string and false.
func safeConvert(data []byte) (string, bool) {
	// Check if the data exceeds the maximum allowed size.
	if len(data) > maxDataSize {
		//log.Printf("Skipping data: size (%d bytes) exceeds limit (%d bytes)", len(data), maxDataSize)
		return "", false
	}
	// Check if the data is valid UTF-8.
	if !utf8.Valid(data) {
		//log.Printf("Skipping data: invalid UTF-8")
		return "", false
	}
	// Safe conversion to string.
	return string(data), true
}

// matches compares the source string with the search pattern.
// If caseSensitive is true, it compares the strings as-is;
// otherwise, it performs a case-insensitive comparison.
func matches(source, pattern string, caseSensitive bool) bool {
	if caseSensitive {
		return strings.Contains(source, pattern)
	}
	return strings.Contains(strings.ToLower(source), strings.ToLower(pattern))
}
