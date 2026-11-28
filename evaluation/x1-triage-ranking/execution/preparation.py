"""X1 immutable evidence. R6-01 uses code-enforced adoption and load-time journal checks."""
import argparse
from contextlib import contextmanager
from datetime import datetime, timedelta, timezone
import fcntl
import json
import os
from pathlib import Path
import subprocess
import sys
import time
import uuid
import copy
import hashlib
from contextvars import ContextVar
import re

# -I deliberately removes the script directory from sys.path.
sys.path.insert(0, str(Path(__file__).resolve().parent))
from verify_inputs import AUTHORITY, MANIFEST, SHARED, exclusive_json, pin, verify


def now():
    return datetime.now(timezone.utc).isoformat().replace('+00:00', 'Z')


def parsed(value):
    date = datetime.fromisoformat(value.replace('Z', '+00:00'))
    if date.utcoffset() != timedelta(0):
        raise ValueError('timestamp is not UTC')
    return date


class RunConfig(dict):
    """Effective values with append-only provenance outside the config key space."""
    amendments = ()


def config_digest(config):
    return hashlib.sha256(json.dumps(config, sort_keys=True, separators=(',', ':')).encode()).hexdigest()


def apply_binding_changes(config, changes):
    allowed = {'binding_record_kind', 'binding_acceptance_keys', 'binding_producer_contract', 'binding_recipe'}
    if not changes or not set(changes) <= allowed:
        raise ValueError('amendment changes outside binding contract')
    if 'binding_record_kind' in changes:
        if changes['binding_record_kind'] != 'scout.x1.binding-acceptance.v2':
            raise ValueError('unsupported binding record kind')
        for record in config['required_records']:
            if record['path'] in ('binding/synthetic.json', 'binding/django.json'):
                record['kind'] = changes['binding_record_kind']
    if 'binding_acceptance_keys' in changes:
        keys = changes['binding_acceptance_keys']
        expected = {'policy_bound_both_revisions', 'deterministic_model_free_proof', 'zero_model_attempts', 'exact_nonempty_issue_id_parity', 'positive_control_expected_failure', 'scanner_health_complete'}
        if not isinstance(keys, list) or len(keys) != len(expected) or set(keys) != expected:
            raise ValueError('unsupported binding acceptance keys')
        config['binding_acceptance_keys'] = keys
    if 'binding_producer_contract' in changes:
        if not isinstance(changes['binding_producer_contract'], str):
            raise ValueError('invalid binding contract')
        config['producer_contracts']['binding'] = changes['binding_producer_contract']
    if 'binding_recipe' in changes:
        if changes['binding_recipe'] != 'preparation/model-free-binding-check-v2.json':
            raise ValueError('unsupported binding recipe')
        config['binding_recipe'] = changes['binding_recipe']


def adoption_paths(config, amendment_path):
    directory = paths(config)['OUT']/'config-adoptions'
    return directory/Path(amendment_path).name, directory/(Path(amendment_path).stem+'.operation.json')


def verify_adoption(config, config_path, amendment_path, review_path):
    receipt_path, command_path = adoption_paths(config, amendment_path)
    receipt = json.loads(receipt_path.read_bytes())
    expected_keys = {'schema', 'adopted_at', 'amendment', 'review', 'command_record'}
    if set(receipt) != expected_keys or receipt['schema'] != 'scout.x1.config-adoption.v1':
        raise ValueError('adopt-command receipt required')
    for key, path in [('amendment', amendment_path), ('review', review_path), ('command_record', command_path)]:
        if receipt[key] != pin(path):
            raise ValueError('adoption receipt '+key+' mismatch')
    command = json.loads(command_path.read_bytes())
    if command.get('schema') != 'scout.x1.config-adoption-operation.v1' or command.get('operation') != 'adopt' or command.get('status') != 'passed':
        raise ValueError('adopt-command operation record required')
    if command.get('adopted_at') != receipt['adopted_at'] or command.get('started_at') != receipt['adopted_at']:
        raise ValueError('adoption timestamp differs from command record')
    if parsed(receipt['adopted_at']) > parsed(now()):
        raise ValueError('future adoption timestamp')
    if command.get('configuration') != pin(config_path) or command.get('writer') != pin(__file__):
        raise ValueError('adoption writer/config pin mismatch')
    if command.get('amendment') != receipt['amendment'] or command.get('review') != receipt['review']:
        raise ValueError('adoption command input mismatch')
    if 'adopt' not in command.get('argv', []):
        raise ValueError('adopt command argv missing')
    verify_tree_pins(command)
    for row in command['prior_journals']:
        record = json.loads(Path(row['path']).read_bytes())
        for key in ('started_at', 'ended_at'):
            if key in record and parsed(record[key]) >= parsed(receipt['adopted_at']):
                raise ValueError('adoption is backdated before an existing journal')
    return receipt


