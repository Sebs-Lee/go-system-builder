#!/usr/bin/env python3
"""Opt-in disposable S0-S11 native Claude flow test; all professional answers simulated.

Real Main sessions use installed production assets/Hooks and native Agents.
This runner supplies only installation, clearly labeled scenario fixtures and
read-only observation. It never writes Runtime, receipts, stage cursors or Worker
answers, and never weakens a domain gate. No real product acceptance is claimed.
"""
import argparse
import hashlib
import json
import os
from pathlib import Path
import shlex
import shutil
import subprocess
import sys
import tarfile
import tempfile
import time

REPO = Path(__file__).resolve().parents[1]

OBSERVE = r'''import json,os,subprocess,sys,time,uuid
from pathlib import Path
artifacts,main,event,i,j=sys.argv[1:6]
raw=sys.stdin.buffer.read()
payload=json.loads(raw)
expected=json.loads((Path(artifacts)/'original-settings.json').read_text())['hooks'][event][int(i)]['hooks'][int(j)]['command']
original=sys.argv[6] if len(sys.argv)>6 else expected
if original!=expected: raise SystemExit('Observer command differs from the frozen original Hook')
def facts():
 try:
  s=json.loads((Path(main)/'.claude/loop-state.json').read_text())
  return {'revision':s.get('revision'),'lifecycle':s.get('lifecycle'),'repair_status':s.get('review',{}).get('repair',{}).get('status')}
 except (OSError,ValueError): return {}
before=facts();start=time.time_ns();clock_start=time.monotonic()
p=subprocess.run(['bash','--noprofile','--norc','-c',original],input=raw,capture_output=True)
row={'event':event,'started_ns':start,'finished_ns':time.time_ns(),'elapsed_seconds':time.monotonic()-clock_start,'project_dir':os.environ.get('CLAUDE_PROJECT_DIR'),'process_cwd':os.getcwd(),'before':before,'after':facts(),'exit_code':p.returncode,'stdout':p.stdout.decode(errors='replace'),'stderr':p.stderr.decode(errors='replace'),'payload':payload}
(Path(artifacts)/('hook-'+str(start)+'-'+uuid.uuid4().hex+'.json')).write_text(json.dumps(row,ensure_ascii=False))
sys.stdout.buffer.write(p.stdout);sys.stderr.buffer.write(p.stderr);sys.exit(p.returncode)
'''

BOUNDARY = '''# User-authorized simulation test override

This disposable project exists ONLY to simulate S0-S11 workflow and native
subagent/form reception. All professional results, findings, tests and human
choices here are synthetic test data, not real engineering/release acceptance.
This user test instruction takes precedence over ordinary product-work guidance.

This test keeps every product byte and file mode unchanged. S6's original canonical integration requires a source commit: in its actual native worktree, the S6 child prepares exactly one empty SIMULATION fixture commit (git commit --allow-empty -m 'SIMULATION source fixture') before its Result. This is disclosed test preparation with no product changes, not implementation or a product commit. S9 supports a genuine whole-batch no-change result: do not create any S9 commit or mode delta. Keep original Main checks and independent Contract verification. Do not fabricate changesets or weaken gates. If platform worktree cleanup interrupts an original consumer, preserve and report the exact recovery boundary rather than manufacturing a change. Native tools remain available: necessary file, directory, JSON and Git inspection is allowed. The boundary is simulated work content, not tool selection.

MAIN: actually use Claude Code native Agent for every delegated responsibility.
Repeat this simulation override in EACH child's task prompt, with exact original
questionnaire/brief references. Main may prepare minimal specifications, original
professional dispatch/Case/plan forms and perform original integration. Main must
NOT author a delegated review/Builder/Investigator answer on the child's behalf.

Professional conclusions are canned test inputs, not questions for the model to re-derive. Pass each original responsibility and its full assertion set by reference; do not replace it with a shortened assertion list or a RED-only subset. Preserve required fields and values in unreceived editable professional JSON; do not add a byte-identical or roundtrip requirement. Original immutable sources and accepted facts keep their original source checks. Do not request real independent judgment, source audit or broad document reading. Normal pass/completed enum values in explicitly labeled mock answers are permitted; never describe them as real executed checks.

S6 child saves its Plan, prepares the one empty SIMULATION fixture commit without git add or any byte/mode change, then saves Result in the same native turn and returns. No Plan-only Stop, Main waiting handshake, private-file polling or extra activation is needed. S9 creates no commit. This test-only fixture action does not exist in production guidance.

EVERY SUBAGENT (including Builder): read the supplied original brief/view and
necessary scenario inputs, simulate the role, directly fill its original
private professional answer file(s), return EXACT engineering references and
state that all conclusions are simulated. Do NOT implement or repair products,
write product source/tests, run real product app/build/test suites, create product commits, create a Team,
launch another Claude process, bind platform IDs, activate or manually consume
answers. No duplicate generic PLAN/readback. Ordinary save is quiet; Start/Stop
and the original domain consumers receive artifacts. Native worktree isolation
is used for S6/S9 private files; read-only roles may use original Main paths.

Do not edit Runtime/journal/receipts, protected governance, settings, AGENTS.md,
this file, or production schemas. Do not bypass gates, inject lifecycle events,
mark unperformed real checks PASS, publish, push, deploy, or perform a real
release. Use canonical schema values while explicitly labeling professional
narrative and evidence as simulation. If an original gate cannot accept a no-op
simulation, report the exact blocker; do not secretly implement products.
'''

