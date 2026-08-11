#!/usr/bin/env python3
"""Curate public Cloud API v2 routes into internal/assets/api-schemas/cloud_v2.json.

Adds every raw-spec path matching one of the given substrings, but ONLY the
publicly usable operations: any operation badged "Internal use only" (or
"Deprecated, ...") is dropped. Paths left with no public operation are skipped.
x-code-samples are stripped. Referenced schemas are pulled in (transitive closure
unioned with the schemas already present). Existing file formatting is preserved.

A substring prefixed with '=' means exact path match.

Usage:
    python3 scripts/curate_cloud_v2.py <raw_spec.json> <path-substring> [<path-substring> ...]

Example:
    curl -s "https://eu.api.ovh.com/v2/publicCloud.json?format=openapi3" > /tmp/raw_v2.json
    python3 scripts/curate_cloud_v2.py /tmp/raw_v2.json storage/file
"""
import json
import sys

REPO = "internal/assets/api-schemas/cloud_v2.json"
METHODS = {"get", "post", "put", "delete", "patch", "head", "options"}
PUBLIC = {"Alpha version", "Beta version", "Stable production version"}


def badge_labels(op):
    return {b.get("label") for b in op.get("x-badges", [])}


def main():
    if len(sys.argv) < 3:
        print(__doc__)
        sys.exit(1)
    raw_path, substrings = sys.argv[1], sys.argv[2:]

    repo = json.load(open(REPO))
    raw = json.load(open(raw_path))
    alls = raw["components"]["schemas"]

    def matches(p):
        for s in substrings:
            if s.startswith("="):
                if p == s[1:]:
                    return True
            elif s in p:
                return True
        return False

    added_paths, skipped_internal = [], []
    for p, node in raw["paths"].items():
        if not matches(p):
            continue
        public_ops = {}
        for k, v in node.items():
            if k not in METHODS:
                continue  # keep non-method keys (parameters, etc.) below
            if badge_labels(v) & PUBLIC:
                v.pop("x-code-samples", None)
                public_ops[k] = v
            else:
                skipped_internal.append(f"{k.upper()} {p}")
        if not public_ops:
            continue
        new_node = {k: v for k, v in node.items() if k not in METHODS}
        new_node.update(public_ops)
        repo["paths"][p] = new_node
        added_paths.append(p)

    def refs(o):
        out = set()
        if isinstance(o, dict):
            for k, v in o.items():
                if k == "$ref" and isinstance(v, str) and "#/components/schemas/" in v:
                    out.add(v.split("/")[-1])
                else:
                    out |= refs(v)
        elif isinstance(o, list):
            for i in o:
                out |= refs(i)
        return out

    seen = set(repo["components"]["schemas"].keys())
    frontier = set()
    for p in added_paths:
        frontier |= refs(repo["paths"][p])
    while frontier - seen:
        name = (frontier - seen).pop()
        seen.add(name)
        if name in alls:
            frontier |= refs(alls[name])

    added_schemas = 0
    for name in seen:
        if name not in repo["components"]["schemas"] and name in alls:
            repo["components"]["schemas"][name] = alls[name]
            added_schemas += 1

    with open(REPO, "w") as f:
        json.dump(repo, f, indent=2, ensure_ascii=False)
        f.write("\n")

    print(f"added paths   ({len(added_paths)}): " + ", ".join(added_paths))
    print(f"skipped internal ops ({len(skipped_internal)}): " + ", ".join(skipped_internal) if skipped_internal else "skipped internal ops (0)")
    print(f"schemas: {len(repo['components']['schemas'])} (+{added_schemas})")


if __name__ == "__main__":
    main()
