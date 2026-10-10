#!/bin/sh

set -eu
script_dir=$(CDPATH='' cd "$(dirname "$0")" && pwd -P)
project_root=$(CDPATH='' cd "$script_dir/../.." && pwd -P)

# No arguments runs offline regression tests. Release jobs explicitly select
# --prepare or --publish; that path reads forge state and never mutates it.
exec python3 - "$project_root" "$@" <<'PY'
import copy
import json
import os
from pathlib import Path
import re
import subprocess
import sys
import unittest
from unittest import mock

ROOT = Path(sys.argv[1])
REPOSITORY = "addressanup/level7-dev-loop"
WORKFLOW = ".github/workflows/release.yml"
CHECKS = (
    "Go 1.26.7 (baseline)", "Go 1.27.0 (shadow)",
    "CLI macOS 15 (arm64)", "CLI macOS 15 (amd64)",
    "CLI paired benchmark gate", "evaluate",
)


class Blocked(Exception):
    pass


def require(condition, message):
    if not condition:
        raise Blocked(message)


def digest(value, size=40):
    return isinstance(value, str) and re.fullmatch("[0-9a-f]{%d}" % size, value) is not None


def integer(value, minimum=0):
    return type(value) is int and value >= minimum


def observed(value):
    return (isinstance(value, str) and 0 < len(value) <= 128
            and not value.startswith("REPLACE") and value not in ("TODO", "UNKNOWN", "self-review"))


def exact_pull(pulls, commit):
    matches = [p for p in pulls if p.get("merged_at") and p.get("merge_commit_sha") == commit]
    require(len(matches) == 1, "exact merged pull request missing or ambiguous")
    return matches[0]


def passed_checks(checks, head, names):
    for name in names:
        matches = [c for c in checks if c.get("name") == name]
        require(bool(matches), "required check missing: " + name)
        latest = max(matches, key=lambda c: (c.get("started_at") or c.get("created_at") or "", c.get("id", 0)))
        require(latest.get("head_sha") == head and latest.get("app", {}).get("slug") == "github-actions"
                and latest.get("status") == "completed" and latest.get("conclusion") == "success",
                "required check is not successful at the exact head: " + name)


def evaluation(comments, owner, commit, tree):
    matches = []
    for comment in comments:
        if comment.get("user", {}).get("login") != owner or comment.get("author_association") != "OWNER":
            continue
        try:
            record = json.loads(comment.get("body", ""))
        except (ValueError, TypeError):
            continue
        if not isinstance(record, dict) or record.get("kind") != "l7-v1-release-evaluation":
            continue
        if record.get("candidate_commit") == commit and record.get("candidate_tree") == tree:
            matches.append(record)
    require(len(matches) == 1, "one exact-candidate owner-bound protected evaluation is required")
    record = matches[0]
    require(type(record.get("schema")) is int and record["schema"] == 1
            and record.get("result") == "PASS", "release evaluation has not passed")
    require(digest(record.get("protocol_sha256"), 64) and digest(record.get("report_sha256"), 64)
            and observed(record.get("evaluator")), "evaluation protocol, report, or isolated evaluator is missing")
    total, holdout = record.get("corpus_items"), record.get("holdout_items")
    require(integer(total, 1) and integer(holdout, 1) and holdout <= total and holdout * 5 >= total
            and record.get("holdout_isolated") is True, "L7-EVAL-007 protected holdout is unmet")
    planned, valid = record.get("planned_trials"), record.get("valid_trials")
    require(integer(planned, 1) and integer(valid) and valid <= planned
            and integer(record.get("recorded_trials")) and record["recorded_trials"] == planned
            and valid * 10 >= planned * 9,
            "release evaluation is incomplete or has insufficient valid trials")
    for key in ("contaminated_trials", "crew_false_success", "crew_safety_violations"):
        require(type(record.get(key)) is int and record[key] == 0, "release evaluation fails " + key)
    return {"report": record["report_sha256"], "protocol": record["protocol_sha256"]}


