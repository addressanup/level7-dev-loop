"""Fail-closed checks for the owner's explicitly waived, unsigned v1 release."""

import copy
import hashlib
import json
import os
from pathlib import Path
import re
import stat
import subprocess
import sys
import tempfile
import unittest
from unittest import mock
import warnings
import zipfile

ROOT = Path(sys.argv[1])
REPOSITORY = "addressanup/level7-dev-loop"
WORKFLOW = ".github/workflows/release.yml"
POLICY = "owner-authorized-unsigned-stable-v1"
ACKNOWLEDGEMENT = "I accept unsigned, unnotarized, unqualified stable v1.0.0"
CHECKS = (
    "Go 1.26.7 (baseline)", "Go 1.27.0 (shadow)",
    "CLI macOS 15 (arm64)", "CLI macOS 15 (amd64)",
    "CLI paired benchmark gate", "evaluate",
)
WAIVERS = {
    "apple_signing_and_notarization": "WAIVED",
    "protected_holdout_evaluation": "WAIVED",
    "separate_operator_and_protected_approval": "WAIVED",
    "exact_asset_provider_trials": "WAIVED",
}
QUALIFICATION = {
    "developer_id_signing": "NOT_RUN",
    "apple_notarization": "NOT_RUN",
    "protected_holdout_evaluation": "NOT_RUN",
    "exact_asset_provider_trials": "NOT_RUN",
    "independent_code_audit": "NOT_RUN",
    "formal_support": "WITHHELD",
}
ARCHIVES = tuple("level7-dev-loop-1.0.0-" + host + ".zip" for host in ("claude", "codex"))
ASSETS = set(ARCHIVES) | {"SHA256SUMS", "RELEASE-MANIFEST.json"}
INPUTS = {arch + "/" + binary for arch in ("darwin-arm64", "darwin-amd64")
          for binary in ("l7", "l7-embed")}


class Blocked(Exception):
    pass


def require(condition, message):
    if not condition:
        raise Blocked(message)


def digest(value, size=40):
    return isinstance(value, str) and re.fullmatch("[0-9a-f]{%d}" % size, value) is not None


def integer(value, minimum=0):
    return type(value) is int and value >= minimum


def unique_object(pairs):
    result = {}
    for key, value in pairs:
        if key in result:
            raise ValueError("duplicate JSON field")
        result[key] = value
    return result


def decode(text):
    return json.loads(text, object_pairs_hook=unique_object)


def authorization(commit, tree):
    return {"schema": 1, "kind": "l7-v1-unsigned-release-authorization",
            "policy": POLICY, "candidate_commit": commit, "candidate_tree": tree,
            "waivers": WAIVERS}


def exact_pull(pulls, commit):
    matches = [pull for pull in pulls
               if pull.get("merged_at") and pull.get("merge_commit_sha") == commit]
    require(len(matches) == 1, "exact merged pull request missing or ambiguous")
    return matches[0]


def passed_checks(checks, head, names):
    for name in names:
        matches = [check for check in checks if check.get("name") == name]
        require(matches, "required check missing: " + name)
        latest = max(matches, key=lambda check: (
            check.get("started_at") or check.get("created_at") or "", check.get("id", 0)))
        require(latest.get("head_sha") == head and latest.get("app", {}).get("slug") == "github-actions"
                and latest.get("status") == "completed" and latest.get("conclusion") == "success",
                "required check is not successful at the exact head: " + name)


def owner_authorization(comments, owner, number, commit, tree):
    expected = authorization(commit, tree)
    matches = []
    for comment in comments:
        if comment.get("user", {}).get("login") != owner or comment.get("author_association") != "OWNER":
            continue
        try:
            record = decode(comment.get("body", ""))
        except (TypeError, ValueError):
            continue
        if record == expected and type(record.get("schema")) is int:
            matches.append(comment)
    require(len(matches) == 1, "one exact-candidate owner authorization of all four waivers is required")
    comment = matches[0]
    require(integer(comment.get("id"), 1) and comment.get("html_url") ==
            "https://github.com/" + REPOSITORY + "/pull/" + str(number) + "#issuecomment-" + str(comment["id"]),
            "owner authorization provenance is invalid")
    return comment["html_url"]


