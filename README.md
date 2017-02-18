# qaz—A CLI tool for Templating & Managing stacks in AWS Cloudformation  [![Release](https://img.shields.io/github/release/daidokoro/qaz.svg)]
[![Build Status](https://travis-ci.org/daidokoro/qaz.svg)](https://travis-ci.org/daidokoro/qaz)
[![Go Report Card](https://goreportcard.com/badge/github.com/daidokoro/qaz)]


Qaz is a Fork of the Bora project by [@pkazmierczak](https://github.com/pkazmierczak) that aims to focus on simplifying the process of deploying infrastructure on AWS via Cloudformation by utilising the Go Templates Library and custom functions to generate diverse and configurable templates.

Qaz focuses on minimal abstraction from the underlying AWS Cloudformation Platform. It instead enhances customisability and re-usability of templates through dynamic template generation via Go Templates and custom template functions.

--

*Features:*

- Advanced template functionality & custom built-in template functions

- Support for Cloudformation templates written in JSON & YAML

- Dynamic deploy script generation utilising the built-in templating functionality

- Single YAML Configuration file for multiple stack templates per environment

- Utilises Go-routines for Multi-stack concurrent Cloudformation requests for *all* appropriate calls

- Support for AWS Profile selection for Multi-AWS account environments

- Cross stack referencing with support for Cloudformation Exports(_Preferred_) & dynamically retrieving stack outputs on deploy

- *Decoupled* build mechanism. Qaz can manage infrastructure by accessing config/templates via S3 or HTTP(S). The tool does not need to be in the same place as the files.

- *Decoupled* stack management. Stacks can be launched individually from different locations and build consistently according to the dependency chain as long as the same configuration file is read.


## Installation

If you have Golang installed:

`go get github.com/daidokoro/qaz`

Pre-build binaries for Darwin and Linux coming soon....

## Requirements
qaz requires:

- AWS credentials, you can read about how to set these up [here](http://blogs.aws.amazon.com/security/post/Tx3D6U6WSFGOK2H/A-New-and-Standardized-Way-to-Manage-Credentials-in-the-AWS-SDKs)


## How It Works!

Qaz uses a main _config.yml_ file as its source-of-truth. The file tells it what stacks it controls and the values to pass to those stacks.

```yaml

# Specify the AWS region code
# qaz will attempt to get it from AWS configuration
# or from the environment. This setting overrides
# every other.
region: "eu-west-1"

# Required: The project name is prepended to the
# stack names at build time to create unique
# identifier on the Cloudformation platform
project: "daidokoro"

# Optional: global values, accisible accross
# all stacks can be define under global
global:


# All stack specific values are defined
# under the "stacks" keyword below.

stacks:
  vpc:
    # Note: "cf" is a required keyword, which tells
    # qaz when to start reading in template values.
    cf:
      cidr: 10.10.0.0/16

  subnet:
    # Note: the "depends_on" keyword is used to list
    # stack dependencies. Any amount can be listed.
    # This key_word must be defined outside of "cf"
    depends_on:
      - vpc

    cf:
      subnets:
        - private: 10.10.0.0/24
        - public: 10.10.2.0/24
```

Note: Config files do not need to be named config.yml

--

## Templates (Getting those values!)

Go has an excellent and expandable templating library which is utilised in this project for additional logic in creating templates. To read more on Go Template see [Here](https://golang.org/pkg/text/template). All features of Go Template are supported in Qaz.

We'll run through some basic tips and tricks to get started.

Note that templates must have the same file name (_extension excluded_) as the stack they reference in config when working with local files, however, this does not apply when dealing with remote templates on S3 or via Http.

--

To access the values in our template we need to use template syntax. In its most basic form, to fetch say the value for `cidr` in my vpc stack config I would do the following:

```yaml
{{ .vpc.cidr }}
```

That's it! Use the generate command to varify the value
`$ qaz generate -c path/to/config -t path/to/template`

--

Go Templates are also capable of looping values, for example, to get the values of both _private_ & _public_ in my *subnets* stack, I would do the following.

```
{{ range $index, $value := .subnets.subnets }} # "range" allows us to loop over items in the template
  {{ range $access, $cidr := $value }} # looping over the key value pairs
    {{$access}} {{$cidr}} # printing output
  {{ end }}
{{ end }} # Closing loops
```

The above should give you the access level and subnets defined above. More examples as well as the full template implementation of this example can be found in the project examples folder.



#### Deploying Stacks

Stacks can be Deployed/Terminated with a single command.

![Alt text](demo/quick_build.gif?raw=true "Quick Build Demo")

The above however, only works when you are using Qaz in the root of your project directory. Alternatively, Qaz offers a few ways fetching both configuration and template files.

Configuration can be retrieved from both Http Get requests & S3.

```
$ qaz deploy -c s3://mybucket/super_config.yml -t vpc::http://someurl/vpc_dev.yml
```

The stack name must be specified using the syntax above. This tells Qaz what values to associate with this stack.

```
$ qaz deploy -c http://mybucket/super_config.yml -t vpc::s3://mybucket/vpc_dev.yml -t subnets::s3://mybucket/subnets.yml
```

You can pass as many `-t` flags as you have stacks, Qaz will deploy all in the correct order and manage the dependency chains as long as the `depends_on` keyword is utilised.

Note that the syntax for specifying stack names with URLs `stackname::url`. The deploy command does not require the stack name syntax when using local files, however the `update` command uses this syntax on *all* `-t --template` arguments. For example:

```
$ qaz deploy -c path/to/config -t path/to/template
$ qaz update -c path/to/config -t vpc::path/to/template
```

Deploy also takes wildcards for local templates. For example:

```
$ qaz deploy -c path/to/config.yml -t "path/*"
```
Quotes are required when using wildcards.

--

 ### Built in Template Functions

Template Functions expand the functionality of Go's Templating library by allowing you to execute external functions to retreive additional information for building your template.

Qaz supports all the Go Template functions as well as some custom ones. These include:

__file:__

A template function for reading values from an external file into a template. For now the file needs to be in the `files` directory in the rood of the project folder.

Example:

```
{{ myfile.txt | File }} # Returns the value of myfile.txt under the files directory
```

__s3_read:__

As the name suggests, this function reads the content of a given s3 key and writes it to the template.

Example:

```
{{ "s3://mybucket/key" | S3Read }} # writes the contents of the object to the template
```

__GET:__

GET implements http GET requests on a given url, and writes the response to the template.

Example
```
{{ "http://localhost" | GET }}
```

--

Qaz also supports Deploy-Time functions which are run just before the template is pushed to AWS. These include:

__stack_output__

stack_output fetches the output value of a given stack and stores the value in your template. This function using the stack name as defined in your project configuration

Example
```
# stackname::output

<< stack_output vpc::vpcid >>
```

__stack_output_ext__

stack_output_ext fetches the output value of a given stack that exists outside of your project/configuration and stores the value in your template. This function requires the full name of the stack as it appears on the AWS Console.

Example
```
# stackname::output

<< stack_output anotherproject-vpc::vpcid >>
```


When using Deploy-Time functions the Template delimiters are different: `<< >>` Qaz identifies items wrapped in these as Deploy-Time functions and only executes them just for before deploying to AWS.

--


See `examples` folder for more on usage. More examples to come.

```
$ qaz

  __ _   __ _  ____
 / _` | / _` ||_  /
| (_| || (_| | / /
\__, | \__,_|/___|
   |_|            

--> Shut up & deploy my templates...!

Usage:
qaz [flags]
qaz [command]

Available Commands:
check       Validates Cloudformation Templates
deploy      Deploys stack(s) to AWS
exports     Prints stack exports
generate    Generates template from configuration values
init        Creates a basic qaz project
invoke      Invoke AWS Lambda Functions
outputs     Prints stack outputs
status      Prints status of deployed/un-deployed stacks
tail        Tail Real-Time AWS Cloudformation events
terminate   Terminates stacks
update      Updates a given stack

Flags:
--debug            Run in debug mode...
-p, --profile string   configured aws profile (default "default")
--version          print current/running version

Use "qaz [command] --help" for more information about a command.

```


--
## Roadmap and status
qaz is in early development.

*TODO:*

- Implement Change-Set management
- More Comprehensive Documentation
- Implement Lambda invoke for API and lambda based event hooks
