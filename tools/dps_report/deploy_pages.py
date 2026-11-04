#!/usr/bin/env python3
"""Deploy the DPS Report site to Cloudflare Pages (https://forever-dps.pages.dev).

    python3 tools/dps_report/deploy_pages.py [dir or file] [--branch main] [--project forever-dps]

The default is dist/dps_report_site (build_artifact.mjs): the page, index.html, and its data files, one per
report mode (data/<mode>.<hash>.json) and one per preset's sample fights, which the page fetches when it needs
them. Every file is uploaded, and after the deploy each one is fetched from the live site and compared. A single
file (dist/dps_report.html, the self-contained page) still deploys as before, as /index.html.

Needs CLOUDFLARE_ACCOUNT_ID, and a token with Account > Cloudflare Pages > Edit: either
CLOUDFLARE_API_TOKEN in the environment, or the cloud environment's network secret for
api.cloudflare.com (path prefix /client/v4/accounts/), which the session proxy adds itself.

Why not `wrangler pages deploy`: the asset upload calls (/client/v4/pages/assets/*) take a
short-lived JWT, and the session proxy breaks them, so they go direct (--noproxy). The
account calls go through the proxy as usual.

A production deploy (--branch main) must come from what master serves: the report's sources
(ui/dps_report, ui/scss/dps_report, tools/dps_report) clean, and HEAD pushed to origin/master.
--unmerged deploys a commit that isn't on master yet, when the user asked for that. The
deployment records the commit it was built from.

A single file is served as /index.html, inside the doctype, <head> and <body> the artifact host
used to add (the site's index.html has them already). After a deploy the script fetches every file
from the live site and compares it byte for byte, apart from the Cloudflare Web Analytics beacon that
Pages injects into every HTML page (enabled 2026-10-28); if one doesn't match, it rolls production
back to the deployment that was live before.
"""

import argparse
import re
import base64
import hashlib
import json
import os
import subprocess
import sys
import tempfile
import time

API = "https://api.cloudflare.com/client/v4"



# Cloudflare Pages injects its Web Analytics beacon into every HTML page it serves: a script from
# static.cloudflareinsights.com, between two "Cloudflare Pages Analytics" comments.
BEACON = re.compile(rb"(<!-- Cloudflare Pages Analytics -->)?\s*<script[^>]*cloudflareinsights[^>]*>\s*</script>\s*(<!-- Cloudflare Pages Analytics -->)?")


def strip_beacon(html):
    return BEACON.sub(b"", html)


# The built file holds only the page's own tags (build_artifact.mjs): the Claude artifact host it was made
# for added the doctype, <head> with the charset and viewport, and <body>. Pages serves the file as it is,
# so we add the same wrapper. It keeps phones at device width and the page out of quirks mode, and it
# gives Pages the </body> it puts the Web Analytics beacon in front of. The tab icon is crossed swords
# (ui/dps_report/favicon.svg, with a 64 px PNG of it for browsers without SVG icons), inline, as
# build_artifact.mjs puts them in the site's index.html.
_ui = os.path.join(os.path.dirname(os.path.abspath(__file__)), "../../ui/dps_report/")
FAVICON = base64.b64encode(open(_ui + "favicon.png", "rb").read())
FAVICON_SVG = base64.b64encode(open(_ui + "favicon.svg", "rb").read())
HEAD = (b'<!doctype html><html lang="en"><head><meta charset="utf-8">'
        b'<meta name="viewport" content="width=device-width, initial-scale=1">'
        b'<link rel="icon" type="image/png" sizes="64x64" href="data:image/png;base64,' + FAVICON + b'">'
        b'<link rel="icon" type="image/svg+xml" href="data:image/svg+xml;base64,' + FAVICON_SVG + b'"></head><body>')


def document(html):
    if html.lstrip()[:9].lower() == b"<!doctype":
        return html
    return HEAD + html + b"</body></html>"

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


# The content type each kind of site file is served with. The sample fights stay gzipped: the page unzips them
# itself (events.ts), so they go out as plain bytes, not with a Content-Encoding.
TYPES = {".html": "text/html", ".json": "application/json", ".gz": "application/gzip", ".js": "application/javascript",
         ".css": "text/css", ".png": "image/png", ".svg": "image/svg+xml", ".jpg": "image/jpeg"}
# One upload call carries at most about this much base64; Pages takes 50 MB a request.
BATCH = 40 * 1024 * 1024