def load_config(path):
    path = Path(path).absolute()
    config = RunConfig(json.loads(path.read_bytes()))
    if config['schema'] != 'scout.x1.run-config.v1':
        raise ValueError('wrong config schema')
    clock_path = paths(config)['OUT']/'clock.json'
    if clock_path.exists():
        clock = json.loads(clock_path.read_bytes())
        original = clock['execution_config']
        if Path(original['path']).absolute() != path:
            raise ValueError('clock original config path differs')
        validate_pin(original)
    prior = pin(path)
    study = paths(config)['STUDY']
    amendments = []
    for amendment_path in sorted((path.parent/'run-config-amendments').glob('*.json')):
        if amendment_path.is_symlink():
            raise ValueError('symlinked config amendment')
        amendment = json.loads(amendment_path.read_bytes())
        if set(amendment) != {'schema', 'amends', 'reason', 'authority', 'changes'} or amendment['schema'] != 'scout.x1.run-config-amendment.v1':
            raise ValueError('invalid config amendment schema')
        if amendment['amends'] != {'path': str(Path(prior['path']).relative_to(study)), 'sha256': prior['sha256']}:
            raise ValueError('config amendment prior digest/path differs')
        authority_ref = amendment['authority']
        if set(authority_ref) != {'path', 'scope'} or Path(authority_ref['path']).is_absolute() or '..' in Path(authority_ref['path']).parts or not authority_ref['scope']:
            raise ValueError('invalid amendment authority reference')
        amendment_pin = pin(amendment_path)
        review_path = study/authority_ref['path']
        approval = review_receipt(review_path, authority_ref['scope'])
        if amendment_pin['sha256'] not in {row['sha256'] for row in approval['reviewed']}:
            raise ValueError('rule-5 review does not bind amendment digest')
        adoption = verify_adoption(config, path, amendment_path, review_path)
        adopted_at = adoption['adopted_at']
        if amendments and parsed(adopted_at) < parsed(amendments[-1]['adopted_at']):
            raise ValueError('amendment adoption order differs from chain order')
        before = config_digest(config)
        apply_binding_changes(config, amendment['changes'])
        prior = pin(amendment_path)
        amendments.append(dict(prior, authority_reference={'path': str(review_path), 'scope': authority_ref['scope']},
                               adoption_review=pin(review_path), authority_record=pin(approval['record']['path']),
                               adoption_receipt=pin(adoption_paths(config, amendment_path)[0]),
                               adoption_command=pin(adoption_paths(config, amendment_path)[1]),
                               adopted_at=adopted_at, prior_effective_sha256=before, effective_sha256=config_digest(config)))
    config.amendments = tuple(amendments)
    for start_path in sorted((paths(config)['OUT']/'journal').glob('*.start.json')):
        verify_journal_configuration(json.loads(start_path.read_bytes()), config, path)
    return config


def effective_configuration(config):
    return {'sha256': config_digest(config), 'value': copy.deepcopy(dict(config)),
            'amendments': copy.deepcopy(list(getattr(config, 'amendments', ()))),
            'adoption_state': {'condition': 'adopted-reviewed-amendments' if getattr(config, 'amendments', ()) else 'no-adopted-amendments',
                               'directory_contents': [Path(row['path']).name for row in getattr(config, 'amendments', ())]}}


def verify_amendment_reviews(config):
    reviews = []
    for amendment in getattr(config, 'amendments', ()):
        validate_pin(amendment)
        validate_pin(amendment['adoption_review'])
        validate_pin(amendment['authority_record'])
        reference = amendment['authority_reference']
        receipt = review_receipt(reference['path'], reference['scope'])
        if amendment['sha256'] not in {row['sha256'] for row in receipt['reviewed']}:
            raise ValueError('rule-5 review does not bind amendment digest')
        adoption = verify_adoption(config, Path(amendment['path']).parent.parent/'run-config.json', amendment['path'], reference['path'])
        if adoption['adopted_at'] != amendment['adopted_at']:
            raise ValueError('amendment adoption timestamp changed')
        reviews.append(dict(pin(receipt['record']['path']), scope=reference['scope']))
    return reviews


def paths(config):
    return {key: Path(value) for key, value in config['paths'].items()}


def authority(config):
    return paths(config)['STUDY']/'reviews/round-9.md'


def clock_value(config, config_path, first_action, timestamp=None):
    timestamp = timestamp or now()
    return dict(schema='scout.x1.preparation-clock.v1', study='X1', preparation_started_at=timestamp,
                gold_freeze_deadline=(parsed(timestamp) + timedelta(seconds=45*86400)).isoformat().replace('+00:00', 'Z'),
                first_action=first_action, operator=os.getlogin() if os.isatty(0) else str(os.getuid()),
                argv=sys.argv, cwd=os.getcwd(), execution_commit=config['execution_commit'],
                authority=pin(authority(config)), manifest=pin(paths(config)['STUDY']/'reviews/rev9-digests.json'),
                execution_config=pin(config_path),
                input_verification=pin(paths(config)['OUT']/'input-verification.json'),
                deferred_submodule_gitlink=config['deferred_to_b_contract']['gitlink'])


def validate_pin(record):
    actual = pin(record['path'])
    if actual['sha256'] != record['sha256'] or ('bytes' in record and actual['bytes'] != record['bytes']):
        raise ValueError('pin mismatch: ' + record['path'])
    return actual


def read_clock(config, config_path, clock):
    expected = paths(config)['OUT']/'clock.json'
    if Path(clock).absolute() != expected.absolute() or expected.is_symlink():
        raise ValueError('noncanonical clock')
    value = json.loads(expected.read_bytes())
    if value['schema'] != 'scout.x1.preparation-clock.v1' or value['study'] != 'X1':
        raise ValueError('wrong clock schema/study')
    if value['execution_commit'] != config['execution_commit']:
        raise ValueError('clock execution commit mismatch')
    if value['first_action'] not in 'abcde' or len(value['first_action']) != 1:
        raise ValueError('invalid first action')
    if parsed(value['gold_freeze_deadline']) - parsed(value['preparation_started_at']) != timedelta(days=45):
        raise ValueError('clock deadline mismatch')
    if parsed(value['preparation_started_at']) > parsed(now()):
        raise ValueError('clock is in the future')
    for field, expected_path, expected_hash in [('authority', authority(config), AUTHORITY), ('manifest', paths(config)['STUDY']/'reviews/rev9-digests.json', MANIFEST), ('execution_config', Path(config_path), pin(config_path)['sha256'])]:
        if Path(value[field]['path']).absolute() != expected_path.absolute() or value[field]['sha256'] != expected_hash:
            raise ValueError('clock ' + field + ' mismatch')
        validate_pin(value[field])
    receipt_path = paths(config)['OUT']/'input-verification.json'
    if Path(value['input_verification']['path']).absolute() != receipt_path.absolute():
        raise ValueError('clock input receipt path mismatch')
    validate_pin(value['input_verification'])
    if value['deferred_submodule_gitlink'] != config['deferred_to_b_contract']['gitlink']:
        raise ValueError('clock submodule gitlink mismatch')
    return value


