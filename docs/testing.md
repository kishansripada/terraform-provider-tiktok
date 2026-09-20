# Testing

This project follows HashiCorp's [provider acceptance-testing framework](https://developer.hashicorp.com/terraform/plugin/testing/acceptance-tests).
It uses `terraform-plugin-framework` for the provider and `terraform-plugin-testing`
for acceptance tests, with protocol 6 provider factories.

```sh
make test       # Go unit, contract, and failure-injection tests
make testacc    # Real Terraform CLI lifecycle tests, with Go race detection
make coverage  # Unit and acceptance coverage; writes ignored coverage.out
make check     # Formatting, vet, unit tests, acceptance tests
```

Install the pinned Go version from go.mod and Terraform 1.14.7 (the CI version).
Tests find Terraform on PATH; `TF_ACC_TERRAFORM_PATH` can select an installed binary.
`TF_ACC=1` enables acceptance tests, following HashiCorp conventions.

Acceptance tests check create/update plans, state values, import equivalence,
refresh/drift reconciliation, dependency references, and destroy cleanup. Terraform
also checks that each applied configuration leaves an empty plan. Local HTTP
fixtures exercise JSON envelopes and replacement semantics against the pinned
request schemas.
Failure tests cover stale plans, partial writes, ambiguous create responses,
missing/deleted objects, account isolation, ownership removal, protected deletion,
and children preventing parent deletion.

CI runs formatting, `go vet`, unit tests, race-enabled acceptance tests, a source
build, and Terraform example validation. A passing fixture suite validates the
provider/Terraform integration, **not TikTok's live behavior**. Input schemas do not
fully specify response shapes, eventual consistency, account permissions, or every
conditional business rule. Before promoting new resource behavior for production,
validate it in an explicitly authorized isolated TikTok test advertiser through
CI. No live tests or sweepers run automatically, and tests never load advertiser
credentials. The production provider cannot redirect its API host to a fixture.

When adding an endpoint, add its pinned request schema and verified route, test
its response handling and ownership/lifecycle rules, and add a real Terraform
acceptance flow before claiming managed-resource support. Do not count catalog
entries, fixture count, or line coverage as proof of live endpoint coverage.
