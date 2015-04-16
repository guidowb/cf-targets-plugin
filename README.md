CF Targets Plugin
=================
This plugin facilitates the use of multiple api targets with the Cloud Foundry CLI.

It originated from the need for a Go play project, and the realization that I was
frequently switching back and forth between development and various test environments,
using tricks like

```
CF_HOME=~/cf-development cf push my-app
CF_HOME=~/cf-production cf push my-app
```

This plugin makes switching a lot less painful by allowing you to save your currently
configured target using a name, then switching back to it by name at any point.


##Usage

Configure and save any number of named targets

```
$ cf api <development-target-url>
$ cf login
...
$ cf save-target development
```

Followed by

```
$ cf api <production-target-url>
$ cf login
...
$ cf save-target production
```

After saving targets, easily switch back and forth between them using:

```
$ cf set-target development
$ cf target
API Endpoint:   <development-target-url>
...
$ cf set-target production
$ cf target
API Endpoint:   <production-target-url>
...
```

View saved targets using

```
$ cf targets
development
production (current)
```


##Installation
#####Install from CLI
  ```
  $ cf add-plugin-repo CF-Community http://plugins.cloudfoundry.org/
  $ cf install-plugin cf-targets-plugin -r CF-Community
  ```
  
  
#####Install from Source (need to have [Go](http://golang.org/dl/) installed)
  ```
  $ go get github.com/cloudfoundry/cli
  $ go get github.com/guidowb/cf-targets-plugin
  $ cd $GOPATH/src/github.com/guidowb/cf-targets-plugin
  $ go build
  $ cf install-plugin cf-targets-plugin
  ```

##Full Command List

| command | usage | description|
| :--------------- |:---------------| :------------|
|`targets`| `cf targets` |list all saved targets|
|`save-target`|`cf save-target [-f] [<name>]`|save the current target for later use|
|`set-target`|`cf set-target [-f] <name>`|restore a previously saved target|
|`delete-target`|`cf delete-target <name>`|delete a previously saved target|