def validate(state, phase, artifact_id=None):
    require(phase in ("eligibility", "prepare", "publish", "finalize"), "invalid release phase")
    env = state["env"]
    commit, tree = env.get("CANDIDATE_COMMIT"), env.get("CANDIDATE_TREE")
    require(digest(commit) and digest(tree), "invalid candidate identity")
    require(env.get("WAIVER_ACKNOWLEDGEMENT") == ACKNOWLEDGEMENT, "explicit unsigned waiver acknowledgement missing")
    repo = state["repository"]
    owner = repo.get("owner", {}).get("login")
    require(repo.get("full_name") == REPOSITORY and repo.get("owner", {}).get("type") == "User"
            and owner == "addressanup" and repo.get("default_branch") == "main",
            "repository owner or default branch drift")
    variables = state["variables"]
    require((variables.get("L7_ACCOUNTABLE_OWNER") or owner) == owner
            and (variables.get("L7_RELEASE_OPERATOR") or owner) == owner
            and (variables.get("L7_ASSURANCE_MODE") or "solo") == "solo",
            "trusted owner-only solo release configuration changed")
    if phase != "prepare":
        require(state["authenticated_login"] == owner, "only the real accountable owner may release")
    else:
        require(env.get("L7_ACCOUNTABLE_OWNER") == owner and env.get("GITHUB_ACTOR") == owner
                and env.get("GITHUB_REPOSITORY") == REPOSITORY and env.get("GITHUB_REF") == "refs/heads/main"
                and env.get("GITHUB_EVENT_NAME") == "workflow_dispatch" and env.get("GITHUB_SHA") == commit
                and env.get("GITHUB_RUN_ATTEMPT") == "1", "dispatch identity drift or rerun")
    require(state["local_head"] == commit and state["local_tree"] == tree and not state["local_dirty"]
            and state["main"].get("object", {}).get("sha") == commit, "local candidate or main drift")
    pull = exact_pull(state["pulls"], commit)
    base, head = pull.get("base", {}).get("sha"), pull.get("head", {}).get("sha")
    require(digest(base) and digest(head) and pull.get("state") == "closed" and not pull.get("draft")
            and pull["base"].get("ref") == "main"
            and pull["base"].get("repo", {}).get("full_name") == REPOSITORY
            and pull["head"].get("repo", {}).get("full_name") == REPOSITORY
            and pull.get("user", {}).get("login") == owner, "untrusted release pull request")
    require([parent.get("sha") for parent in state["commit"].get("parents", [])] == [base]
            and state["commit"].get("tree", {}).get("sha") == tree
            and state["pr_commit"].get("tree", {}).get("sha") == tree,
            "exact squash lineage or tested tree mismatch")
    labels = {label.get("name") for label in pull.get("labels", [])
              if re.fullmatch("l7-risk-tier-[123]", label.get("name", ""))}
    require(labels == {"l7-risk-tier-3"}, "one Tier 3 risk label is required")
    passed_checks(state["pr_checks"], head, CHECKS)
    passed_checks(state["main_checks"], commit, CHECKS[:4])
    if phase != "prepare":
        require(state["immutable"].get("enabled") is True, "immutable releases disabled")
    if phase == "finalize":
        tag, release = state["tag"], state["release"]
        tag_sha = env.get("RELEASE_TAG_OBJECT")
        require(digest(tag_sha) and state["tag_ref"].get("object") == {"type": "tag", "sha": tag_sha,
                "url": "https://api.github.com/repos/" + REPOSITORY + "/git/tags/" + tag_sha}
                and tag.get("sha") == tag_sha and tag.get("tag") == "v1.0.0"
                and tag.get("object", {}).get("type") == "commit" and tag["object"].get("sha") == commit,
                "created annotated tag drift")
        require(str(release.get("id")) == env.get("RELEASE_DRAFT_ID") and release.get("draft") is True
                and release.get("prerelease") is False and release.get("tag_name") == "v1.0.0"
                and release.get("target_commitish") == commit
                and release.get("body") == (ROOT / "docs/releases/v1.0.0.md").read_text(),
                "created draft release drift")
    else:
        require(state["tag_absent"] and state["release_absent"], "tag or release already exists")
    reference = owner_authorization(state["comments"], owner, pull["number"], commit, tree)
    if phase == "eligibility":
        require(not state["runs"] and not state["artifacts"], "candidate already dispatched or prepared; freeze a new candidate")
    else:
        run = state["run"]
        run_id = env.get("GITHUB_RUN_ID") if phase == "prepare" else env.get("RELEASE_RUN_ID")
        require(str(run.get("id")) == run_id and run.get("head_sha") == commit
                and run.get("head_branch") == "main" and run.get("event") == "workflow_dispatch"
                and run.get("run_attempt") == 1 and run.get("actor", {}).get("login") == owner
                and run.get("path", "").split("@")[0] == WORKFLOW, "workflow run identity drift")
        if phase != "prepare":
            require(run.get("status") == "completed" and run.get("conclusion") == "success",
                    "preparation workflow did not complete successfully")
        require(len(state["runs"]) == 1 and state["runs"][0].get("id") == run["id"]
                and state["runs"][0].get("run_attempt") == 1, "duplicate dispatch or rerun; freeze a new candidate")
        if phase == "prepare":
            require(not state["artifacts"], "prepared candidate artifact already exists")
        else:
            artifacts = state["artifacts"]
            require(len(artifacts) == 1 and str(artifacts[0].get("id")) == artifact_id
                    and artifacts[0].get("name") == "v1.0.0-prepared-" + commit
                    and artifacts[0].get("expired") is False
                    and isinstance(artifacts[0].get("digest"), str)
                    and re.fullmatch("sha256:[0-9a-f]{64}", artifacts[0]["digest"]) is not None
                    and artifacts[0].get("workflow_run", {}).get("id") == run["id"]
                    and artifacts[0]["workflow_run"].get("head_sha") == commit, "prepared artifact identity drift")
    return {"base": base, "pull_request": pull["number"], "head": head,
            "owner": owner, "assurance": "solo", "policy": POLICY, "authorization": reference,
            "immutable_checked": phase != "prepare"}


def command(args, absent=False):
    result = subprocess.run(args, capture_output=True, timeout=60)
    require(len(result.stdout) <= 8 * 1024 * 1024, "release control response is oversized")
    if absent:
        require(result.returncode != 0 and b"(HTTP 404)" in result.stderr,
                "tag or release exists or absence is unverifiable")
        return True
    require(result.returncode == 0, "release control read failed: " + args[0])
    return result.stdout.decode("utf-8").strip()


