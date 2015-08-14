package main

import (
	"bytes"
	"flag"
	"fmt"
	"path/filepath"
	"strings"

	realio "io/ioutil"
	realos "os"

	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/configuration/config_helpers"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/plugin"
)

type TargetsPlugin struct {
	configPath  string
	targetsPath string
	currentPath string
	suffix      string
	status      TargetStatus
}

type TargetStatus struct {
	currentHasName     bool
	currentName        string
	currentNeedsSaving bool
	currentNeedsUpdate bool
}

type RealOS struct{}
type OS interface {
	Exit(int)
	Mkdir(string, realos.FileMode)
	Remove(string)
	Symlink(string, string) error
	ReadDir(string) ([]realos.FileInfo, error)
	ReadFile(string) ([]byte, error)
	WriteFile(string, []byte, realos.FileMode) error
}

func (*RealOS) Exit(code int)                                  { realos.Exit(code) }
func (*RealOS) Mkdir(path string, mode realos.FileMode)        { realos.Mkdir(path, mode) }
func (*RealOS) Remove(path string)                             { realos.Remove(path) }
func (*RealOS) Symlink(target string, source string) error     { return realos.Symlink(target, source) }
func (*RealOS) ReadDir(path string) ([]realos.FileInfo, error) { return realio.ReadDir(path) }
func (*RealOS) ReadFile(path string) ([]byte, error)           { return realio.ReadFile(path) }
func (*RealOS) WriteFile(path string, content []byte, mode realos.FileMode) error {
	return realio.WriteFile(path, content, mode)
}

var os OS

func newTargetsPlugin() *TargetsPlugin {
	targetsPath := filepath.Join(filepath.Dir(config_helpers.DefaultFilePath()), "targets")
	os.Mkdir(targetsPath, 0700)
	return &TargetsPlugin{
		configPath:  config_helpers.DefaultFilePath(),
		targetsPath: targetsPath,
		currentPath: filepath.Join(targetsPath, "current"),
		suffix:      "." + filepath.Base(config_helpers.DefaultFilePath()),
	}
}

func (c *TargetsPlugin) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "cf-targets",
		Version: plugin.VersionType{
			Major: 1,
			Minor: 1,
			Build: 0,
		},
		Commands: []plugin.Command{
			{
				Name:     "targets",
				HelpText: "List available targets",
				UsageDetails: plugin.Usage{
					Usage: "cf targets",
				},
			},
			{
				Name:     "set-target",
				HelpText: "Set current target",
				UsageDetails: plugin.Usage{
					Usage: "cf set-target [-f] NAME",
					Options: map[string]string{
						"f": "replace the current target even if it has not been saved",
					},
				},
			},
			{
				Name:     "save-target",
				HelpText: "Save current target",
				UsageDetails: plugin.Usage{
					Usage: "cf save-target [-f] [NAME]",
					Options: map[string]string{
						"f": "save the target even if the specified name already exists",
					},
				},
			},
			{
				Name:     "delete-target",
				HelpText: "Delete a saved target",
				UsageDetails: plugin.Usage{
					Usage: "cf delete-target NAME",
				},
			},
		},
	}
}

func main() {
	os = &RealOS{}
	plugin.Start(newTargetsPlugin())
}

func (c *TargetsPlugin) Run(cliConnection plugin.CliConnection, args []string) {
	defer func() {
		reason := recover()
		if code, ok := reason.(int); ok {
			os.Exit(code)
		} else if reason != nil {
			panic(reason)
		}
	}()

	c.checkStatus()
	if args[0] == "targets" {
		c.TargetsCommand(args)
	} else if args[0] == "set-target" {
		c.SetTargetCommand(args)
	} else if args[0] == "save-target" {
		c.SaveTargetCommand(args)
	} else if args[0] == "delete-target" {
		c.DeleteTargetCommand(args)
	}
}

func (c *TargetsPlugin) TargetsCommand(args []string) {
	if len(args) != 1 {
		c.exitWithUsage("targets")
	}
	targets := c.getTargets()
	if len(targets) < 1 {
		fmt.Println("No targets have been saved yet. To save the current target, use:")
		fmt.Println("   cf save-target NAME")
	} else {
		for _, target := range targets {
			var qualifier string
			if c.isCurrent(target) {
				qualifier = "(current"
				if c.status.currentNeedsSaving {
					qualifier += ", modified"
				} else if c.status.currentNeedsUpdate {
					qualifier += "*"
				}
				qualifier += ")"
			}
			fmt.Println(target, qualifier)
		}
	}
}

func (c *TargetsPlugin) SetTargetCommand(args []string) {
	flagSet := flag.NewFlagSet("set-target", flag.ContinueOnError)
	force := flagSet.Bool("f", false, "force")
	err := flagSet.Parse(args[1:])
	if err != nil || len(flagSet.Args()) != 1 {
		c.exitWithUsage("set-target")
	}
	targetName := flagSet.Arg(0)
	targetPath := c.targetPath(targetName)
	if (!c.targetExists(targetPath)) {
		fmt.Println("Target", targetName, "does not exist.")
		panic(1)
	}
	if *force || !c.status.currentNeedsSaving {
		c.copyContents(targetPath, c.configPath)
		c.linkCurrent(targetPath)
	} else {
		fmt.Println("Your current target has not been saved. Use save-target first, or use -f to discard your changes.")
		panic(1)
	}
	fmt.Println("Set target to", targetName)
}

