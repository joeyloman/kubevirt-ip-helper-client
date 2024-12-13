package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/joeyloman/kubevirt-ip-helper-client/pkg/app"

	"github.com/joeyloman/kubevirt-ip-helper/pkg/util"

	kihclientset "github.com/joeyloman/kubevirt-ip-helper/pkg/generated/clientset/versioned"
)

func usage() {
	fmt.Printf("usage: %s <command> [VirtualMachineNetworkConfig object namespace] [VirtualMachineNetworkConfig object name]\n", os.Args[0])
	os.Exit(1)
}

func main() {
	if len(os.Args) < 2 {
		usage()
	}

	var kubeconfig_file string
	kubeconfig_file = os.Getenv("KUBECONFIG")
	if kubeconfig_file == "" {
		kubeconfig_file = filepath.Join(os.Getenv("HOME"), ".kube", "config")
	}

	var config *rest.Config
	if util.FileExists(kubeconfig_file) {
		// uses kubeconfig
		kubeconfig := flag.String("kubeconfig", kubeconfig_file, "(optional) absolute path to the kubeconfig file")
		flag.Parse()
		config_kube, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
		if err != nil {
			panic(err.Error())
		}
		config = config_kube
	} else {
		// creates the in-cluster config
		config_rest, err := rest.InClusterConfig()
		if err != nil {
			panic(err.Error())
		}
		config = config_rest
	}

	k8s_clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	kih_clientset, err := kihclientset.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	if os.Args[1] == "vmnetcfg-list" {
		if err := app.listVirtualMachineNetworkConfigs(k8s_clientset, kih_clientset); err != nil {
			fmt.Printf("error: %s", err.Error())
		}

		return
	} else if os.Args[1] == "vmnetcfg-cleanup" {
		if err := app.cleanupVirtualMachineNetworkConfigurations(k8s_clientset, kih_clientset); err != nil {
			fmt.Printf("error: %s", err.Error())
		}

		return
	}

	if len(os.Args) < 4 {
		usage()
	}

	if os.Args[1] == "vmnetcfg-clear-status" {
		if err := app.clearVirtualMachineNetworkConfigStatus(k8s_clientset, kih_clientset, os.Args[2], os.Args[3]); err != nil {
			fmt.Printf("error: %s", err.Error())
		}

		return
	} else if os.Args[1] == "vmnetcfg-reset" {
		if err := app.resetVirtualMachineNetworkConfig(k8s_clientset, kih_clientset, os.Args[2], os.Args[3]); err != nil {
			fmt.Printf("error: %s", err.Error())
		}

		return
	}
}