PROMPT = '''在当前隔离项目跑 S0～S11【模拟答卷流转】。先读 CLAUDE.md、simulation/MAIN-FORMS.md 和 simulation/DECISIONS.md。所有专业结论均为给定测试数据，不让角色真的审查/调查/实现/运行产品测试。原 Writer/Schema/源绑定/Gate 正常运行，不改 Runtime、收据或保护资产，不绕过原消费者。不要 push/deploy/真实发布。

角色执行方式：实际用 Claude Code 原生 Agent 创建简短的模拟上下文，分别模拟 document-verifier、Builder、QA、Investigator、Targeted 等原专业责任，general-purpose 可以正常使用。原 manifest/Assignment 的 role_family/owner 保留，平台 actor 不要求同名。不增加自造角色类型或工具白名单。每个 child prompt 给清楚：模拟声明、准确原view/answer或brief、预设测试结论、必要字段要求。child按需使用原生工具读取材料、检查目录/JSON/Git状态，直接填本人原专业JSON并返回准确refs；不要求重新审查产品背景、真实专业判断、实际产品实现/修复/测试或产品提交。不Team/launcher/activate/手工consume。S6/S9使用原生worktree隔离填写private plan/result。**Main不可代填Worker答案**。专业叙述写SIMULATION/test-only，枚举仍用正常pass/completed等；这是mock数据，不称为真实执行的检查结果。工具操作应如实叙述，不能把“未执行产品测试”写成“没有使用任何命令”。

从fresh inactive开始，使用系统REQ-001。场景：仅NFR-001要求simulation/products/artifact.txt保持SIMULATION_ONLY标记，AC-001检查该预置文件；无UI/FR/API/服务行为/共享模型/数据库/部署，适用N/A由NFR/负空间背书，不扩范围。模拟人类明确锁定并启动v1.0.0，dev=simulation-dev，local release=simulation-release，simulation-human认可最小ADR、测试Contract并在S11选择模拟approve。只写隔离Runtime内的测试决定，无外部发布。

若.claude/test-inputs/candidates目录存在，这是事先准备的专业文档测试输入（没有任何Runtime或收据），Main在S0/S2/S3/S4各自阶段复制相应候选文档到实际输出路径，提交所需Git源并填原Main答卷即可，不重新设计产品。没有候选时按现有模板极小填写。正常TASK prospective write paths要包含simulation/products/artifact.txt，不能把“不真正写产品”的测试覆盖指令误写成空scope/永久禁写的专业TASK。产品文件已经预置，**整轮内容字节不修改**；S6 child仅在真实原生树准备一次明确标记的空提交测试输入；S9完整走无改动结果，不改执行位、不制造提交。普通scope是工程数据，测试执行不做实现。TASK Closing Contract用assert `make check` == pass加BE条款条件；make check是准备器已有的只读标记校验，只让Main原集成运行，Worker不做产品测试。其它planning/traceability形状按原模板。Main可以做原登记/集成及这个简单fixture检查。

路线与给定mock数据：S0锁REQ→S1四问原绑定→S2最小架构/ADR→S3仅必要合同→S4单TASK→S5两个独立mock Agent各conclusion=pass/findings=[]，不要求真实判断。S5原阶段出口自动准备两项独立责任及准确原卷；直接携带Hook给的refs原生派发，不创建manifest、不抄Runtime、不登记或手动issue。仅有真实专业分工调整时编辑原清单。→S6 s6 prepare后派模拟Builder（native isolation=worktree；Builder保存Plan后仅运行一次git commit --allow-empty -m SIMULATION-source-fixture作为明确测试夹具，再保存Result，同一回合完成，不Plan-only返回；不git add、不改产品字节或执行位；Worker填模拟卷，可用必要原生工具，不做产品写入或产品测试）：只填专业plan和Result，closing条件用预设satisfied/completed，注明没做真实实现/Worker测试，以现有artifact.txt为evidence。Main执行原task-integrate，实际make check。→S7：Main填原ReviewPlan/注册，原生mock delivery/qa各写原卷；首轮QA按simulation/observations/first-round.txt给一个明确SIMULATION finding，其它Claim mock pass。Finding只给path:simulation/observations/first-round.txt，原接收器自动固定SHA，同一原Finding来源自然进入S8；任何agent都不计算或复制摘要。不要先登记索引别名，不伪造test://或已执行检查。S8新Hypothesis外层仍引用实际Finding/Batch等原ID，来源tuple由原消费者生成，无需Main抄写。→S8：原Case/Hypothesis派mock Investigator，给定supported结论，无真实调查；Main按原表单形成仅fixture路径的模拟Contract并用场景中模拟人类批准。→S9：专业分工后派原生worktree模拟Builder，只写本人worktree的PlanReport/Result。S9不制造任何提交或文件模式差异，proposed_paths=[]；用完整已接收Final的实际空changeset进入影响判断，保原合同断言和独立复验，不能伪造路径已修改、手工替换Final changeset或放松原门。模拟before/after以已有fixture证据作refs，Worker不改产品源码、不做产品测试或产品提交；Main原集成选择make check。收到实际native Stop的candidate后立即走原Main集成，不给已交付Builder重复派legacy RepairResult，不建立until/poll等待Final的循环。→独立mock Targeted直接填写pass的原卷→fresh S7全部适用Claim mock pass，E2E如真的不适用保留系统N/A disposition不造浏览器工作→S10 Main正常填原acceptance/release-audit Manifest，用明确SIMULATION完成结论→S11原应用填写场景预给的模拟approve，终点release_authorized，不rollover。

各阶段读取准确view与必要材料；专业数据是测试输入，不要重新推导真实结论。禁止给mock Agent添加“独立判定/逐项真实性核验”等工作。保存静默，Stop才接收Worker；Main工具也不能抢收。原生返回后沿当前责任和当前Case引用读取原回执；版本化Case文件是不可变来源，不能轮询固定旧版本等待更新；private Plan/Result 也不能当作平台结束或原接收的等待信号，原生返回后读当前候选/回执再集成。原消费者拒绝字段时可原生resume同child修原卷，不能改已接收答案，修订走原命名恢复。不得用stage-answer issue/consume替代正常专业交付或用runtime transition绕过阶段门。Main可以执行原Hook给出的准确issue准备命令；实际发生原系统接收失败时，也可执行原Hook给出的准确consume/recover恢复命令，只续未完成的技术边；不得换源、换责任、改已收事实或为纯系统失败唤醒已履职Worker。持续到S11测试终点或不可恢复的具体系统阻塞，不停在下一步派发提示。若no-op产品来源确实被原接收拒绝，报告准确条件，不实现产品或放松Gate。'''