func (c *TargetsPlugin) SaveTargetCommand(args []string) {
	flagSet := flag.NewFlagSet("save-target", flag.ContinueOnError)
	force := flagSet.Bool("f", false, "force")
	err := flagSet.Parse(args[1:])
	if err != nil || len(flagSet.Args()) > 1 {
		c.exitWithUsage("save-target")
	}
	if len(flagSet.Args()) < 1 {
		c.SaveCurrentTargetCommand(*force)
	} else {
		c.SaveNamedTargetCommand(flagSet.Arg(0), *force)
	}
}

func (c *TargetsPlugin) SaveNamedTargetCommand(targetName string, force bool) {
	targetPath := c.targetPath(targetName)
	if force || !c.targetExists(targetPath) {
		c.copyContents(c.configPath, targetPath)
		c.linkCurrent(targetPath)
	} else {
		fmt.Println("Target", targetName, "already exists. Use -f to overwrite it.")
		panic(1)
	}
	fmt.Println("Saved current target as", targetName)
}

func (c *TargetsPlugin) SaveCurrentTargetCommand(force bool) {
	if !c.status.currentHasName {
		fmt.Println("Current target has not been previously saved. Please provide a name.")
		panic(1)
	}
	targetName := c.status.currentName
	targetPath := c.targetPath(targetName)
	if c.status.currentNeedsSaving && !force {
		fmt.Println("You've made substantial changes to the current target.")
		fmt.Println("Use -f if you intend to overwrite the target named", targetName, "or provide an alternate name")
		panic(1)
	}
	c.copyContents(c.configPath, targetPath)
	fmt.Println("Saved current target as", targetName)
}

func (c *TargetsPlugin) DeleteTargetCommand(args []string) {
	if len(args) != 2 {
		c.exitWithUsage("delete-target")
	}
	targetName := args[1]
	targetPath := c.targetPath(targetName)
	if !c.targetExists(targetPath) {
		fmt.Println("Target", targetName, "does not exist")
		panic(1)
	}
	os.Remove(targetPath)
	if c.isCurrent(targetName) {
		os.Remove(c.currentPath)
	}
	fmt.Println("Deleted target", targetName)
}

func (c *TargetsPlugin) getTargets() []string {
	var targets []string
	files, _ := os.ReadDir(c.targetsPath)
	for _, file := range files {
		filename := file.Name()
		if strings.HasSuffix(filename, c.suffix) {
			targets = append(targets, strings.TrimSuffix(filename, c.suffix))
		}
	}
	return targets
}

func (c *TargetsPlugin) targetExists(targetPath string) bool {
	target := configuration.NewDiskPersistor(targetPath)
	return target.Exists()
}

func (c *TargetsPlugin) checkStatus() {
	currentConfig := configuration.NewDiskPersistor(c.configPath)
	currentTarget := configuration.NewDiskPersistor(c.currentPath)
	if !currentTarget.Exists() {
		os.Remove(c.currentPath)
		c.status = TargetStatus{false, "", true, false}
		return
	}

	name := c.getCurrent()

	configData := core_config.NewData()
	targetData := core_config.NewData()

	err := currentConfig.Load(configData)
	c.checkError(err)
	err = currentTarget.Load(targetData)
	c.checkError(err)

	// Ignore the access-token field, as it changes frequently
	needsUpdate := targetData.AccessToken != configData.AccessToken
	targetData.AccessToken = configData.AccessToken

	currentContent, err := configData.JsonMarshalV3()
	c.checkError(err)
	savedContent, err := targetData.JsonMarshalV3()
	c.checkError(err)
	c.status = TargetStatus{true, name, !bytes.Equal(currentContent, savedContent), needsUpdate}
}

func (c *TargetsPlugin) copyContents(sourcePath, targetPath string) {
	content, err := os.ReadFile(sourcePath)
	c.checkError(err)
	err = os.WriteFile(targetPath, content, 0600)
	c.checkError(err)
}

func (c *TargetsPlugin) linkCurrent(targetPath string) {
	os.Remove(c.currentPath)
	err := os.Symlink(targetPath, c.currentPath)
	c.checkError(err)
}

func (c *TargetsPlugin) targetPath(targetName string) string {
	return filepath.Join(c.targetsPath, targetName+c.suffix)
}

func (c *TargetsPlugin) checkError(err error) {
	if err != nil {
		fmt.Println("Error:", err)
		panic(1)
	}
}

func (c *TargetsPlugin) exitWithUsage(command string) {
	metadata := c.GetMetadata()
	for _, candidate := range metadata.Commands {
		if candidate.Name == command {
			fmt.Println("Usage: " + candidate.UsageDetails.Usage)
			panic(1)
		}
	}
}

func (c *TargetsPlugin) getCurrent() string {
	targetPath, err := filepath.EvalSymlinks(c.currentPath)
	c.checkError(err)
	return strings.TrimSuffix(filepath.Base(targetPath), c.suffix)
}

func (c *TargetsPlugin) isCurrent(target string) bool {
	return c.status.currentHasName && c.status.currentName == target
}
