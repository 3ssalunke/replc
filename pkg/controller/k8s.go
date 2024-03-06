package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/ghodss/yaml"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type K8S struct {
	Clientset *kubernetes.Clientset
}

func NewK8S(kubeConfigPath string) (*K8S, error) {
	// Load Kubernetes client configuration
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		log.Println("Failed to load Kubernetes config:", err)
		return nil, fmt.Errorf("failed to load Kubernetes config %v", err)
	}

	// Create Kubernetes clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Println("Failed to create Kubernetes clientset:", err)
		return nil, fmt.Errorf("failed to create Kubernetes clientset %v", err)
	}

	return &K8S{
		Clientset: clientset,
	}, nil
}

func readAndParseKubeYaml(filePath string, replId string) ([]map[string]interface{}, error) {
	// Read file content
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		log.Println("err", err)
		return nil, err
	}

	// Split YAML documents
	docs := strings.Split(string(fileContent), "---")

	// Parse and process each YAML document
	var parsedDocs []map[string]interface{}
	for _, doc := range docs {
		// Replace occurrences of "service_name" with replId
		docString := strings.ReplaceAll(doc, "service_name", replId)

		// Parse YAML into map[string]interface{}
		var parsedDoc map[string]interface{}
		if err := yaml.Unmarshal([]byte(docString), &parsedDoc); err != nil {
			return nil, err
		}

		// Append parsed document to result
		parsedDocs = append(parsedDocs, parsedDoc)
	}

	return parsedDocs, nil
}

func (c *Controller) CreateK8sResources(k8s *K8S, filePath string, replId string) error {
	// Read and parse Kubernetes YAML manifests
	kubeManifests, err := readAndParseKubeYaml(filePath, replId)
	if err != nil {
		log.Println("Failed to read and parse Kubernetes YAML:", err)
		return fmt.Errorf("failed to read and parse Kubernetes YAML: %v", err)
	}

	// Iterate over Kubernetes manifests and create resources
	for _, manifest := range kubeManifests {
		kindValue, ok := manifest["kind"].(string)
		if !ok {
			// Handle the case where "kind" field is not a string
			log.Println("Failed to parse kind field as string")
			return fmt.Errorf("failed to parse kind field as string")
		}

		// Assuming manifest is a map[string]interface{}
		// Convert manifest to JSON bytes
		manifestBytes, err := json.Marshal(manifest)
		if err != nil {
			log.Println("Failed to marshal manifest:", err)
			return fmt.Errorf("failed to marshal manifest %v", err)
		}

		switch kindValue {
		case "Deployment":
			// Unmarshal JSON bytes into Deployment object
			var deployment appsv1.Deployment
			err = json.Unmarshal(manifestBytes, &deployment)
			if err != nil {
				log.Println("Failed to unmarshal manifest into Deployment object:", err)
				return fmt.Errorf("failed to unmarshal manifest into Deployment object %v", err)
			}
			_, err := k8s.Clientset.AppsV1().Deployments(c.Container.Config.K8S.Namespace).Create(context.TODO(), &deployment, metav1.CreateOptions{})
			if err != nil {
				log.Println("Failed to create Deployment:", err)
				return fmt.Errorf("failed to create Deployment %v", err)
			}
		case "Service":
			// Unmarshal JSON bytes into Service object
			var service corev1.Service
			err = json.Unmarshal(manifestBytes, &service)
			if err != nil {
				log.Println("Failed to unmarshal manifest into Service object:", err)
				return fmt.Errorf("failed to unmarshal manifest into Service object %v", err)
			}
			_, err := k8s.Clientset.CoreV1().Services(c.Container.Config.K8S.Namespace).Create(context.TODO(), &service, metav1.CreateOptions{})
			if err != nil {
				log.Println("Failed to create Service:", err)
				return fmt.Errorf("failed to create Service %v", err)
			}
		case "Ingress":
			// Unmarshal JSON bytes into Ingress object
			var ingress networkingv1.Ingress
			err = json.Unmarshal(manifestBytes, &ingress)
			if err != nil {
				log.Println("Failed to unmarshal manifest into Ingress object:", err)
				return fmt.Errorf("failed to unmarshal manifest into Ingress object: %v", err)
			}
			_, err := k8s.Clientset.NetworkingV1().Ingresses(c.Container.Config.K8S.Namespace).Create(context.TODO(), &ingress, metav1.CreateOptions{})
			if err != nil {
				log.Println("Failed to create Ingress:", err)
				return fmt.Errorf("failed to create Ingress: %v", err)
			}
		default:
			log.Printf("Unsupported kind: %s\n", kindValue)
			return fmt.Errorf("unsupported kind: %s", kindValue)
		}
	}

	log.Println("Resources created successfully")
	return nil
}