def run(cmd, **kwargs):
    return subprocess.run([str(x) for x in cmd], check=True, **kwargs)


def dump(path, obj):
    path.write_text(json.dumps(obj, ensure_ascii=False, indent=2) + '\n')


def file_inventory(root, paths):
    return {str(path): hashlib.sha256((root/path).read_bytes()).hexdigest()
            for path in sorted(paths)}


def tracked_inventory(root):
    paths = subprocess.check_output(['git', 'ls-files', '-z'], cwd=root).decode().split('\0')
    return file_inventory(root, [Path(path) for path in paths if path])


def prepare(base, dry_run=False):
    # A new directory is required: never clean or reuse a live simulation.
    if base.exists():
        raise SystemExit('Preparation requires a new artifacts directory; existing runs are preserved')
    candidates = REPO/'tests/fixtures/native-stage-simulation/documents'
    paths = sorted(candidates.rglob('*')) if candidates.is_dir() else []
    if any(path.is_symlink() or path.is_file() and path.suffix != '.md' for path in paths):
        raise SystemExit('Candidate inputs must be regular Markdown documents without symlinks')
    candidate_files = file_inventory(candidates, [path.relative_to(candidates) for path in paths if path.is_file()])
    if dry_run:
        print(json.dumps({'dry_run': True, 'artifacts': str(base), 'candidate_files': candidate_files,
                          'candidate_inputs_optional': True, 'build_install_model_started': False}, ensure_ascii=False))
        return
    base.mkdir(parents=True, exist_ok=False)
    started_ns = time.time_ns()
    started = time.monotonic()
    completed = False
    try:
        prepare_project(base, candidate_files)
        completed = True
    finally:
        dump(base/'preparation.json', {'started_ns': started_ns, 'finished_ns': time.time_ns(),
             'seconds': round(time.monotonic()-started, 3), 'completed': completed,
             'scope': 'release build, install, fixture and Git preparation; no model request'})