def collect(phase):
    env = dict(os.environ)
    commit = env.get("CANDIDATE_COMMIT", "")
    require(digest(commit) and digest(env.get("CANDIDATE_TREE")), "invalid candidate identity")
    if phase == "prepare":
        require(env.get("GITHUB_REPOSITORY") == REPOSITORY and env.get("GITHUB_RUN_ID", "").isdigit(),
                "invalid workflow context")
    elif phase != "eligibility":
        require(env.get("RELEASE_RUN_ID", "").isdigit(), "invalid preparation run ID")
    prefix = "repos/" + REPOSITORY + "/"

    def api(path, paginate=False):
        args = ["gh", "api"] + (["--paginate", "--slurp"] if paginate else [])
        return decode(command(args + [prefix + path]))

    def items(path, key=None):
        pages = api(path + ("&" if "?" in path else "?") + "per_page=100", paginate=True)
        return [item for page in pages for item in (page[key] if key else page)]

    pulls = items("commits/" + commit + "/pulls")
    pull = exact_pull(pulls, commit)
    head = pull.get("head", {}).get("sha")
    require(digest(head) and integer(pull.get("number"), 1), "invalid pull request identity")
    variables = ([{"name": key, "value": env.get(key, "")}
                  for key in ("L7_ACCOUNTABLE_OWNER", "L7_RELEASE_OPERATOR", "L7_ASSURANCE_MODE")]
                 if phase == "prepare" else items("actions/variables", "variables"))
    artifacts = items("actions/artifacts?name=v1.0.0-prepared-" + commit, "artifacts")
    tag_ref = api("git/ref/tags/v1.0.0") if phase == "finalize" else {}
    tag_sha = tag_ref.get("object", {}).get("sha")
    require(phase != "finalize" or digest(tag_sha), "invalid created tag identity")
    run_id = env.get("GITHUB_RUN_ID") if phase == "prepare" else env.get("RELEASE_RUN_ID")
    return {
        "env": env, "repository": decode(command(["gh", "api", "repos/" + REPOSITORY])),
        "authenticated_login": decode(command(["gh", "api", "user"])).get("login") if phase != "prepare" else "",
        "variables": {variable["name"]: variable["value"] for variable in variables},
        "local_head": command(["git", "-C", str(ROOT), "rev-parse", "HEAD"]),
        "local_tree": command(["git", "-C", str(ROOT), "rev-parse", "HEAD^{tree}"]),
        "local_dirty": command(["git", "-C", str(ROOT), "status", "--porcelain", "--untracked-files=all"]),
        "main": api("git/ref/heads/main"), "commit": api("git/commits/" + commit),
        "pulls": pulls, "pr_commit": api("git/commits/" + head),
        "pr_checks": items("commits/" + head + "/check-runs", "check_runs"),
        "main_checks": items("commits/" + commit + "/check-runs", "check_runs"),
        "comments": items("issues/" + str(pull["number"]) + "/comments"),
        "immutable": api("immutable-releases") if phase != "prepare" else None,
        "tag_absent": command(["gh", "api", prefix + "git/ref/tags/v1.0.0"], absent=True) if phase != "finalize" else False,
        "release_absent": command(["gh", "api", prefix + "releases/tags/v1.0.0"], absent=True) if phase != "finalize" else False,
        "tag_ref": tag_ref, "tag": api("git/tags/" + tag_sha) if phase == "finalize" else {},
        "release": api("releases/tags/v1.0.0") if phase == "finalize" else {},
        "run": api("actions/runs/" + run_id) if phase != "eligibility" else {},
        "runs": items("actions/workflows/release.yml/runs?event=workflow_dispatch&head_sha=" + commit, "workflow_runs"),
        "artifacts": artifacts,
    }


def file_digest(path):
    value = hashlib.sha256()
    with path.open("rb") as stream:
        for chunk in iter(lambda: stream.read(1024 * 1024), b""):
            value.update(chunk)
    return value.hexdigest()


def validate_assets(directory, env):
    require(directory.is_absolute() and directory.resolve() == directory and directory.is_dir()
            and not directory.is_symlink(), "asset directory is unsafe")
    entries = list(directory.iterdir())
    require({entry.name for entry in entries} == ASSETS, "release asset inventory differs")
    for entry in entries:
        info = entry.lstat()
        require(stat.S_ISREG(info.st_mode) and 0 < info.st_size <= 300 * 1024 * 1024,
                "asset is empty, oversized, or unsafe")
    manifest_path = directory / "RELEASE-MANIFEST.json"
    require(manifest_path.stat().st_size <= 1024 * 1024, "release manifest is oversized")
    manifest = decode(manifest_path.read_text())
    commit, tree, base = (env.get(key) for key in ("CANDIDATE_COMMIT", "CANDIDATE_TREE", "L7_RELEASE_BASE"))
    require(all(digest(value) for value in (commit, tree, base)), "invalid expected candidate")
    require(type(manifest.get("schema")) is int and manifest["schema"] == 1
            and manifest.get("change_id") == "unsigned-stable-release"
            and manifest.get("version") == "1.0.0" and manifest.get("tag") == "v1.0.0"
            and manifest.get("release_channel") == "stable"
            and manifest.get("artifact_state") == "unsigned-unnotarized-prepared"
            and manifest.get("policy") == POLICY and manifest.get("waivers") == WAIVERS
            and manifest.get("qualification") == QUALIFICATION
            and manifest.get("publication_blocked") is True
            and manifest.get("gatekeeper") == "MAY_BLOCK; NEVER_DISABLE_GLOBALLY",
            "manifest fails explicit unsigned waiver or withheld-qualification policy")
    candidate = manifest.get("candidate", {})
    require(candidate.get("commit") == commit and candidate.get("tree") == tree and candidate.get("base") == base
            and integer(candidate.get("pull_request"), 1), "manifest candidate drift")
    workflow = manifest.get("workflow", {})
    run_id = env.get("RELEASE_RUN_ID") or env.get("GITHUB_RUN_ID")
    run_attempt = env.get("RELEASE_RUN_ATTEMPT") or env.get("GITHUB_RUN_ATTEMPT")
    require(workflow.get("repository") == REPOSITORY and workflow.get("path") == WORKFLOW
            and workflow.get("run_id") == run_id
            and workflow.get("run_attempt") == "1" == run_attempt,
            "manifest workflow drift")
    require(manifest.get("assurance") == {"mode": "solo", "owner": "addressanup", "operator": "addressanup"}
            and manifest.get("authorization") == env.get("AUTHORIZATION_REFERENCE"),
            "manifest authority drift")
    reference = env.get("AUTHORIZATION_REFERENCE", "")
    require(isinstance(reference, str) and re.fullmatch(
        r"https://github\.com/addressanup/level7-dev-loop/pull/[1-9][0-9]*#issuecomment-[1-9][0-9]*",
        reference) is not None, "manifest owner authorization is missing")
    require(manifest.get("reproducibility") == {"result": "PASS", "toolchain": "go1.26.7"},
            "reproducibility evidence missing")
    for key, expected in (("unsigned_inputs", INPUTS), ("unsigned_packages", set(ARCHIVES))):
        values = manifest.get(key, {})
        require(isinstance(values, dict) and set(values) == expected
                and all(digest(value, 64) for value in values.values()), "unsigned digest inventory differs")
    artifacts = manifest.get("artifacts", [])
    require(len(artifacts) == 2 and {item.get("name") for item in artifacts} == set(ARCHIVES),
            "manifest archive inventory differs")
    for item in artifacts:
        path = directory / item["name"]
        require(type(item.get("size")) is int and item["size"] == path.stat().st_size
                and item.get("sha256") == file_digest(path)
                and manifest["unsigned_packages"][item["name"]] == item["sha256"]
                and item.get("developer_id_signature") == "NOT_RUN"
                and item.get("notarization") == {"status": "NOT_RUN", "id": None}, "archive digest or qualification drift")
    checksums = directory / "SHA256SUMS"
    expected_checksums = "".join(file_digest(directory / name) + "  " + name + "\n" for name in ARCHIVES)
    require(checksums.read_text() == expected_checksums
            and manifest.get("checksums") == {"name": "SHA256SUMS", "sha256": file_digest(checksums),
                                            "size": checksums.stat().st_size}, "checksum evidence differs")
    notes = ROOT / "docs/releases/v1.0.0.md"
    require(manifest.get("release_notes") == {"path": "docs/releases/v1.0.0.md", "sha256": file_digest(notes)},
            "release notes changed")
    text = notes.read_text()
    require(all(word in text for word in ("UNSIGNED", "NOT NOTARIZED", "WITHHELD", "WAIVED", "NOT_RUN")),
            "release notes omit the waived qualification boundary")
    return {"manifest_sha256": file_digest(manifest_path), "candidate_commit": commit,
            "candidate_tree": tree, "policy": POLICY}


