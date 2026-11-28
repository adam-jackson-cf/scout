"""Deterministic audited-seam model-attempt logs and exact evidence comparison.

Private failing stubs record attempts at the audited provider seams, not process
coverage. Registered scanners retain their own product sandbox.
"""
from pathlib import Path


class InterlockFailure(RuntimeError):
    pass


def stub_source(python, log):
    if '\n' in str(python) or ' ' in str(python):
        raise InterlockFailure('stub shebang path unsupported')
    return ('#!'+str(python)+' -ISB\n'+
            'import json, os, sys\n'+
            'row={"argv":sys.argv,"pid":os.getpid(),"ppid":os.getppid(),"kind":"model-stub-attempt","exit_code":97}\n'+
            'with open('+repr(str(log))+',"a") as f:\n'+
            ' f.write(json.dumps(row,separators=(",",":"))+"\\n"); f.flush(); os.fsync(f.fileno())\n'+
            'sys.exit(97)\n')


def make_stubs(directory, python):
    directory = Path(directory); directory.mkdir(parents=True, mode=0o700, exist_ok=False)
    log = directory/'attempts.jsonl'
    with log.open('x'):
        log.chmod(0o600)
    source = stub_source(python, log)
    for name in ('claude', 'codex'):
        target = directory/name
        target.write_text(source); target.chmod(0o700)
    return log


def evaluate(attempts, positive=False, expected_provider=None):
    attempted = list(attempts)
    if positive:
        if expected_provider not in ('claude', 'codex') or not attempted:
            raise InterlockFailure('positive control did not reach a production-selected logging stub')
        prefix = ['-p','--output-format','json'] if expected_provider == 'claude' else ['exec','--sandbox','read-only']
        for row in attempted:
            if not isinstance(row, dict):
                raise InterlockFailure('positive control attempt is not a structured stub log')
            argv = row.get('argv', [])
            if (row.get('kind') != 'model-stub-attempt' or row.get('exit_code') != 97
                    or not isinstance(row.get('pid'), int) or row['pid'] <= 0
                    or not isinstance(row.get('ppid'), int) or row['ppid'] <= 0
                    or not isinstance(argv, list) or not all(isinstance(x, str) for x in argv)
                    or not argv or Path(argv[0]).name != expected_provider
                    or argv[1:1+len(prefix)] != prefix):
                raise InterlockFailure('positive control lacks production provider argv and failing stub evidence')
        try:
            evaluate(attempted)
        except InterlockFailure:
            pass
        else:
            raise InterlockFailure('bound predicate accepted the positive control')
        return {'expected_interlock_failure': True, 'model_attempts': len(attempted),
                'provider': expected_provider, 'stub_exit_code': 97}
    if attempted:
        raise InterlockFailure('model attempt in bound or reference route')
    return {'expected_interlock_failure': False, 'model_attempts': 0}


def compare(bound, reference):
    # Input equivalence is established before looking at findings.
    for name in ('accounted_coverage','scanner_inputs','omissions','mapping_gaps'):
        if bound[name] != reference[name]:
            raise InterlockFailure('scanner input/coverage mismatch: '+name)
    left, right = bound['issue_ids'], reference['issue_ids']
    if left != sorted(set(left)) or right != sorted(set(right)):
        raise InterlockFailure('issue sets not sorted and unique')
    if not left or not right or left != right:
        raise InterlockFailure('nonempty complete issue-ID parity failed')
    if bound['multiplicity'] != reference['multiplicity']:
        raise InterlockFailure('finding multiplicity differs')
    if not bound['scanner_health_complete'] or not reference['scanner_health_complete']:
        raise InterlockFailure('partial/blocked scanner health')
    if not bound['binding_rows_complete']:
        raise InterlockFailure('model binding rows not observed disabled')
    if set(bound['durable_links']) != set(left) or any(not x for x in bound['durable_links'].values()):
        raise InterlockFailure('durable issues do not link to complete occurrence keys')
    return {'equal': True, 'unique_count': len(left), 'multiplicity_count': sum(bound['multiplicity'].values()),
            'bound_durable_issue_count': bound['durable_issue_count'],
            'reference_durable_issue_count': reference['durable_issue_count'],
            'carried_issue_ids': bound['carried_issue_ids']}