def prepare_project(base, candidate_files):
    version = 'native-stage-simulation-' + time.strftime('%Y%m%d-%H%M%S', time.gmtime())
    with (base/'prepare.log').open('w') as log:
        run(['bash', REPO/'packaging/build-release.sh', version, base/'release.tar.gz'], cwd=REPO, stdout=log, stderr=subprocess.STDOUT)
        unpack = base/'release'
        unpack.mkdir()
        with tarfile.open(base/'release.tar.gz') as archive:
            archive.extractall(unpack, filter='data')
        package = unpack/('vibe-coding-loop-template-'+version)
        root = base/'project'
        run([package/'.claude/bin/loop-harness', 'install', '--source', package, '--root', root], stdout=log, stderr=subprocess.STDOUT)
        run(['git', 'init', '-b', 'simulation-dev'], cwd=root, stdout=log, stderr=subprocess.STDOUT)
        run(['git', 'config', 'user.name', 'Simulation Fixture'], cwd=root)
        run(['git', 'config', 'user.email', 'simulation@example.invalid'], cwd=root)
    (root/'.gitignore').write_text('.claude/*\n!.claude/agents/\n!.claude/skills/\n!.claude/settings.json\n')
    (root/'CLAUDE.md').write_text(BOUNDARY)
    (root/'simulation').mkdir(exist_ok=True)
    candidates = REPO/'tests/fixtures/native-stage-simulation/documents'
    if candidates.is_dir():
        shutil.copytree(candidates, root/'.claude/test-inputs/candidates')
        if file_inventory(root/'.claude/test-inputs/candidates', [Path(path) for path in candidate_files]) != candidate_files:
            raise SystemExit('Candidate inputs changed during preparation; preserve this failed run')
        inputs = {'kind':'synthetic Main document inputs; no Runtime, receipts or Worker answers', 'files':candidate_files}
        dump(root/'simulation/candidate-inputs.json', inputs)
        dump(base/'candidate-inputs.json', inputs)
    (root/'simulation/products').mkdir(parents=True)
    (root/'simulation/observations').mkdir()
    (root/'simulation/products/artifact.txt').write_text('SIMULATION_ONLY\n')
    (root/'Makefile').write_text('.PHONY: check\ncheck:\n\tpython3 -c "from pathlib import Path; assert Path(\'simulation/products/artifact.txt\').read_bytes() == b\'SIMULATION_ONLY\\n\'"\n')
    (root/'simulation/observations/first-round.txt').write_text('SIMULATION ONLY: injected first-round QA finding scenario. Expected marker provenance was supplied; synthetic observation models a missing provenance link. No actual product defect or executed test is asserted. Scope: simulation/products/artifact.txt. Use as synthetic encounter and before-fix evidence.\n')
    (root/'simulation/observations/repaired-round.txt').write_text('SIMULATION ONLY: canned repaired-round observation. In the hypothetical repaired outcome provenance is present. This is test fixture data, not executed code or a real check PASS. No product implementation happened.\n')
    (root/'simulation/MAIN-FORMS.md').write_text('''# Explicit synthetic test input and original form shapes

This is read-only test assistance, not Runtime, receipts, a Worker answer or real human consent. Use the actual Hook-provided professional forms and references; internal machine schemas are not author DTOs and are not copied into this scenario. Correct unreceived professional drafts at the same exact path. Original consumers supply immutable evidence and decision identity; do not invent or manually copy machine coordinates.

The S6 child in native worktree saves Plan, creates exactly one labeled empty SIMULATION source fixture commit with no git add or product byte/mode change, saves Result and returns in the same turn. There is no waiting handshake or Plan-only Stop. S9 uses a native worktree and normal platform return/notification, without a required background mode or a private-file polling loop. The S6 child prepares one clearly labeled empty S6 source fixture commit because the original canonical S6 integration requires it; product bytes and mode stay unchanged. S9 is a genuine no-change responsibility: proposed_paths=[], no fixture commit or mode delta, actual empty Final changeset, original Main checks and independent verification still required. Workers fill simulated private answers using native tools as needed, including necessary file, directory, JSON and Git inspection. Do not implement, repair, test or commit real products. Tools are not a stage gate. Preserve original schemas, scope, sources and gates. No explicit --paths fabrication, direct transition bypass, or consume in place of normal professional delivery. Main may execute exact named issue preparation supplied by the original Hook. For an actual original system reception failure, Main may also execute the exact named consume/recover recovery supplied by the original Hook, continuing only unfinished technical work without replacing sources, responsibility or accepted facts, or waking a Worker that has completed its professional work.

S5 normal stage entry prepares the original independent responsibilities and their exact questionnaires automatically. Carry those references into native Agents; do not create/register a manifest or manually issue packets. S7 ReviewPlan is generated: retain every machine field and generated Claim/Assignment; replace all TODO oracles in ONE save with preset SIMULATION descriptions. First-round Finding cites only path:simulation/observations/first-round.txt; the original receiver fixes its actual SHA. No actor computes or transcribes a digest. The original S8 consumer freezes this accepted Finding source without a new index or alias. New Hypothesis outer refs use actual Finding/Batch IDs; do not copy source tuples into professional answers or invent a test:// execution. The original Plan retains every applicable Claim. Main may group compatible QA Claims through the original Assignment fields; delivery remains independent. Native Agent contexts such as general-purpose need no custom actor type or tool whitelist. First round has one canned P2 provenance-gap finding; fresh second round is all canned pass. All findings and checks are explicitly synthetic, not executed professional verification.

S8 fixes the final professional Contract draft and supplies the original decision application. The simulated human choice is pre-given in DECISIONS.md; label it SIMULATION and fill only the presented decision/reason. Runtime/Case/Contract identity, approval hash, evidence and original approval consumption are system responsibilities. Do not compute or copy them, register human_decision, author a raw envelope or invoke approval manually. No editing already accepted files. Use canonical scope directory strings without a trailing slash in new fixture data.

S9 PlanReport output is relative to actual native worktree, not Main. Its professional object has plan, red_checks and proposed_paths. Red checks are explicitly prescribed mock data with normal enum values and real fixture file references; no Worker execution is claimed. Result is the exact original professional object from the prepared brief. Return exact brief and absolute actual Result, Stop receives. Main performs original integration with actual make check. For the no-change batch, change_types and changed_artifacts remain empty from the actual Final; never manufacture a changed path. The original Contract still requires independent verification.

S9 ChangeImpact is the original FLAT DTO, no answer/title/questions wrapper. Preserve every machine field. The accepted no-change batch has change_types=[] and changed_artifacts=[]; do not fill either with invented work. Give a professional decisions row with source_id=<actual Case>, target_id=<actual Contract root assertion slot, e.g. root-1>, relation="requires-verification", rule_id="IM-CONTRACT", decision="reverify", scope=[], rationale="SIMULATION: no product edit proposed; independently observe all original Contract assertions", recovery_evidence=[<actual accepted Final ref>]. Preserve generated required_reverification_ids. The original consumer supplies identities, sources and obligations. The independent Targeted form is also the ORIGINAL FLAT DTO; preserve machine fields and actual Contract assertion IDs, use canned pass observations explicitly labeled synthetic. Main must not fill a Worker answer.

S10 use the original exact Main Manifest shape, explicitly mock professional completion and genuine source receipts. S11 fills the original application with the pre-given SIMULATION approve, ending at isolated release_authorized. No publishing, push, deployment or rollover.
''')
    (root/'simulation/SCENARIO.md').write_text('# Synthetic full-stage scenario\n\n'+PROMPT)
    (root/'simulation/DECISIONS.md').write_text('# Test-only human scenario decisions\n\nIdentity: simulation-human. These are canned simulation choices authorized for this disposable workflow test, not production human consent.\n\n- Lock/start fixture REQ-001 at v1.0.0; development=simulation-dev; release target=local simulation-release.\n- Endorse the minimal fixture-only ADR direction and genuine UI/API/service/browser/deployment N/A.\n- Approve only the synthetic RepairContract for simulation/products/artifact.txt, after its exact draft is presented through the original domain workflow.\n- At S11 choose simulated approve for this fixture candidate only. Do not publish, deploy, push or perform external release.\n')
    with (base/'prepare.log').open('a') as log:
        run(['git', 'add', '.'], cwd=root, stdout=log, stderr=subprocess.STDOUT)
        run(['git', 'commit', '-qm', 'Install current assets and explicitly synthetic scenario fixtures'], cwd=root, stdout=log, stderr=subprocess.STDOUT)
        run(['git', 'branch', 'simulation-release'], cwd=root, stdout=log, stderr=subprocess.STDOUT)
    settings = json.loads((root/'.claude/settings.json').read_text())
    dump(base/'original-settings.json', settings)
    (base/'observe-hook.py').write_text(OBSERVE)
    for event, registrations in settings['hooks'].items():
        for i, registration in enumerate(registrations):
            for j, hook in enumerate(registration['hooks']):
                hook['command'] = shlex.join(['python3',str(base/'observe-hook.py'),str(base),str(root),event,str(i),str(j),hook['command']])
    dump(root/'.claude/settings.json', settings)
    with (base/'prepare.log').open('a') as log:
        run(['git','add','.claude/settings.json'], cwd=root, stdout=log, stderr=subprocess.STDOUT)
        run(['git','commit','-qm','Observe original Hooks without tracking live Runtime'], cwd=root, stdout=log, stderr=subprocess.STDOUT)
    initial = json.loads((root/'.claude/loop-state.json').read_text())
    dump(base/'initial-runtime.json', initial)
    dump(base/'metadata.json', {'scope':'simulation-only S0-S11; real native Claude parent and agents','project':str(root),'created_at':time.strftime('%Y-%m-%dT%H:%M:%SZ', time.gmtime()),'claude_version':subprocess.check_output(['claude','--version'],text=True).strip(),'source_head':subprocess.check_output(['git','rev-parse','HEAD'],cwd=REPO,text=True).strip(),'launcher_sha256':hashlib.sha256((root/'.claude/bin/loop-harness').read_bytes()).hexdigest(),'binary_sha256':hashlib.sha256((root/'.claude/bin/loop-harness-linux-amd64').read_bytes()).hexdigest(),'release_archive_sha256':hashlib.sha256((base/'release.tar.gz').read_bytes()).hexdigest(),'source_worktree_status':subprocess.check_output(['git','status','--short'],cwd=REPO,text=True),'fixture_strategy':'Child empty S6 source fixture only; genuine no-change S9; no product bytes or mode changes; native tools available for simulated answers','fixture_initial_mode':oct((root/'simulation/products/artifact.txt').stat().st_mode & 0o777),'fixture_sha256':hashlib.sha256((root/'simulation/products/artifact.txt').read_bytes()).hexdigest()})
    (base/'prompt.txt').write_text(PROMPT)
    status = subprocess.check_output(['git', 'status', '--porcelain'], cwd=root, text=True)
    if status:
        raise SystemExit('Prepared Git tree is not clean; preserve this failed run')
    dump(base/'prepared-inputs.json', {
        'project_head': subprocess.check_output(['git', 'rev-parse', 'HEAD'], cwd=root, text=True).strip(),
        'tracked_files': tracked_inventory(root),
        'candidate_files': candidate_files,
        'prompt_sha256': hashlib.sha256((base/'prompt.txt').read_bytes()).hexdigest(),
        'note': 'Frozen before the first model request. Never clean fixtures during a run.'})
    print(json.dumps({'prepared':str(base),'project':str(root),'initial_lifecycle':initial.get('lifecycle')},ensure_ascii=False), flush=True)


