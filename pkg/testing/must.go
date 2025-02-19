package testing

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Must is a test helper that fails when error has occurred.
func Must[V any](v V, err error) V {
	GinkgoHelper()
	Expect(err).NotTo(HaveOccurred())
	Expect(v).NotTo(BeNil())

	return v
}
