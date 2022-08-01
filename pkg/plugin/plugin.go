package plugin

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"
	"k8s.io/kubectl/pkg/cmd/util"
)

const replicaCount = 0

type deployment struct {
	Name     string
	Replicas int
}

func RunPlugin(configFlags *genericclioptions.ConfigFlags, cmd *cobra.Command) error {
	factory := util.NewFactory(configFlags)
	clientConfig := factory.ToRawKubeConfigLoader()
	config, err := factory.ToRESTConfig()

	if err != nil {
		return errors.Wrap(err, "failed to read kubeconfig")
	}

	clientset, err := kubernetes.NewForConfig(config)

	namespace, _, err := clientConfig.Namespace()
	if err != nil {
		return errors.WithMessage(err, "Failed getting namespace")
	}

	if getFlagBool(cmd, "all-namespaces") {
		namespace = ""
	}

	// all this shit needs to be a function
	deploymentsClient := clientset.AppsV1().Deployments(namespace)
	if err != nil {
		return errors.Wrap(err, "failed to create clientset")
	}

	fmt.Printf("Listing deployments in namespace %q:\n", namespace)

	list, err := deploymentsClient.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}

	deployments := make(map[string]int32)

	// if this works you might need to check what kind of cluster it is
	// if gke, check autoscaler profile
	// probably needs to be set to not balanced
	// and if gke autopilot, panic
	// also you have to have networkpolicy disabled
	for _, d := range list.Items {
		deployments[d.Name] = *d.Spec.Replicas

		retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			// Retrieve the latest version of Deployment before attempting update
			// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
			result, getErr := deploymentsClient.Get(context.TODO(), d.Name, metav1.GetOptions{})
			if getErr != nil {
				panic(fmt.Errorf("Failed to get latest version of Deployment: %v", getErr))
			}

			*result.Spec.Replicas = replicaCount // reduce replica count
			_, updateErr := deploymentsClient.Update(context.TODO(), result, metav1.UpdateOptions{})
			return updateErr
		})
		if retryErr != nil {
			panic(fmt.Errorf("Update failed: %v", retryErr))
		}
		fmt.Println("Updated deployment...")
	}

	fmt.Println(deployments)

	return nil
}

// Gets the the flag value as a boolean, otherwise returns false if the flag value is nil
func getFlagBool(cmd *cobra.Command, flag string) bool {
	b, err := cmd.Flags().GetBool(flag)
	if err != nil {
		return false
	}
	return b
}
