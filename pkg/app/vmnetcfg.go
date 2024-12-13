package app

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"

	kihclientset "github.com/joeyloman/kubevirt-ip-helper/pkg/generated/clientset/versioned"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"

	// "github.com/ghodss/yaml"
	kihv1 "github.com/joeyloman/kubevirt-ip-helper/pkg/apis/kubevirtiphelper.k8s.binbash.org/v1"
)

func listVirtualMachineNetworkConfigs(k8s_clientset *kubernetes.Clientset, kih_clientset *kihclientset.Clientset) (err error) {
	vmNetCfgObjs, err := kih_clientset.KubevirtiphelperV1().VirtualMachineNetworkConfigs(corev1.NamespaceAll).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("cannot fetch VirtualMachineNetworkConfig objects: %s", err.Error())
	}

	for _, vmnetcfg := range vmNetCfgObjs.Items {
		fmt.Printf("VirtualMachineNetworkConfig: %s/%s\n", vmnetcfg.Namespace, vmnetcfg.Name)
		fmt.Print("  Network interface configuration:\n")
		for _, v := range vmnetcfg.Spec.NetworkConfig {
			for _, s := range vmnetcfg.Status.NetworkConfig {
				if v.MACAddress == s.MACAddress {
					fmt.Printf("    mac: %s, ip: %s, network: %s, status: %s %s\n", v.MACAddress, v.IPAddress, v.NetworkName, s.Status, s.Message)
				}
			}
		}
		fmt.Print("\n")
	}

	return
}

func resetVirtualMachineNetworkConfig(k8s_clientset *kubernetes.Clientset, kih_clientset *kihclientset.Clientset, vmnetcfgNamespace string, vmnetcfgName string) (err error) {
	var netcfgErrorDetected bool = false

	vmnetcfg, err := kih_clientset.KubevirtiphelperV1().VirtualMachineNetworkConfigs(vmnetcfgNamespace).Get(context.TODO(), vmnetcfgName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("cannot fetch VirtualMachineNetworkConfig object: %s", err.Error())
	}

	newVmNetCfgObj := vmnetcfg.DeepCopy()
	vmnetcfgs := []kihv1.NetworkConfig{}
	for _, v := range vmnetcfg.Spec.NetworkConfig {
		netcfg := kihv1.NetworkConfig{}
		netcfg.MACAddress = v.MACAddress
		netcfg.NetworkName = v.NetworkName

		vmnetcfgs = append(vmnetcfgs, netcfg)
	}
	newVmNetCfgObj.Spec.NetworkConfig = vmnetcfgs

	updatedVmNetCfgObj, err := kih_clientset.KubevirtiphelperV1().VirtualMachineNetworkConfigs(vmnetcfgNamespace).Update(context.TODO(), newVmNetCfgObj, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("cannot update the VirtualMachineNetworkConfig object [%s/%s]: %s", vmnetcfgNamespace, vmnetcfgName, err.Error())
	}

	for _, v := range vmnetcfg.Status.NetworkConfig {
		if v.Status == "ERROR" {
			netcfgErrorDetected = true
			break
		}
	}

	if netcfgErrorDetected {
		newUpdatedVmNetCfgObj := updatedVmNetCfgObj.DeepCopy()
		newUpdatedVmNetCfgObj.Status = kihv1.VirtualMachineNetworkConfigStatus{}

		if _, err := kih_clientset.KubevirtiphelperV1().VirtualMachineNetworkConfigs(vmnetcfgNamespace).UpdateStatus(context.TODO(), newUpdatedVmNetCfgObj, metav1.UpdateOptions{}); err != nil {
			return fmt.Errorf("cannot update the VirtualMachineNetworkConfig object status [%s/%s]: %s",
				vmnetcfgNamespace, vmnetcfgName, err.Error())
		}
	}

	fmt.Printf("successfully reset the VirtualMachineNetworkConfig object status [%s/%s]\n", vmnetcfgNamespace, vmnetcfgName)

	return
}

