#!/usr/bin/env python3
"""Download the kabupaten/kota master list from api-wilayah-indonesia (v2).

Usage: python3 scripts/fetch_regencies.py [output.json]
Then set REGENCY_MASTER_PATH to the output path in the backend .env.
"""
import json
import sys
import urllib.request

BASE = "https://www.emsifa.com/api-wilayah-indonesia/v2"


def get(path):
    with urllib.request.urlopen(f"{BASE}/{path}", timeout=30) as response:
        return json.load(response)["data"]


def main():
    output = sys.argv[1] if len(sys.argv) > 1 else "data/regencies.json"
    regencies = []
    for province in get("provinces.json"):
        regencies.extend(get(f"regencies/{province['id']}.json"))
    if not regencies:
        sys.exit("no regencies downloaded")
    with open(output, "w", encoding="utf-8") as handle:
        json.dump([{"id": r["id"], "name": r["name"]} for r in regencies], handle, ensure_ascii=False)
    print(f"wrote {len(regencies)} regencies to {output}")


if __name__ == "__main__":
    main()