_held_operation_locks = ContextVar('x1_operation_locks', default={})


@contextmanager
def operation_lock(config, exclusive=False, blocking=False, read_only=False):
    out = paths(config)['OUT']
    key = str(out.absolute())
    held = _held_operation_locks.get()
    if key in held:
        if exclusive and not held[key]:
            raise ValueError('cannot upgrade a shared operation lock')
        yield
        return
    if not read_only:
        out.mkdir(parents=True, exist_ok=True)
    with (out/'.operation.lock').open('r' if read_only else 'a') as stream:
        fcntl.flock(stream, (fcntl.LOCK_EX if exclusive else fcntl.LOCK_SH) | (0 if blocking else fcntl.LOCK_NB))
        token = _held_operation_locks.set(dict(held, **{key: exclusive}))
        try:
            if (out/'preparation.seal.json').exists():
                raise ValueError('seal already reserved; reviewed recovery required')
            yield
        finally:
            _held_operation_locks.reset(token)


def adopt(config_path, proposed_path, review_path):
    """R6-01 code-enforced choice: no caller-supplied adoption timestamp."""
    config_path = Path(config_path).absolute()
    proposed_path, review_path = Path(proposed_path).absolute(), Path(review_path).absolute()
    base = RunConfig(json.loads(config_path.read_bytes()))
    with operation_lock(base, exclusive=True, blocking=True):
        config = load_config(config_path)
        p = paths(config)
        read_clock(config, config_path, p['OUT']/'clock.json')
        if proposed_path.parent != config_path.parent/'run-config-amendments-proposed' or proposed_path.suffix != '.json' or proposed_path.is_symlink():
            raise ValueError('canonical proposed amendment path required')
        amendment = json.loads(proposed_path.read_bytes())
        if set(amendment) != {'schema', 'amends', 'reason', 'authority', 'changes'} or amendment['schema'] != 'scout.x1.run-config-amendment.v1':
            raise ValueError('invalid config amendment schema')
        previous = pin(config.amendments[-1]['path']) if config.amendments else pin(config_path)
        if amendment['amends'] != {'path': str(Path(previous['path']).relative_to(p['STUDY'])), 'sha256': previous['sha256']}:
            raise ValueError('config amendment prior digest/path differs')
        reference = amendment['authority']
        if set(reference) != {'path', 'scope'} or (p['STUDY']/reference['path']).absolute() != review_path:
            raise ValueError('review path does not match amendment authority')
        approval = review_receipt(review_path, reference['scope'])
        if pin(proposed_path)['sha256'] not in {row['sha256'] for row in approval['reviewed']}:
            raise ValueError('rule-5 review does not bind proposed amendment')
        apply_binding_changes(copy.deepcopy(config), amendment['changes'])
        directory = config_path.parent/'run-config-amendments'
        if config.amendments and proposed_path.name <= Path(config.amendments[-1]['path']).name:
            raise ValueError('amendments must be appended in filename order')
        starts = sorted((p['OUT']/'journal').glob('*.start.json'))
        ends = sorted(path for path in (p['OUT']/'journal').glob('*.json') if not path.name.endswith('.start.json'))
        if {path.name.removesuffix('.start.json') for path in starts} != {path.stem for path in ends}:
            raise ValueError('cannot adopt during a nonterminal operation')
        prior_journals = [pin(path) for path in starts + ends]
        adopted_at = now()
        for row in prior_journals:
            record = json.loads(Path(row['path']).read_bytes())
            for key in ('started_at', 'ended_at'):
                if key in record and parsed(record[key]) >= parsed(adopted_at):
                    raise ValueError('current clock is not later than existing journals')
        destination = directory/proposed_path.name
        receipt_path, command_path = adoption_paths(config, destination)
        if destination.exists() or receipt_path.exists() or command_path.exists():
            raise ValueError('adoption already reserved; reviewed recovery required')
        directory.mkdir(parents=True, exist_ok=True)
        data = proposed_path.read_bytes()
        with destination.open('xb') as stream:
            stream.write(data)
            stream.flush()
            os.fsync(stream.fileno())
        fd = os.open(directory, os.O_RDONLY)
        try:
            os.fsync(fd)
        finally:
            os.close(fd)
        command = {'schema': 'scout.x1.config-adoption-operation.v1', 'operation': 'adopt', 'status': 'passed',
                   'started_at': adopted_at, 'adopted_at': adopted_at, 'argv': sys.argv, 'cwd': os.getcwd(),
                   'writer': pin(__file__), 'configuration': pin(config_path), 'clock': pin(p['OUT']/'clock.json'),
                   'amendment': pin(destination), 'proposal': pin(proposed_path), 'review': pin(review_path),
                   'prior_journals': prior_journals}
        exclusive_json(command_path, command)
        exclusive_json(receipt_path, {'schema': 'scout.x1.config-adoption.v1', 'adopted_at': adopted_at,
                                     'amendment': pin(destination), 'review': pin(review_path), 'command_record': pin(command_path)})
        load_config(config_path)
        return pin(receipt_path)


def producer_pins():
    execution = Path(__file__).resolve().parent
    entry = Path(sys.argv[0]).resolve()
    if not entry.is_file():
        entry = Path(sys.modules['__main__'].__file__).resolve()
    files = {entry}
    for module in tuple(sys.modules.values()):
        filename = getattr(module, '__file__', None)
        if filename:
            path = Path(filename).resolve()
            if path.is_relative_to(execution) and path.is_file():
                files.add(path)
    return [pin(path) for path in sorted(files)]


