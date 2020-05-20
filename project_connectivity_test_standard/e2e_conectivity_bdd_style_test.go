package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/helm"
	http_helper "github.com/gruntwork-io/terratest/modules/http-helper"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const helmChartName = "conn-test"

var helmOptions *helm.Options
var k8sOptions *k8s.KubectlOptions
var namespaceName = fmt.Sprintf("%s-%s", helmChartName, strings.ToLower(random.UniqueId()))

var fqdn = flag.String("fqdn", "", "FQDN that will be used for exposing application")

// func TestConnectivity(t *testing.T) {
func TestWhenISendGetRequest(t *testing.T) {

	t.Cleanup(after(t))
	before(t)

	t.Run("Should return 200 response", func(t *testing.T) {
		url := fmt.Sprintf("http://%s:%s%s", *fqdn, "80", "/")
		http_helper.HttpGetWithRetryWithCustomValidation(t, url, nil, 15, 1*time.Second, func(status int, body string) bool {
			return status == 200
		})
	})

	t.Run("Should return 404 response", func(t *testing.T) {
		url := fmt.Sprintf("http://%s:%s%s", *fqdn, "80", "/error")
		http_helper.HttpGetWithRetryWithCustomValidation(t, url, nil, 5, 1*time.Second, func(status int, body string) bool {
			return status == 404
		})
	})

}

func before(t *testing.T) {
	if *fqdn == "" {
		t.Fatal("-fqdn flag is required")
	}

	k8sOptions = k8s.NewKubectlOptions("", "", namespaceName)

	err := createNamespaceWithIstioEnabled(namespaceName)
	if err != nil {
		t.Fatal(err)
	}

	helmChartPath, err := filepath.Abs(helmChartName)
	if err != nil {
		t.Fatal(err)
	}

	helmOptions = &helm.Options{
		KubectlOptions: k8sOptions,
		SetValues: map[string]string{
			"virtualservice.host.fqdn": *fqdn,
		},
	}

	helm.Install(t, helmOptions, helmChartPath, helmChartName)
	k8s.WaitUntilServiceAvailable(t, k8sOptions, "conn-test", 15, 2*time.Second)
}

func after(t *testing.T) func() {
	return func() {
		helm.Delete(t, helmOptions, helmChartName, true)
		k8s.DeleteNamespace(t, k8sOptions, namespaceName)
	}
}

func createNamespaceWithIstioEnabled(namespace string) error {

	kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube/config")

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	_, err = clientset.CoreV1().Namespaces().Create(&v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
			Labels: map[string]string{
				"istio-injection": "enabled",
			},
		},
	})
	if err != nil {
		return err
	}

	return nil
}
