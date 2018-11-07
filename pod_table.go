package main

import (
	"fmt"
	"github.com/pkg/errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type PodTable struct {
}

func (pt PodTable) Headers() []string {
	return []string{"Cluster", "Namespace", "Pod", "Phase"}
}

func (pt PodTable) GetRowsFromCluster(clusterName string, config *rest.Config) ([][]string, error) {
	var rows [][]string

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not create kubernetes client for cluster %s", clusterName)
	}

	pods, err := client.CoreV1().Pods("").List(metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "Could not list pods for cluster %s", clusterName)
	}

	for _, pod := range pods.Items {
		rows = append(rows, []string{clusterName, pod.ObjectMeta.Namespace, pod.ObjectMeta.Name, fmt.Sprintf("%v", pod.Status.Phase)})
	}

	return rows, nil
}
