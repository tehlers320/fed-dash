package main

import (
	"fmt"
	"github.com/pkg/errors"

	"github.com/kubernetes/client-go/kubernetes/typed/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

type DeploymentTable struct {
}

func (dt DeploymentTable) Headers() []string {
	return []string{"Cluster", "Namespace", "Deployment", "Desired Replicas", "Ready Replicas"}
}

func (dt DeploymentTable) GetRowsFromCluster(clusterName string, config *rest.Config) ([][]string, error) {
	var rows [][]string

	client, err := v1beta1.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not create deployment client for cluster %s", clusterName)
	}

	deployments, err := client.Deployments("").List(metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "Could not list deployments for cluster %s", clusterName)
	}

	for _, deployment := range deployments.Items {
		rows = append(rows, []string{clusterName, deployment.ObjectMeta.Namespace, deployment.ObjectMeta.Name, fmt.Sprintf("%d", deployment.Status.Replicas), fmt.Sprintf("%d", deployment.Status.ReadyReplicas)})
	}

	return rows, nil
}