def site_files(path):
    """{ "/index.html": bytes, "/data/...": bytes } for a site directory, or one file served as /index.html."""
    if os.path.isfile(path):
        return {"/index.html": document(open(path, "rb").read())}
    files = {}
    for root, _, names in os.walk(path):
        for n in sorted(names):
            full = os.path.join(root, n)
            rel = "/" + os.path.relpath(full, path).replace(os.sep, "/")
            data = open(full, "rb").read()
            files[rel] = document(data) if rel == "/index.html" else data
    if "/index.html" not in files:
        sys.exit(f"{path} has no index.html")
    return files


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("path", nargs="?", default="dist/dps_report_site", help="the site directory, or one HTML file")
    ap.add_argument("--project", default="forever-dps")
    ap.add_argument("--branch", default="main", help="main is production; anything else is a preview")
    ap.add_argument("--unmerged", action="store_true", help="allow a production deploy of a commit not on origin/master")
    a = ap.parse_args()

    git = lambda *args: subprocess.run(["git", *args], capture_output=True, text=True).stdout.strip()
    head = git("rev-parse", "HEAD")
    dirty = git("status", "--porcelain", "--", "ui/dps_report", "ui/scss/dps_report", "tools/dps_report")
    if a.branch == "main":
        if dirty:
            sys.exit(f"the report's sources have uncommitted changes; commit and push first:\n{dirty}")
        subprocess.run(["git", "fetch", "-q", "origin", "master"], check=False)
        on_master = subprocess.run(["git", "merge-base", "--is-ancestor", head, "origin/master"]).returncode == 0
        if not on_master and not a.unmerged:
            sys.exit(f"HEAD {head[:10]} is not on origin/master: push it first (or pass --unmerged)")

    account = os.environ.get("CLOUDFLARE_ACCOUNT_ID")
    if not account:
        sys.exit("CLOUDFLARE_ACCOUNT_ID is not set")
    token = os.environ.get("CLOUDFLARE_API_TOKEN")
    proj = f"{API}/accounts/{account}/pages/projects/{a.project}"

    files = site_files(a.path)
    entries = {}
    for rel, data in files.items():
        ext = os.path.splitext(rel)[1]
        b64 = base64.b64encode(data).decode()
        # Any 32-hex key works; wrangler's is blake3, this is the same length from sha256.
        key = hashlib.sha256(b64.encode() + ext.lstrip(".").encode()).hexdigest()[:32]
        entries[rel] = {"key": key, "b64": b64, "type": TYPES.get(ext, "application/octet-stream")}
    keys = sorted({e["key"] for e in entries.values()})

    project = curl([proj], auth=token)
    prod = (project.get("canonical_deployment") or {}).get("id")
    jwt = curl([f"{proj}/upload-token"], auth=token)["jwt"]

    with tempfile.TemporaryDirectory() as tmp:
        hashes = json.dumps({"hashes": keys})
        missing = set(assets("check-missing", hashes, jwt) or [])
        todo = [e for e in {e["key"]: e for e in entries.values()}.values() if e["key"] in missing]
        # in batches, each under the request limit
        batches, cur, size = [], [], 0
        for e in todo:
            if cur and size + len(e["b64"]) > BATCH:
                batches.append(cur)
                cur, size = [], 0
            cur.append(e)
            size += len(e["b64"])
        if cur:
            batches.append(cur)
        for i, batch in enumerate(batches):
            up = os.path.join(tmp, f"upload{i}.json")
            with open(up, "w") as f:
                json.dump([{"key": e["key"], "value": e["b64"], "metadata": {"contentType": e["type"]}, "base64": True} for e in batch], f)
            assets("upload", "@" + up, jwt)
        print(f"{len(files)} files, {len(todo)} uploaded in {len(batches)} call(s), {len(keys) - len(todo)} already there")
        assets("upsert-hashes", hashes, jwt)
        if assets("check-missing", hashes, jwt):
            sys.exit("the upload didn't stick; not deploying")

    manifest = {rel: e["key"] for rel, e in entries.items()}
    dep = curl(["-X", "POST", "-F", "manifest=" + json.dumps(manifest), "-F", f"branch={a.branch}",
                "-F", f"commit_hash={head}", "-F", f"commit_dirty={'true' if dirty else 'false'}",
                f"{proj}/deployments"], auth=token)
    for _ in range(60):
        stage = curl([f"{proj}/deployments/{dep['id']}"], auth=token)["latest_stage"]
        if stage["status"] in ("success", "failure"):
            break
        time.sleep(5)
    if stage["status"] != "success":
        sys.exit(f"deployment {dep['id']} ended at {stage['name']} {stage['status']}")

    live = f"https://{a.project}.pages.dev" if dep["environment"] == "production" else dep["url"]
    # Every file, as the live site serves it. The page is checked last: the data files it names are there by then.
    bad = []
    for rel in sorted(files, key=lambda r: r == "/index.html"):
        want = strip_beacon(files[rel]) if rel.endswith(".html") else files[rel]
        url = live + ("/" if rel == "/index.html" else rel)
        for _ in range(12):
            # A deployment's own preview URL can stall partway through a file, so each fetch has a time limit.
            got = subprocess.run(["curl", "-sS", "--max-time", "120", url], capture_output=True).stdout
            if (strip_beacon(got) if rel.endswith(".html") else got) == want:
                break
            time.sleep(5)
        else:
            bad.append(rel)
    if not bad:
        size = sum(len(d) for d in files.values())
        print(f"deployed {len(files)} files, {size} bytes, each checked live; {dep['environment']}: {live}/ ({dep['url']})")
        return
    if dep["environment"] == "production" and prod:
        curl(["-X", "POST", f"{proj}/deployments/{prod}/rollback"], auth=token)
        sys.exit(f"{live} didn't serve {', '.join(bad[:5])}; rolled production back to {prod}")
    sys.exit(f"{live} didn't serve {', '.join(bad[:5])}")


if __name__ == "__main__":
    main()