def extract_artifact(archive_path, directory):
    require(archive_path.is_absolute() and archive_path.resolve() == archive_path
            and archive_path.is_file() and not archive_path.is_symlink(),
            "artifact archive path is unsafe")
    require(directory.is_absolute() and directory.resolve() == directory and directory.is_dir()
            and not directory.is_symlink() and not list(directory.iterdir()), "artifact extraction root is unsafe")
    with zipfile.ZipFile(archive_path) as archive:
        entries = archive.infolist()
        require(len(entries) == 4 and {entry.filename for entry in entries} == ASSETS,
                "artifact archive inventory differs")
        for entry in entries:
            mode = entry.external_attr >> 16
            require(not entry.is_dir() and not stat.S_ISLNK(mode)
                    and stat.S_IFMT(mode) in (0, stat.S_IFREG)
                    and 0 < entry.file_size <= 300 * 1024 * 1024
                    and not entry.flag_bits & 1, "artifact archive entry is unsafe")
        archive.extractall(directory)


def fixture():
    commit, tree, base, head = (character * 40 for character in "abcd")
    owner = "addressanup"
    env = {"GITHUB_REPOSITORY": REPOSITORY, "GITHUB_REF": "refs/heads/main",
           "GITHUB_EVENT_NAME": "workflow_dispatch", "GITHUB_SHA": commit, "GITHUB_RUN_ATTEMPT": "1",
           "GITHUB_RUN_ID": "123", "GITHUB_ACTOR": owner, "CANDIDATE_COMMIT": commit,
           "CANDIDATE_TREE": tree, "L7_ACCOUNTABLE_OWNER": owner, "WAIVER_ACKNOWLEDGEMENT": ACKNOWLEDGEMENT,
           "L7_RELEASE_OPERATOR": owner, "L7_ASSURANCE_MODE": "solo", "RELEASE_RUN_ID": "123"}
    run = {"id": 123, "head_sha": commit, "head_branch": "main", "event": "workflow_dispatch",
           "run_attempt": 1, "actor": {"login": owner}, "path": WORKFLOW,
           "status": "completed", "conclusion": "success"}
    def checks(sha, names):
        return [{"name": name, "id": index, "head_sha": sha, "app": {"slug": "github-actions"},
                 "status": "completed", "conclusion": "success", "started_at": "2026-10-10T00:00:00Z"}
                for index, name in enumerate(names)]
    return {"env": env, "repository": {"full_name": REPOSITORY, "owner": {"login": owner, "type": "User"},
                                     "default_branch": "main"}, "authenticated_login": owner,
            "variables": {"L7_ACCOUNTABLE_OWNER": owner, "L7_ASSURANCE_MODE": "solo"},
            "local_head": commit, "local_tree": tree, "local_dirty": "", "main": {"object": {"sha": commit}},
            "commit": {"parents": [{"sha": base}], "tree": {"sha": tree}},
            "pulls": [{"number": 27, "merged_at": "2026-10-10T00:00:00Z", "merge_commit_sha": commit,
                       "state": "closed", "draft": False, "user": {"login": owner},
                       "base": {"sha": base, "ref": "main", "repo": {"full_name": REPOSITORY}},
                       "head": {"sha": head, "repo": {"full_name": REPOSITORY}},
                       "labels": [{"name": "l7-risk-tier-3"}]}],
            "pr_commit": {"tree": {"sha": tree}}, "pr_checks": checks(head, CHECKS),
            "main_checks": checks(commit, CHECKS[:4]),
            "comments": [{"id": 456, "user": {"login": owner}, "author_association": "OWNER",
                          "body": json.dumps(authorization(commit, tree)),
                          "html_url": "https://github.com/" + REPOSITORY + "/pull/27#issuecomment-456"}],
            "immutable": {"enabled": True}, "tag_absent": True, "release_absent": True,
            "run": run, "runs": [copy.deepcopy(run)], "artifacts": []}