@contextmanager
def journal_publication(directory):
    """Serialize snapshots and publication without serializing operation bodies."""
    directory.mkdir(parents=True, exist_ok=True)
    descriptor = os.open(directory, os.O_RDONLY | os.O_DIRECTORY)
    try:
        fcntl.flock(descriptor, fcntl.LOCK_EX)
        yield
    finally:
        os.close(descriptor)


class Journal:
    """One immutable start and one immutable terminal record per operation."""
    def __init__(self, config, config_path, clock, action, source_pins=(), environment=None):
        self.config, self.config_path, self.clock_path = config, config_path, Path(clock)
        with operation_lock(config, blocking=True):
            self.clock = read_clock(config, config_path, clock)
            self.directory = paths(config)['OUT']/'journal'
            with journal_publication(self.directory):
                if isinstance(config, RunConfig) and effective_configuration(load_config(config_path)) != effective_configuration(config):
                    raise ValueError('adopted config changed before journal reservation')
                self.id = uuid.uuid4().hex
                self.started_at, self.monotonic = now(), time.monotonic()
                self.base = dict(schema='scout.x1.operation.v1', id=self.id, action=action,
                                 started_at=self.started_at, argv=sys.argv, cwd=os.getcwd(),
                                 environment=environment or {}, configuration=pin(config_path), clock=pin(clock),
                                 source_pins=list(source_pins), effective_configuration=effective_configuration(config),
                                 producer_pins=producer_pins())
                self.start_path = self.directory/(self.id+'.start.json')
                self.end_path = self.directory/(self.id+'.json')
                self.base['prior_journals'] = [pin(path) for path in sorted(self.directory.glob('*.json'))]
                exclusive_json(self.start_path, dict(self.base, status='started'))

    def finish(self, status, outputs=(), reason=None, exit_code=0):
        with operation_lock(self.config, blocking=True):
            record = dict(self.base, ended_at=now(), monotonic_duration_seconds=time.monotonic()-self.monotonic,
                          elapsed_since_clock_seconds=(parsed(now())-parsed(self.clock['preparation_started_at'])).total_seconds(),
                          output_pins=list(outputs), status=status, exit_code=exit_code, failure_reason=reason,
                          start_record=pin(self.start_path))
            with journal_publication(self.directory):
                exclusive_json(self.end_path, record)
            return pin(self.end_path)


def documentary_gate(config):
    p = paths(config)
    receipt = verify(p['ROOT'], p['STUDY']/'reviews/rev9-digests.json', authority(config))
    if receipt['status'] != 'passed':
        raise ValueError('canonical documentary inputs fail: ' + json.dumps(receipt['failures']))
    return receipt


def start(config, config_path, action):
    current_gate = documentary_gate(config)
    p = paths(config)
    actual = subprocess.check_output(['git', '-C', str(p['ROOT']), 'rev-parse', 'HEAD'], text=True).strip()
    if actual != config['execution_commit']:
        raise ValueError('execution HEAD differs')
    gate = json.loads((p['OUT']/'input-verification.json').read_bytes())
    if gate['status'] != 'passed':
        raise ValueError('passing input receipt required')
    if gate.get('deferred_to_b', []) != current_gate.get('deferred_to_b', []):
        raise ValueError('input receipt deferrals differ from canonical gate')
    verify_input_receipt(gate)
    with operation_lock(config, exclusive=True):
        exclusive_json(p['OUT']/'clock.json', clock_value(config, config_path, action))


def verify_input_receipt(value, root=None):
    """Shared whole-file observations and deferred rows are not immutable pins."""
    root = Path(root) if root is not None else Path(next(iter(value.get('path_mapping', {}).values()), Path(__file__).resolve().parents[3]))
    shared = {str((root/path).absolute()) for path in SHARED}
    def immutable(child):
        if isinstance(child, dict):
            if 'path' in child and 'sha256' in child and str(Path(child['path']).absolute()) in shared:
                return None
            return {key: immutable(item) for key, item in child.items()}
        if isinstance(child, list):
            return [immutable(item) for item in child]
        return child
    verify_tree_pins(immutable({key: child for key, child in value.items() if key != 'deferred_to_b'}))


def verify_tree_pins(value):
    """Verify explicit path/hash records recursively; never trust aggregate pass flags."""
    found = []
    if isinstance(value, dict):
        if 'path' in value and 'sha256' in value:
            found.append(validate_pin(value))
        for child in value.values():
            found.extend(verify_tree_pins(child))
    elif isinstance(value, list):
        for child in value:
            found.extend(verify_tree_pins(child))
    return found


def review_receipt(path, scope, study=None):
    value = json.loads(Path(path).read_bytes())
    if set(value) != {'scope', 'record', 'reviewed'} or value['scope'] != scope or not value['reviewed']:
        raise ValueError('invalid scoped review receipt: ' + str(path))
    validate_pin(value['record'])
    receipt_path = Path(path).absolute()
    run = next((parent for parent in receipt_path.parents if parent.name == 'preparation-run'), None)
    study = Path(study) if study is not None else (run.parent if run else Path(__file__).resolve().parent.parent)
    record_path = Path(value['record']['path']).resolve()
    if not any(record_path.is_relative_to((study/directory).resolve()) for directory in ('reviews', 'preparation-reviews')):
        raise ValueError('authority record outside study review directories')
    text = record_path.read_text()
    if '**Zero issues. Signed off.**' not in text.splitlines():
        raise ValueError('exact zero-issue verdict line required')
    scopes = list(re.finditer(r'^Signed-off scope: ([^\n]+)$', text, re.M))
    matches = [index for index, match in enumerate(scopes) if match.group(1) == scope]
    if len(matches) != 1:
        raise ValueError('signed-off scope missing or ambiguous')
    index = matches[0]
    block = text[scopes[index].end():scopes[index+1].start() if index+1 < len(scopes) else len(text)]
    for record in value['reviewed']:
        validate_pin(record)
        if not re.search(r'(?<![0-9a-f])'+re.escape(record['sha256'])+r'(?![0-9a-f])', block):
            raise ValueError('scope block does not name reviewed digest')
    return value


