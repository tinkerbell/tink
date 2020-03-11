## rover workflow

Workflow operations.

### Synopsis

Workflow operations:
```shell
  create      create a workflow
  data        get workflow data
  delete      delete a workflow
  events      show all events for a workflow
  get         get a workflow
  list        list all workflows
  state       get the current workflow context
```

### Options

```
  -h, --help   help for target
```

### Examples

 - Create a workflow using a template and a target
 ```shell
  $ rover workflow create -t <template-uuid> -r <target-uuid>
  $ rover workflow create -t edb80a56-b1f2-4502-abf9-17326324192b -r 9356ae1d-6165-4890-908d-7860ed04b421
 ```

### See Also

 - [rover hardware](hardware.md) - Hardware (worker) data operations 
 - [rover target](target.md) - Target operations
 - [rover template](template.md) - Template operations
 