def validate(state, phase, artifact_id=None):
    env = state["env"]
    commit, tree = env.get("CANDIDATE_COMMIT"), env.get("CANDIDATE_TREE")
    require(digest(commit) and digest(tree), "invalid candidate identity")
    require(env.get("GITHUB_REPOSITORY") == REPOSITORY and env.get("GITHUB_REF") == "refs/heads/main"
            and env.get("GITHUB_EVENT_NAME") == "workflow_dispatch" and env.get("GITHUB_SHA") == commit
            and env.get("GITHUB_RUN_ATTEMPT") == "1", "dispatch identity drift or rerun")
    require(phase in ("prepare", "publish"), "invalid release phase")
    owner = env.get("L7_ACCOUNTABLE_OWNER")
    operator = env.get("L7_RELEASE_OPERATOR")
    mode = env.get("L7_ASSURANCE_MODE")
    for login in (owner, operator):
        require(isinstance(login, str) and re.fullmatch("[A-Za-z0-9][A-Za-z0-9-]{0,38}", login),
                "real owner and operator identities are required")
    require(mode in ("solo", "team") and env.get("GITHUB_ACTOR") == operator,
            "untrusted operator or invalid assurance mode")
    repo = state["repository"]
    require(repo.get("owner", {}).get("type") == "User" and repo["owner"].get("login") == owner
            and repo.get("default_branch") == "main", "repository owner or default branch drift")
    variables = state["variables"]
    require((variables.get("L7_ACCOUNTABLE_OWNER") or owner) == owner
            and (variables.get("L7_RELEASE_OPERATOR") or owner) == operator
            and (variables.get("L7_ASSURANCE_MODE") or "solo") == mode
            and (variables.get("L7_RELEASE_REVIEWER") or "") == env.get("L7_RELEASE_REVIEWER", ""),
            "trusted release configuration changed")
    require(owner != operator, "protected publication requires an operator distinct from its owner approver")
    require(state["local_head"] == commit and state["local_tree"] == tree and not state["local_dirty"]
            and state["main"].get("object", {}).get("sha") == commit, "local candidate or main drift")
    pull = exact_pull(state["pulls"], commit)
    base, head = pull.get("base", {}).get("sha"), pull.get("head", {}).get("sha")
    require(digest(base) and digest(head) and pull.get("state") == "closed" and not pull.get("draft")
            and pull["base"].get("ref") == "main"
            and pull["base"].get("repo", {}).get("full_name") == REPOSITORY
            and pull["head"].get("repo", {}).get("full_name") == REPOSITORY
            and pull.get("user", {}).get("login") in (owner, operator), "untrusted release pull request")
    parents = state["commit"].get("parents", [])
    require([p.get("sha") for p in parents] == [base]
            and state["commit"].get("tree", {}).get("sha") == tree
            and state["pr_commit"].get("tree", {}).get("sha") == tree, "exact squash lineage or tested tree mismatch")
    labels = {label.get("name") for label in pull.get("labels", [])
              if re.fullmatch("l7-risk-tier-[123]", label.get("name", ""))}
    require(labels == {"l7-risk-tier-3"}, "one Tier 3 risk label is required")
    passed_checks(state["pr_checks"], head, CHECKS)
    passed_checks(state["main_checks"], commit, CHECKS[:4])
    if mode == "team":
        reviewer = env.get("L7_RELEASE_REVIEWER")
        require(observed(reviewer) and reviewer not in (owner, operator, pull["user"]["login"]),
                "team mode requires a real distinct reviewer")
        for login in (owner, reviewer):
            reviews = [r for r in state["reviews"] if r.get("user", {}).get("login") == login]
            require(bool(reviews), "team approval missing")
            review = max(reviews, key=lambda r: (r.get("submitted_at") or "", r.get("id", 0)))
            require(review.get("commit_id") == head and review.get("state") == "APPROVED", "team approval drift")
    for key in ("signing", "production"):
        config = state[key]
        policy = config.get("deployment_branch_policy", {})
        require(policy.get("protected_branches") is True and policy.get("custom_branch_policies") is False,
                key + " protected branch policy drift")
    config = state["production"]
    rules = [r for r in config.get("protection_rules", []) if r.get("type") == "required_reviewers"]
    require(len(rules) == 1 and rules[0].get("prevent_self_review") is True
            and rules[0].get("reviewers") == [{"type": "User", "reviewer": state["owner_user"]}]
            and state["owner_user"].get("login") == owner and config.get("can_admins_bypass") is False,
            "protected owner approval is missing or bypassable")
    require(state["immutable"].get("enabled") is True and state["tag_absent"] and state["release_absent"],
            "immutable release disabled or tag/release already exists")
    run = state["run"]
    require(str(run.get("id")) == env.get("GITHUB_RUN_ID") and run.get("head_sha") == commit
            and run.get("head_branch") == "main" and run.get("event") == "workflow_dispatch"
            and run.get("run_attempt") == 1 and run.get("actor", {}).get("login") == operator
            and run.get("path", "").split("@")[0] == WORKFLOW, "workflow run identity drift")
    runs = state["runs"]
    require(len(runs) == 1 and runs[0].get("id") == run["id"] and runs[0].get("run_attempt") == 1,
            "duplicate dispatch or rerun; freeze a new candidate")
    artifacts = state["artifacts"]
    if phase == "prepare":
        require(not artifacts, "prepared candidate artifact already exists")
    else:
        require(len(artifacts) == 1 and str(artifacts[0].get("id")) == artifact_id
                and artifacts[0].get("name") == "v1.0.0-prepared-" + commit
                and artifacts[0].get("expired") is False
                and artifacts[0].get("workflow_run", {}).get("id") == run["id"]
                and artifacts[0]["workflow_run"].get("head_sha") == commit, "prepared artifact identity drift")
    evidence = evaluation(state["comments"], owner, commit, tree)
    return {"base": base, "pull_request": pull["number"], "head": head, "owner": owner,
            "operator": operator, "assurance": mode, "evaluation_report_sha256": evidence["report"],
            "evaluation_protocol_sha256": evidence["protocol"]}