def verify_journal_configuration(record, config, config_path):
    """Check the exact chain adopted at operation start, not the later seal time."""
    canonical = list(getattr(config, 'amendments', ()))
    expected_amendments = [row for row in canonical if parsed(row['adopted_at']) <= parsed(record['started_at'])]
    observed = record.get('effective_configuration')
    if observed is None:
        if expected_amendments:
            raise ValueError('post-adoption journal lacks effective config')
        return
    if observed['amendments'] != expected_amendments:
        raise ValueError('journal amendment chain/adoption ordering differs')
    expected_digest = expected_amendments[-1]['effective_sha256'] if expected_amendments else config_digest(json.loads(Path(config_path).read_bytes()))
    if observed['sha256'] != expected_digest or config_digest(observed['value']) != expected_digest:
        raise ValueError('journal effective config digest differs')
    expected_state = {'condition': 'adopted-reviewed-amendments' if expected_amendments else 'no-adopted-amendments',
                      'directory_contents': [Path(row['path']).name for row in expected_amendments]}
    if observed.get('adoption_state') != expected_state:
        raise ValueError('journal adoption precondition/directory listing differs')


def verify_deferred_rows(config, input_receipt, closure):
    from verify_inputs import working_pin
    rows = input_receipt.get('deferred_to_b', [])
    observed = closure.get('deferred_working_files_verified', [])
    if len(observed) != len(rows):
        raise ValueError('closure has outstanding deferred working-file rows')
    expected = {str(paths(config)['ROOT']/row['path']): row for row in rows}
    if len(expected) != len(rows) or {row['path'] for row in observed} != set(expected):
        raise ValueError('closure deferred working-file row set differs')
    for row in observed:
        actual = working_pin(paths(config)['ROOT'], expected[row['path']])
        if row['sha256'] != actual['sha256'] or row['bytes'] != actual['bytes']:
            raise ValueError('closure deferred working-file digest differs')


def verify_output_accounting(out, canonical_pins, journal_output_pins=(), reserved=()):
    referenced = {str(Path(row['path']).absolute()) for row in canonical_pins}
    journal_outputs = []
    for row in verify_tree_pins(list(journal_output_pins)):
        path = str(Path(row['path']).absolute())
        if path not in referenced:
            journal_outputs.append(dict(row, role='journal-output'))
            referenced.add(path)
    referenced.update(str(Path(path).absolute()) for path in reserved)
    for path in Path(out).rglob('*'):
        if path.is_symlink():
            raise ValueError('symlink in sealed run output')
        if path.is_file() and path.name not in ('.operation.lock', 'preparation.seal.json') and str(path.absolute()) not in referenced:
            raise ValueError('unaccounted run output: ' + str(path))
    return journal_outputs


def verify_scoped_evidence(config, reviewed):
    p = paths(config)
    required = {
        'execution-implementation': [config['source_manifest'], 'closure.json'],
        'resolved-development-boundaries': [],
        'binding-acceptance': ['binding/synthetic.json', 'binding/django.json'],
        'identifier-completeness-delta': [row['path'] for row in config['required_records'] if row['path'].startswith('identifiers/')] + list(config['required_files']),
    }
    for repo in config['repositories']:
        scope = 'resolved-development-boundaries' if repo.get('role') == 'development' else 'execution-implementation'
        required[scope].append('objects/'+repo['repository'].replace('/', '__')+'.json')
    for scope, names in required.items():
        if not {pin(p['OUT']/name)['sha256'] for name in names} <= reviewed.get(scope, set()):
            raise ValueError('scoped reviews do not bind all required evidence: '+scope)


