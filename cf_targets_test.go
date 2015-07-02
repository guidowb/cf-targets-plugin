package main

import (
	realos "os"

	"github.com/cloudfoundry/cli/plugin/fakes"
	. "github.com/cloudfoundry/cli/testhelpers/io"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type FakeOS struct {
	exitCalled                 bool
	exitCalledWithCode         int
	mkdirCalled                int
	mkdirCalledWithPath        string
	mkdirCalledWithMode        realos.FileMode
	removeCalled               int
	removeCalledWithPath       string
	symlinkCalled              int
	symlinkCalledWithTarget    string
	symlinkCalledWithSource    string
	readdirCalled              int
	readdirCalledWithPath      string
	readfileCalled             int
	readfileCalledWithPath     string
	writefileCalled            int
	writefileCalledWithPath    string
	writefileCalledWithContent []byte
	writefileCalledWithMode    realos.FileMode
	readdirShouldReturn        []realos.FileInfo
	readfileShouldReturn       []byte
}

func (os *FakeOS) Exit(code int) {
	os.exitCalled = true
	os.exitCalledWithCode = code
}

func (os *FakeOS) Mkdir(path string, mode realos.FileMode) {
	os.mkdirCalled++
	os.mkdirCalledWithPath = path
	os.mkdirCalledWithMode = mode
}

func (os *FakeOS) Remove(path string) {
	os.removeCalled++
	os.removeCalledWithPath = path
}

func (os *FakeOS) Symlink(target string, source string) error {
	os.symlinkCalled++
	os.symlinkCalledWithTarget = target
	os.symlinkCalledWithSource = source
	return nil
}

func (os *FakeOS) ReadDir(path string) ([]realos.FileInfo, error) {
	os.readdirCalled++
	os.readdirCalledWithPath = path
	return os.readdirShouldReturn, nil
}

func (os *FakeOS) ReadFile(path string) ([]byte, error) {
	os.readfileCalled++
	os.readfileCalledWithPath = path
	return os.readfileShouldReturn, nil
}

func (os *FakeOS) WriteFile(path string, content []byte, mode realos.FileMode) error {
	os.writefileCalled++
	os.writefileCalledWithPath = path
	os.writefileCalledWithContent = content
	os.writefileCalledWithMode = mode
	return nil
}

var _ = Describe("TargetsPlugin", func() {
	Describe("Command Syntax", func() {
		var fakeCliConnection *fakes.FakeCliConnection
		var targetsPlugin *TargetsPlugin
		var fakeOS FakeOS

		BeforeEach(func() {
			fakeOS = FakeOS{}
			os = &fakeOS
			fakeCliConnection = &fakes.FakeCliConnection{}
			targetsPlugin = &TargetsPlugin{}
		})

		It("displays usage when targets called with too many arguments", func() {
			output := CaptureOutput(func() {
				targetsPlugin.Run(fakeCliConnection, []string{"targets", "blah"})
			})
			Expect(fakeOS.exitCalled).To(Equal(true))
			Expect(fakeOS.exitCalledWithCode).To(Equal(1))
			Expect(output).To(ContainSubstrings([]string{"Usage:", "cf", "targets"}))
		})

		It("displays usage when set-target called with too many arguments", func() {
			output := CaptureOutput(func() {
				targetsPlugin.Run(fakeCliConnection, []string{"set-target", "blah", "blah"})
			})
			Expect(fakeOS.exitCalled).To(Equal(true))
			Expect(fakeOS.exitCalledWithCode).To(Equal(1))
			Expect(output).To(ContainSubstrings([]string{"Usage:", "cf", "set-target", "[-f]", "NAME"}))
		})

		It("displays usage when set-target called with too few arguments", func() {
			output := CaptureOutput(func() {
				targetsPlugin.Run(fakeCliConnection, []string{"set-target"})
			})
			Expect(fakeOS.exitCalled).To(Equal(true))
			Expect(fakeOS.exitCalledWithCode).To(Equal(1))
			Expect(output).To(ContainSubstrings([]string{"Usage:", "cf", "set-target", "[-f]", "NAME"}))
		})

		It("displays usage when set-target called with unsupported option", func() {
			output := CaptureOutput(func() {
				targetsPlugin.Run(fakeCliConnection, []string{"set-target", "blah", "-k"})
			})
			Expect(fakeOS.exitCalled).To(Equal(true))
			Expect(fakeOS.exitCalledWithCode).To(Equal(1))
			Expect(output).To(ContainSubstrings([]string{"Usage:", "cf", "set-target", "[-f]", "NAME"}))
		})

		It("displays usage when save-target called with too many arguments", func() {
			output := CaptureOutput(func() {
				targetsPlugin.Run(fakeCliConnection, []string{"save-target", "blah", "blah"})
			})
			Expect(fakeOS.exitCalled).To(Equal(true))
			Expect(fakeOS.exitCalledWithCode).To(Equal(1))
			Expect(output).To(ContainSubstrings([]string{"Usage:", "cf", "save-target", "[-f]", "[NAME]"}))
		})

		It("displays usage when save-target called with unsupported option", func() {
			output := CaptureOutput(func() {
				targetsPlugin.Run(fakeCliConnection, []string{"save-target", "blah", "-k"})
			})
			Expect(fakeOS.exitCalled).To(Equal(true))
			Expect(fakeOS.exitCalledWithCode).To(Equal(1))
			Expect(output).To(ContainSubstrings([]string{"Usage:", "cf", "save-target", "[-f]", "[NAME]"}))
		})

		It("displays usage when delete-target called with too few arguments", func() {
			output := CaptureOutput(func() {
				targetsPlugin.Run(fakeCliConnection, []string{"delete-target"})
			})
			Expect(fakeOS.exitCalled).To(Equal(true))
			Expect(fakeOS.exitCalledWithCode).To(Equal(1))
			Expect(output).To(ContainSubstrings([]string{"Usage:", "cf", "delete-target", "NAME"}))
		})

		It("displays usage when delete-target called with too many arguments", func() {
			output := CaptureOutput(func() {
				targetsPlugin.Run(fakeCliConnection, []string{"delete-target", "blah", "blah"})
			})
			Expect(fakeOS.exitCalled).To(Equal(true))
			Expect(fakeOS.exitCalledWithCode).To(Equal(1))
			Expect(output).To(ContainSubstrings([]string{"Usage:", "cf", "delete-target", "NAME"}))
		})
	})
})
