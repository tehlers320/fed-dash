package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/pkg/errors"

	"github.com/kubernetes-sigs/federation-v2/pkg/client/clientset/versioned/typed/core/v1alpha1"
	crv1alpha1 "github.com/kubernetes/cluster-registry/pkg/client/clientset/versioned/typed/clusterregistry/v1alpha1"
	"github.com/olekukonko/tablewriter"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, "Go to /deployments /pods or /jobs") })
	http.HandleFunc("/deployments", func(w http.ResponseWriter, r *http.Request) { getTable(w, DeploymentTable{}) })
	http.HandleFunc("/pods", func(w http.ResponseWriter, r *http.Request) { getTable(w, PodTable{}) })
	http.HandleFunc("/jobs", func(w http.ResponseWriter, r *http.Request) { getTable(w, JobTable{}) })
	http.ListenAndServe(":8080", nil)
	fmt.Println("Done")
}

func getTable(w http.ResponseWriter, resource ResourceTable) {
	fmt.Fprint(w, "Current pages: /deployments, /pods, /jobs\n\n")
	table := tablewriter.NewWriter(w)
	table.SetHeader(resource.Headers())
	clusterConfigs, err := getClusterConfigs()
	if err != nil {
		fmt.Fprintf(w, "%s", err)
		return
	}
	numClusters := len(clusterConfigs)
	for clusterIndex, clusterConfig := range clusterConfigs {
		rows, err := resource.GetRowsFromCluster(clusterConfig.Name, clusterConfig.Config)
		if err != nil {
			fmt.Fprintf(w, "%s", err)
			return
		}
		for _, row := range rows {
			table.Append(row)
		}

		if clusterIndex < numClusters-1 {
			table.Append(make([]string, len(resource.Headers())))
		}
	}
	table.Render()
}

type ClusterConfig struct {
	Name   string
	Config *rest.Config
}

func getClusterConfigs() ([]ClusterConfig, error) {
	var configs []ClusterConfig

	var config *rest.Config
	var err error
	if os.Getenv("OUTSIDE_CLUSTER") == "TRUE" {
		config, err = clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
		if err != nil {
			return nil, errors.Wrap(err, "Could not read KUBECONFIG")
		}
	} else {
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, errors.Wrap(err, "Could not get InClusterConfig")
		}
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "Could not create kubernetes client from config")
	}

	fedClient, err := v1alpha1.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "Could not create federation client from config")
	}

	clusterRegistryClient, err := crv1alpha1.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "Could not create cluster registry client")
	}

	clusters, err := fedClient.FederatedClusters("").List(metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "Could not list federated clusters")
	}

	for _, cluster := range clusters.Items {
		secret, err := client.CoreV1().Secrets("federation-system").Get(cluster.Spec.SecretRef.Name, metav1.GetOptions{})
		if err != nil {
			return nil, errors.Wrapf(err, "Could not get secret for cluster %s", cluster.Name)
		}

		ca, caFound := secret.Data["ca.crt"]
		if !caFound {
			return nil, errors.Errorf("Could not get ca.crt from secret for cluster %s", cluster.Name)
		}

		token, tokenFound := secret.Data["token"]
		if !tokenFound {
			return nil, errors.Errorf("Could not get token from secret for cluster %s", cluster.Name)
		}

		clusterInfo, err := clusterRegistryClient.Clusters("kube-multicluster-public").Get(cluster.Name, metav1.GetOptions{})
		if err != nil {
			return nil, errors.Wrapf(err, "Could not get cluster registry information for cluster %s", cluster.Name)
		}

		clusterAddress := clusterInfo.Spec.KubernetesAPIEndpoints.ServerEndpoints[0].ServerAddress

		clusterConfig, err := clientcmd.BuildConfigFromFlags(clusterAddress, "")
		clusterConfig.CAData = ca
		clusterConfig.BearerToken = string(token)
		if err != nil {
			return nil, errors.Wrapf(err, "Could not build config for cluster %s", cluster.Name)
		}

		configs = append(configs, ClusterConfig{Name: cluster.Name, Config: clusterConfig})
	}

	return configs, nil
}