func clearVirtualMachineNetworkConfigStatus(k8s_clientset *kubernetes.Clientset, kih_clientset *kihclientset.Clientset, vmNetCfgNamespace string, vmNetCfgName string) (err error) {
	vmNetCfgObj, err := kih_clientset.KubevirtiphelperV1().VirtualMachineNetworkConfigs(vmNetCfgNamespace).Get(context.TODO(), vmNetCfgName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("cannot fetch VirtualMachineNetworkConfig objects: %s", err.Error())
	}

	updatedVmNetCfgObj := vmNetCfgObj.DeepCopy()
	updatedVmNetCfgObj.Status = kihv1.VirtualMachineNetworkConfigStatus{}

	if _, err := kih_clientset.KubevirtiphelperV1().VirtualMachineNetworkConfigs(vmNetCfgNamespace).UpdateStatus(context.TODO(), updatedVmNetCfgObj, metav1.UpdateOptions{}); err != nil {
		return fmt.Errorf("cannot update status of VirtualMachineNetworkConfig [%s/%s]: %s",
			vmNetCfgNamespace, vmNetCfgName, err.Error())
	}

	fmt.Printf("successfully cleared the status of VirtualMachineNetworkConfig [%s/%s]\n", vmNetCfgNamespace, vmNetCfgName)

	return
}

func checkForVirtualMachineInstance(k8s_clientset *kubernetes.Clientset, vmNamespace string, vmName string) bool {
	_, err := k8s_clientset.RESTClient().Get().AbsPath("/apis/kubevirt.io/v1").Namespace(vmNamespace).Resource("virtualmachines").Name(vmName).DoRaw(context.TODO())
	if err != nil {
		if strings.Contains(err.Error(), "the server could not find the requested resource") {
			return true
		}
	}

	return false
}

func backupVirtualMachineNetworkConfiguration(kih_clientset *kihclientset.Clientset, vmNamespace string, vmName string) (err error) {
	vmNetCfgObj, err := kih_clientset.KubevirtiphelperV1().VirtualMachineNetworkConfigs(vmNamespace).Get(context.TODO(), vmName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("error while fetching VirtualMachineNetworkConfig object: %s", err.Error())
	}

	scheme := runtime.NewScheme()
	err = kihv1.AddToScheme(scheme)
	if err != nil {
		return err
	}
	codec := serializer.NewCodecFactory(scheme).LegacyCodec(kihv1.SchemeGroupVersion)
	jsonData, err := runtime.Encode(codec, vmNetCfgObj)
	if err != nil {
		return err
	}
	yamlData, err := yaml.JSONToYAML(jsonData)
	if err != nil {
		return err
	}

	filename := fmt.Sprintf("%s_%s.yaml", vmNamespace, vmName)
	fmt.Printf("writing vmnetcfg object backup to: %s\n", filename)

	return os.WriteFile(filename, yamlData, 0644)
}

func cleanupVirtualMachineNetworkConfigurations(k8s_clientset *kubernetes.Clientset, kih_clientset *kihclientset.Clientset) (err error) {
	vmNetCfgObjs, err := kih_clientset.KubevirtiphelperV1().VirtualMachineNetworkConfigs(corev1.NamespaceAll).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("cannot fetch VirtualMachineNetworkConfig objects: %s", err.Error())
	}

	for _, vmnetcfg := range vmNetCfgObjs.Items {
		if checkForVirtualMachineInstance(k8s_clientset, vmnetcfg.Namespace, vmnetcfg.Name) {
			fmt.Printf("\n%s/%s has no vm, backup and cleanup the vmnetcfg object? (y/n) ", vmnetcfg.Namespace, vmnetcfg.Name)
			reader := bufio.NewReader(os.Stdin)
			input, err := reader.ReadString('\n')
			if err == nil {
				input = strings.TrimSpace(input)
				if input == "y" {
					if err = backupVirtualMachineNetworkConfiguration(kih_clientset, vmnetcfg.Namespace, vmnetcfg.Name); err != nil {
						fmt.Printf("failed to write backup file: %s\n", err.Error())
					} else {
						fmt.Printf("backup succeeded, removing vmnetcfg object..")
						if err := kih_clientset.KubevirtiphelperV1().VirtualMachineNetworkConfigs(vmnetcfg.Namespace).Delete(context.TODO(), vmnetcfg.Name, metav1.DeleteOptions{}); err != nil {
							fmt.Printf("error while deleting VirtualMachineNetworkConfig object %s/%s: %s",
								vmnetcfg.Namespace, vmnetcfg.Name, err.Error())
						}
						fmt.Printf("done!\n")
					}

				}
			}
		}
	}

	return
}
