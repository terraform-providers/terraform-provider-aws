<!--- See what makes a good Pull Request at : https://github.com/hashicorp/terraform-provider-aws/blob/master/docs/CONTRIBUTING.md --->

<!--- Please keep this note for the community --->

### Community Note

* Please vote on this pull request by adding a 👍 [reaction](https://blog.github.com/2016-03-10-add-reactions-to-pull-requests-issues-and-comments/) to the original pull request comment to help the community and maintainers prioritize this request
* Please do not leave "+1" or other comments that do not add relevant new information or questions, they generate extra noise for pull request followers and do not help prioritize the request

<!--- Thank you for keeping this note for the community --->

<!--- If your PR fully resolves and should automatically close the linked issue, use Closes. Otherwise, use Relates --->
Relates OR Closes #0000

Release note for [CHANGELOG](https://github.com/hashicorp/terraform-provider-aws/blob/master/CHANGELOG.md):
<!--
If change is not user facing, just write "NONE" in the release-note block below.
-->

```release-note

```

Output from acceptance testing:

<!--
Replace TestAccXXX with a pattern that matches the tests affected by this PR.

For more information on the `-run` flag, see the `go test` documentation at https://tip.golang.org/cmd/go/#hdr-Testing_flags.
-->
```
$ make testacc TESTARGS='-run=TestAccXXX'

...
```
