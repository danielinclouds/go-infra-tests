package main_test

import (
	"fmt"
	"time"

	http_helper "github.com/gruntwork-io/terratest/modules/http-helper"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Connectivity", func() {

	Context("When I send Get request on / path to application", func() {
		It("Then I get 200 response", func() {
			url := fmt.Sprintf("http://%s:%s%s", *fqdn, "80", "/")
			err := http_helper.HttpGetWithRetryWithCustomValidationE(NewTestingT(), url, nil, 15, 1*time.Second, func(status int, body string) bool {
				return status == 200
			})
			Expect(err).To(BeNil())
		})
	})

	Context("When I send Get request on /hello path to application", func() {
		It("Then I get 404 response", func() {
			url := fmt.Sprintf("http://%s:%s%s", *fqdn, "80", "/hello")
			err := http_helper.HttpGetWithRetryWithCustomValidationE(NewTestingT(), url, nil, 5, 1*time.Second, func(status int, body string) bool {
				return status == 404
			})
			Expect(err).To(BeNil())
		})
	})
})
