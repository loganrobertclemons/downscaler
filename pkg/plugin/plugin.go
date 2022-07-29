package plugin

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/kubectl/pkg/cmd/util"
)

func RunPlugin(configFlags *genericclioptions.ConfigFlags, cmd *cobra.Command) error {
	factory := util.NewFactory(configFlags)
	clientConfig := factory.ToRawKubeConfigLoader()
	config, err := factory.ToRESTConfig()

	if err != nil {
		return errors.Wrap(err, "failed to read kubeconfig")
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return errors.Wrap(err, "failed to create clientset")
	}

	namespace, _, err := clientConfig.Namespace()
	if err != nil {
		return errors.WithMessage(err, "Failed getting namespace")
	}

	if getFlagBool(cmd, "all-namespaces") {
		namespace = ""
	}

	// build the pod defination we want to deploy
	pod := getPodObject()

	pods, err := clientset.CoreV1().Pods(namespace).Create(context.TODO(), pod, metav1.CreateOptions{})
	if err != nil {
		panic(err)
	}
	fmt.Println("Created Pod: ", pods.Name)

	return nil
}

// func getPodObject() *core.Pod {
// 	return &core.Pod{
// 		ObjectMeta: metav1.ObjectMeta{
// 			Name: "my-test-pod",
// 			Labels: map[string]string{
// 				"app": "demo",
// 			},
// 		},
// 		Spec: core.PodSpec{
// 			Containers: []core.Container{
// 				{
// 					Name:            "busybox",
// 					Image:           "busybox",
// 					ImagePullPolicy: core.PullIfNotPresent,
// 					Command: []string{
// 						"sleep",
// 						"3600",
// 					},
// 				},
// 			},
// 		},
	}
}

// Gets the the flag value as a boolean, otherwise returns false if the flag value is nil
func getFlagBool(cmd *cobra.Command, flag string) bool {
	b, err := cmd.Flags().GetBool(flag)
	if err != nil {
		return false
	}
	return b
}
