package main

import (
	"fmt"
	"github.com/pkg/errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/typed/batch/v1"
	"k8s.io/client-go/rest"
)

type JobTable struct {
}

func (jt JobTable) Headers() []string {
	return []string{"Cluster", "Namespace", "Pod", "Parallelism", "Desired Completions", "Active", "Succeeded", "Failed"}
}

func (jt JobTable) GetRowsFromCluster(clusterName string, config *rest.Config) ([][]string, error) {
	var rows [][]string

	client, err := v1.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not create jobs client for cluster %s", clusterName)
	}

	jobs, err := client.Jobs("").List(metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "Could not list jobs for cluster %s", clusterName)
	}

	for _, job := range jobs.Items {
		rows = append(rows, []string{clusterName, job.ObjectMeta.Namespace, job.ObjectMeta.Name, fmt.Sprintf("%d", *job.Spec.Parallelism), fmt.Sprintf("%d", job.Spec.Completions), fmt.Sprintf("%d", job.Status.Active), fmt.Sprintf("%d", job.Status.Succeeded), fmt.Sprintf("%d", job.Status.Failed)})
	}

	return rows, nil
}