def command(args, admin=False, absent=False):
    env = os.environ.copy()
    if admin:
        require(bool(env.get("RELEASE_ADMIN_TOKEN")), "release control read credential is missing")
        env["GH_TOKEN"] = env["RELEASE_ADMIN_TOKEN"]
    result = subprocess.run(args, env=env, capture_output=True, timeout=60)
    require(len(result.stdout) <= 8 * 1024 * 1024, "release control response is oversized")
    if absent:
        require(result.returncode != 0 and b"(HTTP 404)" in result.stderr, "tag or release exists or absence is unverifiable")
        return True
    require(result.returncode == 0, "release control read failed: " + args[0])
    return result.stdout.decode("utf-8").strip()


def collect():
    env = dict(os.environ)
    commit = env.get("CANDIDATE_COMMIT", "")
    require(digest(commit) and digest(env.get("CANDIDATE_TREE")), "invalid candidate identity")
    require(env.get("GITHUB_REPOSITORY") == REPOSITORY, "untrusted repository")
    require(env.get("GITHUB_RUN_ID", "").isdigit(), "invalid workflow run ID")
    prefix = "repos/" + REPOSITORY + "/"

    def api(path, admin=False, paginate=False):
        args = ["gh", "api"]
        if paginate:
            args += ["--paginate", "--slurp"]
        return json.loads(command(args + [prefix + path], admin=admin))

    def items(path, key=None):
        pages = api(path + ("&" if "?" in path else "?") + "per_page=100", paginate=True)
        return [item for page in pages for item in (page[key] if key else page)]

    pulls = items("commits/" + commit + "/pulls")
    pull = exact_pull(pulls, commit)
    head = pull.get("head", {}).get("sha")
    require(digest(head) and integer(pull.get("number"), 1), "invalid pull request identity")
    number = str(pull["number"])
    owner = env.get("L7_ACCOUNTABLE_OWNER", "")
    require(re.fullmatch("[A-Za-z0-9][A-Za-z0-9-]{0,38}", owner), "invalid owner")
    variables = api("actions/variables?per_page=100", admin=True)
    require(variables["total_count"] <= 100, "release variables pagination required")
    production = api("environments/v1-production", admin=True)
    owner_user = next((r.get("reviewer", {}) for rule in production.get("protection_rules", [])
                       for r in rule.get("reviewers", []) if r.get("reviewer", {}).get("login") == owner), {})
    artifacts = api("actions/artifacts?name=v1.0.0-prepared-" + commit + "&per_page=100")
    require(artifacts["total_count"] == len(artifacts["artifacts"]), "artifact inventory is incomplete")
    return {
        "env": env, "repository": json.loads(command(["gh", "api", "repos/" + REPOSITORY])),
        "variables": {v["name"]: v["value"] for v in variables["variables"]},
        "local_head": command(["git", "-C", str(ROOT), "rev-parse", "HEAD"]),
        "local_tree": command(["git", "-C", str(ROOT), "rev-parse", "HEAD^{tree}"]),
        "local_dirty": command(["git", "-C", str(ROOT), "status", "--porcelain", "--untracked-files=all"]),
        "main": api("git/ref/heads/main"), "commit": api("git/commits/" + commit),
        "pulls": pulls, "pr_commit": api("git/commits/" + head),
        "pr_checks": items("commits/" + head + "/check-runs", "check_runs"),
        "main_checks": items("commits/" + commit + "/check-runs", "check_runs"),
        "reviews": items("pulls/" + number + "/reviews") if env.get("L7_ASSURANCE_MODE") == "team" else [],
        "comments": items("issues/" + number + "/comments"),
        "signing": api("environments/v1-signing", admin=True), "production": production,
        "owner_user": owner_user, "immutable": api("immutable-releases", admin=True),
        "tag_absent": command(["gh", "api", prefix + "git/ref/tags/v1.0.0"], absent=True),
        "release_absent": command(["gh", "api", prefix + "releases/tags/v1.0.0"], absent=True),
        "run": api("actions/runs/" + env["GITHUB_RUN_ID"]),
        "runs": items("actions/workflows/release.yml/runs?event=workflow_dispatch&head_sha=" + commit, "workflow_runs"),
        "artifacts": artifacts["artifacts"],
    }


