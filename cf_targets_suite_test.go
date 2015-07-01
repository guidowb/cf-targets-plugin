package main

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestTargets(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Targets Suite")
}