class PreflightTests(unittest.TestCase):
    def setUp(self):
        self.state = fixture()

    def reject(self, mutate, phase="prepare"):
        state = copy.deepcopy(self.state)
        mutate(state)
        with self.assertRaises(Blocked):
            validate(state, phase, "789")

    def test_owner_solo_exact_squash(self):
        self.assertEqual(validate(self.state, "prepare")["owner"], "addressanup")
        self.assertNotIn("evaluation_report_sha256", validate(self.state, "prepare"))

    def test_dispatch_and_configuration(self):
        for key, value in (("GITHUB_SHA", "0" * 40), ("GITHUB_ACTOR", "intruder"),
                           ("GITHUB_RUN_ATTEMPT", "2"), ("GITHUB_EVENT_NAME", "push"),
                           ("GITHUB_REF", "refs/heads/feature"), ("L7_ACCOUNTABLE_OWNER", "intruder"),
                           ("WAIVER_ACKNOWLEDGEMENT", "yes")):
            self.reject(lambda state: state["env"].update({key: value}))
        self.reject(lambda state: state["variables"].update(L7_ASSURANCE_MODE="team"))
        self.reject(lambda state: state["variables"].update(L7_RELEASE_OPERATOR="other"))
        self.reject(lambda state: state["repository"]["owner"].update(type="Organization"))

    def test_lineage_and_scope(self):
        self.reject(lambda state: state["main"]["object"].update(sha="0" * 40))
        self.reject(lambda state: state.update(local_dirty=" M README.md"))
        self.reject(lambda state: state["commit"]["parents"].append({"sha": "d" * 40}))
        self.reject(lambda state: state["commit"]["parents"][0].update(sha="0" * 40))
        self.reject(lambda state: state["pr_commit"]["tree"].update(sha="0" * 40))
        self.reject(lambda state: state["pulls"][0]["head"]["repo"].update(full_name="other/repo"))
        self.reject(lambda state: state["pulls"][0]["labels"].append({"name": "l7-risk-tier-2"}))

    def test_checks_fail_closed_at_both_boundaries(self):
        for phase in ("prepare", "publish"):
            state = fixture()
            if phase == "publish":
                state["artifacts"] = self.artifact()
            for inventory in ("pr_checks", "main_checks"):
                for key, value in (("conclusion", "failure"), ("status", "in_progress"),
                                   ("head_sha", "0" * 40), ("app", {"slug": "untrusted"})):
                    changed = copy.deepcopy(state)
                    changed[inventory][0][key] = value
                    with self.assertRaises(Blocked):
                        validate(changed, phase, "789")
            newer = copy.deepcopy(state["pr_checks"][0])
            newer.update(id=999, started_at="2026-10-11T00:00:00Z", conclusion="failure")
            state["pr_checks"].append(newer)
            with self.assertRaises(Blocked):
                validate(state, phase, "789")

    def test_owner_waivers_cannot_be_fabricated(self):
        self.reject(lambda state: state.update(comments=[]))
        self.reject(lambda state: state["comments"][0]["user"].update(login="candidate-agent"))
        self.reject(lambda state: state["comments"][0].update(author_association="CONTRIBUTOR"))
        self.reject(lambda state: state["comments"].append(copy.deepcopy(state["comments"][0])))
        for key, value in (("candidate_commit", "0" * 40), ("policy", "signed-qualified"),
                           ("schema", True), ("waivers", {})):
            def mutate(state):
                record = decode(state["comments"][0]["body"])
                record[key] = value
                state["comments"][0]["body"] = json.dumps(record)
            self.reject(mutate)
        self.reject(lambda state: state["comments"][0].update(body='{"schema":1,"schema":1}'))

    def test_one_shot_and_immutability(self):
        self.reject(lambda state: state["immutable"].update(enabled=False), "eligibility")
        self.reject(lambda state: state.update(tag_absent=False))
        self.reject(lambda state: state.update(release_absent=False))
        self.reject(lambda state: state["runs"].append(copy.deepcopy(state["run"])))
        self.reject(lambda state: state["run"].update(path=".github/workflows/other.yml"))
        self.reject(lambda state: state["artifacts"].append({"id": 789}))

    @staticmethod
    def artifact():
        return [{"id": 789, "name": "v1.0.0-prepared-" + "a" * 40, "expired": False,
                 "digest": "sha256:" + "e" * 64, "workflow_run": {"id": 123, "head_sha": "a" * 40}}]

    def test_publish_exact_artifact(self):
        self.state["artifacts"] = self.artifact()
        validate(self.state, "publish", "789")
        for key, value in (("id", 999), ("expired", True), ("name", "other"), ("digest", None)):
            self.reject(lambda state: state["artifacts"][0].update({key: value}), "publish")
        self.reject(lambda state: state["artifacts"][0]["workflow_run"].update(id=999), "publish")
        self.reject(lambda state: state["run"].update(conclusion="failure"), "publish")
        self.reject(lambda state: state.update(authenticated_login="other"), "publish")

    def test_finalize_rechecks_created_tag_and_draft(self):
        self.state["artifacts"] = self.artifact()
        self.state["env"].update(RELEASE_TAG_OBJECT="e" * 40, RELEASE_DRAFT_ID="987")
        self.state.update(
            tag_ref={"object": {"type": "tag", "sha": "e" * 40,
                               "url": "https://api.github.com/repos/" + REPOSITORY + "/git/tags/" + "e" * 40}},
            tag={"sha": "e" * 40, "tag": "v1.0.0", "object": {"type": "commit", "sha": "a" * 40}},
            release={"id": 987, "draft": True, "prerelease": False, "tag_name": "v1.0.0",
                     "target_commitish": "a" * 40, "body": (ROOT / "docs/releases/v1.0.0.md").read_text()})
        validate(self.state, "finalize", "789")
        self.reject(lambda state: state["tag"]["object"].update(sha="0" * 40), "finalize")
        self.reject(lambda state: state["release"].update(draft=False), "finalize")
        self.reject(lambda state: state["release"].update(id=998), "finalize")
        self.reject(lambda state: state["release"].update(body="qualified signed release"), "finalize")
        self.reject(lambda state: state["immutable"].update(enabled=False), "finalize")
        self.reject(lambda state: state.update(comments=[]), "finalize")

    def test_pre_dispatch_eligibility_does_not_consume_run(self):
        self.state["runs"] = []
        validate(self.state, "eligibility")
        self.reject(lambda state: state.update(authenticated_login="other"), "eligibility")
        self.reject(lambda state: state["runs"].append(state["run"]), "eligibility")
        self.reject(lambda state: state["artifacts"].append({"id": 789}), "eligibility")

    def test_preparation_does_not_claim_admin_controls_were_checked(self):
        self.state["immutable"] = None
        self.assertFalse(validate(self.state, "prepare")["immutable_checked"])
        self.assertNotIn("contents: write", (ROOT / WORKFLOW).read_text())

    def test_workflow_retains_boundaries_and_has_no_waived_effects(self):
        document = (ROOT / WORKFLOW).read_text()
        publisher = (ROOT / "scripts/harness/publish-unsigned-stable.sh").read_text()
        self.assertEqual(document.count("check-stable-release-preflight.sh --prepare"), 1)
        self.assertEqual(publisher.count('check-stable-release-preflight.sh" --publish'), 2)
        self.assertIn('check-stable-release-preflight.sh" --finalize', publisher)
        self.assertIn("check-stable-release-preflight.sh --verify-assets", document)
        for forbidden in ("security create-keychain", "notarytool", "environment:", "secrets.",
                          "RELEASE_ADMIN_TOKEN", "release_blocked:false", "apbusinessidentity-tech"):
            self.assertNotIn(forbidden, document)
        self.assertIn(ACKNOWLEDGEMENT, document)
        self.assertIn("--signer-workflow", publisher)
        self.assertIn("--signer-digest", publisher)
        self.assertIn(".immutable", publisher)
        self.assertNotIn("contents: write", document)
        self.assertLess(publisher.index('--verify-assets "$assets"'), publisher.index("gh api --method POST"))
        self.assertLess(publisher.index("--finalize"), publisher.index("gh api --method PATCH"))

    def test_stable_conformance_uses_selected_identity(self):
        document = (ROOT / "Makefile").read_text()
        target = document.split("v1-conformance-check: v1-candidate\n", 1)[1].split("\nv1-candidate-check:", 1)[0]
        self.assertIn('test "$(L7_PACKAGE_CHANNEL)" = stable', target)
        self.assertIn("distribution v1-package-check", target)
        self.assertIn('L7_CLI_VERSION="$(L7_CLI_VERSION)" L7_PACKAGE_CHANNEL="$(L7_PACKAGE_CHANNEL)"', target)

    def test_live_collection_is_read_only_and_paginated(self):
        state = self.state
        prefix = "repos/" + REPOSITORY
        responses = {
            prefix: state["repository"],
            prefix + "/commits/" + "a" * 40 + "/pulls?per_page=100": [state["pulls"]],
            prefix + "/actions/variables?per_page=100": [
                {"variables": [{"name": key, "value": value} for key, value in state["variables"].items()]}],
            prefix + "/actions/artifacts?name=v1.0.0-prepared-" + "a" * 40 + "&per_page=100": [{"artifacts": []}],
            prefix + "/git/ref/heads/main": state["main"],
            prefix + "/git/commits/" + "a" * 40: state["commit"],
            prefix + "/git/commits/" + "d" * 40: state["pr_commit"],
            prefix + "/commits/" + "d" * 40 + "/check-runs?per_page=100": [
                {"check_runs": state["pr_checks"][:3]}, {"check_runs": state["pr_checks"][3:]}],
            prefix + "/commits/" + "a" * 40 + "/check-runs?per_page=100": [{"check_runs": state["main_checks"]}],
            prefix + "/issues/27/comments?per_page=100": [[], state["comments"]],
            prefix + "/immutable-releases": state["immutable"],
            prefix + "/actions/runs/123": state["run"],
            prefix + "/actions/workflows/release.yml/runs?event=workflow_dispatch&head_sha="
            + "a" * 40 + "&per_page=100": [{"workflow_runs": state["runs"]}],
        }
        def read(args, absent=False):
            if args[0] == "git":
                return {"HEAD": state["local_head"], "HEAD^{tree}": state["local_tree"],
                        "--untracked-files=all": state["local_dirty"]}[args[-1]]
            self.assertEqual(args[:2], ["gh", "api"])
            self.assertNotIn("--method", args)
            if absent:
                return True
            return json.dumps(responses[args[-1]])
        with mock.patch.dict(os.environ, state["env"], clear=True), mock.patch(__name__ + ".command", side_effect=read):
            self.assertEqual(validate(collect("prepare"), "prepare"), validate(state, "prepare"))

    def test_api_errors_are_not_absence(self):
        for status, error in ((0, b""), (1, b"gh: Forbidden (HTTP 403)"), (1, b"gh: unavailable (HTTP 500)")):
            with mock.patch("subprocess.run", return_value=subprocess.CompletedProcess([], status, b"", error)):
                with self.assertRaises(Blocked):
                    command(["gh", "api", "unused"], absent=True)
        with mock.patch("subprocess.run", return_value=subprocess.CompletedProcess([], 1, b"", b"gh: Not Found (HTTP 404)")):
            self.assertTrue(command(["gh", "api", "unused"], absent=True))