def collect_cc2_delta(config, config_path, clock_path, passed):
    """Require the append-only delta and both of its separately scoped reviews."""
    p = paths(config)
    implementation_scope = 'identifier-delta-cc2-implementation'
    output_scope = 'identifier-delta-cc2-output'
    review_path = p['STUDY']/'preparation-reviews/identifiers-delta-cc2-implementation.json'
    implementation = review_receipt(review_path, implementation_scope, study=p['STUDY'])
    code = [pin(p['EXEC']/name) for name in (
        'identifiers/cc2_delta.py', 'identifiers/cc2_readers.py', 'identifiers/cc2_schemas.py',
        'tests/test_cc2_delta.py', 'identifiers/projection.py', 'preparation.py', 'verify_inputs.py')]
    code.append(pin(p['STUDY']/'preparation-reviews/identifiers-delta-cc2-REQUEST.md'))
    source = pin(p['STUDY']/'preparation/exposure-sources.json')
    baseline_path = p['OUT']/'journal/906389d7212a401b8c7f781d37960d84.json'
    baseline = json.loads(baseline_path.read_bytes())
    prior_outputs = [pin(p['OUT']/'identifiers'/name) for name in (
        'manifest.json', 'source-coverage.json', 'delta-review.json', 'unresolved.json', 'familiarity.jsonl')]
    if baseline.get('action') != 'd' or baseline.get('status') != 'passed' or baseline.get('exit_code') != 0 or baseline.get('output_pins') != prior_outputs:
        raise ValueError('CC2 baseline identifier journal/output mismatch')
    prior = [pin(baseline_path), *prior_outputs]
    controls = code + [source, pin(config_path)] + prior
    if not {(row['path'], row['sha256']) for row in controls} <= {(row['path'], row['sha256']) for row in implementation['reviewed']}:
        raise ValueError('CC2 implementation review omits bound controls')
    output_specs = (
        ('manifest.json', 'scout.x1.identifier-delta-cc2.v1'),
        ('coverage.json', 'scout.x1.identifier-delta-cc2-coverage.v1'),
        ('familiarity.jsonl', 'scout.x1.familiarity-delta-cc2.v1'))
    outputs, values, sources = [], [], verify_tree_pins(implementation)
    for name, kind in output_specs:
        path = p['OUT']/'identifiers-delta-cc2'/name
        with path.open(encoding='utf-8') as stream:
            value = json.loads(stream.readline()) if name.endswith('.jsonl') else json.load(stream)
        if value.get('schema') != kind or value.get('status') != 'passed' or value.get('clock') != pin(clock_path):
            raise ValueError('CC2 delta schema/status/clock mismatch: '+name)
        if value.get('failures') or value.get('unmeasured') or value.get('development_freshness_filter') is not False:
            raise ValueError('CC2 delta failed/unmeasured/freshness evidence: '+name)
        if value.get('implementation_review') != pin(review_path) or value.get('sources_manifest') != source or value.get('predecessors') != prior:
            raise ValueError('CC2 delta implementation/source/baseline chain mismatch: '+name)
        sources.extend(verify_tree_pins(value))
        outputs.append(pin(path, kind))
        values.append(value)
    if values[0].get('merge_semantics') != 'union_with_frozen_v4_and_existing_X1_identifiers_never_replace' or values[0].get('unknown_freshness') != 'ineligible' or values[0].get('candidate_exclusions_computed') is not False:
        raise ValueError('CC2 delta must preserve conservative append-union semantics')
    if values[2].get('entry_kind') != 'source_table' or values[2].get('sources') != values[0].get('sources'):
        raise ValueError('CC2 familiarity source-table mismatch')
    output_pins = [pin(row['path']) for row in outputs]
    producers = [(path, record) for path, record in passed.items()
                 if record.get('action') == 'd' and record.get('output_pins') == output_pins]
    if len(producers) != 1:
        raise ValueError('CC2 delta lacks unique passed action-d producer')
    producer_path, producer = producers[0]
    if not {(row['path'], row['sha256']) for row in code + [source]} <= {(row['path'], row['sha256']) for row in producer.get('source_pins', [])}:
        raise ValueError('CC2 delta producer omits implementation/source pins')
    output_review_path = p['OUT']/'reviews/identifier-delta-cc2-output.json'
    output_review = review_receipt(output_review_path, output_scope, study=p['STUDY'])
    required = output_pins + [pin(producer_path)]
    if not {(row['path'], row['sha256']) for row in required} <= {(row['path'], row['sha256']) for row in output_review['reviewed']}:
        raise ValueError('CC2 output review omits delta outputs/producer journal')
    sources.extend(verify_tree_pins(output_review))
    outputs.extend([pin(review_path, 'scoped-review'), pin(output_review_path, 'scoped-review')])
    reviews = [dict(pin(value['record']['path']), scope=scope) for value, scope in
               ((implementation, implementation_scope), (output_review, output_scope))]
    return sources, outputs, reviews


def collect_journals(config, config_path, clock_path, clock, pending_seal=None):
    out = paths(config)['OUT']
    operations, outputs, output_pins, deviations, passed = [], [], [], [], {}
    starts = sorted(path for path in (out/'journal').glob('*.start.json') if path.name != str(pending_seal)+'.start.json')
    ends = sorted(path for path in (out/'journal').glob('*.json') if not path.name.endswith('.start.json'))
    if not starts or {path.name.removesuffix('.start.json') for path in starts} != {path.stem for path in ends}:
        raise ValueError('missing/nonterminal operation records')
    for path in ends:
        value = json.loads(path.read_bytes())
        if value['id'] != path.stem:
            raise ValueError('operation id/stem mismatch')
        if value['clock'] != pin(clock_path):
            raise ValueError('stale operation clock')
        if parsed(value['started_at']) < parsed(clock['preparation_started_at']) or parsed(value['ended_at']) < parsed(value['started_at']):
            raise ValueError('invalid operation timeline')
        start_path = out/'journal'/(value['id']+'.start.json')
        if Path(value['start_record']['path']) != start_path:
            raise ValueError('noncanonical operation start path')
        start_record = json.loads(start_path.read_bytes())
        if any(start_record.get(key) != value.get(key) for key in ('id', 'action', 'started_at', 'argv', 'cwd', 'configuration', 'clock', 'source_pins', 'environment', 'effective_configuration', 'producer_pins', 'prior_journals')):
            raise ValueError('terminal/start operation record mismatch')
        verify_journal_configuration(value, config, config_path)
        verify_tree_pins(value)
        output_pins.extend(value.get('output_pins', []))
        outputs.extend([pin(path, 'journal'), pin(start_path, 'journal-start')])
        if value['status'] != 'passed' or value['exit_code'] != 0:
            recovery_path = out/'reviews'/('operation-recovery-'+value['id']+'.json')
            if not recovery_path.is_file():
                raise ValueError('failed operation lacks operation-recovery review')
            recovery = review_receipt(recovery_path, 'operation-recovery')
            required = [pin(start_path), pin(path)] + verify_tree_pins(value.get('output_pins', []))
            if not {row['sha256'] for row in required} <= {row['sha256'] for row in recovery['reviewed']}:
                raise ValueError('operation-recovery review omits operation evidence')
            outputs.extend([pin(recovery_path, 'operation-recovery'), pin(recovery['record']['path'], 'operation-recovery-authority')])
            deviations.extend(dict(pin(item), reason='reviewed operation recovery') for item in (start_path, path, recovery_path, recovery['record']['path']))
        else:
            operations.append({'id': value['id'], 'action': value['action'], 'record': pin(path)})
            passed[str(path)] = value
            if not value.get('producer_pins'):
                deviations.append(dict(pin(path), reason='producer bytes unrecorded'))
    if not set('abcd') <= {row['action'] for row in operations}:
        raise ValueError('required action journals missing')
    return operations, outputs, output_pins, deviations, passed


