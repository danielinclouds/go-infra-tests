package main_test

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var namespaceName string
var k8sOptions *k8s.KubectlOptions
var helmOptions *helm.Options
var fqdn = flag.String("fqdn", "", "FQDN that will be used for exposing application")

const helmChartName = "conn-test"

type TestingT struct {
	GinkgoTInterface
	desc GinkgoTestDescription
}

func NewTestingT() TestingT {
	return TestingT{GinkgoT(), CurrentGinkgoTestDescription()}
}

func (i TestingT) Helper() {

}
func (i TestingT) Name() string {
	return i.desc.FullTestText
}

func TestProjectConnectivityTest(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ProjectConnectivityTest Suite")
}

var _ = BeforeSuite(func() {

	if *fqdn == "" {
		log.Println("-fqdn flag is required")
		os.Exit(1)
	}

	namespaceName = fmt.Sprintf("%s-%s", helmChartName, strings.ToLower(random.UniqueId()))
	err := createNamespaceWithIstioEnabled(namespaceName)
	if err != nil {
		panic(err)
	}

	helmChartPath, err := filepath.Abs(helmChartName)
	if err != nil {
		panic(err)
	}

	k8sOptions = k8s.NewKubectlOptions("", "", namespaceName)
	helmOptions = &helm.Options{
		KubectlOptions: k8sOptions,
		SetValues: map[string]string{
			"virtualservice.host.fqdn": *fqdn,
		},
	}

	helm.Install(NewTestingT(), helmOptions, helmChartPath, helmChartName)
	k8s.WaitUntilServiceAvailable(NewTestingT(), k8sOptions, "conn-test", 15, 2*time.Second)
})

var _ = AfterSuite(func() {
	defer k8s.DeleteNamespace(NewTestingT(), k8sOptions, namespaceName)
	helm.Delete(NewTestingT(), helmOptions, helmChartName, true)
})

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