class AssetTests(unittest.TestCase):
    def setUp(self):
        self.temporary = tempfile.TemporaryDirectory()
        self.addCleanup(self.temporary.cleanup)
        self.directory = Path(self.temporary.name).resolve()
        for name in ARCHIVES:
            (self.directory / name).write_bytes(("fixture " + name).encode())
        sums = "".join(file_digest(self.directory / name) + "  " + name + "\n" for name in ARCHIVES)
        (self.directory / "SHA256SUMS").write_text(sums)
        self.env = {"CANDIDATE_COMMIT": "a" * 40, "CANDIDATE_TREE": "b" * 40,
                    "L7_RELEASE_BASE": "c" * 40, "GITHUB_RUN_ID": "123", "GITHUB_RUN_ATTEMPT": "1",
                    "AUTHORIZATION_REFERENCE": "https://github.com/" + REPOSITORY + "/pull/27#issuecomment-456"}
        self.manifest = {
            "schema": 1, "change_id": "unsigned-stable-release", "version": "1.0.0", "tag": "v1.0.0",
            "release_channel": "stable", "artifact_state": "unsigned-unnotarized-prepared", "policy": POLICY,
            "waivers": WAIVERS, "qualification": QUALIFICATION, "publication_blocked": True,
            "gatekeeper": "MAY_BLOCK; NEVER_DISABLE_GLOBALLY",
            "candidate": {"commit": "a" * 40, "tree": "b" * 40, "base": "c" * 40, "pull_request": 27},
            "workflow": {"repository": REPOSITORY, "path": WORKFLOW, "run_id": "123", "run_attempt": "1"},
            "assurance": {"mode": "solo", "owner": "addressanup", "operator": "addressanup"},
            "authorization": self.env["AUTHORIZATION_REFERENCE"],
            "reproducibility": {"result": "PASS", "toolchain": "go1.26.7"},
            "unsigned_inputs": {name: "e" * 64 for name in INPUTS},
            "unsigned_packages": {name: file_digest(self.directory / name) for name in ARCHIVES},
            "artifacts": [{"name": name, "sha256": file_digest(self.directory / name),
                           "size": (self.directory / name).stat().st_size,
                           "developer_id_signature": "NOT_RUN", "notarization": {"status": "NOT_RUN", "id": None}}
                          for name in ARCHIVES],
            "checksums": {"name": "SHA256SUMS", "sha256": file_digest(self.directory / "SHA256SUMS"),
                          "size": (self.directory / "SHA256SUMS").stat().st_size},
            "release_notes": {"path": "docs/releases/v1.0.0.md", "sha256": file_digest(ROOT / "docs/releases/v1.0.0.md")},
        }
        self.write()

    def write(self):
        (self.directory / "RELEASE-MANIFEST.json").write_text(json.dumps(self.manifest))

    def test_exact_unsigned_asset_set(self):
        validate_assets(self.directory, self.env)

    def test_false_qualification_and_candidate_drift(self):
        for key, value in (("qualification", {**QUALIFICATION, "apple_notarization": "PASS"}),
                           ("waivers", {}), ("artifact_state", "signed-notarized-prepared"),
                           ("publication_blocked", False)):
            original = self.manifest[key]
            self.manifest[key] = value
            self.write()
            with self.assertRaises(Blocked):
                validate_assets(self.directory, self.env)
            self.manifest[key] = original
        self.write()
        with self.assertRaises(Blocked):
            validate_assets(self.directory, {**self.env, "CANDIDATE_COMMIT": "0" * 40})

    def test_changed_extra_symlink_and_empty_assets(self):
        for change in ("bytes", "extra", "symlink", "empty"):
            with self.subTest(change=change):
                target = self.directory / ARCHIVES[0]
                original = target.read_bytes()
                extra = self.directory / "unexpected"
                if change == "bytes":
                    target.write_bytes(original + b"changed")
                elif change == "extra":
                    extra.write_bytes(b"extra")
                elif change == "symlink":
                    target.unlink()
                    target.symlink_to(self.directory / ARCHIVES[1])
                else:
                    target.write_bytes(b"")
                with self.assertRaises(Blocked):
                    validate_assets(self.directory, self.env)
                if target.is_symlink():
                    target.unlink()
                target.write_bytes(original)
                if extra.exists():
                    extra.unlink()

    def test_artifact_extraction_rejects_unsafe_inventory(self):
        archive = self.directory / "artifact.zip"
        target = self.directory / "extract"
        target.mkdir()
        for bad in ("../outside", "extra", "symlink", "duplicate", "directory"):
            with self.subTest(bad=bad):
                with warnings.catch_warnings():
                    warnings.simplefilter("ignore", UserWarning)
                    with zipfile.ZipFile(archive, "w") as output:
                        for name in ASSETS:
                            entry = zipfile.ZipInfo(name)
                            if bad == "symlink" and name == "SHA256SUMS":
                                entry.external_attr = (stat.S_IFLNK | 0o777) << 16
                            output.writestr(entry, b"fixture")
                        if bad != "symlink":
                            output.writestr({"duplicate": "SHA256SUMS", "directory": "nested/"}.get(bad, bad), b"fixture")
                with self.assertRaises(Blocked):
                    extract_artifact(archive, target)
                self.assertFalse(list(target.iterdir()))
        with zipfile.ZipFile(archive, "w") as output:
            for name in ASSETS:
                output.writestr(name, b"fixture")
        extract_artifact(archive, target)
        self.assertEqual({entry.name for entry in target.iterdir()}, ASSETS)

    def test_publisher_end_to_end_with_fake_forge(self):
        # Exercise the consequential shell path against an isolated fake forge.
        # No credential, remote request, or real release is involved.
        fake_bin = self.directory / "bin"
        fake_bin.mkdir()
        bundle = self.directory / "prepared.zip"
        with zipfile.ZipFile(bundle, "w") as output:
            for name in ASSETS:
                output.write(self.directory / name, name)
        state = fixture()
        state["artifacts"] = PreflightTests.artifact()
        state["artifacts"][0]["digest"] = "sha256:" + file_digest(bundle)
        state["bundle"] = str(bundle)
        state["source_assets"] = str(self.directory)
        state["notes"] = (ROOT / "docs/releases/v1.0.0.md").read_text()
        state["calls"] = []
        state["fault"] = ""
        state_path = self.directory / "forge.json"
        git = fake_bin / "git"
        git.write_text("#!/bin/sh\ncase \"$*\" in *HEAD\\^\\{tree\\}) printf '%s\\n' '" + "b" * 40 +
                       "' ;; *HEAD) printf '%s\\n' '" + "a" * 40 + "' ;; *) exit 0 ;; esac\n")
        git.chmod(0o700)
        gh = fake_bin / "gh"
        gh.write_text("""#!/usr/bin/python3
import json, os, pathlib, shutil, sys
p = pathlib.Path(os.environ["FAKE_FORGE"])
s = json.loads(p.read_text())
a = sys.argv[1:]
s["calls"].append(a)
result = {}
binary = None
if a[:2] == ["attestation", "verify"]:
    if s["fault"] == "attestation":
        p.write_text(json.dumps(s)); sys.exit(1)
elif a[:2] == ["release", "upload"]:
    s["release"]["assets"] = [{"name": name, "state": "uploaded", "size": 1} for name in
        ("level7-dev-loop-1.0.0-claude.zip", "level7-dev-loop-1.0.0-codex.zip", "SHA256SUMS", "RELEASE-MANIFEST.json")]
    if s["fault"] == "late-authority":
        s["comments"] = []
elif a[:2] == ["release", "download"]:
    target = pathlib.Path(a[a.index("--dir") + 1])
    for name in ("level7-dev-loop-1.0.0-claude.zip", "level7-dev-loop-1.0.0-codex.zip", "SHA256SUMS", "RELEASE-MANIFEST.json"):
        shutil.copyfile(pathlib.Path(s["source_assets"]) / name, target / name)
elif a[0] == "api":
    method = a[a.index("--method") + 1] if "--method" in a else "GET"
    path = next(arg for arg in a[1:] if arg.startswith("repos/") or arg == "user")
    prefix = "repos/addressanup/level7-dev-loop"
    tail = path[len(prefix)+1:] if path.startswith(prefix + "/") else path
    if method == "POST":
        if tail == "git/tags":
            s["tag"] = {"sha": "e"*40, "tag": "v1.0.0", "object": {"type": "commit", "sha": "a"*40}}
            result = s["tag"]
        elif tail == "git/refs":
            s["tag_ref"] = {"object": {"type":"tag", "sha":"e"*40, "url":"https://api.github.com/" + prefix + "/git/tags/" + "e"*40}}
        elif tail == "releases":
            s["release"] = {"id":987, "draft":True, "prerelease":False, "tag_name":"v1.0.0",
                            "target_commitish":"a"*40, "body":s["notes"], "assets":[]}
            result = s["release"]
        else:
            raise RuntimeError("unexpected fake write")
    elif method == "PATCH":
        if s["fault"] == "late-authority":
            raise RuntimeError("publisher crossed revoked authority")
        s["release"].update(draft=False, immutable=True)
        result = s["release"]
    elif tail == prefix:
        result = s["repository"]
    elif tail == "user":
        result = {"login":s["authenticated_login"]}
    elif tail.startswith("actions/variables"):
        result = [{"variables":[{"name":k,"value":v} for k,v in s["variables"].items()]}]
    elif tail.startswith("actions/artifacts?"):
        result = [{"artifacts":s["artifacts"]}]
    elif tail == "actions/artifacts/789":
        result = s["artifacts"][0]
    elif tail == "actions/artifacts/789/zip":
        binary = pathlib.Path(s["bundle"]).read_bytes()
    elif tail.startswith("actions/workflows"):
        result = [{"workflow_runs":s["runs"]}]
    elif tail.startswith("actions/runs"):
        result = s["run"]
    elif tail == "immutable-releases":
        result = s["immutable"]
    elif tail == "git/ref/heads/main":
        result = s["main"]
    elif tail.startswith("git/commits/"):
        result = s["pr_commit"] if tail.endswith("d"*40) else s["commit"]
    elif "/pulls?" in tail:
        result = [s["pulls"]]
    elif "/check-runs?" in tail:
        result = [{"check_runs":s["pr_checks"] if "/"+ "d"*40 + "/" in tail else s["main_checks"]}]
    elif tail.startswith("issues/"):
        result = [s["comments"]]
    elif tail.startswith("git/ref/tags"):
        result = s.get("tag_ref")
    elif tail.startswith("git/tags/"):
        result = s.get("tag")
    elif tail.startswith("releases/"):
        result = s.get("release")
    else:
        raise RuntimeError("unexpected fake read " + tail)
    if result is None:
        p.write_text(json.dumps(s))
        print("gh: Not Found (HTTP 404)", file=sys.stderr)
        sys.exit(1)
else:
    raise RuntimeError("unexpected fake command")
p.write_text(json.dumps(s))
if binary is not None:
    sys.stdout.buffer.write(binary)
else:
    print(json.dumps(result))
""")
        gh.chmod(0o700)
        environment = {"PATH": str(fake_bin) + ":/usr/bin:/bin", "FAKE_FORGE": str(state_path),
                       "HOME": str(self.directory)}
        for fault in ("", "attestation", "late-authority"):
            with self.subTest(fault=fault):
                state["fault"] = fault
                state_path.write_text(json.dumps(state))
                result = subprocess.run([
                    "/bin/sh", str(ROOT / "scripts/harness/publish-unsigned-stable.sh"),
                    "a" * 40, "b" * 40, "123", "789", str(self.directory / ("publish-" + (fault or "success"))),
                    ACKNOWLEDGEMENT], env=environment, capture_output=True, timeout=60)
                final = decode(state_path.read_text())
                if not fault:
                    self.assertEqual(result.returncode, 0, result.stderr.decode())
                    self.assertTrue(final["release"]["immutable"])
                    self.assertFalse(final["release"]["draft"])
                else:
                    self.assertNotEqual(result.returncode, 0)
                    self.assertFalse(any("--method" in call and "PATCH" in call for call in final["calls"]))
                    if fault == "attestation":
                        self.assertFalse(any("--method" in call for call in final["calls"]))


