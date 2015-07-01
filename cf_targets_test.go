package main

import (
	"github.com/cloudfoundry/cli/plugin/fakes"
	. "github.com/cloudfoundry/cli/testhelpers/io"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	fakeExitCalled         bool
	fakeExitCalledWithCode int
)

func fakeExit(code int) {
	fakeExitCalled = true
	fakeExitCalledWithCode = code
}

var _ = Describe("TargetsPlugin", func() {
	Describe("Run()", func() {
		var fakeCliConnection *fakes.FakeCliConnection
		var targetsPlugin *TargetsPlugin

		BeforeEach(func() {
			fakeCliConnection = &fakes.FakeCliConnection{}
			targetsPlugin = &TargetsPlugin{exit: fakeExit}
		})

		It("displays usage when targets called with too many arguments", func() {
			output := CaptureOutput(func() {
				targetsPlugin.Run(fakeCliConnection, []string{"targets", "blah"})
			})

			Expect(fakeExitCalled).To(Equal(true))
			Expect(fakeExitCalledWithCode).To(Equal(1))
			Expect(output).To(ContainSubstrings([]string{"Usage:", "cf", "targets"}))
		})
	})
})