def launch(base, env_file, resume, prompt_file, driver=False):
    root = base/'project'
    if not env_file.is_file():
        raise SystemExit('Provider configuration file is missing; never supply its contents in logs')
    index = len(list(base.glob('claude-run-*.jsonl'))) + 1
    if index == 1:
        frozen = json.loads((base/'prepared-inputs.json').read_text())
        if (subprocess.check_output(['git', 'status', '--porcelain'], cwd=root, text=True)
                or subprocess.check_output(['git', 'rev-parse', 'HEAD'], cwd=root, text=True).strip() != frozen['project_head']
                or tracked_inventory(root) != frozen['tracked_files']):
            raise SystemExit('Prepared Git inputs drifted before the first request; prepare a fresh run')
        candidate_root = root/'.claude/test-inputs/candidates'
        actual_candidates = file_inventory(candidate_root, [path.relative_to(candidate_root) for path in candidate_root.rglob('*') if path.is_file()])
        if actual_candidates != frozen['candidate_files']:
            raise SystemExit('Prepared candidate inputs drifted; prepare a fresh run')
        if hashlib.sha256((base/'prompt.txt').read_bytes()).hexdigest() != frozen['prompt_sha256']:
            raise SystemExit('Prepared scenario prompt drifted; prepare a fresh run')
    prompt = prompt_file.read_text() if prompt_file else (base/'prompt.txt').read_text()
    dump(base/f'request-{index}.json', {'prompt_sha256': hashlib.sha256(prompt.encode()).hexdigest(),
         'prompt_source': str(prompt_file) if prompt_file else 'prepared prompt.txt',
         'resume': resume, 'test_driver': driver})
    args = ['claude','-p','--permission-mode','bypassPermissions','--output-format','stream-json','--verbose','--include-hook-events','--setting-sources','user,project','--strict-mcp-config','--mcp-config','{"mcpServers":{}}','--no-chrome','--max-turns','240']
    if driver:
        system = "You are a disposable workflow simulation test driver. Use the installed original forms, native Agents and domain Writers with exact current references. Professional conclusions are supplied synthetic inputs, not real product verification. Give each child the simulation instruction and its original task/answer references; the child fills its own original professional answers. Native tools remain available for necessary reading, file handling and metadata checks. Do not implement, repair, test or commit real products, edit Runtime/receipts/governance, bypass stage gates, publish or deploy. S6/S9 use native worktree isolation. Main performs original integration and the read-only fixture check; professional observations remain explicitly synthetic. Continue from the current recorded responsibility through S11 simulated authorization or report a precise blocker. Runtime and original journal are authority."
        (base/f'test-driver-system-{index}.txt').write_text(system)
        args += ['--system-prompt',system]
    if resume:
        args += ['--resume',resume]
    script = 'source "$1" >/dev/null 2>&1 && shift && exec "$@"'
    started_ns=time.time_ns()
    started=time.monotonic()
    with (base/f'claude-run-{index}.jsonl').open('w') as out, (base/f'claude-run-{index}.stderr.log').open('w') as err:
        try:
            result=subprocess.run(['bash','--noprofile','--norc','-c',script,'native-stage-simulation',str(env_file),*args],cwd=root,input=prompt,text=True,stdout=out,stderr=err,timeout=3600)
            code=result.returncode
        except subprocess.TimeoutExpired:
            code=124
        except KeyboardInterrupt:
            code=130
    dump(base/f'run-{index}.json',{'exit_code':code,'seconds':round(time.monotonic()-started,3),'started_ns':started_ns,'finished_ns':time.time_ns(),'external_interference':'driver interrupted by SIGINT' if code == 130 else 'driver timeout' if code == 124 else 'signal exit observed' if code < 0 else 'not inferred; inspect external intervention records','timing_scope':'model process wall time includes tools, Hooks and rework; not pure inference','resume':resume,'effort':'provider/default','test_driver':driver,'simulation_only':True})
    summarize(base)
    raise SystemExit(code)