def collect_seal(config, config_path, clock_path, pending_seal=None):
    p = paths(config)
    if 'OPENROUTER_API_KEY' in os.environ:
        raise ValueError('key present; preparation seal refused')
    clock = read_clock(config, config_path, clock_path)
    amendment_reviews = verify_amendment_reviews(config)
    gate = documentary_gate(config)
    outputs, sources = [], list(gate['pins'])
    for amendment in getattr(config, 'amendments', ()):
        sources.append(pin(amendment['path'], 'config-amendment'))
        reference = amendment['authority_reference']
        outputs.append(pin(reference['path'], 'config-amendment-review'))
        outputs.extend([pin(amendment['adoption_receipt']['path'], 'config-adoption'),
                        pin(amendment['adoption_command']['path'], 'config-adoption-operation')])
    sources.extend(dict(row, role='config-amendment-authority') for row in amendment_reviews)
    inputs = json.loads((p['OUT']/'input-verification.json').read_bytes())
    if inputs['status'] != 'passed':
        raise ValueError('input receipt failed')
    verify_input_receipt(inputs, p['ROOT'])
    if gate.get('deferred_to_b'):
        raise ValueError('uninitialised submodule at seal; action b has not completed')
    outputs.append(pin(p['OUT']/'input-verification.json', 'input-verification'))
    records = {}
    for spec in config['required_records']:
        path = p['OUT']/spec['path']
        value = json.loads(path.read_bytes())
        if value['schema'] != spec['kind'] or value.get('status') != 'passed':
            raise ValueError('missing/passing record schema: ' + str(path))
        if value.get('key_present', False) or value.get('unmeasured', []) or value.get('failures', []):
            raise ValueError('failed/unmeasured/present-key evidence: ' + str(path))
        if value.get('clock') != pin(clock_path):
            raise ValueError('record clock mismatch: ' + str(path))
        sources.extend(verify_tree_pins(value))
        outputs.append(pin(path, spec['kind']))
        records[spec['path']] = value
    for repo in config['repositories']:
        record = records['objects/'+repo['repository'].replace('/', '__')+'.json']
        if record['expected_tip'] != repo['tip'] or record['actual_tip'] != repo['tip'] or (repo['c0'] and record['actual_c0'] != repo['c0']) or not record['first_parent_membership']:
            raise ValueError('substituted object boundary')
        if not all(record[key] for key in ('no_shallow', 'no_promisor', 'no_alternates', 'no_replacements')):
            raise ValueError('unsafe object store')
        from fetch_objects import Git, inspect_store, first_parent_boundary
        fetch_journal = json.loads(Path(record['journal']['path']).read_bytes())
        git = Git(record['git_identity']['path'], fetch_journal['environment']['HOME'])
        inspect_store(git, Path(record['database_realpath']), repo)
        boundary, derivation, edges, tip_header, c0_header = first_parent_boundary(git, Path(record['database_realpath']), repo['tip'], repo['c0'])
        if (boundary, derivation, edges, tip_header, c0_header) != (record['actual_c0'], record['c0_derivation'], record['checked_edge_count'], record['tip_header'], record['c0_header']):
            raise ValueError('retained object headers/boundaries changed')
        git.run(Path(record['database_realpath']), 'fsck', '--full', '--strict', '--no-reflogs', repo['tip'], boundary)
    for fixture in ('synthetic', 'django'):
        record = records['binding/'+fixture+'.json']
        if record.get('key_present') is not False:
            raise ValueError('explicit absent-key binding evidence required')
        if record.get('runtime_closure') != pin(p['OUT']/'closure.json'):
            raise ValueError('binding runtime closure mismatch')
        expected_keys = config.get('binding_acceptance_keys', ['policy_bound_both_revisions', 'complete_process_observation', 'zero_model_attempts', 'exact_nonempty_issue_id_parity', 'positive_control_expected_failure', 'scanner_health_complete'])
        if record.get('acceptance') != dict.fromkeys(expected_keys, True):
            raise ValueError('binding acceptance incomplete')
    django = records['objects/django__django.json']
    if records['binding/django.json'].get('revisions') != {'head': django['actual_tip'], 'parent': django['django_first_parent']['sha']}:
        raise ValueError('Django binding pair substituted')
    from attest_closure import verify_installed_closure
    verify_installed_closure(p['OUT']/'closure.json')
    if not records['closure.json'].get('retained_files'):
        raise ValueError('runtime closure needs absolute retained-file pins for seal recheck')
    verify_deferred_rows(config, inputs, records['closure.json'])
    for name in config['required_files']:
        outputs.append(pin(p['OUT']/name, 'identifier-output'))
    manifest_path = p['OUT']/config['source_manifest']
    source_manifest = json.loads(manifest_path.read_bytes())
    expected_sources = {str(x.absolute()) for x in p['EXEC'].rglob('*') if x.is_file() and '__pycache__' not in x.parts}
    if {x['path'] for x in source_manifest['files']} != expected_sources:
        raise ValueError('execution source closure incomplete or stale')
    sources.extend(verify_tree_pins(source_manifest))
    outputs.append(pin(manifest_path, 'execution-source-closure'))
    reviews = []
    reviewed = {}
    for spec in config['required_reviews']:
        path = p['OUT']/spec['path']
        value = review_receipt(path, spec['scope'], study=p['STUDY'])
        reviewed.setdefault(spec['scope'], set()).update(row['sha256'] for row in value['reviewed'])
        reviews.append(dict(pin(value['record']['path']), scope=spec['scope']))
        outputs.append(pin(path, 'scoped-review'))
    operations, journal_outputs, journal_output_pins, deviations, passed = collect_journals(config, config_path, clock_path, clock, pending_seal)
    outputs.extend(journal_outputs)
    passed_outputs = {(row['path'], row['sha256']) for operation in passed.values() for row in verify_tree_pins(operation.get('output_pins', []))}
    passed_starts = {operation['start_record']['path'] for operation in passed.values()}
    for name, record in records.items():
        actual = pin(p['OUT']/name)
        if record.get('journal', {}).get('path') not in passed and record.get('journal_start', {}).get('path') not in passed_starts and (actual['path'], actual['sha256']) not in passed_outputs:
            raise ValueError('required record lacks passed producer operation: '+name)
    cc2_sources, cc2_outputs, cc2_reviews = collect_cc2_delta(config, config_path, clock_path, passed)
    sources.extend(cc2_sources)
    outputs.extend(cc2_outputs)
    reviews.extend(cc2_reviews)
    verify_scoped_evidence(config, reviewed)
    reserved = [Path(clock_path), Path(config_path)]
    if pending_seal:
        reserved.append(p['OUT']/'journal'/(pending_seal+'.start.json'))
    outputs.extend(verify_output_accounting(p['OUT'], sources + outputs, journal_output_pins, reserved))
    sources = [dict(x, role=x.get('role', 'canonical-input')) for x in sources]
    stamp = now()
    return dict(schema='scout.x1.preparation-seal.v1', stage='preparation', created_at=stamp, status='sealed',
                authority=pin(authority(config)), scoped_reviews=reviews + amendment_reviews,
                predecessors=[dict(x, kind='signed-input') for x in gate['pins']], clock=pin(clock_path),
                execution_commit=config['execution_commit'], execution_root=str(p['ROOT']), historical_path_mapping=config['historical_path_mapping'],
                argv=sys.argv, cwd=os.getcwd(), config=dict(pin(config_path), effective_configuration=effective_configuration(config)), sources=sources, outputs=outputs,
                operations=operations, started_at=clock['preparation_started_at'], ended_at=stamp,
                elapsed_since_first_action_seconds=(parsed(stamp)-parsed(clock['preparation_started_at'])).total_seconds(), refusals_or_deviations=deviations)


