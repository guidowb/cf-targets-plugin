package main

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestTargets(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CF Plugin Targets Suite")
}
