package app

import (
	"context"
	"fmt"
	"log"
	"strings"

	yaml "gopkg.in/yaml.v3"

	"k8s.io/client-go/kubernetes"

	kihclientset "github.com/joeyloman/kubevirt-ip-helper/pkg/generated/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ListIPPools(k8s_clientset *kubernetes.Clientset, kih_clientset *kihclientset.Clientset) (err error) {
	IPPoolObjs, err := kih_clientset.KubevirtiphelperV1().IPPools().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("cannot fetch IPPool objects: %s", err.Error())
	}

	fmt.Printf("IPPool list:\n\n")
	for _, ippool := range IPPoolObjs.Items {
		fmt.Printf("  %s\n", ippool.Name)
	}
	fmt.Print("\n")

	return
}

func ShowIPPool(k8s_clientset *kubernetes.Clientset, kih_clientset *kihclientset.Clientset, ippool string) (err error) {
	IPPoolObj, err := kih_clientset.KubevirtiphelperV1().IPPools().Get(context.TODO(), ippool, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("cannot fetch IPPool object: %s", err.Error())
	}

	fmt.Printf("IPPool: %s\n\n", IPPoolObj.Name)

	fmt.Printf("  Configuration:\n")
	specYaml, err := yaml.Marshal(IPPoolObj.Spec)
	if err != nil {
		log.Fatal(err)
	}
	// indent the parsed yaml
	fmt.Printf("    %s", strings.Replace(string(specYaml), "\n", "\n    ", -1))
	fmt.Print("\n")

	fmt.Printf("  Status:\n")
	statusYaml, err := yaml.Marshal(IPPoolObj.Status)
	if err != nil {
		log.Fatal(err)
	}
	// indent the parsed yaml
	fmt.Printf("    %s", strings.Replace(string(statusYaml), "\n", "\n    ", -1))
	fmt.Print("\n")

	return
}
