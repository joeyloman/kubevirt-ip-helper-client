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
	fmt.Printf("usage: %s <command> [namespace] [name]\n\n", os.Args[0])
	fmt.Printf("  commands:\n\n" +
		"    ippool-list                              : lists all IPPool objects\n" +
		"    ippool-show <name>                       : shows all IPPool details\n" +
		"\n" +
		"    vmnetcfg-list                            : lists all VirtualMachineNetworkConfig objects\n" +
		"    vmnetcfg-clear-status <namespace> <name> : clears the status of a VirtualMachineNetworkConfig object (in case of errors and this needs to be cleared)\n" +
		"    vmnetcfg-reset <namespace> <name>        : resets the VirtualMachineNetworkConfig object network configuration (in case you want to allocate a new IP)\n" +
		"    vmnetcfg-cleanup                         : cleans up VirtualMachineNetworkConfig object orphans (this is interactive with a backup option)\n\n")
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

	fmt.Printf("Running on clustername: %s\n\n", config.Host)

	// no args
	if os.Args[1] == "ippool-list" {
		if err := app.ListIPPools(k8s_clientset, kih_clientset); err != nil {
			fmt.Printf("error: %s", err.Error())
		}

		return
	} else if os.Args[1] == "vmnetcfg-list" {
		if err := app.ListVirtualMachineNetworkConfigs(k8s_clientset, kih_clientset); err != nil {
			fmt.Printf("error: %s", err.Error())
		}

		return
	} else if os.Args[1] == "vmnetcfg-cleanup" {
		if err := app.CleanupVirtualMachineNetworkConfigurations(k8s_clientset, kih_clientset); err != nil {
			fmt.Printf("error: %s", err.Error())
		}

		return
	}

	// 1 arg
	if len(os.Args) < 3 {
		usage()
	}

	if os.Args[1] == "ippool-show" {
		if err := app.ShowIPPool(k8s_clientset, kih_clientset, os.Args[2]); err != nil {
			fmt.Printf("error: %s", err.Error())
		}

		return
	}

	// 2 args
	if len(os.Args) < 4 {
		usage()
	}

	if os.Args[1] == "vmnetcfg-clear-status" {
		if err := app.ClearVirtualMachineNetworkConfigStatus(k8s_clientset, kih_clientset, os.Args[2], os.Args[3]); err != nil {
			fmt.Printf("error: %s", err.Error())
		}

		return
	} else if os.Args[1] == "vmnetcfg-reset" {
		if err := app.ResetVirtualMachineNetworkConfig(k8s_clientset, kih_clientset, os.Args[2], os.Args[3]); err != nil {
			fmt.Printf("error: %s", err.Error())
		}

		return
	}
}