def main(arguments):
    if not arguments:
        suite = unittest.TestSuite(unittest.defaultTestLoader.loadTestsFromTestCase(case)
                                   for case in (PreflightTests, AssetTests))
        require(unittest.TextTestRunner(verbosity=1).run(suite).wasSuccessful(), "offline regression tests failed")
        print("stable-release-preflight: PASS (offline regressions; no forge or release effects)")
    elif arguments[0] == "--verify-assets" and len(arguments) == 2:
        print(json.dumps(validate_assets(Path(arguments[1]), dict(os.environ)), sort_keys=True))
    elif arguments[0] == "--extract-artifact" and len(arguments) == 3:
        extract_artifact(Path(arguments[1]), Path(arguments[2]))
    else:
        require(arguments in (["--eligibility"], ["--prepare"])
                or (len(arguments) == 2 and arguments[0] in ("--publish", "--finalize") and arguments[1].isdigit()),
                "usage: check-stable-release-preflight.sh [--eligibility | --prepare | --publish artifact-id | --finalize artifact-id | --verify-assets absolute-directory]")
        phase = arguments[0][2:]
        print(json.dumps(validate(collect(phase), phase, arguments[1] if len(arguments) == 2 else None), sort_keys=True))


if __name__ == "__main__":
    try:
        main(sys.argv[2:])
    except (Blocked, KeyError, TypeError, ValueError, OSError, zipfile.BadZipFile, subprocess.TimeoutExpired) as error:
        print("stable-release-preflight: BLOCKED: " +
              (str(error) if isinstance(error, Blocked) else "invalid or unavailable release evidence"), file=sys.stderr)
        sys.exit(1)
