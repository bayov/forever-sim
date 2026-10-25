#!/usr/bin/env python3
"""Deploy dist/dps_report.html to Cloudflare Pages (https://forever-dps.pages.dev).

    python3 tools/dps_report/deploy_pages.py [file] [--branch main] [--project forever-dps]

Needs CLOUDFLARE_ACCOUNT_ID, and a token with Account > Cloudflare Pages > Edit: either
CLOUDFLARE_API_TOKEN in the environment, or the cloud environment's network secret for
api.cloudflare.com (path prefix /client/v4/accounts/), which the session proxy adds itself.

Why not `wrangler pages deploy`: the asset upload calls (/client/v4/pages/assets/*) take a
short-lived JWT, and the session proxy breaks them, so they go direct (--noproxy). The
account calls go through the proxy as usual.

The deploy serves the file as /index.html. After a production deploy the script fetches the
live site and compares it byte for byte; if it doesn't match, it rolls production back to
the deployment that was live before.
"""

import argparse
import base64
import hashlib
import json
import os
import subprocess
import sys
import tempfile
import time

API = "https://api.cloudflare.com/client/v4"


def curl(args, direct=False, auth=None):
    cmd = ["curl", "-sS", "--fail-with-body", "--retry", "3"]
    if direct:
        cmd += ["--noproxy", "*"]
    if auth:
        cmd += ["-H", f"Authorization: Bearer {auth}"]
    out = subprocess.run(cmd + args, capture_output=True, check=False)
    try:
        body = json.loads(out.stdout)
    except ValueError:
        sys.exit(f"curl {args[-1]}: {out.stderr.decode().strip()} {out.stdout[:300]!r}")
    if not body.get("success"):
        sys.exit(f"{args[-1]}: {body.get('errors')}")
    return body.get("result")


def assets(path, payload, jwt):
    # Direct first; fall back to the proxy for networks that don't allow a direct connection.
    data = ["-H", "Content-Type: application/json", "--data-binary", payload]
    url = f"{API}/pages/assets/{path}"
    try:
        return curl(["-X", "POST", *data, url], direct=True, auth=jwt)
    except SystemExit as e:
        print(f"direct {path} failed ({e}); retrying through the proxy", file=sys.stderr)
        return curl(["-X", "POST", *data, url], auth=jwt)


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("file", nargs="?", default="dist/dps_report.html")
    ap.add_argument("--project", default="forever-dps")
    ap.add_argument("--branch", default="main", help="main is production; anything else is a preview")
    a = ap.parse_args()

    account = os.environ.get("CLOUDFLARE_ACCOUNT_ID")
    if not account:
        sys.exit("CLOUDFLARE_ACCOUNT_ID is not set")
    token = os.environ.get("CLOUDFLARE_API_TOKEN")
    proj = f"{API}/accounts/{account}/pages/projects/{a.project}"

    content = open(a.file, "rb").read()
    b64 = base64.b64encode(content).decode()
    # Any 32-hex key works; wrangler's is blake3, this is the same length from sha256.
    key = hashlib.sha256(b64.encode() + b"html").hexdigest()[:32]

    project = curl([proj], auth=token)
    prod = (project.get("canonical_deployment") or {}).get("id")
    jwt = curl([f"{proj}/upload-token"], auth=token)["jwt"]

    with tempfile.TemporaryDirectory() as tmp:
        hashes = json.dumps({"hashes": [key]})
        if assets("check-missing", hashes, jwt):
            up = os.path.join(tmp, "upload.json")
            with open(up, "w") as f:
                json.dump([{"key": key, "value": b64, "metadata": {"contentType": "text/html"}, "base64": True}], f)
            assets("upload", "@" + up, jwt)
        assets("upsert-hashes", hashes, jwt)
        if assets("check-missing", hashes, jwt):
            sys.exit("the upload didn't stick; not deploying")

    dep = curl(["-X", "POST", "-F", "manifest=" + json.dumps({"/index.html": key}), "-F", f"branch={a.branch}",
                f"{proj}/deployments"], auth=token)
    for _ in range(60):
        stage = curl([f"{proj}/deployments/{dep['id']}"], auth=token)["latest_stage"]
        if stage["status"] in ("success", "failure"):
            break
        time.sleep(5)
    if stage["status"] != "success":
        sys.exit(f"deployment {dep['id']} ended at {stage['name']} {stage['status']}")

    live = f"https://{a.project}.pages.dev/" if dep["environment"] == "production" else dep["url"] + "/"
    for _ in range(12):
        got = subprocess.run(["curl", "-sS", live], capture_output=True).stdout
        if got == content:
            print(f"deployed {len(content)} bytes, {dep['environment']}: {live} ({dep['url']})")
            return
        time.sleep(5)
    if dep["environment"] == "production" and prod:
        curl(["-X", "POST", f"{proj}/deployments/{prod}/rollback"], auth=token)
        sys.exit(f"{live} didn't serve the new file; rolled production back to {prod}")
    sys.exit(f"{live} didn't serve the new file")


if __name__ == "__main__":
    main()