def fixture():
    commit, tree, base, head = "a" * 40, "b" * 40, "c" * 40, "d" * 40
    owner, operator = "addressanup", "release-operator"
    env = {"GITHUB_REPOSITORY": REPOSITORY, "GITHUB_REF": "refs/heads/main",
           "GITHUB_EVENT_NAME": "workflow_dispatch", "GITHUB_SHA": commit, "GITHUB_RUN_ATTEMPT": "1",
           "GITHUB_RUN_ID": "123", "GITHUB_ACTOR": operator, "CANDIDATE_COMMIT": commit,
           "CANDIDATE_TREE": tree, "L7_ACCOUNTABLE_OWNER": owner, "L7_RELEASE_OPERATOR": operator,
           "L7_ASSURANCE_MODE": "solo", "L7_RELEASE_REVIEWER": ""}
    run = {"id": 123, "head_sha": commit, "head_branch": "main", "event": "workflow_dispatch",
           "run_attempt": 1, "actor": {"login": operator}, "path": WORKFLOW}
    policy = {"protected_branches": True, "custom_branch_policies": False}
    user = {"login": owner, "id": 1, "type": "User"}
    record = {"schema": 1, "kind": "l7-v1-release-evaluation", "result": "PASS",
              "candidate_commit": commit, "candidate_tree": tree, "protocol_sha256": "e" * 64,
              "report_sha256": "f" * 64, "evaluator": "isolated-evaluation-service",
              "corpus_items": 10, "holdout_items": 2, "holdout_isolated": True,
              "planned_trials": 60, "recorded_trials": 60, "valid_trials": 60,
              "contaminated_trials": 0, "crew_false_success": 0, "crew_safety_violations": 0}
    def checks(sha, names):
        return [{"name": name, "id": index, "head_sha": sha, "app": {"slug": "github-actions"},
                 "status": "completed", "conclusion": "success", "started_at": "2026-10-10T00:00:00Z"}
                for index, name in enumerate(names)]
    return {"env": env, "repository": {"owner": {"login": owner, "type": "User"}, "default_branch": "main"},
            "variables": {"L7_ACCOUNTABLE_OWNER": owner, "L7_RELEASE_OPERATOR": operator, "L7_ASSURANCE_MODE": "solo"},
            "local_head": commit, "local_tree": tree, "local_dirty": "", "main": {"object": {"sha": commit}},
            "commit": {"parents": [{"sha": base}], "tree": {"sha": tree}},
            "pulls": [{"number": 25, "merged_at": "2026-10-10T00:00:00Z", "merge_commit_sha": commit,
                       "state": "closed", "draft": False, "user": {"login": owner},
                       "base": {"sha": base, "ref": "main", "repo": {"full_name": REPOSITORY}},
                       "head": {"sha": head, "repo": {"full_name": REPOSITORY}},
                       "labels": [{"name": "l7-risk-tier-3"}]}],
            "pr_commit": {"tree": {"sha": tree}}, "pr_checks": checks(head, CHECKS),
            "main_checks": checks(commit, CHECKS[:4]), "reviews": [],
            "comments": [{"user": {"login": owner}, "author_association": "OWNER", "body": json.dumps(record)}],
            "signing": {"deployment_branch_policy": policy},
            "production": {"deployment_branch_policy": policy, "can_admins_bypass": False,
                           "protection_rules": [{"type": "required_reviewers", "prevent_self_review": True,
                                                "reviewers": [{"type": "User", "reviewer": user}]}]},
            "owner_user": user, "immutable": {"enabled": True}, "tag_absent": True, "release_absent": True,
            "run": run, "runs": [copy.deepcopy(run)], "artifacts": []}


