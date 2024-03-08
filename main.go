package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"path/filepath"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// generateRandomSuffix generates a random 6 digit number as a string
func generateRandomSuffix() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%03d", rand.Intn(1000))
}

// createPod creates a new pod with the specified arguments
func createPod(namespace, podName, mirrorType, podInterface, destinationIP, vxlanID, containerID string) error {
	// config, err := rest.InClusterConfig()
	// if err != nil {
	// 	return err
	// }

	// clientset, err := kubernetes.NewForConfig(config)
	// if err != nil {
	// 	return err
	// }

	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("captured-%s-%s", podName, generateRandomSuffix()),
			Namespace: namespace,
			Labels: map[string]string{
				"dn-vtap": "capturing",
			},
		},
		Spec: corev1.PodSpec{
			HostNetwork: true,
			NodeName:    "k8s-w1.5g.dn.th-koeln.de",
			Volumes: []corev1.Volume{
				{
					Name: "hostproc",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/proc",
						},
					},
				},
				{
					Name: "var-crio",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/var/run/crio/crio.sock",
						},
					},
				},
			},

			Containers: []corev1.Container{
				{
					Name:  "network-mirror",
					Image: "quay.io/s1061123/kokotap:latest", // Specify the image you want to use
					/*Env: []corev1.EnvVar{
						{Name: "MIRROR_TYPE", Value: mirrorType},
						{Name: "MIRROR_INTERFACE", Value: mirrorInterface},
						{Name: "DESTINATION_IP", Value: destinationIP},
						{Name: "VXLAN_ID", Value: vxlanID},
					},
					*/
					Command: []string{"/bin/kokotap_pod"}, // Specify the binary you want to run
					SecurityContext: &corev1.SecurityContext{
						Privileged: &[]bool{true}[0],
					},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "hostproc",
							MountPath: "/hostproc",
						},
						{
							Name:      "var-crio",
							MountPath: "/var/run/crio/crio.sock",
						},
					},
					Args: []string{"--procprefix=/host",
						"mode",
						"sender",
						//TODO: Add a function to get containerID from podName
						"--containerid=" + containerID,
						"--mirrortype=" + mirrorType,
						"--mirrorif=" + podInterface,
						"--ifname=mirror",
						"--vxlan-egressip=192.168.1.108",
						"--vxlan-ip=" + destinationIP,
						"--vxlan-id=" + vxlanID,
						"--vxlan-port=4789",
					},
					// Add any other container specifications here
				},
			},
			// Specify any other pod specifications here
		},
	}

	// Create Pod
	_, err = clientset.CoreV1().Pods(namespace).Create(context.TODO(), pod, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	fmt.Printf("Pod %s is created successfully in namespace %s\n", pod.ObjectMeta.Name, namespace)
	return nil
}

// Example usage
func main() {
	ns := "default"
	pod := "ueransim-gnb-ues-6c7d5c7bfb-2nd9t"
	mType := "both"
	PodInt := "eth0"
	destIP := "192.168.1.109"
	vID := "1100"
	contID := "cri-o://3ba9777f0b218d6364a33f79b28a91e3a8e48aca7d55c73c2109e49e7e993285"
	err := createPod(ns, pod, mType, PodInt, destIP, vID, contID)
	if err != nil {
		fmt.Printf("Error creating pod: %s\n", err)
	}
}

/* package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func main() {

	// Define default path to kubeconfig file in home directory

	homedir := os.Getenv("HOME")
	kubeconfig := filepath.Join(homedir, ".kube", "config")

	// Build the client config from the kubeconfig file
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatal(err)
	}

	// Create a new clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	// Get all available Namespaces in the cluster

	ns, err := clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Fatal(err)
	}
	for _, n := range ns.Items {
		fmt.Printf("NAMESPACE %s \n", n.Name)
		pods, err1 := clientset.CoreV1().Pods(n.Name).List(context.TODO(), metav1.ListOptions{})
		if err1 != nil {
			log.Fatal(err1)
		}
		for _, pod := range pods.Items {
			fmt.Printf("Namespace: %s, Pod: %s has the following annotations: \n", n.Name, pod.Name)
			for _, a := range pod.Spec.Containers {
				fmt.Printf("Container Name: %s, Container Image: %s\n", a.Name, a.Image)

			}
		}
	}

*/
//
