package main

import (
	"github.com/pkg/errors"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type ConfigMapTable struct {
}

func (cmt ConfigMapTable) Headers() []string {
	return []string{"Cluster", "Namespace", "ConfigMap", "Data"}
}

func (cmt ConfigMapTable) GetRowsFromCluster(clusterName string, config *rest.Config) ([][]string, error) {
	var rows [][]string

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not create kubernetes client for cluster %s", clusterName)
	}

	configmap, err := client.CoreV1().ConfigMaps("").List(metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "Could not list configmaps for cluster %s", clusterName)
	}

	for _, configmap := range configmap.Items {
		ns := configmap.ObjectMeta.Namespace
		name := configmap.ObjectMeta.Name
		var dataNames []string
		for k, _ := range configmap.Data {
			dataNames = append(dataNames, k)
		}
		datas := strings.Join(dataNames, ", ")
		rows = append(rows, []string{clusterName, ns, name, datas})
	}

	return rows, nil
}
