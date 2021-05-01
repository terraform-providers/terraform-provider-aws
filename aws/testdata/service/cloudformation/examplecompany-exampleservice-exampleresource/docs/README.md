# ExampleCompany::ExampleService::ExampleResource

An example resource schema demonstrating some basic constructs and validation rules.

## Syntax

To declare this entity in your AWS CloudFormation template, use the following syntax:

### JSON

<pre>
{
    "Type" : "ExampleCompany::ExampleService::ExampleResource",
    "Properties" : {
        "<a href="#title" title="Title">Title</a>" : <i>String</i>,
        "<a href="#coversheetincluded" title="CoverSheetIncluded">CoverSheetIncluded</a>" : <i>Boolean</i>,
        "<a href="#duedate" title="DueDate">DueDate</a>" : <i>String</i>,
        "<a href="#approvaldate" title="ApprovalDate">ApprovalDate</a>" : <i>String</i>,
        "<a href="#memo" title="Memo">Memo</a>" : <i><a href="memo.md">Memo</a></i>,
        "<a href="#secondcopyofmemo" title="SecondCopyOfMemo">SecondCopyOfMemo</a>" : <i><a href="memo.md">Memo</a></i>,
        "<a href="#testcode" title="TestCode">TestCode</a>" : <i>String</i>,
        "<a href="#authors" title="Authors">Authors</a>" : <i>[ String, ... ]</i>,
        "<a href="#tags" title="Tags">Tags</a>" : <i>[ <a href="tag.md">Tag</a>, ... ]</i>
    }
}
</pre>

### YAML

<pre>
Type: ExampleCompany::ExampleService::ExampleResource
Properties:
    <a href="#title" title="Title">Title</a>: <i>String</i>
    <a href="#coversheetincluded" title="CoverSheetIncluded">CoverSheetIncluded</a>: <i>Boolean</i>
    <a href="#duedate" title="DueDate">DueDate</a>: <i>String</i>
    <a href="#approvaldate" title="ApprovalDate">ApprovalDate</a>: <i>String</i>
    <a href="#memo" title="Memo">Memo</a>: <i><a href="memo.md">Memo</a></i>
    <a href="#secondcopyofmemo" title="SecondCopyOfMemo">SecondCopyOfMemo</a>: <i><a href="memo.md">Memo</a></i>
    <a href="#testcode" title="TestCode">TestCode</a>: <i>String</i>
    <a href="#authors" title="Authors">Authors</a>: <i>
      - String</i>
    <a href="#tags" title="Tags">Tags</a>: <i>
      - <a href="tag.md">Tag</a></i>
</pre>

## Properties

#### Title

The title of the TPS report is a mandatory element.

_Required_: Yes

_Type_: String

_Minimum_: <code>20</code>

_Maximum_: <code>250</code>

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)

#### CoverSheetIncluded

Required for all TPS Reports submitted after 2/19/1999

_Required_: No

_Type_: Boolean

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)

#### DueDate

_Required_: No

_Type_: String

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)

#### ApprovalDate

_Required_: No

_Type_: String

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)

#### Memo

_Required_: No

_Type_: <a href="memo.md">Memo</a>

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)

#### SecondCopyOfMemo

_Required_: No

_Type_: <a href="memo.md">Memo</a>

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)

#### TestCode

_Required_: Yes

_Type_: String

_Allowed Values_: <code>NOT_STARTED</code> | <code>CANCELLED</code>

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)

#### Authors

_Required_: No

_Type_: List of String

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)

#### Tags

An array of key-value pairs to apply to this resource.

_Required_: No

_Type_: List of <a href="tag.md">Tag</a>

_Update requires_: [No interruption](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/using-cfn-updating-stacks-update-behaviors.html#update-no-interrupt)

## Return Values

### Ref

When you pass the logical ID of this resource to the intrinsic `Ref` function, Ref returns the TPSCode.

### Fn::GetAtt

The `Fn::GetAtt` intrinsic function returns a value for a specified attribute of this type. The following are the available attributes and sample return values.

For more information about using the `Fn::GetAtt` intrinsic function, see [Fn::GetAtt](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/intrinsic-function-reference-getatt.html).

#### TPSCode

A TPS Code is automatically generated on creation and assigned as the unique identifier.