def seal(config, config_path, clock_path, check=False):
    with operation_lock(config, exclusive=True, read_only=check):
        if check:
            return collect_seal(config, config_path, clock_path)
        # Reserve before any validation. Failure leaves the canonical reservation intact.
        destination = paths(config)['OUT']/'preparation.seal.json'
        with destination.open('x') as stream:
            stream.flush()
            os.fsync(stream.fileno())
            fd = os.open(destination.parent, os.O_RDONLY)
            try:
                os.fsync(fd)
            finally:
                os.close(fd)
            journal = Journal(config, config_path, clock_path, 'e', source_pins=[pin(config_path)])
            try:
                value = collect_seal(config, config_path, clock_path, pending_seal=journal.id)
            except BaseException as exc:
                journal.finish('refused', reason=str(exc), exit_code=1)
                raise
            # The seal-validation operation closes before its immutable record is
            # hashed into the seal. Reservation remains the crash/final-write guard.
            terminal = journal.finish('passed')
            value['outputs'].extend([dict(terminal, role='journal'), pin(journal.start_path, 'journal-start')])
            value['operations'].append({'id': journal.id, 'action': 'e', 'record': terminal})
            value['ended_at'] = now()
            value['elapsed_since_first_action_seconds'] = (parsed(value['ended_at'])-parsed(value['started_at'])).total_seconds()
            for row in value['outputs'] + value['sources'] + [value['clock'], value['config']]:
                path = Path(row['path'])
                # Only freeze execution/run evidence; source/product/runtime files remain untouched.
                if path.is_relative_to(paths(config)['OUT']) or path.is_relative_to(paths(config)['EXEC']):
                    mode = 0o555 if path.stat().st_mode & 0o111 else 0o444
                    path.chmod(mode)
                    row['mode'] = oct(mode)
            json.dump(value, stream, sort_keys=True, indent=2)
            stream.write('\n')
            stream.flush()
            os.fsync(stream.fileno())
        destination.chmod(0o444)
        fd = os.open(destination.parent, os.O_RDONLY)
        try:
            os.fsync(fd)
        finally:
            os.close(fd)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    commands = parser.add_subparsers(dest='command', required=True)
    adoption = commands.add_parser('adopt')
    adoption.add_argument('--config', type=Path, default=Path(__file__).resolve().parent/'run-config.json')
    adoption.add_argument('--amendment', type=Path, required=True)
    adoption.add_argument('--review', type=Path, required=True)
    for name in ('start', 'seal', 'journal'):
        cmd = commands.add_parser(name)
        cmd.add_argument('--config', required=True, type=Path)
        if name == 'start':
            cmd.add_argument('--first-action', required=True, choices=list('abcde'))
        else:
            cmd.add_argument('--clock', required=True, type=Path)
        if name == 'journal':
            cmd.add_argument('--action', choices=list('abcde'), required=True)
            cmd.add_argument('argv', nargs=argparse.REMAINDER)
        if name == 'seal':
            cmd.add_argument('--check', action='store_true')
    args = parser.parse_args()
    if args.command == 'adopt':
        print(json.dumps(adopt(args.config, args.amendment, args.review), sort_keys=True))
        return 0
    config = load_config(args.config)
    if args.command == 'start':
        start(config, args.config, args.first_action)
    elif args.command == 'seal':
        seal(config, args.config, args.clock, check=args.check)
    else:
        if not args.argv:
            raise ValueError('journal requires command argv')
        with operation_lock(config):
            env = {'PATH': '/usr/bin:/bin', 'LANG': 'C', 'LC_ALL': 'C'}
            journal = Journal(config, args.config, args.clock, args.action, environment=env)
            try:
                # Clean allowlist, no inherited credentials/config.
                code = subprocess.run(args.argv, env=env, check=False).returncode
                journal.finish('passed' if code == 0 else 'failed', exit_code=code)
                return code
            except BaseException as exc:
                journal.finish('refused', reason=str(exc), exit_code=1)
                raise
    return 0


if __name__ == '__main__':
    try:
        sys.exit(main())
    except (OSError, ValueError, KeyError, subprocess.CalledProcessError) as exc:
        print('REFUSED: ' + str(exc), file=sys.stderr)
        sys.exit(1)
