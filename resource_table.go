package main

import (
	"k8s.io/client-go/rest"
)

type ResourceTable interface {
	Headers() []string
	GetRowsFromCluster(clusterName string, config *rest.Config) ([][]string, error)
}
