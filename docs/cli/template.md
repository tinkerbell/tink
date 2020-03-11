## rover template

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
  $ rover template create -n <template-name> -p <path-to-template>
  $ rover template create -n sample -p /tmp/sample.tmpl
 ``` 

 - List all the templates 
 ```shell
  $ rover template list
 ```

 - Update the name of an existing template
 ```shell
  $ rover template update <template-uuid> -n <new-name>
  $ rover template update edb80a56-b1f2-4502-abf9-17326324192b -n new-sample-template
 ```

 - Update an existing template and keep the same name
 ```shell
  $ rover template update <template-uuid> -p <path-to-new-template-file>
  $ rover template update edb80a56-b1f2-4502-abf9-17326324192b -p /tmp/new-sample-template.tmpl
 ```

### See Also

 - [rover hardware](hardware.md) - Hardware (worker) data operations 
 - [rover target](target.md) - Target operations
 - [rover workflow](workflow.md) - Workflow operations