def summarize(base):
    rows=[]
    for path in base.glob('hook-*.json'):
        try:
            row = json.loads(path.read_text())
            row['observation_file'] = path.name
            rows.append(row)
        except (OSError,ValueError): pass
    rows.sort(key=lambda x:x['started_ns'])
    cursors=[];actors={};ordinary=[]
    def cursor(facts):
        life=facts.get('lifecycle') or {}
        return (life.get('state'),life.get('phase'))
    for row in rows:
        for label in ('before','after'):
            c=cursor(row[label])
            if not cursors or cursors[-1]!=list(c): cursors.append(list(c))
        actor=row['payload'].get('agent_id')
        if actor:
            info=actors.setdefault(actor,{'starts':0,'stops':0,'events':0,'stage_at_start':None,'repair_status_at_start':None,'start_observations':[],'stop_diagnostics':[]})
            info['events']+=1
            if row['event']=='SubagentStart':
                info['starts']+=1;info['stage_at_start']=cursor(row['before']);info['repair_status_at_start']=row['before'].get('repair_status');info['start_observations'].append({'at_ns':row['started_ns'],'stage':cursor(row['before']),'repair_status':row['before'].get('repair_status')})
            elif row['event']=='SubagentStop': info['stops']+=1;info['stop_diagnostics'].append({'exit':row['exit_code'],'message':row['stderr'][-2200:]})
            else: ordinary.append(row)
    revision_windows=[]
    for row in ordinary:
        if row['before'].get('revision') == row['after'].get('revision'): continue
        overlaps=[{'file':other['observation_file'], 'event':other['event'], 'actor':other['payload'].get('agent_id'), 'before':other['before'].get('revision'), 'after':other['after'].get('revision')} for other in rows if other is not row and (other['event'] in ('SubagentStart','SubagentStop','SessionStart','PostToolUse') or other['event']=='PreToolUse' and not other['payload'].get('agent_id')) and other['started_ns'] <= row['finished_ns'] and other['finished_ns'] >= row['started_ns'] and other['before'].get('revision') != other['after'].get('revision')]
        revision_windows.append({'file':row['observation_file'], 'event':row['event'], 'actor':row['payload'].get('agent_id'), 'before':row['before'].get('revision'), 'after':row['after'].get('revision'), 'overlapping_events':overlaps})
    dump(base/'concurrent-revision-windows.json', {'windows':revision_windows, 'note':'Before/after samples span concurrent original Hook processes; a revision delta is not attribution. Inspect the exact original journal transaction before alleging an ordinary Worker write.'})
    streams=[]
    for path in sorted(base.glob('claude-run-*.jsonl'), key=lambda p:int(p.stem.rsplit('-',1)[1])):
        for line in path.read_text().splitlines():
            try: streams.append(json.loads(line))
            except ValueError: pass
    native_calls=[];native_messages=[];results=[];sessions=[];latest_activity=[]
    hook_started={};hook_responses={};hook_failures=[]
    for row in streams:
        if row.get('type') == 'system' and row.get('subtype') == 'hook_started': hook_started[row.get('hook_id')]=row
        if row.get('type') == 'system' and row.get('subtype') == 'hook_response':
            hook_responses[row.get('hook_id')]=row
            if row.get('exit_code') not in (None, 0) or row.get('outcome') != 'success':
                hook_failures.append({key:row.get(key) for key in ('hook_id','hook_name','hook_event','exit_code','outcome','stderr')})
        if row.get('session_id') and row['session_id'] not in sessions: sessions.append(row['session_id'])
        if row.get('type')=='result': results.append({k:row.get(k) for k in ('session_id','is_error','subtype','result','num_turns','duration_ms','duration_api_ms')})
        if row.get('type')=='assistant':
            for part in row.get('message',{}).get('content',[]):
                if part.get('type')=='text': latest_activity.append({'text':part['text'][-350:]})
                elif part.get('type')=='tool_use': latest_activity.append({'tool':part.get('name'),'input_preview':str(part.get('input'))[:180]})
                if isinstance(part,dict) and part.get('type')=='tool_use' and part.get('name') in ('Agent','Task'):
                    native_calls.append({'input':part.get('input'), 'tool_use_id':part.get('id')})
                if isinstance(part,dict) and part.get('type')=='tool_use' and part.get('name')=='SendMessage':
                    native_messages.append({'input':part.get('input'),'tool_use_id':part.get('id')})
    writes=[];commands=[];violations=[]
    for row in rows:
        actor=row['payload'].get('agent_id')
        if not actor or row['event']!='PreToolUse': continue
        payload=row['payload']; tool=payload.get('tool_name'); inp=payload.get('tool_input') or {}
        if tool in ('Write','Edit','MultiEdit'):
            path=inp.get('file_path') or inp.get('path') or ''
            private=('/.claude/submissions/' in path or '/.claude/assignments/native/' in path or path.startswith('.claude/submissions/') or path.startswith('.claude/assignments/native/'))
            item={'actor':actor,'tool':tool,'path':path,'private_answer_path':private}
            writes.append(item)
            # Retain non-private writes for review; a path alone does not prove product work.
        if tool=='Bash':
            cmd=inp.get('command') or ''
            item={'actor':actor,'command':cmd}
            commands.append(item)
            forbidden=any(x in cmd for x in ('git commit','git add ','go test','go build','npm test','npm run','pytest','cargo test','make '))
            if forbidden: violations.append(item)
    dump(base/'worker-tool-audit.json',{'writes':writes,'bash_commands':commands,'potential_boundary_violations':violations,'note':'Read-only observations, not tool restrictions. Commands and non-private writes require review of actual work content; raw original Hook payloads are authoritative.'})
    root=base/'project'
    state=json.loads((root/'.claude/loop-state.json').read_text())
    journal_rows=[]
    journal_path=root/'.claude/loop-events.jsonl'
    if journal_path.exists():
        for line in journal_path.read_text().splitlines():
            try: entry=json.loads(line)
            except ValueError: continue
            if entry.get('event') == 'transition_committed':
                journal_rows.append({key:entry.get(key) for key in ('sequence','event_id','transition_id','before_revision','after_revision','from','to','request_id','evidence_ids')})
    evidence_drift=[]
    for entry in state.get('evidence', []):
        if entry.get('status') != 'valid': continue
        ref=entry.get('path'); expected=entry.get('sha256')
        if not ref or not expected: continue
        target=root/ref
        try: actual=hashlib.sha256(target.read_bytes()).hexdigest()
        except OSError: actual=None
        if actual != expected:
            evidence_drift.append({'id':entry.get('id'), 'kind':entry.get('kind'), 'path':ref, 'expected_sha256':expected, 'actual_sha256':actual})
    dump(base/'registered-evidence-source-audit.json', {'valid_source_drift':evidence_drift, 'note':'Read-only test-input hygiene audit; this never repairs evidence, invalidates an entry or changes Runtime.'})
    fixture_changed=hashlib.sha256((root/'simulation/products/artifact.txt').read_bytes()).hexdigest()!=json.loads((base/'metadata.json').read_text())['fixture_sha256']
    s1_bindings=[x['observation_file'] for x in rows if x['event']=='PostToolUse' and cursor(x['before'])[0]=='inactive' and cursor(x['after'])==('planning','design') and 'S1 binding recorded through original TR-001' in x.get('stdout','')]
    data={'simulation_only':True,'original_s1_binding_observations':s1_bindings,'platform_hooks':{'started':len(hook_started),'responses':len(hook_responses),'unreturned_started_ids':sorted(set(hook_started)-set(hook_responses)),'non_success':hook_failures},'journal_transitions':journal_rows,'latest_activity':latest_activity[-3:],'native_actor_count':len(actors),'cursors_observed':cursors,'native_calls':native_calls,'native_messages':native_messages,'actors':actors,'ordinary_worker_events':len(ordinary),'ordinary_worker_nonzero_exit':sum(x['exit_code']!=0 for x in ordinary),'ordinary_worker_nonempty_stdout':sum(bool(x['stdout']) for x in ordinary),'ordinary_worker_revision_changes':len(revision_windows),'ordinary_revision_windows_with_overlapping_events':sum(bool(x['overlapping_events']) for x in revision_windows),'sessions':sessions,'results':results,'final_revision':state.get('revision'),'final_lifecycle':state.get('lifecycle'),'fixture_changed':fixture_changed,'fixture_mode_changed':oct((root/'simulation/products/artifact.txt').stat().st_mode & 0o777)!=json.loads((base/'metadata.json').read_text())['fixture_initial_mode'],'worker_private_answer_writes':sum(x['private_answer_path'] for x in writes),'worker_other_writes':sum(not x['private_answer_path'] for x in writes),'worker_bash_commands':len(commands),'potential_boundary_violations':violations,'registered_evidence_source_drift':evidence_drift,'gate_bypass_by_runner':False}
    intervals = sorted((row['started_ns'], row['finished_ns']) for row in rows)
    merged = []
    for start, end in intervals:
        if merged and start <= merged[-1][1]:
            merged[-1][1] = max(merged[-1][1], end)
        else:
            merged.append([start, end])
    data['measurement'] = {
        'preparation': json.loads((base/'preparation.json').read_text()) if (base/'preparation.json').exists() else None,
        'model_process_runs': [json.loads(path.read_text()) for path in sorted(base.glob('run-*.json'))],
        'hook_interval_sum_seconds': round(sum(end-start for start, end in intervals)/1e9, 6),
        'hook_interval_union_seconds': round(sum(end-start for start, end in merged)/1e9, 6),
        'rework': 'Use journal transitions and original responsibility records; no duration inferred from lifecycle alone',
        'external_interference': 'Not automatically attributable; retain SIGINT, preparation failures and external edits separately',
        'boundary': 'Hook intervals overlap and are nested in model process wall time. Do not add these categories, subtract to infer model time, or claim single-variable gains across different scenarios. Provider API duration, when reported, is separate from inference-only time.'}
    dump(base/'summary.json',data)
    for name in ('loop-state.json','loop-events.jsonl','hook-decisions.jsonl'):
        p=root/'.claude'/name
        if p.exists(): (base/('final-'+name)).write_bytes(p.read_bytes())
    print(json.dumps({k:data[k] for k in ('cursors_observed','ordinary_worker_events','native_actor_count','final_lifecycle','fixture_changed','latest_activity')},ensure_ascii=False),flush=True)


