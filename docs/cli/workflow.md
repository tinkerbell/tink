## tink workflow

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
  -h, --help   help for workflow
```

### Examples

 - Create a workflow using a template and hardware devices
 ```shell
  $ tink workflow create -t <template-uuid> -r <hardware_input_in_json_format>
  $ tink workflow create -t edb80a56-b1f2-4502-abf9-17326324192b -r {"device_1": "mac/IP"}
 ```
 #### Note:
 1. The key used in the above command which is "device_1" should be in sync with "worker" field in the template.
    Click [here](../concepts.md) to check the template structure.
 2. These keys can only contain letter, numbers and underscore.

### See Also

 - [tink hardware](hardware.md) - Hardware (worker) data operations
 - [tink template](template.md) - Template operations

