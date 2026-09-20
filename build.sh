#!/usr/bin/env bash
set -euo pipefail

provider_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
platform="$(go env GOOS)_$(go env GOARCH)"
plugin_dir="$provider_dir/.terraform/plugins/registry.terraform.io/singingbuddy/tiktok/0.2.0/$platform"
mkdir -p "$plugin_dir"
cd "$provider_dir"
CGO_ENABLED=0 go build -trimpath -buildvcs=false -o "$plugin_dir/terraform-provider-tiktok_v0.2.0" .
# This provider is installed from source. Terraform's development
# override deliberately avoids stale release checksums between local builds.
cat > "$provider_dir/.terraform/development.tfrc" <<EOF
provider_installation {
  dev_overrides {
    "singingbuddy/tiktok" = "$plugin_dir"
  }
  direct {}
}
EOF
cat > "$provider_dir/.terraform/install.tfrc" <<EOF
provider_installation {
  filesystem_mirror {
    path = "$provider_dir/.terraform/plugins"
    include = ["singingbuddy/tiktok"]
  }
  direct {
    exclude = ["singingbuddy/tiktok"]
  }
}
EOF
echo "Built TikTok provider. Set TF_CLI_CONFIG_FILE=$provider_dir/.terraform/install.tfrc"
