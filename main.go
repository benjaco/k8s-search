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

// KubernetesSecret represents the structure of a Kubernetes Secret YAML file
type KubernetesSecret struct {
	APIVersion string            `yaml:"apiVersion"`
	Kind       string            `yaml:"kind"`
	Metadata   Metadata          `yaml:"metadata"`
	Type       string            `yaml:"type,omitempty"`
	StringData map[string]string `yaml:"stringData,omitempty"`
}

// KubernetesConfig represents the structure of a Kubernetes ConfigMap YAML file
type KubernetesConfig struct {
	APIVersion string            `yaml:"apiVersion"`
	Kind       string            `yaml:"kind"`
	Metadata   Metadata          `yaml:"metadata"`
	Data       map[string]string `yaml:"data,omitempty"`
}

// DeployedData represents the structure of a deployed Kubernetes Secret or ConfigMap
type DeployedData struct {
	Type      string
	Name      string
	Namespace string
	Data      map[string]string
}

// Metadata holds the metadata information for Kubernetes resources
type Metadata struct {
	Name      string `yaml:"name"`
	Namespace string `yaml:"namespace"`
}

func main() {
	// Define command-line flags
	findPatterenPtr := flag.String("string", "", "Directory to scan for config and secret YAML files")
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
	findPatteren = strings.ToLower(findPatteren)

	log.Printf("Seaching >>%s<<", findPatteren)

	clientset, err := getKubernetesClient()
	if err != nil {
		log.Fatalf("Failed to create Kubernetes client: %v", err)
	}

	secrets := clientset.CoreV1().Secrets("")
	secretList, err := secrets.List(context.Background(), metav1.ListOptions{})
	if err != nil {
		log.Fatalf("Failed to secretList secrets: %v", err)
	}
	for _, item := range secretList.Items {
		if strings.Contains(strings.ToLower(item.Name), findPatteren) {
			log.Printf("Found secret with name: %s in namespace %s", item.Name, item.Namespace)

		}

		// log.Printf("Searchig secret in namespace: %s %s", item.Namespace, item.Name)
		for name, stringdata := range item.StringData {
			// log.Printf("Seaching secret in namespace: %s; Secret: %s - %s : %s", item.Namespace, item.Name, name, stringdata)

			if strings.Contains(strings.ToLower(name), findPatteren) || strings.Contains(strings.ToLower(stringdata), findPatteren) {
				log.Printf("Found secret in namespace: %s; Secret: %s - %s : %s", item.Namespace, item.Name, name, stringdata)
			}
		}
		for name, stringdata := range item.Data {
			str, v := safeConvert(stringdata)
			if v == true {
				//log.Printf("Seaching secret in namespace: %s; Secret: %s - %s : %s", item.Namespace, item.Name, name, str)

				//println(strings.Contains(strings.ToLower(name), findPatteren))
				//println(strings.Contains(strings.ToLower(str), findPatteren))

				if strings.Contains(strings.ToLower(name), findPatteren) || strings.Contains(strings.ToLower(str), findPatteren) {
					log.Printf("Found secret in namespace: %s; Secret: %s - %s : %s", item.Namespace, item.Name, name, string(stringdata))
				}
			} else {
				//log.Printf("Seaching secret in namespace: %s; Secret: %s - %s : %s", item.Namespace, item.Name, name, str)
				if *verbosePtr {
					log.Printf("Value could not be decoded: %s; Secret: %s - %s", item.Namespace, item.Name, name)
				}
			}

		}
	}

	configs := clientset.CoreV1().ConfigMaps("")
	configMapList, err := configs.List(context.Background(), metav1.ListOptions{})
	if err != nil {
		log.Fatalf("Failed to configMapList secrets: %v", err)
	}
	for _, item := range configMapList.Items {
		for name, stringdata := range item.Data {
			if strings.Contains(strings.ToLower(name), findPatteren) || strings.Contains(strings.ToLower(stringdata), findPatteren) {
				log.Printf("Found config in namespace: %s; %s : %s", item.Namespace, name, stringdata)
			}
		}
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

const maxDataSize = 10000 // Set your maximum allowed data size in bytes

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