def verify(base):
    summarize(base)
    s=json.loads((base/'summary.json').read_text())
    root=base/'project'
    journal=(root/'.claude/loop-events.jsonl').read_text()
    cursors=[tuple(x) for x in s['cursors_observed']]
    def seen(state, phase=None):
        return any(a==state and (phase is None or b==phase) for a,b in cursors)
    stages={
        'S0':seen('inactive') and (root/'docs/requirements/REQ-001.md').exists(),
        'S1':bool(s.get('original_s1_binding_observations')) or 'TR-001' in journal or 'TR001' in journal,
        'S2':seen('planning','design'), 'S3':seen('planning','contracts'),
        'S4':seen('planning','tasks'), 'S5':seen('document_verification'),
        'S6':seen('building'), 'S7':seen('verification'),
        'S8':seen('bug_resolution','investigation'),
        'S9':any(a=='bug_resolution' and b not in (None,'investigation') for a,b in cursors),
        'S10':seen('acceptance') and seen('release_audit'),
        'S11':seen('awaiting_human_release') and seen('release_authorized')}
    checks={'all_native_actors_stopped':all(x['stops']>0 for x in s['actors'].values()),'all_stages_observed':all(stages.values()),'test_terminal':s['final_lifecycle'].get('state')=='release_authorized','native_agents_dispatched':len(s['native_calls'])>=6,'native_stop_observed':sum(x['stops'] for x in s['actors'].values())>=6,'ordinary_worker_nonblocking':s['ordinary_worker_nonzero_exit']==0,'ordinary_worker_stdout_quiet':s['ordinary_worker_nonempty_stdout']==0,'fixture_unchanged':not s['fixture_changed'],'fixture_mode_unchanged':not s['fixture_mode_changed'],'registered_evidence_sources_intact':not s['registered_evidence_source_drift']}
    result={'simulation_only':True,'mechanical_checks_passed':all(checks.values()),'stages':stages,'checks':checks,'worker_activity_needing_review':{'command_signals':s['potential_boundary_violations'],'non_private_write_count':s['worker_other_writes'],'audit_file':'worker-tool-audit.json'},'boundary':'Mechanical checks cover workflow, source and fixture facts only; work scope requires raw-record review. Commands, path categories and heuristic command substrings do not establish product development. This is not product or professional acceptance.'}
    dump(base/'verification.json',result)
    print(json.dumps(result,ensure_ascii=False),flush=True)
    raise SystemExit(0 if result['mechanical_checks_passed'] else 1)


if __name__=='__main__':
    p=argparse.ArgumentParser(description=__doc__)
    p.add_argument('action',choices=['prepare','run','summarize','verify'])
    p.add_argument('--artifacts',type=Path,required=True)
    p.add_argument('--env-file',type=Path,default=Path.home()/'.claude/minimax')
    p.add_argument('--dry-run', action='store_true', help='Validate preparation inputs without building, installing or starting a model')
    p.add_argument('--resume')
    p.add_argument('--driver',action='store_true',help='Small native Main simulation prompt; original native tools and Agents remain available')
    p.add_argument('--prompt-file',type=Path)
    args=p.parse_args()
    base=args.artifacts.resolve()
    if args.dry_run and args.action != 'prepare': p.error('--dry-run applies only to prepare')
    if args.action=='prepare': prepare(base, args.dry_run)
    elif args.action=='run': launch(base,args.env_file.expanduser(),args.resume,args.prompt_file,args.driver)
    elif args.action=='summarize': summarize(base)
    else: verify(base)