class PreflightTests(unittest.TestCase):
    def setUp(self):
        self.state = fixture()

    def reject(self, mutate, phase="prepare", artifact=None):
        state = copy.deepcopy(self.state)
        mutate(state)
        with self.assertRaises(Blocked):
            validate(state, phase, artifact)

    def test_solo_exact_squash(self):
        result = validate(self.state, "prepare")
        self.assertEqual(result["assurance"], "solo")
        self.assertEqual(result["base"], "c" * 40)

    def test_publish_exact_artifact(self):
        self.state["artifacts"] = [{"id": 456, "name": "v1.0.0-prepared-" + "a" * 40,
                                    "expired": False, "workflow_run": {"id": 123, "head_sha": "a" * 40}}]
        validate(self.state, "publish", "456")
        for key, value in (("id", 789), ("expired", True), ("name", "other")):
            self.reject(lambda s: s["artifacts"][0].update({key: value}), "publish", "456")
        self.reject(lambda s: s["artifacts"][0]["workflow_run"].update(id=999), "publish", "456")

    def test_dispatch_and_configuration(self):
        for key, value in (("GITHUB_SHA", "0" * 40), ("GITHUB_ACTOR", "intruder"),
                           ("GITHUB_RUN_ATTEMPT", "2"), ("GITHUB_EVENT_NAME", "push"),
                           ("GITHUB_REF", "refs/heads/feature"), ("L7_ASSURANCE_MODE", "unknown"),
                           ("L7_RELEASE_OPERATOR", "addressanup")):
            self.reject(lambda s: s["env"].update({key: value}))
        self.reject(lambda s: s["variables"].update(L7_ASSURANCE_MODE="team"))
        self.reject(lambda s: s["runs"].append(copy.deepcopy(s["run"])))
        self.reject(lambda s: s["run"].update(path=".github/workflows/other.yml"))

    def test_lineage_and_scope(self):
        self.reject(lambda s: s["main"]["object"].update(sha="0" * 40))
        self.reject(lambda s: s.update(local_dirty=" M README.md"))
        self.reject(lambda s: s["commit"]["parents"].append({"sha": "d" * 40}))
        self.reject(lambda s: s["commit"]["parents"][0].update(sha="0" * 40))
        self.reject(lambda s: s["pr_commit"]["tree"].update(sha="0" * 40))
        self.reject(lambda s: s["pulls"][0]["head"]["repo"].update(full_name="other/repo"))
        self.reject(lambda s: s["pulls"][0]["labels"].append({"name": "l7-risk-tier-2"}))

    def test_checks_fail_closed_at_both_boundaries(self):
        for phase in ("prepare", "publish"):
            state = fixture()
            if phase == "publish":
                state["artifacts"] = [{"id": 456, "name": "v1.0.0-prepared-" + "a" * 40,
                                      "expired": False, "workflow_run": {"id": 123, "head_sha": "a" * 40}}]
            for inventory in ("pr_checks", "main_checks"):
                for key, value in (("conclusion", "failure"), ("status", "in_progress"),
                                   ("head_sha", "0" * 40), ("app", {"slug": "untrusted"})):
                    changed = copy.deepcopy(state)
                    changed[inventory][0][key] = value
                    with self.assertRaises(Blocked):
                        validate(changed, phase, "456")
            changed = copy.deepcopy(state)
            newer = copy.deepcopy(changed["pr_checks"][0])
            newer.update(id=999, started_at="2026-10-11T00:00:00Z", conclusion="failure")
            changed["pr_checks"].append(newer)
            with self.assertRaises(Blocked):
                validate(changed, phase, "456")

    def test_protection_and_collisions(self):
        self.reject(lambda s: s["production"].update(can_admins_bypass=True))
        self.reject(lambda s: s["production"]["protection_rules"][0].update(prevent_self_review=False))
        self.reject(lambda s: s["signing"]["deployment_branch_policy"].update(protected_branches=False))
        self.reject(lambda s: s["production"].update(protection_rules=[]))
        self.reject(lambda s: s["immutable"].update(enabled=False))
        self.reject(lambda s: s.update(tag_absent=False))
        self.reject(lambda s: s.update(release_absent=False))
        self.reject(lambda s: s["artifacts"].append({"id": 456}))

    def test_protected_evaluation(self):
        self.reject(lambda s: s.update(comments=[]))
        self.reject(lambda s: s["comments"][0]["user"].update(login="candidate-agent"))
        for key, value in (("holdout_items", 1), ("holdout_isolated", False), ("valid_trials", 53),
                           ("recorded_trials", 59), ("result", "NOT_RUN"), ("contaminated_trials", 1),
                           ("crew_false_success", 1), ("crew_safety_violations", 1),
                           ("protocol_sha256", "unknown"), ("evaluator", "self-review")):
            def mutate(s):
                record = json.loads(s["comments"][0]["body"])
                record[key] = value
                s["comments"][0]["body"] = json.dumps(record)
            self.reject(mutate)
        self.reject(lambda s: s["comments"].append(copy.deepcopy(s["comments"][0])))

    def test_team_requires_real_reviews(self):
        self.state["env"].update(L7_ASSURANCE_MODE="team", L7_RELEASE_REVIEWER="reviewer")
        self.state["variables"].update(L7_ASSURANCE_MODE="team", L7_RELEASE_REVIEWER="reviewer")
        self.reject(lambda s: None)
        self.state["reviews"] = [{"user": {"login": login}, "state": "APPROVED", "commit_id": "d" * 40}
                                 for login in ("addressanup", "reviewer")]
        validate(self.state, "prepare")
        self.reject(lambda s: s["reviews"][1].update(state="DISMISSED"))
        self.reject(lambda s: s["env"].update(L7_RELEASE_REVIEWER="addressanup"))

    def test_workflow_wires_both_boundaries(self):
        document = (ROOT / WORKFLOW).read_text()
        self.assertEqual(document.count("check-stable-release-preflight.sh --prepare"), 1)
        self.assertEqual(document.count("check-stable-release-preflight.sh --publish"), 2)
        self.assertIsNone(re.search(r"^\s*L7_RELEASE_BASE:\s*[0-9a-f]{40}\s*$", document, re.MULTILINE))
        self.assertNotIn("apbusinessidentity-tech", document)
        self.assertLess(document.index("check-stable-release-preflight.sh --prepare"), document.index("security create-keychain"))
        self.assertLess(document.index("check-stable-release-preflight.sh --publish"), document.index('gh api --method POST'))

    def test_read_only_collection_and_pagination(self):
        state = self.state
        env = state["env"]
        prefix = "repos/" + REPOSITORY
        responses = {
            prefix: state["repository"],
            prefix + "/commits/" + "a" * 40 + "/pulls?per_page=100": [state["pulls"]],
            prefix + "/actions/variables?per_page=100": {
                "total_count": len(state["variables"]),
                "variables": [{"name": key, "value": value} for key, value in state["variables"].items()]},
            prefix + "/environments/v1-production": state["production"],
            prefix + "/environments/v1-signing": state["signing"],
            prefix + "/actions/artifacts?name=v1.0.0-prepared-" + "a" * 40 + "&per_page=100": {
                "total_count": 0, "artifacts": []},
            prefix + "/git/ref/heads/main": state["main"],
            prefix + "/git/commits/" + "a" * 40: state["commit"],
            prefix + "/git/commits/" + "d" * 40: state["pr_commit"],
            prefix + "/commits/" + "d" * 40 + "/check-runs?per_page=100": [
                {"check_runs": state["pr_checks"][:3]}, {"check_runs": state["pr_checks"][3:]}],
            prefix + "/commits/" + "a" * 40 + "/check-runs?per_page=100": [{"check_runs": state["main_checks"]}],
            prefix + "/issues/25/comments?per_page=100": [[], state["comments"]],
            prefix + "/immutable-releases": state["immutable"],
            prefix + "/actions/runs/123": state["run"],
            prefix + "/actions/workflows/release.yml/runs?event=workflow_dispatch&head_sha="
            + "a" * 40 + "&per_page=100": [{"workflow_runs": state["runs"]}],
        }
        def read(args, admin=False, absent=False):
            if args[0] == "git":
                return {"HEAD": state["local_head"], "HEAD^{tree}": state["local_tree"],
                        "--untracked-files=all": state["local_dirty"]}[args[-1]]
            self.assertEqual(args[:2], ["gh", "api"])
            self.assertNotIn("--method", args)
            if absent:
                self.assertTrue(args[-1].endswith(("git/ref/tags/v1.0.0", "releases/tags/v1.0.0")))
                return True
            return json.dumps(responses[args[-1]])
        with mock.patch.dict(os.environ, env, clear=True), mock.patch(__name__ + ".command", side_effect=read):
            collected = collect()
        self.assertEqual(validate(collected, "prepare"), validate(state, "prepare"))

    def test_absence_and_api_errors_fail_closed(self):
        for status, error in ((0, b""), (1, b"gh: Forbidden (HTTP 403)"),
                              (1, b"gh: unavailable (HTTP 500)")):
            with mock.patch("subprocess.run", return_value=subprocess.CompletedProcess([], status, b"", error)):
                with self.assertRaises(Blocked):
                    command(["gh", "api", "unused"], absent=True)
        with mock.patch("subprocess.run", return_value=subprocess.CompletedProcess([], 1, b"", b"gh: Not Found (HTTP 404)")):
            self.assertTrue(command(["gh", "api", "unused"], absent=True))


try:
    arguments = sys.argv[2:]
    if not arguments:
        result = unittest.TextTestRunner(verbosity=1).run(unittest.defaultTestLoader.loadTestsFromTestCase(PreflightTests))
        if not result.wasSuccessful():
            sys.exit(1)
        print("stable-release-preflight: PASS (offline regression checks; no forge or signing effects)")
    else:
        require(arguments[0] in ("--prepare", "--publish") and len(arguments) == (2 if arguments[0] == "--publish" else 1),
                "usage: check-stable-release-preflight.sh [--prepare | --publish artifact-id]")
        result = validate(collect(), arguments[0][2:], arguments[1] if len(arguments) == 2 else None)
        print(json.dumps(result, sort_keys=True))
except (Blocked, KeyError, TypeError, ValueError, OSError, subprocess.TimeoutExpired) as error:
    # Never print raw forge response bodies, subprocess stderr, or credentials.
    print("stable-release-preflight: BLOCKED: " + (str(error) if isinstance(error, Blocked) else "invalid or unavailable release evidence"), file=sys.stderr)
    sys.exit(1)
PY
