package kubernetes

import (
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func NewClient() kubernetes.Interface {
	configPath := filepath.Join(
		os.Getenv("HOME"), ".kube", "config",
	)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		configPath = ""
	}
	config, err := clientcmd.BuildConfigFromFlags("", configPath)
	if err != nil {
		panic(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	return clientset
}
