## tinkerbell template

Template operations.

### Synopsis

Template operations:
```shell
  create      create a workflow template 
  delete      delete a template
  get         get a template
  list        list all saved templates
  update      update a template
```

### Options

```
  -h, --help   help for target
```

### Examples

 - The following command creates a workflow template using the `sample.tmpl` file and save it as `sample`. It returns a UUID for the newly created template.
 ```shell
  $ tinkerbell template create -n <template-name> -p <path-to-template>
  $ tinkerbell template create -n sample -p /tmp/sample.tmpl
 ``` 

 - List all the templates 
 ```shell
  $ tinkerbell template list
 ```

 - Update the name of an existing template
 ```shell
  $ tinkerbell template update <template-uuid> -n <new-name>
  $ tinkerbell template update edb80a56-b1f2-4502-abf9-17326324192b -n new-sample-template
 ```

 - Update an existing template and keep the same name
 ```shell
  $ tinkerbell template update <template-uuid> -p <path-to-new-template-file>
  $ tinkerbell template update edb80a56-b1f2-4502-abf9-17326324192b -p /tmp/new-sample-template.tmpl
 ```

### See Also

 - [tinkerbell hardware](hardware.md) - Hardware (worker) data operations 
 - [tinkerbell target](target.md) - Target operations
 - [tinkerbell workflow](workflow.md) - Workflow operations

