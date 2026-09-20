# Terraform provider for TikTok Ads

An independent, community-maintained Terraform provider for TikTok Marketing API
v1.3. It calls the REST API directly; no MCP server, Node.js process, or SDK is
needed at runtime. This project is not affiliated with or endorsed by TikTok.

## Support

| Terraform type | Behavior |
| --- | --- |
| `tiktok_campaign` | Manage regular auction campaigns |
| `tiktok_smart_plus_campaign` | Manage Upgraded Smart+ campaigns |
| `tiktok_inventory` | Read regular and Smart+ campaigns, ad groups, and ads |

Ad groups and ads are read-only. Campaign attributes are native HCL fields derived
from the pinned schemas in [campaigns.json](campaigns.json). The source snapshot
in [spec/operations.json](spec/operations.json) records its provenance and retrieval
date. These are MCP JSON input schemas, not a complete OpenAPI response contract.

## Install from Git

Requires Go 1.27.1 and Terraform 1.14 or newer (below 2.0). A C compiler is needed
for race-enabled development tests, but not for the provider build.

```sh
git clone --branch v0.1.0 https://github.com/kishansripada/terraform-provider-tiktok.git
cd terraform-provider-tiktok
bash build.sh
export TF_CLI_CONFIG_FILE="$PWD/.terraform/install.tfrc"
terraform -chdir=examples/basic init -backend=false
terraform -chdir=examples/basic validate
```

The build installs a local Terraform filesystem mirror. Keep `TF_CLI_CONFIG_FILE`
set when running Terraform in your own configuration directory. The provider
uses `singingbuddy/tiktok` as its existing Terraform identity for state
compatibility. It is **not published on the Terraform Registry**; the mirror is
required. GitHub is the source distribution channel for this release.

To pin it inside another Git project:

```sh
git submodule add https://github.com/kishansripada/terraform-provider-tiktok.git vendor/terraform-provider-tiktok
git -C vendor/terraform-provider-tiktok checkout v0.1.0
git add .gitmodules vendor/terraform-provider-tiktok
bash vendor/terraform-provider-tiktok/build.sh
export TF_CLI_CONFIG_FILE="$PWD/vendor/terraform-provider-tiktok/.terraform/install.tfrc"
```

Commit the submodule pointer. Subsequent clones and CI checkouts need
`git submodule update --init` (or `git clone --recurse-submodules`). To upgrade,
review the new release, check out its tag in the submodule, and commit its new
pointer. Source builds can have different platform/toolchain checksums; regenerate
the consumer's provider lock entry deliberately when upgrading. Never remove
Terraform state to upgrade the provider. For iterative provider development,
`build.sh` also creates `.terraform/development.tfrc` with a development override.

## Configure

Set `TIKTOK_MARKETING_ACCESS_TOKEN` in your environment to an authorized Marketing
API token for the advertiser. Do not put tokens in HCL or commit them.

```hcl
terraform {
  required_providers {
    tiktok = {
      source  = "singingbuddy/tiktok"
      version = "0.1.0"
    }
  }
}

provider "tiktok" {
  advertiser_id = "1234567890123456789" # Replace with your advertiser ID.
  allow_writes  = false                 # Explicit opt-in to remote changes.
  allow_delete  = false                 # Separate opt-in to terminal deletion.
}

data "tiktok_inventory" "account" {}
```

See [examples/basic/main.tf](examples/basic/main.tf) for a disabled campaign
example. Review a plan before enabling writes. Import an existing campaign using
`advertiser_id/campaign_id`, for example:

```sh
terraform import tiktok_campaign.example '1234567890123456789/9876543210987654321'
```

Both resource types support the same import identity format. Import and refresh
still need API credentials. Inventory's `resources_json` contains raw API objects
and may contain overlapping regular/Smart+ results. Protect Terraform state and
saved plans as account data.

## Lifecycle and limitations

- Writes and deletion default to disabled. New campaigns require explicit
  `operation_status = "DISABLE"`; enabling is a separate apply.
- Omitted fields are unmanaged. Removing a field does not reset it remotely.
- Immutable changes and missing/deleted campaigns fail instead of silently
  replacing or recreating campaigns. Reach & Frequency creation is unsupported.
- Pause is applied before settings; enable follows successful settings read-back.
  The provider checks drift before updates and verifies resulting values.
- Deletion checks all four child collections, including unmanaged children.
  Use Terraform `prevent_destroy` for an additional configuration guard.
- Writes are not retried. Partial failures preserve observed state. Ambiguous
  creates return a blocking `pending-...` identity: inspect TikTok and import the
  verified campaign before retrying. A process killed before Terraform persists
  its response can still lose that identity; there is no durable write journal.
- TikTok has no atomic compare-and-swap for these operations; other writers can
  race between a check and a write. Account eligibility and permissions also
  constrain which settings the API accepts.

This provider does not include any organization's deployment approval policy,
backend credentials, campaign configuration, or state.

## Development

```sh
gofmt -w *.go
go vet ./...
TF_ACC=1 go test -race ./...
bash build.sh
```

Acceptance tests run real Terraform against local HTTP fixtures and do not need
TikTok credentials or mutate live ads. The production provider has no API-host
override. Tests cover both campaign families, import, drift, lifecycle sequencing,
validation, isolation, pagination, and recovery from partial failures.

Contributions should include a fixture regression test for behavior changes.
Run the commands above before opening a pull request. Never include tokens,
account exports, Terraform state, or real campaign identifiers in a report.

## License and upstream schemas

Provider code is MIT licensed; see [LICENSE](LICENSE). Schema descriptions are
sourced from TikTok's official MCP tool metadata, with provenance retained in
[NOTICE](NOTICE). TikTok API access remains subject to TikTok's applicable terms.
