# ttsum

> ttsum is a plugin that helps summarize taints and tolerations across resources

## Use cases

- Helps cluster admins make modifications to taints & tolerations

- List nodes with/without a specific taint

- List resources with/without a specific toleration

## Install

```text
$ go install github.com/eytan-avisror/ttsum
```

## Usage from command line

Either download/install a released binary or add as a plugin to kubectl via Krew

```text
$ ttsum
ttsum helps summarize tainted nodes and tolerating resources

Usage:
  ttsum [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  taints      taints summarizes taints for nodes, and whether they will accept a toleration
  tolerations tolerations summarizes tolerations for a resource
  version     Version of ttsum

Flags:
  -h, --help   help for ttsum

Use "ttsum [command] --help" for more information about a command.
```

Get all tolerations belonging to deployments in a namespace

```text
$ ttsum tolerations apps/v1 deployments -n eytan-avisror
NAMESPACE    	NAME                        	TOLERATIONS
eytan-avisror	nginx      	                  Equal(app=web:NoSchedule)
eytan-avisror	mysql                       	Equal(app=db:NoSchedule)
```

Get all tolerations belonging to deployments in a namespace with a matcher

```text
$ ttsum tolerations apps/v1 deployments -n eytan-avisror --match "Equal(app=web:NoSchedule)"
NAMESPACE    	NAME                        	TOLERATIONS
eytan-avisror	nginx      	                  Equal(app=web:NoSchedule)

$ ttsum tolerations apps/v1 deployments -n eytan-avisror --no-match "Equal(app=web:NoSchedule)"
NAMESPACE    	NAME                        	TOLERATIONS
eytan-avisror	mysql                       	Equal(app=db:NoSchedule)
```

List node taints

```text
$ ttsum taints
NAME                         	  TAINTS
ip-10-82-155-58.ec2.internal 	  app=web:NoSchedule
ip-10-82-253-196.ec2.internal	  app=web:NoSchedule
ip-10-82-204-150.ec2.internal	  app=web:NoSchedule
ip-10-82-183-247.ec2.internal  	app=web:NoSchedule
ip-10-82-215-83.ec2.internal 	  app=web:NoSchedule
ip-10-82-198-233.ec2.internal  	app=db:NoSchedule
ip-10-82-190-200.ec2.internal	  app=db:NoSchedule
```

Similarly you can use a match selector
```text
$ ttsum taints --match "app=web:NoSchedule)"
NAME                         	  TAINTS
ip-10-82-155-58.ec2.internal 	  app=web:NoSchedule
ip-10-82-253-196.ec2.internal	  app=web:NoSchedule
ip-10-82-204-150.ec2.internal	  app=web:NoSchedule
ip-10-82-183-247.ec2.internal  	app=web:NoSchedule
ip-10-82-215-83.ec2.internal 	  app=web:NoSchedule

$ ttsum taints --match "app=db:NoSchedule)"
NAME                         	  TAINTS
ip-10-82-198-233.ec2.internal  	app=db:NoSchedule
ip-10-82-190-200.ec2.internal	  app=db:NoSchedule
```
