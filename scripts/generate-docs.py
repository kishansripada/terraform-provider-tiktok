#!/usr/bin/env python3
"""Regenerate the coverage index and compact field references from pinned schemas."""
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent

def read(name):
    return json.loads((ROOT / name).read_text())

def type_name(p):
    if p['type'] == 'array':
        return 'list(' + type_name(p['items']) + ')'
    return p['type']

queries = read('queries.json')['operations']
objects = read('objects.json')
campaigns = read('campaigns.json')
kinds = ['campaign', 'smart_plus_campaign', 'adgroup', 'smart_plus_adgroup', 'ad', 'smart_plus_ad']
managed = {k + suffix for k in kinds for suffix in ['_create', '_update', '_get', '_status_update']}
for kind in kinds:
    source = campaigns if kind.endswith('campaign') else objects
    create, update = source[kind + '_create'], source[kind + '_update']
    if kind == 'ad':
        create, update = (dict(op['properties']['creatives']['items']) for op in (create, update))
        for op in (create, update):
            op['properties'] = dict(op['properties'], adgroup_id={'type': 'string'})
        create['required'] = create.get('required', []) + ['adgroup_id']
    props = dict(create['properties'], **update['properties'])
    for key in ['advertiser_id', 'request_id', 'campaign_id' if kind.endswith('campaign') else 'smart_plus_ad_id' if kind == 'smart_plus_ad' else kind.removeprefix('smart_plus_') + '_id']:
        props.pop(key, None)
    text = f'# tiktok_{kind}\n\nManage one {kind.replace("_", " ")} through the TikTok Marketing API.\n\n'
    text += 'Import: `terraform import tiktok_' + kind + '.example "ADVERTISER_ID/RESOURCE_ID"`.\n\n'
    text += 'New resources must set `operation_status = "DISABLE"`. Enable in a later apply.\n'
    text += 'Fields are optional in the Terraform schema to support partial ownership and imports; creation requirements below still apply. Conditional requirements and account eligibility are also enforced by TikTok.\n\n'
    text += 'Computed attributes: `id`, `advertiser_id`' + (', `observed_json` (raw last response)' if not kind.endswith('campaign') else '') + '.\n\n'
    text += '| Attribute | Type | Required to create | Update endpoint accepts |\n| --- | --- | --- | --- |\n'
    for key, p in sorted(props.items()):
        mutable = key == 'operation_status' or key in update['properties'] and key not in ['campaign_id', 'adgroup_id']
        text += f'| `{key}` | `{type_name(p)}` | {"Yes" if key in create.get("required", []) or key == "operation_status" else "No"} | {"Yes" if mutable else "No"} |\n'
    text += '\nFull field descriptions, nested object fields, enum values, and conditional rules are pinned in [' + ('campaigns.json' if kind.endswith('campaign') else 'objects.json') + '](../../' + ('campaigns.json' if kind.endswith('campaign') else 'objects.json') + ').\n'
    text += '\nSee [lifecycle limitations](../../README.md#lifecycle-and-limitations) before applying changes.\n'
    (ROOT / 'docs/resources' / (kind + '.md')).write_text(text)
text = '# API coverage\n\n'
text += f'This release has **6 managed resource types** and **{len(queries)} allowlisted read-only query operations**, plus the fully paginated six-collection inventory. It does **not** cover the entire TikTok API.\n\n'
text += 'The table audits the 379-operation MCP snapshot retrieved on 2026-09-20. That snapshot is not a complete OpenAPI response specification or a guarantee of all TikTok endpoints. Routes for generic queries are checked against the pinned official SDK in [spec/routes.json](../spec/routes.json).\n\n'
text += '“Resource” means the operation participates in a managed lifecycle. “Query” means one GET request through `tiktok_query`, with explicit parameters and pagination. “Not implemented” means no provider execution support; having a schema in the snapshot does not implement it.\n\n'
text += 'Unimplemented areas include persistent catalog/audience/pixel/Business Center writes as well as one-off payments, messages, event delivery, and media uploads. These need separate lifecycle designs and response-contract verification; they are not hidden behind a generic write escape hatch.\n\n'
text += '| Operation | Support | Query route |\n| --- | --- | --- |\n'
for op in sorted(read('spec/catalog.json')['operations'], key=lambda o:o['name']):
    n=op['name']; support=[]
    if n in managed: support.append('Resource')
    if n in queries: support.append('Query')
    text += f'| `{n}` | {", ".join(support) or "Not implemented"} | ' + ('`'+queries[n]['path']+'`' if n in queries else '—') + ' |\n'
(ROOT / 'docs/coverage.md').write_text(text)
