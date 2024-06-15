package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/retry"
)

func main() {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err.Error())
	}
	kubeconfigpath := filepath.Join(home, ".kube/config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigpath)
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)

	if err != nil {
		panic(err.Error())
	}
	pod, err := clientset.CoreV1().Pods("default").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	for i, pods := range pod.Items {
		fmt.Println("pods present in clusture are", i, pods.Name)
	}

	podcreate := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "curd-pod",
			Namespace: "default",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "nginx-container",
					Image: "nginx:latest",
				},
			},
		},
	}

	//pod creation
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {

		newpod, err := clientset.CoreV1().Pods("default").Create(context.Background(), podcreate, metav1.CreateOptions{})
		fmt.Println("new pod is created", newpod.Name)

		//pod update
		//find the pod which we have to update
		findpod, err := clientset.CoreV1().Pods("default").Get(context.TODO(), "curd-pod", metav1.GetOptions{})
		if err != nil {
			panic(err.Error())
		}
		findpod.Spec.Containers[0].Image = "nginx:1.25.4"

		updatedpod, updterr := clientset.CoreV1().Pods("default").Update(context.TODO(), findpod, metav1.UpdateOptions{})
		fmt.Println("pod is updated ", updatedpod.Name)
		return updterr

	})
	if retryErr != nil {
		panic(retryErr.Error())
	}
	podclient := clientset.CoreV1().Pods("default")

	deleteerr := podclient.Delete(context.Background(), "curd-pod", metav1.DeleteOptions{})
	if deleteerr != nil {
		panic(deleteerr.Error())
	}
}
